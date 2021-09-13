package proxy

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

type Proxy struct {
	dest     string
	local    string
	timeout  time.Duration
	logger   Logger
	toStop   chan struct{}
	Done     chan struct{}
	failFast bool
	running  bool
	notified bool
	mu       *sync.Mutex
}

func New(dest, local string, timeout time.Duration, logger Logger, failFast bool) *Proxy {
	return &Proxy{
		dest:     dest,
		local:    local,
		timeout:  timeout,
		logger:   logger,
		toStop:   make(chan struct{}, 2),
		Done:     make(chan struct{}),
		failFast: failFast,
		mu:       &sync.Mutex{},
	}
}

func (p *Proxy) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.running || p.notified {
		return errors.New("already stopped")
	}

	p.toStop <- struct{}{}
	p.notified = true
	p.logger.Info("notify proxy to stop")

	// consume a conn
	testConn, err := net.Dial("tcp", p.local)
	if err != nil {
		return err
	}
	return testConn.Close()
}

func (p *Proxy) Run() error {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return errors.New("already running")
	}
	p.mu.Unlock()

	go p.doRun()
	return nil
}

func (p *Proxy) doRun() {
	go p.run()

	select {
	case <-p.Done:
		p.logger.Info("proxy stopped")
		p.running = false
		return
	}
}

func (p *Proxy) run() {
	// check destination alive first
	testConn, err := net.DialTimeout("tcp", p.dest, p.timeout)
	if err != nil {
		if p.failFast {
			p.logger.Fatal(err)
		} else {
			p.logger.Errorf("%v", err)
			goto bind
		}
	}
	_ = testConn.Close()

bind:
	// bind local port
	ln, err := net.Listen("tcp", p.local)
	if err != nil {
		p.logger.Fatal(err)
	}
	defer func() { _ = ln.Close() }()

	p.running = true
	p.logger.Info("proxy has been started")

	// accept connections
	for {
		select {
		case <-p.toStop:
			p.Done <- struct{}{}
			return
		default:
			conn, err := ln.Accept()
			if err != nil {
				p.logger.Errorf("%v: ", err)
				continue
			}

			// to proxy
			go p.proxy(conn)
		}
	}
}

func (p *Proxy) proxy(inConn net.Conn) {
	errChan := make(chan error, 2)

	connClose := func(conn net.Conn) { _ = conn.Close() }
	connDup := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errChan <- err
	}

	defer connClose(inConn)

	outConn, err := net.DialTimeout("tcp", p.dest, p.timeout)
	if err != nil {
		p.logger.Errorf("%v", err)
		return
	}
	defer connClose(outConn)

	go connDup(inConn, outConn)
	go connDup(outConn, inConn)
	<-errChan
}

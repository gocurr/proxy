package proxy

import (
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
	inConn   net.Conn
	outConn  net.Conn
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

func (p *Proxy) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.running || p.notified {
		return
	}
	p.toStop <- struct{}{}
	p.notified = true
	p.logger.Infof("notify proxy to stop")
}

func (p *Proxy) Run() {
	go p.doRun()
}

func (p *Proxy) doRun() {
	go p.run()

	for {
		select {
		case <-p.Done:
			p.logger.Infof("Done")
			return
		}
	}
}

func (p *Proxy) run() {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		p.logger.Errorf("already running")
		return
	}
	p.mu.Unlock()

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

	// accept for connections
	for {
		select {
		case <-p.toStop:
			p.logger.Infof("proxy stopped")
			p.Done <- struct{}{}
			return
		default:
			conn, err := ln.Accept()
			if err != nil {
				p.logger.Errorf("%v: ", err)
				continue
			}

			p.inConn = conn
			// to proxy
			go p.proxy()
		}

	}
}

func (p *Proxy) proxy() {
	errc := make(chan error, 2)

	connClose := func(conn net.Conn) { _ = conn.Close() }
	connDup := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errc <- err
	}

	defer connClose(p.inConn)

	var err error
	p.outConn, err = net.DialTimeout("tcp", p.dest, p.timeout)
	if err != nil {
		p.logger.Errorf("%v", err)
		return
	}
	defer connClose(p.outConn)

	go connDup(p.inConn, p.outConn)
	go connDup(p.outConn, p.inConn)
	<-errc
}

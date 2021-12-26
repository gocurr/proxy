package proxy

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Proxy struct {
	name       string
	local      string
	remote     string
	timeout    time.Duration
	notifyDone chan struct{}
	done       chan struct{}
	burst      chan struct{}
	failFast   bool // when remote is invalid
	logger     Logger
	mu         sync.Mutex // protects the remaining
	running    bool
}

func New(name, local, remote string, timeout time.Duration, failFast bool, logger Logger) *Proxy {
	return &Proxy{
		name:       name,
		local:      local,
		remote:     remote,
		timeout:    timeout,
		logger:     logger,
		notifyDone: make(chan struct{}, 2),
		done:       make(chan struct{}),
		burst:      make(chan struct{}),
		failFast:   failFast,
	}
}

func (p *Proxy) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return fmt.Errorf("%s is already done", p.name)
	}

	p.notifyDone <- struct{}{}

	// consume a conn
	testConn, err := net.Dial("tcp", p.local)
	if err != nil {
		return err
	}
	return testConn.Close()
}

func (p *Proxy) Run() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return fmt.Errorf("%s is running", p.name)
	}

	go p.doRun()

	<-p.burst

	return nil
}

func (p *Proxy) doRun() {
	go p.run()

	<-p.done
	p.logger.Infof("%s is just done", p.name)
	p.running = false
	return
}

func (p *Proxy) run() {
	// check destination invalid first
	testConn, err := net.DialTimeout("tcp", p.remote, p.timeout)
	if err != nil {
		p.logger.Errorf("%v", err)
		if p.failFast {
			return
		}
	} else {
		_ = testConn.Close()
	}

	// bind local port
	ln, err := net.Listen("tcp", p.local)
	if err != nil {
		p.logger.Errorf("%v", err)
		return
	}
	defer func() { _ = ln.Close() }()

	p.running = true
	p.logger.Info(fmt.Sprintf("%s is running", p.name))
	p.burst <- struct{}{}

	// accept connections
	for {
		select {
		case <-p.notifyDone:
			p.done <- struct{}{}
			return
		default:
			conn, err := ln.Accept()
			if err != nil {
				p.logger.Errorf("%v", err)
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
		errChan <- p.hijack(dst, src)
	}

	defer connClose(inConn)

	outConn, err := net.DialTimeout("tcp", p.remote, p.timeout)
	if err != nil {
		p.logger.Errorf("%v", err)
		return
	}
	defer connClose(outConn)

	go connDup(inConn, outConn)
	go connDup(outConn, inConn)
	<-errChan
}

var errInvalidWrite = errors.New("invalid write result")

const defaultBufSize = 32 << 10 // 32 KB

func (p *Proxy) hijack(dst io.Writer, src io.Reader) (err error) {
	buf := make([]byte, defaultBufSize)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			b := buf[0:nr]
			p.logger.Infof("%s", string(b))
			nw, ew := dst.Write(b)
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errInvalidWrite
				}
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return err
}

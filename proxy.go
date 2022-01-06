package proxy

import (
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Proxy struct {
	name     string
	local    string
	remote   string
	timeout  time.Duration
	failFast bool // when remote is invalid
	logger   Logger
	mu       sync.Mutex // protects the remaining fields

	// Note: notifyDone must be a buffered channel.
	// In the endless for-loop, once the default case is selected,
	// code "select { case <-p.notifyDone: ...}" maybe not prepared.
	notifyDone chan struct{}
	done       chan struct{}
	burst      chan error
	running    bool
}

var ipPattern = `/^((\d|[1-9]\d|1\d\d|2[0-4]\d|25[0-5])\.){3}(\d|[1-9]\d|1\d\d|2[0-4]\d|25[0-5])$`
var ipReg = regexp.MustCompile(ipPattern)

var errIp = errors.New("proxy: bad ip format")
var errAddr = errors.New("proxy: bad addr format")
var errPort = errors.New("proxy: bad port format")
var errTimeout = errors.New("proxy: timeout must be greater than 0")

func portCheck(s string) error {
	a, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	if a < 0 || a > 65535 {
		return errPort
	}
	return nil
}

func reAddr(s string) (string, error) {
	ipPort := strings.Split(s, ":")
	// Only port.
	if len(ipPort) == 1 {
		if err := portCheck(s); err != nil {
			return "", err
		} else {
			return "127.0.0.1:" + s, nil
		}
	}

	if len(ipPort) != 2 {
		return "", errAddr
	}

	// ip:port
	if ipReg.MatchString(ipPort[0]) {
		return "", errIp
	}
	if err := portCheck(ipPort[1]); err != nil {
		return "", err
	}
	return s, nil
}

func New(name, local, remote string, timeout time.Duration, failFast bool, logger Logger) (*Proxy, error) {
	local, err := reAddr(local)
	if err != nil {
		return nil, err
	}

	remote, err = reAddr(remote)
	if err != nil {
		return nil, err
	}

	if remote == local {
		return nil, errAddr
	}

	if timeout <= 0 {
		return nil, errTimeout
	}

	// Check local-port first.
	lConn, lErr := net.DialTimeout("tcp", local, timeout)
	if lErr == nil {
		_ = lConn.Close()
		return nil, errLocal
	}
	// Check remote address invalid then.
	rConn, rErr := net.DialTimeout("tcp", remote, timeout)
	if rErr != nil {
		if failFast {
			return nil, rErr
		}
	} else {
		_ = rConn.Close()
	}

	return &Proxy{
		name:       name,
		local:      local,
		remote:     remote,
		timeout:    timeout,
		logger:     logger,
		notifyDone: make(chan struct{}, 1),
		done:       make(chan struct{}),
		burst:      make(chan error),
		failFast:   failFast,
	}, nil
}

func (p *Proxy) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return fmt.Errorf("proxy: %s is done already", p.name)
	}

	p.notifyDone <- struct{}{}

	// Wait for connection refused.
	var try = 0
	for {
		try++
		testConn, err := net.Dial("tcp", p.local)
		if err != nil {
			return nil
		}
		_ = testConn.Close()
		const maxTries = 3
		if try > maxTries {
			return fmt.Errorf("proxy: tried %d times to wait for connection refused", try)
		}
	}
}

func (p *Proxy) Run() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return fmt.Errorf("proxy: %s was running", p.name)
	}

	go p.doRun()

	err := <-p.burst
	return err
}

func (p *Proxy) doRun() {
	go p.run()

	<-p.done
	p.logger.Infof("%s is just done", p.name)
	p.running = false
	return
}

var errLocal = errors.New("proxy: local-port is in use")

func (p *Proxy) run() {
	// Check local-port first.
	lConn, lErr := net.DialTimeout("tcp", p.local, p.timeout)
	if lErr == nil {
		_ = lConn.Close()
		p.burst <- errLocal
		return
	}
	// Check remote address invalid then.
	rConn, rErr := net.DialTimeout("tcp", p.remote, p.timeout)
	if rErr != nil {
		if p.failFast {
			p.burst <- rErr
			return
		}
	} else {
		_ = rConn.Close()
	}

	// Bind local port.
	ln, err := net.Listen("tcp", p.local)
	if err != nil {
		p.burst <- err
		return
	}
	defer func() { _ = ln.Close() }()

	p.running = true
	p.logger.Infof("%s is running", p.name)
	p.burst <- nil

	// Accept connections.
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

			// To proxy.
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

var errInvalidWrite = errors.New("proxy: invalid write result")

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

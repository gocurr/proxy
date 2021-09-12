package proxy

import (
	"io"
	"net"
	"time"
)

type Proxy struct {
	Dest    string
	Local   string
	Timeout time.Duration
	logger  Logger
}

func New(dest, local string, timeout time.Duration, logger Logger) *Proxy {
	return &Proxy{
		Dest:    dest,
		Local:   local,
		Timeout: timeout,
		logger:  logger,
	}
}

func (p *Proxy) Run() {
	// check destination alive first
	testConn, err := net.DialTimeout("tcp", p.Dest, p.Timeout)
	if err != nil {
		p.logger.Fatal(err)
	}
	_ = testConn.Close()

	// bind local port
	ln, err := net.Listen("tcp", p.Local)
	if err != nil {
		p.logger.Fatal(err)
	}

	// accept for connections
	for {
		conn, err := ln.Accept()
		if err != nil {
			p.logger.Errorf("%v: ", err)
			continue
		}

		// to proxy
		go p.proxy(conn)
	}
}

func (p *Proxy) proxy(inConn net.Conn) {
	errc := make(chan error, 2)

	connClose := func(conn net.Conn) { _ = conn.Close() }
	connDup := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errc <- err
	}

	defer connClose(inConn)

	outConn, err := net.DialTimeout("tcp", p.Dest, p.Timeout)
	if err != nil {
		p.logger.Errorf("%v", err)
		return
	}
	defer connClose(outConn)

	go connDup(inConn, outConn)
	go connDup(outConn, inConn)
	<-errc
}

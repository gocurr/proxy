package proxy_test

import (
	"github.com/gocurr/proxy"
	"testing"
	"time"
)

var err error
var p *proxy.Proxy

func init() {
	p, err = proxy.New(mysql, local, remote, time.Second, false, proxy.Discard)
	if err != nil {
		panic(err)
	}
}

func Test_Proxy(t *testing.T) {
	for i := 0; i < 10; i++ {
		run(t)
		stop(t)
	}
}

func run(t *testing.T) {
	if err := p.Run(); err != nil {
		t.Fatal(err)
	}
	if tryConn() != nil {
		t.Fatal("cannot add")
	}
}

func stop(t *testing.T) {
	if err := p.Stop(); err != nil {
		t.Fatal(err)
	}
	if tryConn() == nil {
		t.Fatal("cannot remove")
	}
}

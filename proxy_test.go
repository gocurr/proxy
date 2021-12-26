package proxy_test

import (
	"github.com/gocurr/proxy"
	"net"
	"testing"
	"time"
)

var manager = proxy.NewManager(3*time.Second, false, proxy.Logrus)
var local = "127.0.0.1:3307"
var remote = "127.0.0.1:3306"
var mysql = "mysql"

func Test_Proxy(t *testing.T) {
	for i := 0; i < 100; i++ {
		add(t)
		remove(t)
	}
}

func add(t *testing.T) {
	err := manager.Add(mysql, local, remote)
	if err != nil {
		t.Fatal(err)
	}
	if tryConn() != nil {
		t.Fatal("cannot add")
	}
}

func remove(t *testing.T) {
	if err := manager.Remove(mysql); err != nil {
		t.Fatal(err)
	}
	if tryConn() == nil {
		t.Fatal("cannot remove")
	}
}

func tryConn() error {
	conn, err := net.Dial("tcp", local)
	if err != nil {
		return err
	}
	return conn.Close()
}

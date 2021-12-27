package proxy_test

import (
	"github.com/gocurr/proxy"
	"net"
	"strings"
	"testing"
	"time"
)

var manager = proxy.NewManager(3*time.Second, false, proxy.Discard)
var local = "3307"
var remote = "3306"
var mysql = "mysql"

func Test_Manager(t *testing.T) {
	for i := 0; i < 10; i++ {
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
	address := local
	if !strings.Contains(local, ":") {
		address = "127.0.0.1:" + local
	}
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	return conn.Close()
}

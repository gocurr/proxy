package proxy_test

import (
	"github.com/gocurr/proxy"
	"net"
	"testing"
	"time"
)

var manager = proxy.NewManager(3*time.Second, false, proxy.Logrus)

func Test_Proxy(t *testing.T) {
	for i := 0; i < 5; i++ {
		add(t)
		remove(t)
	}
}

func add(t *testing.T) {
	err := manager.Add("mysql", "127.0.0.1:3307", "127.0.0.1:3306")
	if err != nil {
		panic(err)
	}
	if tryConn() != nil {
		t.Fatal("cannot add")
	}
}

func remove(t *testing.T) {
	if err := manager.Remove("mysql"); err != nil {
		panic(err)
	}
	if tryConn() == nil {
		t.Fatal("cannot remove")
	}
}

func tryConn() error {
	time.Sleep(1 * time.Second)
	conn, err := net.Dial("tcp", "127.0.0.1:3307")
	if err != nil {
		return err
	}
	return conn.Close()
}

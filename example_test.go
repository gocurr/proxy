package proxy_test

import (
	"github.com/gocurr/proxy"
	"net"
	"testing"
	"time"
)

func Test_Proxy(t *testing.T) {
	manager := proxy.NewManager(3*time.Second, false, proxy.Discard)
	manager = proxy.NewManager(3*time.Second, false, proxy.Logrus)
	err := manager.Add("mysql", "127.0.0.1:3307", "127.0.0.1:3306")
	if err != nil {
		panic(err)
	}

	time.Sleep(1 * time.Second)
	testConn, err := net.Dial("tcp", "127.0.0.1:3307")
	if err != nil {
		panic(err)
	}
	_ = testConn.Close()
}

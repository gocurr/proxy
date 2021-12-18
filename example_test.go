package proxy_test

import (
	"github.com/gocurr/proxy"
	"net/http"
	"testing"
	"time"
)

func Test_Proxy(t *testing.T) {
	manager := proxy.NewManager(3*time.Second, false, proxy.Logrus)
	err := manager.Add("mysql", "127.0.0.1:3307", "127.0.0.1:3306")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/inner", manager.HttpProxyCtrl("xxx"))
	_ = http.ListenAndServe(":9000", nil)
}

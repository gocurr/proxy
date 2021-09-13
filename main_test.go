package proxy

import (
	"net/http"
	"testing"
	"time"
)

func Test_main(t *testing.T) {
	manager := NewManager()
	err := manager.Add("mysql", "127.0.0.1:3307", "127.0.0.1:3306", 3*time.Second, DefaultLogger{}, false)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/inner", manager.HttpProxyCtrl("xxx"))
	_ = http.ListenAndServe(":9000", nil)
}

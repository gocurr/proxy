package proxy

import (
	"net/http"
	"testing"
)

func Test_main(t *testing.T) {
	proxys := NewProxys()
	err := proxys.Add("mysql", "127.0.0.1:8888", "127.0.0.1:9090")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/inner", proxys.HttpProxyCtrl("xxx"))
	_ = http.ListenAndServe(":9000", nil)
}

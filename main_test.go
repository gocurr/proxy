package proxy

import (
	"net/http"
	"testing"
	"time"
)

func Test_main(t *testing.T) {
	p := New("localhost:9090", "localhost:9999", time.Second*3, DefaultLogger{}, false)

	_ = p.Run()

	http.HandleFunc("/inner", p.HttpProxyCtrl("xxx", false))
	_ = http.ListenAndServe(":9000", nil)
}

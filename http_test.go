package proxy

import (
	"log"
	"net/http"
	"testing"
)

func Test_Http(t *testing.T) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("access: %v", r.RemoteAddr)
		_, _ = w.Write([]byte("ok"))
	})
	_ = http.ListenAndServe(":9090", nil)
}

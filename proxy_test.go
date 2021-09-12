package proxy

import (
	"testing"
	"time"
)

func Test_Proxy(t *testing.T) {
	p := New("127.0.0.1:9090", "127.0.0.1:9999", 3*time.Second, DefaultLogger{})
	p.Run()

	select {}
}

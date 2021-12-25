# TCP Proxy in Go

To download, run:

```bash
go get -u github.com/gocurr/proxy
```

Import it in your program as:

```go
import "github.com/gocurr/proxy"
```

It requires Go 1.11 or later due to usage of Go Modules.

- To start a proxy:

```go
package main

import (
	"github.com/gocurr/proxy"
	"time"
)

func main() {
	p := proxy.New("mysql", "127.0.0.1:3306", "127.0.0.1:3307", time.Second, false, proxy.Logrus)
	if err := p.Run(); err != nil {
		panic(err)
	}

	// when you want to stop the proxy
	if err := p.Stop(); err != nil {
		panic(err)
	}
}
```

- To start a proxy-manager:

```go
package main

import (
	"github.com/gocurr/proxy"
	"net/http"
	"time"
)

func main() {
	manager := proxy.NewManager(3*time.Second, false, proxy.Logrus)
	if err := manager.Add("httpserver", "127.0.0.1:9091", "127.0.0.1:9090"); err != nil {
		panic(err)
	}

	http.HandleFunc("/", manager.HttpProxyCtrl("xxx"))
	_ = http.ListenAndServe(":9000", nil)
}
```

```bash
curl -X POST -H "Content-Type:application/json" -d '{"token":"xxx", "type":"details"}' http://127.0.0.1:9000/proxy
```

```bash
curl -X POST -H "Content-Type:application/json" -d '{"token":"xxx", "type":"start", "name":"mysql"}' http://127.0.0.1:9000/proxy
```

```bash
curl -X POST -H "Content-Type:application/json" -d '{"token":"xxx", "type":"stop", "name":"mysql"}' http://127.0.0.1:9000/proxy
```

```bash
curl -X POST -H "Content-Type:application/json" -d '{"token":"xxx", "type":"insert", "name":"http-proxy", "local":"127.0.0.1:9091", "remote":"127.0.0.1:80"}' http://127.0.0.1:9000/proxy
```

```bash
curl -X POST -H "Content-Type:application/json" -d '{"token":"xxx", "type":"delete", "name":"http-proxy"}' http://127.0.0.1:9000/proxy
```



# proxy

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
p := New(name, local, remote, 3*time.Second, DefaultLogger{}, false)
err := p.Run()
err := p.Stop()
```

- To start a proxy-manager:

```go
manager := NewManager()
err := manager.Add("mysql", "127.0.0.1:3309", "127.0.0.1:3306")
err := manager.Remove("mysql")
details := manager.Details()

http.HandleFunc("/proxy", manager.HttpProxyCtrl("xxx"))
_ = http.ListenAndServe(":9000", nil)
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
curl -X POST -H "Content-Type:application/json" -d '{"token":"xxx", "type":"add", "name":"mysql", "local":"127.0.0.1:8888", "remote":"127.0.0.1:9090"}' http://127.0.0.1:9000/proxy
```

```bash
curl -X POST -H "Content-Type:application/json" -d '{"token":"xxx", "type":"remove", "name":"mysql"}' http://127.0.0.1:9000/proxy
```



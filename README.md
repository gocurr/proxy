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
var err error
p := proxy.New(name, local, remote, 3*time.Second, proxy.DefaultLogger{}, false)
err = p.Run()
if err != nil {
// ...
}

err = p.Stop()
if err != nil {
// ...
}
```

- To start a proxy-manager:

```go
var err error
manager := proxy.NewManager()
err = manager.Add("mysql", "127.0.0.1:3307", "127.0.0.1:3306", 3*time.Second, DefaultLogger{}, false)
if err != nil {
// ...
}

err = manager.Remove("mysql")
if err != nil {
// ...
}

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
curl -X POST -H "Content-Type:application/json" -d '{"token":"xxx", "type":"add", "name":"http-proxy", "local":"127.0.0.1:9091", "remote":"127.0.0.1:80"}' http://127.0.0.1:9000/proxy
```

```bash
curl -X POST -H "Content-Type:application/json" -d '{"token":"xxx", "type":"remove", "name":"http-proxy"}' http://127.0.0.1:9000/proxy
```



package main

import (
	"github.com/gocurr/proxy"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func main() {
	manager := proxy.NewManager(3*time.Second, false, proxy.Logrus)
	if err := manager.Add("mysql", "127.0.0.1:3307", "127.0.0.1:3306"); err != nil {
		panic(err)
	}

	http.HandleFunc("/proxy", manager.HttpProxyCtrl("xxx"))
	log.Fatalln(http.ListenAndServe(":9000", nil))
}

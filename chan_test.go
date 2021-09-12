package proxy

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func Test_Chan(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)

	c := make(chan struct{}, 2)
	go func() {
		select {
		case <-c:
			fmt.Println("sleep for 1 sec")
			time.Sleep(1 * time.Second)
			fmt.Println("get")
		}
		wg.Done()
	}()

	go func() {
		time.Sleep(2 * time.Second)
		c <- struct{}{}
		fmt.Println("notified")
		wg.Done()
	}()

	wg.Wait()
}

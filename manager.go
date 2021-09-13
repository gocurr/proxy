package proxy

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type Proxys struct {
	mu   sync.Mutex
	dict map[string]*Proxy
}

func NewProxys() *Proxys {
	return &Proxys{
		mu:   sync.Mutex{},
		dict: make(map[string]*Proxy),
	}
}

func (ps *Proxys) Add(name, local, remote string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	p := New(name, local, remote, 3*time.Second, DefaultLogger{}, false)
	err := ps.add(name, p)
	if err != nil {
		return err
	}
	return p.Run()
}

func (ps *Proxys) add(name string, p *Proxy) error {
	_, ok := ps.dict[name]
	if ok {
		return errors.New(fmt.Sprintf("proxy: %s exists", name))
	}

	ps.dict[name] = p
	return nil
}

func (ps *Proxys) Remove(name string) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	_ = ps.dict[name].Stop()

	return ps.remove(name)
}

func (ps *Proxys) remove(name string) error {
	_, ok := ps.dict[name]
	if !ok {
		return errors.New(fmt.Sprintf("proxy: %s dose not exists", name))
	}

	delete(ps.dict, name)
	return nil
}

type Detail struct {
	Name    string
	Local   string
	Remote  string
	Running bool
}

func (ps *Proxys) Details() []*Detail {
	var details []*Detail
	for name, p := range ps.dict {
		detail := Detail{
			Name:    name,
			Local:   p.local,
			Remote:  p.remote,
			Running: p.running,
		}
		details = append(details, &detail)
	}
	return details
}

func (ps *Proxys) Exists(name string) bool {
	_, exists := ps.dict[name]
	return exists
}

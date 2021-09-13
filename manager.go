package proxy

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type Manager struct {
	mu   sync.RWMutex
	dict map[string]*Proxy
}

func NewManager() *Manager {
	return &Manager{
		mu:   sync.RWMutex{},
		dict: make(map[string]*Proxy),
	}
}

func (m *Manager) Add(name, local, remote string, timeout time.Duration, logger Logger, failFast bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p := New(name, local, remote, timeout, logger, failFast)
	err := m.add(name, p)
	if err != nil {
		return err
	}
	return p.Run()
}

func (m *Manager) add(name string, p *Proxy) error {
	_, ok := m.dict[name]
	if ok {
		return errors.New(fmt.Sprintf("proxy: %s exists", name))
	}

	m.dict[name] = p
	return nil
}

func (m *Manager) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_ = m.dict[name].Stop()

	return m.remove(name)
}

func (m *Manager) remove(name string) error {
	_, ok := m.dict[name]
	if !ok {
		return errors.New(fmt.Sprintf("proxy: %s dose not exists", name))
	}

	delete(m.dict, name)
	return nil
}

type Detail struct {
	Name    string
	Local   string
	Remote  string
	Running bool
}

func (m *Manager) Details() []*Detail {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var details []*Detail
	for name, p := range m.dict {
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

func (m *Manager) Exists(name string) bool {
	_, exists := m.dict[name]
	return exists
}

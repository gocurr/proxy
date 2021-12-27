package proxy

import (
	"fmt"
	"sync"
	"time"
)

type Manager struct {
	mu       sync.RWMutex
	dict     map[string]*Proxy
	timeout  time.Duration
	failFast bool
	logger   Logger
}

func NewManager(timeout time.Duration, failFast bool, logger Logger) *Manager {
	return &Manager{
		mu:       sync.RWMutex{},
		dict:     make(map[string]*Proxy),
		timeout:  timeout,
		failFast: failFast,
		logger:   logger,
	}
}

func (m *Manager) Add(name, local, remote string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, err := New(name, local, remote, m.timeout, m.failFast, m.logger)
	if err != nil {
		return err
	}
	if err := m.add(name, p); err != nil {
		return err
	}
	return p.Run()
}

func (m *Manager) add(name string, p *Proxy) error {
	_, ok := m.dict[name]
	if ok {
		return fmt.Errorf("proxy: %s exists", name)
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

func (m *Manager) Register(name, local, remote string, timeout int, failFast, discard bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var logger Logger = Logrus
	if discard {
		logger = Discard
	}

	proxy, err := New(name, local, remote, time.Duration(timeout)*time.Second, failFast, logger)
	if err != nil {
		return err
	}
	m.dict[name] = proxy
	return nil
}

func (m *Manager) remove(name string) error {
	_, ok := m.dict[name]
	if !ok {
		return fmt.Errorf("proxy: %s dose not exists", name)
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

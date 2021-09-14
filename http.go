package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type P struct {
	Token  string `json:"token"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Local  string `json:"local"`
	Remote string `json:"remote"`
}

func (m *Manager) HttpProxyCtrl(token string) func(http.ResponseWriter, *http.Request) {
	return m.ctrl(token)
}

func (m *Manager) ctrl(token string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			handleErr("method-check", errors.New(fmt.Sprintf("illegal method: %s", r.Method)), w)
			return
		}

		p, err := parameter(r)
		if err != nil {
			handleErr("parameter", err, w)
			return
		}

		if ok := tokenValid(token, p.Token); !ok {
			handleErr("tokenValid", errors.New("token invalid"), w)
			return
		}

		if p.Type == "" {
			handleErr("paramter", errors.New("type is nil"), w)
			return
		}

		switch p.Type {
		case "details":
			m.details(w)
		case "start":
			m.start(w, p)
		case "stop":
			m.stop(w, p)
		case "insert":
			m.insert(w, p)
		case "delete":
			m.delete(w, p)
		default:
			handleErr("check-type", errors.New(fmt.Sprintf("unknow type %s", p.Type)), w)
		}
	}
}

func (m *Manager) start(w http.ResponseWriter, p *P) {
	if !m.Exists(p.Name) {
		handleErr("manager.Exists", errors.New(fmt.Sprintf("proxy: %s not exist", p.Name)), w)
		return
	}

	proxy := m.dict[p.Name]
	err := proxy.Run()
	handleErr("start", err, w)
}

func (m *Manager) stop(w http.ResponseWriter, p *P) {
	if !m.Exists(p.Name) {
		handleErr("manager.Exists", errors.New(fmt.Sprintf("proxy: %s not exist", p.Name)), w)
		return
	}

	proxy := m.dict[p.Name]
	err := proxy.Stop()
	handleErr("stop", err, w)
}

func (m *Manager) delete(w http.ResponseWriter, p *P) {
	if !m.Exists(p.Name) {
		handleErr("manager.Exists", errors.New(fmt.Sprintf("proxy: %s exists", p.Name)), w)
		return
	}

	err := m.Remove(p.Name)
	handleErr("proxys.Remove", err, w)
}

func parameter(r *http.Request) (*P, error) {
	var p P
	all, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.Body.Close() }()

	err = json.Unmarshal(all, &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (m *Manager) insert(w http.ResponseWriter, p *P) {
	if m.Exists(p.Name) {
		handleErr("ps.Exists", errors.New(fmt.Sprintf("proxy: %s exists", p.Name)), w)
		return
	}
	if !strings.ContainsAny(p.Local, ":") ||
		!strings.ContainsAny(p.Remote, ":") {
		handleErr("ip check", errors.New(fmt.Sprintf("proxy bad format: %s %s ", p.Local, p.Remote)), w)
		return
	}
	err := m.Add(p.Name, p.Local, p.Remote)
	handleErr("proxys.Add", err, w)
}

func (m *Manager) details(w http.ResponseWriter) {
	details := m.Details()
	if len(details) == 0 {
		details = []*Detail{}
	}
	bytes, err := json.Marshal(details)
	if err != nil {
		handleErr("detail", err, w)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(bytes)
}

func tokenValid(token string, other string) bool {
	return token == other
}

func handleErr(method string, err error, w http.ResponseWriter) {
	msg := "ok"
	if err != nil {
		msg = err.Error()
	}
	_, _ = w.Write([]byte(fmt.Sprintf("%s: %s", method, msg)))
}

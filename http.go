package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const (
	tokenLiteral = "token"
)

var (
	tokenNotValidErr = errors.New("token not valid")
)

func (m *Manager) HttpProxyCtrl(token string) func(http.ResponseWriter, *http.Request) {
	return m.ctrl(token)
}

func (m *Manager) ctrl(token string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := tokenValid(token, r); err != nil {
			handleErr("tokenValid", err, w)
			return
		}

		typ, err := parameter("type", r)
		if err != nil {
			handleErr("paramter", err, w)
			return
		}

		if typ == "details" {
			m.detailsFunc(w)
			return
		}

		name, err := parameter("name", r)
		if err != nil {
			handleErr("parameter", err, w)
			return
		}
		switch typ {
		case "start":
			m.startFunc(w, name, err)
		case "stop":
			m.stopFunc(w, name, err)
		case "add":
			m.addFunc(w, r, name)
		case "remove":
			m.removeFunc(w, name, err)
		default:
			handleErr("check-type", errors.New(fmt.Sprintf("unknow type %s", typ)), w)
		}
	}
}

func (m *Manager) startFunc(w http.ResponseWriter, name string, err error) {
	if !m.Exists(name) {
		handleErr("ps.Exists", errors.New(fmt.Sprintf("proxy: %s not exist", name)), w)
		return
	}

	p := m.dict[name]
	err = p.Run()
	handleErr("start", err, w)
}

func (m *Manager) stopFunc(w http.ResponseWriter, name string, err error) {
	if !m.Exists(name) {
		handleErr("ps.Exists", errors.New(fmt.Sprintf("proxy: %s not exist", name)), w)
		return
	}

	p := m.dict[name]
	err = p.Stop()
	handleErr("stop", err, w)
}

func (m *Manager) removeFunc(w http.ResponseWriter, name string, err error) {
	if !m.Exists(name) {
		handleErr("ps.Exists", errors.New(fmt.Sprintf("proxy: %s exists", name)), w)
		return
	}

	err = m.Remove(name)
	handleErr("proxys.Remove", err, w)
}

func (m *Manager) addFunc(w http.ResponseWriter, r *http.Request, name string) {
	if m.Exists(name) {
		handleErr("ps.Exists", errors.New(fmt.Sprintf("proxy: %s exists", name)), w)
		return
	}
	local, err := parameter("local", r)
	if err != nil {
		handleErr("paramter", err, w)
		return
	}
	remote, err := parameter("remote", r)
	if err != nil {
		handleErr("paramter", err, w)
		return
	}
	err = m.Add(name, local, remote)
	handleErr("proxys.Add", err, w)
}

func (m *Manager) detailsFunc(w http.ResponseWriter) {
	details := m.Details()
	bytes, err := json.Marshal(details)
	if err != nil {
		handleErr("detail", err, w)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(bytes)
}

func parameter(name string, r *http.Request) (string, error) {
	values := r.URL.Query()
	val, ok := values[name]
	if !ok || len(val) < 1 {
		return "", errors.New(fmt.Sprintf(`parameter "%s" not found`, name))
	}

	return val[0], nil
}

func tokenValid(token string, r *http.Request) error {
	t, err := parameter(tokenLiteral, r)
	if err != nil || t != token {
		return tokenNotValidErr
	}

	return nil
}

func handleErr(method string, err error, w http.ResponseWriter) {
	msg := "ok"
	if err != nil {
		msg = err.Error()
	}
	_, _ = w.Write([]byte(fmt.Sprintf("%s: %s", method, msg)))
}

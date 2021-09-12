package proxy

import (
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

func (p *Proxy) HttpProxyCtrl(token string, logging bool) func(http.ResponseWriter, *http.Request) {
	return p.ctrl(token, logging)
}

func (p *Proxy) ctrl(token string, logging bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := tokenValid(token, r); err != nil {
			p.handleErr("tokenValid", err, w, logging)
			return
		}

		typ, err := parameter("type", r)
		if err != nil {
			p.handleErr("paramter", err, w, logging)
			return
		}

		switch typ {
		case "start":
			p.Run()
			p.handleErr("start", nil, w, logging)
		case "stop":
			p.Stop()
			p.handleErr("stop", nil, w, logging)
		default:
			p.handleErr("check-type", errors.New(fmt.Sprintf("unknow type %s", typ)), w, logging)
		}
	}
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

func (p *Proxy) handleErr(method string, err error, w http.ResponseWriter, logging bool) {
	msg := "ok"
	if err != nil {
		msg = err.Error()
		if logging {
			p.logger.Errorf("%s %v", method, err)
		}
	}
	_, _ = w.Write([]byte(msg))
}

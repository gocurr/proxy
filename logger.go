package proxy

import log "github.com/sirupsen/logrus"

type Logger interface {
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
}

var Logrus DefaultLogger

type DefaultLogger struct {
}

func (d DefaultLogger) Info(args ...interface{}) {
	log.Info(args...)
}

func (d DefaultLogger) Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func (d DefaultLogger) Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func (d DefaultLogger) Fatal(args ...interface{}) {
	log.Fatal(args...)
}

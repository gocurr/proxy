package proxy

import (
	log "github.com/sirupsen/logrus"
	"os"
)

type Logger interface {
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
}

var Discard discardLogger

type discardLogger struct {
}

func (d discardLogger) Info(...interface{}) {

}

func (d discardLogger) Infof(string, ...interface{}) {

}

func (d discardLogger) Errorf(string, ...interface{}) {

}

func (d discardLogger) Fatal(...interface{}) {
	os.Exit(1)
}

var Logrus defaultLogger

type defaultLogger struct {
}

func (d defaultLogger) Info(args ...interface{}) {
	log.Info(args...)
}

func (d defaultLogger) Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func (d defaultLogger) Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func (d defaultLogger) Fatal(args ...interface{}) {
	log.Fatal(args...)
}

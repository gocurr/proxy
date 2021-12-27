package proxy

import (
	log "github.com/sirupsen/logrus"
)

type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// Discard discards logs
var Discard discardLogger

type discardLogger struct{}

func (d discardLogger) Infof(string, ...interface{}) {}

func (d discardLogger) Errorf(string, ...interface{}) {}

// Logrus default proxy logger
var Logrus defaultLogger

type defaultLogger struct{}

func (d defaultLogger) Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func (d defaultLogger) Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

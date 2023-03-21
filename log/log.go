// influxclean log package provides log abstraction
//
// Author: Tesifonte Belda
// License: The MIT License (MIT)

package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	log   *logrus.Logger
	debug bool
}

func NewLogger(debug bool) *Logger {
	var l = &Logger{
		debug: debug,
	}
	var log = logrus.New()
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.TextFormatter{
		DisableColors:   true,
		FullTimestamp:   false,
		TimestampFormat: "2006/01/02 15:04:05",
	})
	l.log = log
	if debug {
		log.SetLevel(logrus.DebugLevel)
	}
	return l
}

func (l *Logger) SetLevel(level logrus.Level) {
	l.log.SetLevel(level)
}

func (l *Logger) Debug(template string) {
	l.log.Debug(template)
}

func (l *Logger) Debugf(template string, args ...interface{}) {
	l.log.Debugf(template, args...)
}

func (l *Logger) Info(template string) {
	l.log.Info(template)
}

func (l *Logger) Infof(template string, args ...interface{}) {
	l.log.Infof(template, args...)
}

func (l *Logger) Warnf(template string, args ...interface{}) {
	l.log.Warnf(template, args...)
}

func (l *Logger) Errorf(template string, args ...interface{}) {
	l.log.Errorf(template, args...)
}

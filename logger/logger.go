package logger

import (
	"fmt"
	"log"
)

type Logger interface {
	Error(f string, v ...interface{})
	Warn(f string, v ...interface{})
	Info(f string, v ...interface{})
	Debug(f string, v ...interface{})
}

func New(warn, info, debug bool) Logger {
	return &logger{
		warn:  warn || info || debug,
		info:  info || debug,
		debug: debug,
	}
}

type logger struct {
	warn, debug, info bool
}

func (l *logger) Error(f string, v ...interface{}) {
	l.log("ERROR", f, v...)
}

func (l *logger) Warn(f string, v ...interface{}) {
	if l.warn {
		l.log("WARN", f, v...)
	}
}

func (l *logger) Info(f string, v ...interface{}) {
	if l.info {
		l.log("INFO", f, v...)
	}
}

func (l *logger) Debug(f string, v ...interface{}) {
	if l.debug {
		l.log("DEBUG", f, v...)
	}
}

func (l *logger) log(prefix, f string, v ...interface{}) {
	log.Println(prefix, fmt.Sprintf(f, v...))
}

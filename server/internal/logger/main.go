package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

type LoggerType string

const (
	Console LoggerType = "console"
	File    LoggerType = "file"
)

type Logger interface {
	Info(args ...any)
	Success(args ...any)
	Warning(args ...any)
	Error(scope string, args ...any)
	Panic(args ...any)
	Println(args ...any)
}

func NewLogger(loggerType LoggerType, fnScope string) Logger {
	switch loggerType {
	case File:
		file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic("cannot open log file: " + err.Error())
		}
		return newWriterLogger(file, fnScope, false)
	default:
		return newWriterLogger(os.Stdout, fnScope, true)
	}
}

type writerLogger struct {
	w        io.Writer
	fnScope  string
	useColor bool
}

func newWriterLogger(w io.Writer, fnScope string, useColor bool) Logger {
	return &writerLogger{
		w:        w,
		fnScope:  fnScope,
		useColor: useColor,
	}
}

func (l *writerLogger) log(tag, colorCode string, msg string) {
	if l.useColor {
		fmt.Fprintf(l.w, "%s[%s] %s%s\n", colorCode, tag, msg, reset)
	} else {
		fmt.Fprintf(l.w, "[%s] %s\n", tag, msg)
	}
}

const (
	blue    = "\033[34m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	red     = "\033[31m"
	magenta = "\033[35m"
	reset   = "\033[0m"
)

func (l *writerLogger) Info(args ...any) {
	l.log("💬INFO", blue, fmt.Sprint(args...))
}

func (l *writerLogger) Success(args ...any) {
	l.log("🎉SUCCESS", green, fmt.Sprint(args...))
}

func (l *writerLogger) Warning(args ...any) {
	l.log("⚠️WARNING", yellow, fmt.Sprint(args...))
}

func (l *writerLogger) Error(scope string, args ...any) {
	fullMsg := fmt.Sprintf("in %s[%s] → %s", l.fnScope, scope, fmt.Sprint(args...))
	l.log("🚨ERROR", red, fullMsg)
}

func (l *writerLogger) Panic(args ...any) {
	msg := fmt.Sprint(args...)
	if l.useColor {
		log.Panicf(magenta+"❌[PANIC] %s"+reset+"\n", msg)
	} else {
		log.Panicf("❌[PANIC] %s\n", msg)
	}
}

func (l *writerLogger) Println(args ...any) {
	msg := fmt.Sprint(args...)
	log.Println(msg)
}

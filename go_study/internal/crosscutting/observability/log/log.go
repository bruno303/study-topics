package log

import (
	"context"
	"fmt"
	"sync"
	"time"
)

var (
	log  Logger = NewDefaultLogger(LevelInfo)
	once sync.Once
)

type Level uint

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Debug(ctx context.Context, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, err error)
	SetLevel(l Level) error
}

func Log() Logger {
	return log
}

func SetLogger(lg Logger) {
	once.Do(func() {
		log = lg
	})
}

type DefaultLogger struct {
	level Level
}

func NewDefaultLogger(l Level) *DefaultLogger {
	return &DefaultLogger{level: l}
}

func (l *DefaultLogger) Info(ctx context.Context, msg string, args ...any) {
	if l.level > LevelInfo {
		return
	}
	fmt.Printf("%v - INFO  - "+msg+"\n", enrichArgs(args)...)
}

func (l *DefaultLogger) Debug(ctx context.Context, msg string, args ...any) {
	if l.level > LevelDebug {
		return
	}
	fmt.Printf("%v - DEBUG - "+msg+"\n", enrichArgs(args)...)
}

func (l *DefaultLogger) Warn(ctx context.Context, msg string, args ...any) {
	if l.level > LevelWarn {
		return
	}
	fmt.Printf("%v - WARN  - "+msg+"\n", enrichArgs(args)...)
}

func (l *DefaultLogger) Error(ctx context.Context, msg string, err error) {
	fmt.Printf("%v - ERROR - "+msg+" - Error %v\n", time.Now().String(), err)
}

func (dl *DefaultLogger) SetLevel(l Level) error {
	dl.level = l
	return nil
}

func enrichArgs(args []any) []any {
	richArgs := make([]any, 1, len(args)+1)
	richArgs[0] = time.Now().String()
	return append(richArgs, args...)
}

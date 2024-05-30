package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
)

type SlogLogger struct {
	log      *slog.Logger
	levelVar *slog.LevelVar
}

func DefaultSlogLogger() *SlogLogger {
	cfg := LogConfig{
		Level:  slog.LevelInfo,
		Output: OUTPUT_TEXT,
		CtxKey: CtxKey,
	}
	return NewSlogLogger(&cfg)
}

func NewSlogLogger(cfg *LogConfig) *SlogLogger {
	levelVar := &slog.LevelVar{}
	levelVar.Set(cfg.Level.Level())
	handlerOpts := slog.HandlerOptions{
		AddSource: false,
		Level:     levelVar,
	}
	var handler slog.Handler
	if cfg.Output == OUTPUT_JSON {
		handler = slog.NewJSONHandler(os.Stdout, &handlerOpts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, &handlerOpts)
	}

	return &SlogLogger{
		log:      slog.New(handler),
		levelVar: levelVar,
	}
}

func (l *SlogLogger) SetLevel(level slog.Level) {
	l.levelVar.Set(level)
}

func (l *SlogLogger) doLog(ctx context.Context, level slog.Level, msg string, args ...any) {
	if l.log.Enabled(ctx, level) {
		l.log.Log(ctx, level, fmt.Sprintf(msg, args...), l.getArgs(getMap(ctx))...)
	}
}

func (l *SlogLogger) Warn(ctx context.Context, msg string, args ...any) {
	l.doLog(ctx, slog.LevelWarn, msg, args...)
}

func (l *SlogLogger) Info(ctx context.Context, msg string, args ...any) {
	l.doLog(ctx, slog.LevelInfo, msg, args...)
}

func (l *SlogLogger) Debug(ctx context.Context, msg string, args ...any) {
	l.doLog(ctx, slog.LevelDebug, msg, args...)
}

func (l *SlogLogger) Error(ctx context.Context, msg string, err error) {
	if l.log.Enabled(ctx, slog.LevelError) {
		l.log.ErrorContext(ctx, fmt.Sprintf("%s: %v", msg, err), l.getArgs(getMap(ctx))...)
	}
}

func (l *SlogLogger) AddToCtx(ctx context.Context, key string, value any) context.Context {
	m := getMap(ctx)
	if m != nil {
		m.Store(key, value)
		return ctx
	}
	m = &sync.Map{}
	m.Store(key, value)
	return context.WithValue(ctx, CtxKey, m)
}

func (l *SlogLogger) getArgs(m *sync.Map) []any {
	logArgs := []any{}
	if m == nil {
		return logArgs
	}
	m.Range(func(key, value any) bool {
		if value == nil {
			logArgs = append(logArgs, key, "")
		} else {
			logArgs = append(logArgs, key, value)
		}
		return true
	})
	return logArgs
}

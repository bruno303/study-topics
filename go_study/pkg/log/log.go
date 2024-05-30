package log

import (
	"context"
	"log/slog"
	"sync"
)

type OutputType int
type LogCtxKey struct{ Key string }
type LogConfig struct {
	Level  slog.Leveler
	Output OutputType
	CtxKey any
}

var (
	logger Logger
	CtxKey = LogCtxKey{Key: "logCtx"}
)

const (
	OUTPUT_TEXT OutputType = iota
	OUTPUT_JSON
)

type Logger interface {
	Error(ctx context.Context, msg string, err error)
	Warn(ctx context.Context, msg string, args ...any)
	Info(ctx context.Context, msg string, args ...any)
	Debug(ctx context.Context, msg string, args ...any)
	SetLevel(level slog.Level)
	AddToCtx(ctx context.Context, key string, value any) context.Context
}

func SetDefaultLogger(lg Logger) {
	logger = lg
}

func SetLevel(level slog.Level) {
	logger.SetLevel(level)
}

func Error(ctx context.Context, msg string, err error) {
	logger.Error(ctx, msg, err)
}

func Warn(ctx context.Context, msg string, args ...any) {
	logger.Warn(ctx, msg, args...)
}

func Info(ctx context.Context, msg string, args ...any) {
	logger.Info(ctx, msg, args...)
}

func Debug(ctx context.Context, msg string, args ...any) {
	logger.Debug(ctx, msg, args...)
}

func AddToCtx(ctx context.Context, key string, value any) context.Context {
	return logger.AddToCtx(ctx, key, value)
}

func getMap(ctx context.Context) *sync.Map {
	val := ctx.Value(CtxKey)
	if val == nil {
		return nil
	}
	valMap, ok := val.(*sync.Map)
	if !ok {
		return nil
	}
	return valMap
}

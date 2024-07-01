package slog

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/log"
)

type (
	ExtractFunc func(context.Context) []any
	SlogAdapter struct {
		logger                *slog.Logger
		level                 *slog.LevelVar
		extractAdditionalInfo func(context.Context) []any
	}
	SlogAdapterOpts struct {
		Level                 log.Level
		FormatJson            bool
		ExtractAdditionalInfo func(context.Context) []any
	}
)

func NewSlogAdapter(opts SlogAdapterOpts) SlogAdapter {
	level := toSlogLevel(opts.Level)
	levelVar := &slog.LevelVar{}
	levelVar.Set(level)

	if opts.ExtractAdditionalInfo == nil {
		opts.ExtractAdditionalInfo = func(context.Context) []any { return nil }
	}

	handlerOpts := &slog.HandlerOptions{
		AddSource: false,
		Level:     levelVar,
	}

	var handler slog.Handler
	if opts.FormatJson {
		handler = slog.NewJSONHandler(os.Stdout, handlerOpts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, handlerOpts)
	}
	return SlogAdapter{
		logger:                slog.New(handler),
		level:                 levelVar,
		extractAdditionalInfo: opts.ExtractAdditionalInfo,
	}
}

func (l SlogAdapter) Info(ctx context.Context, msg string, args ...any) {
	if !l.logger.Enabled(ctx, slog.LevelInfo) {
		return
	}

	l.logger.InfoContext(ctx, fmt.Sprintf(msg, args...), l.extractAdditionalInfo(ctx)...)
}

func (l SlogAdapter) Debug(ctx context.Context, msg string, args ...any) {
	if !l.logger.Enabled(ctx, slog.LevelDebug) {
		return
	}
	l.logger.DebugContext(ctx, fmt.Sprintf(msg, args...), l.extractAdditionalInfo(ctx)...)
}

func (l SlogAdapter) Warn(ctx context.Context, msg string, args ...any) {
	if !l.logger.Enabled(ctx, slog.LevelWarn) {
		return
	}
	l.logger.WarnContext(ctx, fmt.Sprintf(msg, args...), l.extractAdditionalInfo(ctx)...)
}

func (l SlogAdapter) Error(ctx context.Context, msg string, err error) {
	additionalData := l.extractAdditionalInfo(ctx)
	additionalData = append(additionalData, "err", err.Error())
	l.logger.ErrorContext(ctx, msg, additionalData...)
}

func (la SlogAdapter) SetLevel(l log.Level) error {
	la.level.Set(toSlogLevel(l))
	return nil
}

func toSlogLevel(l log.Level) slog.Level {
	switch l {
	case log.LevelInfo:
		return slog.LevelInfo
	case log.LevelWarn:
		return slog.LevelWarn
	case log.LevelDebug:
		return slog.LevelDebug
	case log.LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

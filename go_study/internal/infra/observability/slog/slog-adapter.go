package slog

import (
	"context"
	"fmt"
	"log/slog"
	"main/internal/crosscutting/observability/log"
	"main/internal/crosscutting/observability/trace"
	"os"
)

type SlogAdapter struct {
	logger *slog.Logger
	level  *slog.LevelVar
}

type SlogAdapterOpts struct {
	Level      log.Level
	FormatJson bool
}

func NewSlogAdapter(opts SlogAdapterOpts) SlogAdapter {
	level := toSlogLevel(opts.Level)
	levelVar := &slog.LevelVar{}
	levelVar.Set(level)

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
		logger: slog.New(handler),
		level:  levelVar,
	}
}

func (l SlogAdapter) Info(ctx context.Context, msg string, args ...any) {
	if !l.logger.Enabled(ctx, slog.LevelInfo) {
		return
	}

	l.logger.InfoContext(ctx, fmt.Sprintf(msg, args...), getAdditionalData(ctx)...)
}

func (l SlogAdapter) Debug(ctx context.Context, msg string, args ...any) {
	if !l.logger.Enabled(ctx, slog.LevelDebug) {
		return
	}
	l.logger.DebugContext(ctx, fmt.Sprintf(msg, args...), getAdditionalData(ctx)...)
}

func (l SlogAdapter) Warn(ctx context.Context, msg string, args ...any) {
	if !l.logger.Enabled(ctx, slog.LevelWarn) {
		return
	}
	l.logger.WarnContext(ctx, fmt.Sprintf(msg, args...), getAdditionalData(ctx)...)
}

func (l SlogAdapter) Error(ctx context.Context, msg string, err error) {
	additionalData := getAdditionalData(ctx)
	additionalData = append(additionalData, "err", err.Error())
	l.logger.ErrorContext(ctx, msg, additionalData...)
}

func (la SlogAdapter) SetLevel(l log.Level) error {
	la.level.Set(toSlogLevel(l))
	return nil
}

func getAdditionalData(ctx context.Context) []any {
	var additionalLogData []any
	traceData := trace.ExtractTraceIds(ctx)
	if traceData.IsValid {
		additionalLogData = []any{
			"traceId", traceData.TraceId,
			"spanId", traceData.SpanId,
		}
	}
	return additionalLogData
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

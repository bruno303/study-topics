package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"main/pkg/log"
)

func main() {
	ctx := context.Background()
	configureLog()
	ctx = log.AddToCtx(log.AddToCtx(ctx, "traceId", "1234567789"), "spanId", "987654321")

	name := flag.String("name", "world", "Inform your name")
	flag.Parse()

	log.Debug(ctx, "Hello, debug")
	log.Info(ctx, "Hello, %s", *name)
	log.Warn(ctx, "Warning")
	log.Error(ctx, "Error while trying this", errors.New("test"))

	log.Info(context.Background(), "*****************************************************")

	log.SetLevel(slog.LevelWarn)
	log.Debug(ctx, "Hello, debug")
	log.Info(ctx, "Hello, %s", *name)
	log.Warn(ctx, "Warning")
	log.Error(ctx, "Error while trying this", errors.New("test"))
}

func configureLog() {
	cfg := log.LogConfig{
		Level:  slog.LevelDebug,
		Output: log.OUTPUT_TEXT,
		CtxKey: log.CtxKey,
	}
	log.SetDefaultLogger(log.NewSlogLogger(&cfg))
}

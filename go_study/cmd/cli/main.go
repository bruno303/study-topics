package main

import (
	"context"
	"errors"
	"flag"
	"strings"

	"github.com/bruno303/study-topics/go-study/internal/config"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/log"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	correlationid "github.com/bruno303/study-topics/go-study/internal/infra/observability/correlation-id"
	"github.com/bruno303/study-topics/go-study/internal/infra/observability/otel"
	"github.com/bruno303/study-topics/go-study/internal/infra/observability/slog"
)

func main() {
	ctx := context.Background()
	cfg := config.LoadConfig()
	shutdown, err := otel.SetupOTelSDK(ctx, cfg)
	panicIfErr(err)
	defer shutdown(ctx)

	configureLog(cfg)
	trace.SetTracer(otel.NewOtelTracerAdapter())

	ctx, end := trace.Trace(ctx, trace.NameConfig("Main", "Execution"))
	defer end()

	name := flag.String("name", "world", "Inform your name")
	flag.Parse()

	log.Log().Debug(ctx, "Hello, debug")
	log.Log().Info(ctx, "Hello, %s", *name)
	log.Log().Warn(ctx, "Warning")
	log.Log().Error(ctx, "Error while trying this", errors.New("test"))

	log.Log().Error(context.Background(), "*****************************************************", errors.New(""))

	if err := log.Log().SetLevel(log.LevelDebug); err != nil {
		panic(err)
	}

	log.Log().Debug(ctx, "Hello, debug")
	log.Log().Info(ctx, "Hello, %s", *name)
	log.Log().Warn(ctx, "Warning")
	log.Log().Error(ctx, "Error while trying this", errors.New("test"))
}

func configureLog(cfg *config.Config) {
	l := slog.NewSlogAdapter(slog.SlogAdapterOpts{
		Level:                 log.LevelWarn,
		FormatJson:            strings.ToUpper(cfg.Application.Log.Format) == "JSON",
		ExtractAdditionalInfo: logExtractor(),
	})
	log.SetLogger(l)
	// log.SetLogger(log.NewDefaultLogger(log.LevelWarn))
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func logExtractor() func(context.Context) []any {
	return func(ctx context.Context) []any {
		additionalLogData := make([]any, 0, 6)
		traceData := trace.ExtractTraceIds(ctx)
		if traceData.IsValid {
			additionalLogData = append(additionalLogData, "traceId", traceData.TraceId, "spanId", traceData.SpanId)
		}
		if correlationId, ok := correlationid.Get(ctx); ok {
			additionalLogData = append(additionalLogData, "correlationId", correlationId)
		}
		return additionalLogData
	}
}

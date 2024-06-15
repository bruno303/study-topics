package main

import (
	"context"
	"fmt"
	"main/internal/config"
	"main/internal/crosscutting/observability/log"
	"main/internal/crosscutting/observability/trace"
	"main/internal/infra/api/hello"
	correlationid "main/internal/infra/observability/correlation-id"
	"main/internal/infra/observability/otel"
	"main/internal/infra/observability/slog"
	"main/internal/infra/utils/shutdown"
	"net/http"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	ctx := context.Background()
	initialize(ctx)
}

func initialize(ctx context.Context) {
	cfg := config.LoadConfig()
	otelShutdown := configureObservability(ctx, cfg)
	shutdown.CreateListener(func() {
		otelShutdown(context.Background())
	})

	container := NewContainer(ctx, cfg)

	startKafkaConsumers(container)
	startProducer(container)
	startApi(ctx, cfg, container)

	log.Log().Info(ctx, "Application started")

	shutdown.AwaitAll()
}

func configureObservability(ctx context.Context, cfg *config.Config) func(context.Context) error {
	toSlogLevel := func(source string) (log.Level, error) {
		switch strings.ToUpper(source) {
		case "INFO":
			return log.LevelInfo, nil
		case "DEBUG":
			return log.LevelDebug, nil
		case "WARN":
			return log.LevelWarn, nil
		case "ERROR":
			return log.LevelError, nil
		default:
			return log.LevelInfo, fmt.Errorf("LogLevel %s is invalid", source)
		}
	}

	logLevel, err := toSlogLevel(cfg.Application.Log.Level)
	panicIfErr(err)
	log.SetLogger(
		slog.NewSlogAdapter(
			slog.SlogAdapterOpts{
				Level:                 logLevel,
				FormatJson:            strings.ToUpper(cfg.Application.Log.Format) == "JSON",
				ExtractAdditionalInfo: logExtractor(),
			},
		),
	)
	trace.SetTracer(otel.NewOtelTracerAdapter())

	otelShutdown, err := otel.SetupOTelSDK(ctx, cfg)
	panicIfErr(err)
	return otelShutdown
}

func startApi(ctx context.Context, cfg *config.Config, container *Container) {
	router := http.NewServeMux()
	hello.SetupApi(cfg, router, container.Services.HelloService, container.Repositories.HelloRepository)
	srv := &http.Server{Addr: ":8080", Handler: otelhttp.NewHandler(router, "/")}

	shutdown.CreateListener(func() {
		log.Log().Info(ctx, "Stopping API")
		if err := srv.Shutdown(context.Background()); err != nil {
			panic(err)
		}
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Log().Error(ctx, "Got error", err)
			}
		}
	}()
}

func startKafkaConsumers(container *Container) {
	for _, cons := range container.Kafka.Consumers {
		if err := cons.Start(); err != nil {
			panic(err)
		}
	}
}

func startProducer(container *Container) {
	container.Workers.HelloProducerWorker.Start()
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

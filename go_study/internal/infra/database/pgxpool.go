package database

import (
	"context"
	"fmt"
	"os"

	"github.com/bruno303/study-topics/go-study/internal/config"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/log"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace/attr"
	"github.com/bruno303/study-topics/go-study/internal/infra/utils/shutdown"
	"github.com/bruno303/study-topics/go-study/pkg/utils/array"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(config *config.Config) *pgxpool.Pool {
	ctx := context.Background()
	cfg := config.Database
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DatabaseName)

	pgxCfg, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		log.Log().Error(ctx, "Unable to parse connection string", err)
		os.Exit(1)
	}
	pgxCfg.ConnConfig.Tracer = &pgxTracer{}

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		log.Log().Error(ctx, "Unable to create connection pool", err)
		os.Exit(1)
	}
	if err = pool.Ping(ctx); err != nil {
		log.Log().Error(ctx, fmt.Sprintf("Unable to connect to %s:%d", cfg.Host, cfg.Port), err)
		os.Exit(1)
	}

	shutdown.CreateListener(func() {
		log.Log().Info(ctx, "Closing database pool")
		pool.Close()
	})

	return pool
}

type pgxTracer struct{}

func (p *pgxTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	trace.InjectAttributes(ctx, attr.New("Query", data.CommandTag.String()))
	if data.Err != nil {
		trace.InjectError(ctx, data.Err)
	}
	trace.GetTracer().EndTrace(ctx)
}

func (p *pgxTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	ctx, _ = trace.Trace(ctx, trace.NameConfig("PgxPool", "TraceQueryStart"))

	queryArgs := array.Join(array.Map(data.Args, func(a any) string {
		return fmt.Sprintf("%v", a)
	}), ",")

	trace.InjectAttributes(ctx,
		attr.New("SQL", data.SQL),
		attr.New("Args", queryArgs),
	)

	return ctx
}

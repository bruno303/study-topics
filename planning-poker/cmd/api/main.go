package main

import (
	"context"
	"net/http"
	"planning-poker/internal/infra"
	"planning-poker/internal/planningpoker"
	"strings"

	"github.com/bruno303/go-toolkit/pkg/log"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	ctx := context.Background()

	configureLogging()
	logger := log.NewLogger("main")

	container := NewContainer()

	r := mux.NewRouter()

	// mux := http.NewServeMux()
	planningpoker.ConfigurePlanningPokerAPI(r, container.Hub)
	infra.ConfigureInfraAPI(r)

	err := r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			logger.Debug(ctx, "ROUTE: %v", pathTemplate)
		}
		pathRegexp, err := route.GetPathRegexp()
		if err == nil {
			logger.Debug(ctx, "Path regexp: %v", pathRegexp)
		}
		queriesTemplates, err := route.GetQueriesTemplates()
		if err == nil {
			logger.Debug(ctx, "Queries templates: %v", strings.Join(queriesTemplates, ","))
		}
		queriesRegexps, err := route.GetQueriesRegexp()
		if err == nil {
			logger.Debug(ctx, "Queries regexps: %v", strings.Join(queriesRegexps, ","))
		}
		methods, err := route.GetMethods()
		if err == nil {
			logger.Debug(ctx, "Methods: %v", strings.Join(methods, ","))
		}
		logger.Debug(ctx, "")
		return nil
	})

	if err != nil {
		logger.Error(ctx, "Error walking routes", err)
	}

	logger.Info(ctx, "Starting server on :8080")
	if err = http.ListenAndServe(":8080", handlers.CORS(
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedHeaders([]string{"content-type", "authorization", "origin"}),
		handlers.AllowCredentials(),
	)(r)); err != nil {
		logger.Error(ctx, "Error starting server", err)
	}
}

func configureLogging() {
	log.SetLoggerFactory(func(name string) log.Logger {
		return log.NewSlogAdapter(
			log.SlogAdapterOpts{
				Level:                 log.LevelInfo,
				FormatJson:            false,
				ExtractAdditionalInfo: func(context.Context) []any { return []any{} },
				Source:                name,
			},
		)
	})
}

package middleware

import (
	"main/internal/crosscutting/observability/log"
	correlationid "main/internal/infra/observability/correlation-id"
	"net/http"
)

type CorrelationIdMiddleware struct {
	BaseMiddleware
}

func (m *CorrelationIdMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	correlationId := correlationid.Generate()
	log.Log().Debug(r.Context(), "CorrelationId generated: %s", correlationId)
	ctx := correlationid.Set(r.Context(), correlationId)
	m.Next(rw, r.WithContext(ctx))
}

func NewCorrelationIdMiddleware() *CorrelationIdMiddleware {
	return &CorrelationIdMiddleware{}
}

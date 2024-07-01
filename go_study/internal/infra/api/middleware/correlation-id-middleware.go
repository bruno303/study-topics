package middleware

import (
	"net/http"

	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/log"
	correlationid "github.com/bruno303/study-topics/go-study/internal/infra/observability/correlation-id"
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

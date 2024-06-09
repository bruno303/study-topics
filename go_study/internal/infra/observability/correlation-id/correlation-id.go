package correlationid

import (
	"context"

	"github.com/google/uuid"
)

type correlationIdContextKey struct {
	Key string
}

var (
	correlationIdKey = correlationIdContextKey{Key: "RequestIdContextKey"}
)

func Set(ctx context.Context, correlationId string) context.Context {
	return context.WithValue(ctx, correlationIdKey, correlationId)
}

func Get(ctx context.Context) (string, bool) {
	key := ctx.Value(correlationIdKey)
	if key == nil {
		return "", false
	}
	keyStr, bool := key.(string)
	return keyStr, bool
}

func Generate() string {
	return uuid.NewString()
}

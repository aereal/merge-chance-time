package logging

import (
	"context"
	"net/http"

	stackdriverlog "github.com/yfuruyama/stackdriver-request-context-log"
)

type keyType struct{}

var ctxKey = &keyType{}

func WithLogger(cfg *stackdriverlog.Config) func(http.Handler) http.Handler {
	mw := stackdriverlog.RequestLogging(cfg)
	return func(next http.Handler) http.Handler {
		return mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := stackdriverlog.RequestContextLogger(r)
			ctx := context.WithValue(r.Context(), ctxKey, logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		}))
	}
}

func GetLogger(ctx context.Context) *stackdriverlog.ContextLogger {
	return ctx.Value(ctxKey).(*stackdriverlog.ContextLogger)
}

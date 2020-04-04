package logging

import (
	"context"
	"log"
	"net/http"

	stackdriverlog "github.com/yfuruyama/stackdriver-request-context-log"
)

type keyType struct{}

var ctxKey = &keyType{}

func WithLogger(projectID string) func(http.Handler) http.Handler {
	cfg := stackdriverlog.NewConfig(projectID)
	mw := stackdriverlog.RequestLogging(cfg)
	return func(next http.Handler) http.Handler {
		return mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := stackdriverlog.RequestContextLogger(r)
			log.Printf("logger ok?=%v", logger != nil)
			ctx := context.WithValue(r.Context(), ctxKey, logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		}))
	}
}

func GetLogger(ctx context.Context) *stackdriverlog.ContextLogger {
	return ctx.Value(ctxKey).(*stackdriverlog.ContextLogger)
}

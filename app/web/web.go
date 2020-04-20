package web

import (
	"fmt"
	"net/http"

	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"github.com/aereal/merge-chance-time/app/config"
	"github.com/aereal/merge-chance-time/logging"
	"github.com/dimfeld/httptreemux/v5"
	"github.com/rs/cors"
	"go.opencensus.io/plugin/ochttp"
)

type Handler func(router *httptreemux.TreeMux)

func New(onGAE bool, cfg *config.Config, handlers ...Handler) *Web {
	return &Web{
		onGAE:     onGAE,
		projectID: cfg.GCPProjectID,
		handlers:  handlers,
	}
}

type Web struct {
	onGAE     bool
	projectID string
	handlers  []Handler
}

func (w *Web) Server(port string) *http.Server {
	var (
		host                 = "localhost"
		handler http.Handler = w.handler()
	)

	if w.onGAE {
		host = ""
		handler = &ochttp.Handler{
			Handler:     handler,
			Propagation: &propagation.HTTPFormat{},
		}
	}

	return &http.Server{
		Addr:    fmt.Sprintf("%s:%s", host, port),
		Handler: handler,
	}
}

func (w *Web) handler() http.Handler {
	router := httptreemux.New()
	mw := cors.New(cors.Options{
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowCredentials: true,
		AllowedHeaders:   []string{"*"},
		Debug:            true,
	})
	router.UseHandler(logging.WithLogger(w.projectID))
	router.UseHandler(mw.Handler)
	router.UseHandler(withDefaultHeaders)
	router.UsingContext().Handler(http.MethodGet, "/", http.HandlerFunc(handleRoot))
	for _, h := range w.handlers {
		h(router)
	}
	return router
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}

func withDefaultHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-frame-options", "deny")
		w.Header().Set("x-xss-protection", "1; mode=block")
		w.Header().Set("x-content-type-options", "nosniff")
		w.Header().Set("content-security-policy", "default-src 'none'")
		next.ServeHTTP(w, r)
	})
}

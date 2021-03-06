package web

import (
	"fmt"
	"net/http"

	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"github.com/99designs/gqlgen/graphql"
	"github.com/aereal/merge-chance-time/app/adapter/githubapps"
	"github.com/aereal/merge-chance-time/app/authz"
	"github.com/aereal/merge-chance-time/app/config"
	"github.com/aereal/merge-chance-time/authflow"
	"github.com/aereal/merge-chance-time/logging"
	"github.com/aereal/merge-chance-time/usecase"
	"github.com/dimfeld/httptreemux/v5"
	"github.com/rs/cors"
	stackdriverlog "github.com/yfuruyama/stackdriver-request-context-log"
	"go.opencensus.io/plugin/ochttp"
)

type Handler func(router *httptreemux.TreeMux)

func New(onGAE bool, cfg *config.Config, ghAdapter githubapps.GitHubAppsAdapter, uc usecase.Usecase, af authflow.GitHubAuthFlow, authorizer authz.Authorizer, es graphql.ExecutableSchema) *Web {
	return &Web{
		onGAE:               onGAE,
		projectID:           cfg.GCPProjectID,
		adminOrigin:         cfg.AdminOrigin.String(),
		githubWebhookSecret: cfg.GitHubAppConfig.WebhookSecret,
		ghAdapter:           ghAdapter,
		usecase:             uc,
		githubAuthFlow:      af,
		authorizer:          authorizer,
		es:                  es,
	}
}

type Web struct {
	onGAE               bool
	projectID           string
	adminOrigin         string
	githubWebhookSecret []byte
	ghAdapter           githubapps.GitHubAppsAdapter
	usecase             usecase.Usecase
	githubAuthFlow      authflow.GitHubAuthFlow
	authorizer          authz.Authorizer
	es                  graphql.ExecutableSchema
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
		AllowedOrigins: []string{w.adminOrigin},
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
	})
	cfg := stackdriverlog.NewConfig(w.projectID)
	router.UseHandler(logging.WithLogger(cfg))
	router.UseHandler(mw.Handler)
	router.UseHandler(withDefaultHeaders)
	router.UsingContext().Handler(http.MethodGet, "/", http.HandlerFunc(handleRoot))

	group := router.UsingContext().NewContextGroup("/app")
	group.POST("/webhook", w.handleWebhook())
	group.POST("/cron", w.handleCron())

	auth := router.UsingContext().NewContextGroup("/auth")
	auth.GET("/start", w.handleGetAuthStart())
	auth.GET("/callback", w.handleGetAuthCallback())

	srv := w.newHandler()
	apiGroup := router.UsingContext().NewContextGroup("/api")
	apiGroup.UseHandler(w.authorizer.Middleware())
	apiGroup.Handler(http.MethodOptions, "/query", srv)
	apiGroup.Handler(http.MethodPost, "/query", srv)

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

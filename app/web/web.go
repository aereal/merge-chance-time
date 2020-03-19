package web

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"github.com/dimfeld/httptreemux/v5"
	ctxlog "github.com/yfuruyama/stackdriver-request-context-log"
	"go.opencensus.io/plugin/ochttp"
)

func New(onGAE bool, projectID string) *Web {
	return &Web{
		onGAE:     onGAE,
		projectID: projectID,
	}
}

type Web struct {
	onGAE     bool
	projectID string
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
	cfg := ctxlog.NewConfig(w.projectID)
	router := httptreemux.New()
	handle := ctxlog.RequestLogging(cfg)
	router.UsingContext().Handler(http.MethodGet, "/", handle(http.HandlerFunc(handleRoot)))
	router.UsingContext().Handler(http.MethodPost, "/webhook", handle(http.HandlerFunc(handleWebhook)))
	return router
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	read, _ := ioutil.ReadAll(r.Body)
	logger := ctxlog.RequestContextLogger(r)
	logger.Infof("request body = %s", string(read))
	fmt.Fprintln(w, "OK")
}

package web

import (
	"fmt"
	"net/http"
	"strings"

	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"github.com/aereal/merge-chance-time/app/adapter/githubapps"
	"github.com/dimfeld/httptreemux/v5"
	"github.com/google/go-github/v30/github"
	ctxlog "github.com/yfuruyama/stackdriver-request-context-log"
	"go.opencensus.io/plugin/ochttp"
)

func New(onGAE bool, projectID string, ghAdapter *githubapps.GitHubAppsAdapter, githubWebhookSecret []byte) *Web {
	return &Web{
		onGAE:               onGAE,
		projectID:           projectID,
		githubWebhookSecret: githubWebhookSecret,
		ghAdapter:           ghAdapter,
	}
}

type Web struct {
	onGAE               bool
	projectID           string
	ghAdapter           *githubapps.GitHubAppsAdapter
	githubWebhookSecret []byte
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
	router.UsingContext().Handler(http.MethodPost, "/webhook", handle(http.HandlerFunc(w.handleWebhook)))
	router.UsingContext().Handler(http.MethodGet, "/cron", handle(http.HandlerFunc(w.handleCron)))
	return router
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}

func (c *Web) handleCron(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("x-appengine-cron") != "true" {
		http.Error(w, "invalid request", 400)
		return
	}
	fmt.Fprintln(w, "OK cron")
}

func (c *Web) handleWebhook(w http.ResponseWriter, r *http.Request) {
	logger := ctxlog.RequestContextLogger(r)

	payloadBytes, err := github.ValidatePayload(r, c.githubWebhookSecret)
	if err != nil {
		err = fmt.Errorf("failed to validate incoming payload: %w", err)
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	logger.Infof("webhook request body = %s", string(payloadBytes))
	payload, err := github.ParseWebHook(github.WebHookType(r), payloadBytes)
	if err != nil {
		err = fmt.Errorf("failed to parse incoming payload: %w", err)
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logger.Infof("webhook payload = %#v", payload)

	switch p := payload.(type) {
	case *github.PullRequestEvent:
		c.onPullRequest(w, r, p)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

func (c *Web) onPullRequest(w http.ResponseWriter, r *http.Request, payload *github.PullRequestEvent) {
	logger := ctxlog.RequestContextLogger(r)
	logger.Infof("Pull Request Event: %#v", payload)
	action := payload.GetAction()
	if action != "opened" && action != "synchronize" {
		logger.Warnf("Received action is %q skipping", action)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	ctx := r.Context()
	ghClient := c.ghAdapter.NewInstallationClient(payload.Installation.GetID())
	after := payload.GetAfter()
	logger.Infof("after commit = %q", after)
	fullName := payload.GetRepo().GetFullName()
	names := strings.Split(fullName, "/")
	if len(names) < 2 {
		http.Error(w, fmt.Sprintf("invalid repo.fullName: %q", fullName), http.StatusBadRequest)
		return
	}
	_, _, err := ghClient.Repositories.CreateStatus(ctx, names[0], names[1], after, &github.RepoStatus{
		State:   github.String("success"),
		Context: github.String("merge-chance-time"),
	})
	if err != nil {
		err = fmt.Errorf("failed to create status: %w", err)
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

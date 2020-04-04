package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"github.com/aereal/merge-chance-time/app/adapter/githubapps"
	"github.com/aereal/merge-chance-time/domain/model"
	"github.com/aereal/merge-chance-time/domain/repo"
	"github.com/aereal/merge-chance-time/logging"
	"github.com/aereal/merge-chance-time/usecase"
	"github.com/dimfeld/httptreemux/v5"
	"github.com/google/go-github/v30/github"
	"go.opencensus.io/plugin/ochttp"
)

func New(onGAE bool, projectID string, ghAdapter *githubapps.GitHubAppsAdapter, githubWebhookSecret []byte, repo *repo.Repository, uc *usecase.Usecase) *Web {
	return &Web{
		onGAE:               onGAE,
		projectID:           projectID,
		githubWebhookSecret: githubWebhookSecret,
		ghAdapter:           ghAdapter,
		repo:                repo,
		usecase:             uc,
	}
}

type Web struct {
	onGAE               bool
	projectID           string
	ghAdapter           *githubapps.GitHubAppsAdapter
	githubWebhookSecret []byte
	repo                *repo.Repository
	usecase             *usecase.Usecase
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
	router.UsingContext().Handler(http.MethodGet, "/", http.HandlerFunc(handleRoot))
	router.UsingContext().Handler(http.MethodPost, "/webhook", http.HandlerFunc(w.handleWebhook))
	router.UsingContext().Handler(http.MethodPost, "/cron", w.handleCron())
	return logging.WithLogger(w.projectID)(router)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}

func (c *Web) handleCron() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logging.GetLogger(ctx)

		logger.Infof("headers = %#v", r.Header)

		var payload PubSubPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			logger.Warnf("cannot read request: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "cannot read request: %+v\n", err)
			return
		}

		logger.Infof("payload.subscription=%q payload.message.id=%q publishTime=%q data=%q", payload.Subscription, payload.Message.ID, payload.Message.PublishTime, string(payload.Message.Data))

		baseTime := time.Time(payload.Message.PublishTime).Round(time.Minute)
		notice, err := c.usecase.NotifyEvent(ctx, baseTime)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("content-type", "application/json")
			json.NewEncoder(w).Encode(struct{ Error string }{err.Error()})
			return
		}

		w.Header().Set("content-type", "application/json")
		json.NewEncoder(w).Encode(notice)
	})
}

func (c *Web) handleWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.GetLogger(ctx)

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
	ctx := r.Context()
	logger := logging.GetLogger(ctx)
	logger.Infof("Pull Request Event: %#v", payload)
	action := payload.GetAction()
	if action != "opened" && action != "synchronize" {
		logger.Warnf("Received action is %q skipping", action)
		w.WriteHeader(http.StatusNoContent)
		return
	}
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

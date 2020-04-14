package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"github.com/aereal/merge-chance-time/app/adapter/githubapps"
	"github.com/aereal/merge-chance-time/app/authz"
	"github.com/aereal/merge-chance-time/app/config"
	"github.com/aereal/merge-chance-time/domain/repo"
	"github.com/aereal/merge-chance-time/logging"
	"github.com/aereal/merge-chance-time/usecase"
	"github.com/dimfeld/httptreemux/v5"
	"github.com/google/go-github/v30/github"
	"github.com/rs/cors"
	"go.opencensus.io/plugin/ochttp"
)

func New(onGAE bool, cfg *config.Config, ghAdapter *githubapps.GitHubAppsAdapter, repo *repo.Repository, uc *usecase.Usecase, authorizer *authz.Authorizer) *Web {
	return &Web{
		onGAE:                 onGAE,
		projectID:             cfg.GCPProjectID,
		githubWebhookSecret:   cfg.GitHubAppConfig.WebhookSecret,
		githubAppClientID:     cfg.GitHubAppConfig.ClientID,
		githubAppClientSecret: cfg.GitHubAppConfig.ClientSecret,
		ghAdapter:             ghAdapter,
		repo:                  repo,
		usecase:               uc,
		authorizer:            authorizer,
	}
}

type Web struct {
	onGAE                 bool
	projectID             string
	ghAdapter             *githubapps.GitHubAppsAdapter
	githubWebhookSecret   []byte
	githubAppClientID     string
	githubAppClientSecret string
	repo                  *repo.Repository
	usecase               *usecase.Usecase
	authorizer            *authz.Authorizer
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
	router.UsingContext().GET("/auth/callback", w.handleGetAuthCallback())
	router.UsingContext().GET("/api/user/installed_repos", w.handleGetUserInstalledRepos())
	loggingMW := logging.WithLogger(w.projectID)
	corsMW := cors.AllowAll()
	return corsMW.Handler(loggingMW(withDefaultHeaders(router)))
}

func (c *Web) handleGetAuthCallback() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		qs := r.URL.Query()

		w.Header().Set("content-type", "application/json")

		accessToken, err := c.ghAdapter.CreateUserAccessToken(ctx, qs.Get("code"), qs.Get("state"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(struct{ Error string }{err.Error()})
			return
		}
		cryptedToken, err := c.authorizer.IssueAuthenticationToken(&authz.AppClaims{AccessToken: accessToken})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(struct{ Error string }{fmt.Sprintf("cannot issue token: %+v", err)})
			return
		}

		escapedToken := url.QueryEscape(cryptedToken)
		http.Redirect(w, r, fmt.Sprintf("http://localhost:3000/auth/callback?accessToken=%s", escapedToken), http.StatusSeeOther)
	})
}

func (c *Web) handleGetUserInstalledRepos() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		w.Header().Set("content-type", "application/json")
		token := strings.Replace(r.Header.Get("authorization"), "Bearer ", "", 1)
		claims, err := c.authorizer.AuthenticateWithToken(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(struct{ Error string }{fmt.Sprintf("%+v", err)})
			return
		}

		ghClient := c.ghAdapter.NewUserClient(ctx, claims.AccessToken)
		userInstallations, _, err := ghClient.Apps.ListUserInstallations(ctx, nil)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(struct{ Error string }{fmt.Sprintf("%+v", err)})
			return
		}

		payload := struct {
			Repositories []*github.Repository `json:"repositories"`
		}{[]*github.Repository{}}
		for _, userInst := range userInstallations {
			rs, _, err := ghClient.Apps.ListUserRepos(ctx, userInst.GetID(), nil)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(struct{ Error string }{fmt.Sprintf("%+v", err)})
				return
			}
			payload.Repositories = append(payload.Repositories, rs...)
		}

		json.NewEncoder(w).Encode(payload)
	})
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
		if payload.Message == nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Header().Set("content-type", "application/json")
			json.NewEncoder(w).Encode(struct{ Error string }{"Invalid payload format"})
			return
		}

		logger.Infof("payload.subscription=%q payload.message.id=%q publishTime=%q data=%q", payload.Subscription, payload.Message.ID, payload.Message.PublishTime, string(payload.Message.Data))

		baseTime := time.Time(payload.Message.PublishTime)
		err := c.usecase.UpdateChanceTime(ctx, c.ghAdapter, baseTime)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("content-type", "application/json")
			json.NewEncoder(w).Encode(struct{ Error string }{err.Error()})
			return
		}

		w.WriteHeader(http.StatusNoContent)
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

	err := c.usecase.UpdatePullRequestCommitStatus(ctx, ghClient, payload.GetPullRequest())
	if err == usecase.ErrConfigNotFound {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("content-type", "application/json")
		json.NewEncoder(w).Encode(struct{ Error string }{err.Error()})
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("content-type", "application/json")
		json.NewEncoder(w).Encode(struct{ Error string }{err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

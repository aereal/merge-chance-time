package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"firebase.google.com/go/auth"
	"github.com/aereal/merge-chance-time/app/adapter/githubapps"
	"github.com/aereal/merge-chance-time/app/authz"
	"github.com/aereal/merge-chance-time/domain/repo"
	"github.com/aereal/merge-chance-time/logging"
	"github.com/aereal/merge-chance-time/usecase"
	"github.com/dimfeld/httptreemux/v5"
	"github.com/google/go-github/v30/github"
	"github.com/rs/cors"
	"go.opencensus.io/plugin/ochttp"
)

func New(onGAE bool, projectID string, ghAdapter *githubapps.GitHubAppsAdapter, githubWebhookSecret []byte, repo *repo.Repository, uc *usecase.Usecase, authClient *auth.Client, httpClient *http.Client, authorizer *authz.Authorizer) *Web {
	return &Web{
		onGAE:               onGAE,
		projectID:           projectID,
		githubWebhookSecret: githubWebhookSecret,
		ghAdapter:           ghAdapter,
		repo:                repo,
		usecase:             uc,
		authClient:          authClient,
		httpClient:          httpClient,
		authorizer:          authorizer,
	}
}

type Web struct {
	onGAE               bool
	projectID           string
	ghAdapter           *githubapps.GitHubAppsAdapter
	githubWebhookSecret []byte
	repo                *repo.Repository
	usecase             *usecase.Usecase
	authClient          *auth.Client
	httpClient          *http.Client
	authorizer          *authz.Authorizer
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
	if !w.onGAE {
		group := router.NewGroup("/api")
		group.UsingContext().Handler(http.MethodGet, "/config", w.handleListRepositoryConfigs())
		group.UsingContext().Handler(http.MethodGet, "/repos/:owner/:name/config", w.handleGetRepositoryConfig())
		group.UsingContext().Handler(http.MethodPut, "/repos/:owner/:name/config", w.handlePutRepositoryConfig())
		group.UsingContext().GET("/installations", w.handleListInstallations())
		group.UsingContext().GET("/me", w.handleGetMe())
	}
	loggingMW := logging.WithLogger(w.projectID)
	corsMW := cors.AllowAll()
	return corsMW.Handler(loggingMW(withDefaultHeaders(router)))
}

func (c *Web) handleGetAuthCallback() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logging.GetLogger(ctx)
		qs := r.URL.Query()
		logger.Infof("query = %#v", qs)

		w.Header().Set("content-type", "application/json")

		params := url.Values{}
		params.Set("client_id", "")
		params.Set("client_secret", "")
		params.Set("code", qs.Get("code"))
		authReq, err := http.NewRequest(http.MethodPost, "https://github.com/login/oauth/access_token", strings.NewReader(params.Encode()))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(struct{ Error string }{err.Error()})
			return
		}
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		resp, err := c.httpClient.Do(authReq.WithContext(ctx))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(struct{ Error string }{err.Error()})
			return
		}
		body, _ := ioutil.ReadAll(resp.Body)
		respBody, _ := url.ParseQuery(string(body))

		accessToken := respBody.Get("access_token")
		logger.Infof("access_token = %q", accessToken)
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

func (c *Web) handleGetMe() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		w.Header().Set("content-type", "application/json")
		idToken := strings.Replace(r.Header.Get("authorization"), "Bearer ", "", 1)
		token, err := c.authClient.VerifyIDToken(ctx, idToken)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(struct{ Error string }{fmt.Sprintf("%+v", err)})
			return
		}

		userInst, _, err := c.ghAdapter.NewAppClient().Apps.FindUserInstallation(ctx, token.Claims["name"].(string))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(struct{ Error string }{fmt.Sprintf("%+v", err)})
			return
		}
		installationClient := c.ghAdapter.NewInstallationClient(userInst.GetID())
		repos, _, err := installationClient.Apps.ListUserRepos(ctx, userInst.GetID(), &github.ListOptions{PerPage: 20})
		if err != nil {
			// w.WriteHeader(http.StatusUnauthorized)
			// json.NewEncoder(w).Encode(struct{ Error string }{fmt.Sprintf("%+v", err)})
			// return
			logging.GetLogger(ctx).Infof("failed to list user repos: %+v", err)
		}
		if repos == nil {
			repos = []*github.Repository{}
		}

		payload := struct {
			Token           *auth.Token            `json:"token"`
			Claims          map[string]interface{} `json:"claims"`
			Repositories    []*github.Repository   `json:"repositories"`
			OrgRepositories []*github.Repository   `json:"org_repositories"`
		}{token, token.Claims, repos, []*github.Repository{}}

		orgRepos, _, err := installationClient.Repositories.ListByOrg(ctx, "oneetyan", nil)
		if err != nil {
			logging.GetLogger(ctx).Infof("failed to list repos by org: %s", err)
		}
		payload.OrgRepositories = append(payload.OrgRepositories, orgRepos...)

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

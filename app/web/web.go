package web

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"github.com/bradleyfalzon/ghinstallation"
	"github.com/dgrijalva/jwt-go"
	"github.com/dimfeld/httptreemux/v5"
	"github.com/google/go-github/v30/github"
	ctxlog "github.com/yfuruyama/stackdriver-request-context-log"
	"go.opencensus.io/plugin/ochttp"
)

func New(onGAE bool, projectID string, githubAppID int, githubWebhookSecret []byte, githubAppPrivateKey *rsa.PrivateKey, httpClient *http.Client) *Web {
	return &Web{
		onGAE:               onGAE,
		projectID:           projectID,
		githubAppID:         githubAppID,
		githubWebhookSecret: githubWebhookSecret,
		githubAppPrivateKey: githubAppPrivateKey,
		httpClient:          httpClient,
	}
}

type Web struct {
	onGAE               bool
	projectID           string
	githubAppID         int
	githubWebhookSecret []byte
	githubAppPrivateKey *rsa.PrivateKey
	httpClient          *http.Client
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
	return router
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
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
	case *github.InstallationRepositoriesEvent:
		c.onInstallationRepositoriesEvent(w, r, p)
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
	ghClient := c.createInstallationClient(payload.Installation)
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

func (c *Web) onInstallationRepositoriesEvent(w http.ResponseWriter, r *http.Request, payload *github.InstallationRepositoriesEvent) {
	logger := ctxlog.RequestContextLogger(r)

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, "https://api.github.com/app", nil)
	if err != nil {
		err = fmt.Errorf("failed to build request: %w", err)
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Add("accept", "application/vnd.github.machine-man-preview+json")
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": now.Unix(),
		"exp": now.Add(time.Minute * 9).Unix(),
		"iss": c.githubAppID,
	})
	tokenStr, err := token.SignedString(c.githubAppPrivateKey)
	if err != nil {
		err = fmt.Errorf("failed to sign JWT: %w", err)
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Add("authorization", fmt.Sprintf("Bearer %s", tokenStr))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to request: %w", err)
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, _ := ioutil.ReadAll(resp.Body)
	logger.Infof("response body = %s", string(b))

	w.WriteHeader(http.StatusNoContent)
}

func (c *Web) createAppClient() *github.Client {
	tr := ghinstallation.NewAppsTransportFromPrivateKey(c.httpClient.Transport, int64(c.githubAppID), c.githubAppPrivateKey)
	return github.NewClient(&http.Client{Transport: tr})
}

func (c *Web) createInstallationClient(inst *github.Installation) *github.Client {
	atr := ghinstallation.NewAppsTransportFromPrivateKey(c.httpClient.Transport, int64(c.githubAppID), c.githubAppPrivateKey)
	itr := ghinstallation.NewFromAppsTransport(atr, inst.GetID())
	return github.NewClient(&http.Client{Transport: itr})
}

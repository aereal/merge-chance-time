package web

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
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
	default:
		fmt.Fprintln(w, "OK")
	}
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

	fmt.Fprintln(w, "OK")
}

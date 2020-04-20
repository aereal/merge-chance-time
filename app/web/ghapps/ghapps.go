package ghapps

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aereal/merge-chance-time/app/adapter/githubapps"
	"github.com/aereal/merge-chance-time/app/config"
	"github.com/aereal/merge-chance-time/logging"
	"github.com/aereal/merge-chance-time/usecase"
	"github.com/dimfeld/httptreemux/v5"
	"github.com/google/go-github/v30/github"
)

type PubSubPayload struct {
	Message      *PubSubMessage `json:"message"`
	Subscription string         `json:"subscription"`
}

type PubSubMessage struct {
	Data        json.RawMessage `json:"data"`
	ID          string          `json:"messageId"`
	PublishTime PublishTime     `json:"publishTime"`
}

type PublishTime time.Time

func (t *PublishTime) UnmarshalText(text []byte) error {
	parsed, err := time.ParseInLocation(time.RFC3339Nano, string(text), time.Local)
	if err != nil {
		return err
	}
	*t = PublishTime(parsed)
	return nil
}

func (t PublishTime) MarshalText() ([]byte, error) {
	return []byte(time.Time(t).Format(time.RFC3339Nano)), nil
}

func (t PublishTime) String() string {
	return time.Time(t).Format(time.RFC3339Nano)
}

func New(cfg *config.GitHubAppConfig, ghAdapter *githubapps.GitHubAppsAdapter, uc *usecase.Usecase) (*Web, error) {
	if cfg == nil {
		return nil, fmt.Errorf("cfg is nil")
	}
	if ghAdapter == nil {
		return nil, fmt.Errorf("ghAdapter is nil")
	}
	if uc == nil {
		return nil, fmt.Errorf("uc is nil")
	}
	return &Web{
		githubWebhookSecret: cfg.WebhookSecret,
		ghAdapter:           ghAdapter,
		usecase:             uc,
	}, nil
}

type Web struct {
	githubWebhookSecret []byte
	ghAdapter           *githubapps.GitHubAppsAdapter
	usecase             *usecase.Usecase
}

func (a *Web) Routes() func(router *httptreemux.TreeMux) {
	return func(router *httptreemux.TreeMux) {
		group := router.UsingContext().NewContextGroup("/app")
		group.POST("/webhook", a.handleWebhook)
		group.POST("/cron", a.handleCron())
	}
}

func (c *Web) handleCron() http.HandlerFunc {
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
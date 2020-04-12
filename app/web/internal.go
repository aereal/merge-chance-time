package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/aereal/merge-chance-time/domain/model"
	"github.com/aereal/merge-chance-time/domain/repo"
	"github.com/aereal/merge-chance-time/usecase"
	"github.com/dimfeld/httptreemux/v5"
)

func getMe(ctx context.Context) error {
	return nil
}

func (c *Web) handlePutRepositoryConfig() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ghClient := c.ghAdapter.NewAppClient()
		params := httptreemux.ContextParams(ctx)

		if ct := r.Header.Get("content-type"); ct != "application/json" {
			json.NewEncoder(w).Encode(struct{ Error string }{Error: fmt.Sprintf("Invalid content-type: %s", ct)})
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		w.Header().Set("content-type", "application/json")

		owner, name := params["owner"], params["name"]
		err := c.usecase.PutRepositoryConfig(ctx, ghClient, owner, name, r.Body)
		if errors.Is(err, usecase.ErrInvalidInput) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(struct{ Error string }{
				Error: fmt.Sprintf("failed to decode request body: %s", err),
			})
			return
		}
		if errors.Is(err, usecase.ErrInstallationNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(struct{ Error string }{Error: "Not installed"})
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(struct{ Error string }{Error: err.Error()})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

type RepositoryConfigAggr struct {
	// Repository *github.Repository      `json:"repo"`
	Config *model.RepositoryConfig `json:"config"`
}

func (c *Web) handleGetRepositoryConfig() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ghClient := c.ghAdapter.NewAppClient()
		params := httptreemux.ContextParams(ctx)

		owner, name := params["owner"], params["name"]
		w.Header().Set("content-type", "application/json")
		inst, _, err := ghClient.Apps.FindRepositoryInstallation(ctx, owner, name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(struct{ Error string }{Error: err.Error()})
			return
		}
		if inst == nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(struct{ Error string }{Error: "Not installed"})
			return
		}

		cfg, err := c.repo.GetRepositoryConfig(ctx, owner, name)
		if err == repo.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(struct{ Error string }{Error: "repository config not found"})
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(struct{ Error string }{Error: err.Error()})
			return
		}

		json.NewEncoder(w).Encode(&RepositoryConfigAggr{
			Config: cfg,
		})
	})
}

func (c *Web) handleListRepositoryConfigs() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		w.Header().Set("content-type", "application/json")
		configs, err := c.repo.ListRepositoryConfigs(ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(struct{ Error string }{Error: err.Error()})
			return
		}

		json.NewEncoder(w).Encode(configs)
	})
}

func (c *Web) handleListInstallations() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ghClient := c.ghAdapter.NewAppClient()
		w.Header().Set("content-type", "application/json")
		installations, _, err := ghClient.Apps.ListInstallations(ctx, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(struct{ Error string }{err.Error()})
			return
		}

		json.NewEncoder(w).Encode(installations)
	})
}

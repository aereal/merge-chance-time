package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aereal/merge-chance-time/app/adapter/githubapps"
	"github.com/aereal/merge-chance-time/app/authz"
	"github.com/dimfeld/httptreemux/v5"
	"github.com/google/go-github/v30/github"
)

func New(ghAdapter *githubapps.GitHubAppsAdapter, authorizer *authz.Authorizer) (*Web, error) {
	return &Web{
		ghAdapter:  ghAdapter,
		authorizer: authorizer,
	}, nil
}

type Web struct {
	ghAdapter  *githubapps.GitHubAppsAdapter
	authorizer *authz.Authorizer
}

func (s *Web) Routes() func(router *httptreemux.TreeMux) {
	return func(router *httptreemux.TreeMux) {
		group := router.UsingContext().NewContextGroup("/api")
		group.GET("/user/installed_repos", s.handleGetUserInstalledRepos())
		group.OPTIONS("/user/installed_repos", s.handleGetUserInstalledRepos())
	}
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

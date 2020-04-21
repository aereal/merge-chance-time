package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/aereal/gqlgen-tracer-opencensus/tracer"
	"github.com/aereal/merge-chance-time/app/adapter/githubapps"
	"github.com/aereal/merge-chance-time/app/authz"
	"github.com/dimfeld/httptreemux/v5"
	"github.com/google/go-github/v30/github"
)

func New(ghAdapter *githubapps.GitHubAppsAdapter, authorizer *authz.Authorizer, es graphql.ExecutableSchema) (*Web, error) {
	if es == nil {
		return nil, fmt.Errorf("graphql.ExecutableSchema is nil")
	}
	return &Web{
		ghAdapter:  ghAdapter,
		authorizer: authorizer,
		es:         es,
	}, nil
}

type Web struct {
	ghAdapter  *githubapps.GitHubAppsAdapter
	authorizer *authz.Authorizer
	es         graphql.ExecutableSchema
}

func (s *Web) Routes() func(router *httptreemux.TreeMux) {
	srv := s.newHandler()
	return func(router *httptreemux.TreeMux) {
		group := router.UsingContext().NewContextGroup("/api")
		group.UseHandler(s.authorizer.Middleware())
		group.GET("/user/installed_repos", s.handleGetUserInstalledRepos())
		group.OPTIONS("/user/installed_repos", s.handleGetUserInstalledRepos())
		group.Handler(http.MethodOptions, "/query", srv)
		group.Handler(http.MethodPost, "/query", srv)
	}
}

func (s *Web) newHandler() *handler.Server {
	srv := handler.New(s.es)
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.POST{})
	srv.Use(extension.Introspection{})
	srv.Use(tracer.Tracer{})
	return srv
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

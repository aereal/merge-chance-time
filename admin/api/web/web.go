package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/aereal/merge-chance-time/app/adapter/githubapps"
	"github.com/aereal/merge-chance-time/app/authz"
	"github.com/aereal/merge-chance-time/authflow"
	"github.com/aereal/merge-chance-time/logging"
	"github.com/dimfeld/httptreemux/v5"
	"github.com/google/go-github/v30/github"
)

func New(ghAdapter *githubapps.GitHubAppsAdapter, authorizer *authz.Authorizer, githubAuthFlow *authflow.GitHubAuthFlow) (*Web, error) {
	return &Web{
		ghAdapter:      ghAdapter,
		authorizer:     authorizer,
		githubAuthFlow: githubAuthFlow,
	}, nil
}

type Web struct {
	ghAdapter      *githubapps.GitHubAppsAdapter
	authorizer     *authz.Authorizer
	githubAuthFlow *authflow.GitHubAuthFlow
}

func (s *Web) Routes() func(router *httptreemux.TreeMux) {
	return func(router *httptreemux.TreeMux) {
		group := router.UsingContext().NewContextGroup("/api")
		group.GET("/auth/start", s.handleGetAuthStart())
		group.GET("/auth/callback", s.handleGetAuthCallback())
		group.GET("/user/installed_repos", s.handleGetUserInstalledRepos())
		group.OPTIONS("/user/installed_repos", s.handleGetUserInstalledRepos())
	}
}

func (c *Web) handleGetAuthStart() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		initiatorURL, err := getInitiatorURL(r)
		if err != nil {
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(struct{ Error string }{err.Error()})
			return
		}

		nextURL, err := c.githubAuthFlow.NewAuthorizeURL(ctx, buildCurrentOrigin(r), initiatorURL.String())
		if err != nil {
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(struct{ Error string }{err.Error()})
			return
		}
		http.Redirect(w, r, nextURL, http.StatusSeeOther)
	})
}

func (c *Web) handleGetAuthCallback() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		qs := r.URL.Query()

		w.Header().Set("content-type", "application/json")

		initiatorURL, err := c.githubAuthFlow.NavigateAuthCompletion(ctx, qs.Get("code"), qs.Get("state"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(struct{ Error string }{err.Error()})
			return
		}

		http.Redirect(w, r, initiatorURL.String(), http.StatusSeeOther)
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

func buildCurrentOrigin(r *http.Request) string {
	host := r.Host
	if forwardedHost := r.Header.Get("x-forwarded-host"); forwardedHost != "" {
		host = forwardedHost
	}
	proto := "http"
	if r.TLS != nil {
		proto = "https"
	}
	if forwardedProto := r.Header.Get("x-forwarded-proto"); forwardedProto != "" {
		proto = forwardedProto
	}
	return fmt.Sprintf("%s://%s", proto, host)
}

func getInitiatorURL(r *http.Request) (*url.URL, error) {
	raw := r.URL.Query().Get("initiator_url")
	if raw == "" {
		return nil, fmt.Errorf("initiator_url is empty")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	referrer, err := url.Parse(r.Referer())
	if err != nil {
		return nil, fmt.Errorf("referrer is invalid URL: %w", err)
	}
	initiatorOrigin := origin(parsed)
	referrerOrigin := origin(referrer)
	logger := logging.GetLogger(r.Context())
	logger.Infof("initiatorOrigin=%s referrerOrigin=%s", initiatorOrigin, referrerOrigin)
	if initiatorOrigin != referrerOrigin {
		return nil, fmt.Errorf("origin of initiator_url and referrer are different")
	}
	return parsed, nil
}

func origin(url *url.URL) string {
	return fmt.Sprintf("%s://%s", url.Scheme, url.Host)
}

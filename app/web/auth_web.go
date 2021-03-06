package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/aereal/merge-chance-time/logging"
)

func (c *Web) handleGetAuthStart() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logging.GetLogger(ctx)

		initiatorURL, err := getInitiatorURL(r)
		if err != nil {
			logger.Errorf("getInitiatorURL: %v", err)
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
	rawReferrer := r.Referer()
	if rawReferrer == "" {
		return nil, fmt.Errorf("referer is empty")
	}
	referrer, err := url.Parse(rawReferrer)
	if err != nil {
		return nil, fmt.Errorf("referrer is invalid URL: %w", err)
	}
	initiatorOrigin := origin(parsed)
	referrerOrigin := origin(referrer)
	logger := logging.GetLogger(r.Context())
	logger.Infof("initiator=%s initiatorOrigin=%s referrer=%s referrerOrigin=%s", parsed, initiatorOrigin, referrer, referrerOrigin)
	if initiatorOrigin != referrerOrigin {
		return nil, fmt.Errorf("origin of initiator_url and referrer are different")
	}
	return parsed, nil
}

func origin(url *url.URL) string {
	return fmt.Sprintf("%s://%s", url.Scheme, url.Host)
}

package auth

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/aereal/merge-chance-time/authflow"
	"github.com/aereal/merge-chance-time/logging"
	"github.com/golang/mock/gomock"
	stackdriverlog "github.com/yfuruyama/stackdriver-request-context-log"
)

func TestStart(t *testing.T) {
	cfg := stackdriverlog.NewConfig("")
	cfg.ContextLogOut, cfg.RequestLogOut = ioutil.Discard, ioutil.Discard
	mw := logging.WithLogger(cfg)

	cases := []struct {
		name          string
		statusCode    int
		pathQuery     string
		referrer      string
		buildAuthflow func(ctrl *gomock.Controller) authflow.GitHubAuthFlow
	}{
		{
			name:       "ok",
			statusCode: http.StatusSeeOther,
			pathQuery:  "/?initiator_url=http%3A%2F%2Fexample.com%2Finit",
			referrer:   "http://example.com/init",
			buildAuthflow: func(ctrl *gomock.Controller) authflow.GitHubAuthFlow {
				af := authflow.NewMockGitHubAuthFlow(ctrl)
				af.EXPECT().
					NewAuthorizeURL(
						gomock.Any(),
						gomock.AssignableToTypeOf(""),
						gomock.Eq("http://example.com/init"),
					).
					Times(1)
				return af
			},
		},
		{
			name:       "mismatch origin",
			statusCode: http.StatusBadRequest,
			pathQuery:  "/?initiator_url=http%3A%2F%2Fevilexample.com%2Finit",
			referrer:   "http://example.com/init",
			buildAuthflow: func(ctrl *gomock.Controller) authflow.GitHubAuthFlow {
				af := authflow.NewMockGitHubAuthFlow(ctrl)
				return af
			},
		},
		{
			name:       "no referrer",
			statusCode: http.StatusBadRequest,
			pathQuery:  "/?initiator_url=http%3A%2F%2Fexample.com%2Finit",
			referrer:   "",
			buildAuthflow: func(ctrl *gomock.Controller) authflow.GitHubAuthFlow {
				af := authflow.NewMockGitHubAuthFlow(ctrl)
				return af
			},
		},
		{
			name:       "no initiator_url",
			statusCode: http.StatusBadRequest,
			pathQuery:  "/",
			referrer:   "",
			buildAuthflow: func(ctrl *gomock.Controller) authflow.GitHubAuthFlow {
				af := authflow.NewMockGitHubAuthFlow(ctrl)
				return af
			},
		},
		{
			name:       "NewAuthorizeURL returns error",
			statusCode: http.StatusInternalServerError,
			pathQuery:  "/?initiator_url=http%3A%2F%2Fexample.com%2Finit",
			referrer:   "http://example.com/init",
			buildAuthflow: func(ctrl *gomock.Controller) authflow.GitHubAuthFlow {
				af := authflow.NewMockGitHubAuthFlow(ctrl)
				af.EXPECT().
					NewAuthorizeURL(
						gomock.Any(),
						gomock.AssignableToTypeOf(""),
						gomock.Eq("http://example.com/init"),
					).
					Times(1).
					Return("", fmt.Errorf("oops"))
				return af
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			w := &Web{
				githubAuthFlow: c.buildAuthflow(ctrl),
			}
			srv := httptest.NewServer(mw(w.handleGetAuthStart()))
			defer srv.Close()

			req, err := http.NewRequest(http.MethodGet, srv.URL+c.pathQuery, nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("referer", c.referrer)
			client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("resp=%#v", resp)
			if resp.StatusCode != c.statusCode {
				t.Errorf("status code expected=%d got=%d", c.statusCode, resp.StatusCode)
			}
		})
	}
}

func TestCallback(t *testing.T) {
	cfg := stackdriverlog.NewConfig("")
	cfg.ContextLogOut, cfg.RequestLogOut = ioutil.Discard, ioutil.Discard
	mw := logging.WithLogger(cfg)

	cases := []struct {
		name          string
		statusCode    int
		pathQuery     string
		buildAuthflow func(ctrl *gomock.Controller) authflow.GitHubAuthFlow
		redirectURL   *url.URL
	}{
		{
			name:        "ok",
			statusCode:  http.StatusSeeOther,
			pathQuery:   "/?code=xxx",
			redirectURL: mustURI("http://example.com/"),
			buildAuthflow: func(ctrl *gomock.Controller) authflow.GitHubAuthFlow {
				af := authflow.NewMockGitHubAuthFlow(ctrl)
				af.EXPECT().
					NavigateAuthCompletion(gomock.Any(), gomock.Eq("xxx"), gomock.Eq("")).
					Times(1).
					Return(mustURI("http://example.com/"), nil)
				return af
			},
		},
		{
			name:        "ok w/state",
			statusCode:  http.StatusSeeOther,
			pathQuery:   "/?code=xxx&state=yyy",
			redirectURL: mustURI("http://example.com/"),
			buildAuthflow: func(ctrl *gomock.Controller) authflow.GitHubAuthFlow {
				af := authflow.NewMockGitHubAuthFlow(ctrl)
				af.EXPECT().
					NavigateAuthCompletion(gomock.Any(), gomock.Eq("xxx"), gomock.Eq("yyy")).
					Times(1).
					Return(mustURI("http://example.com/"), nil)
				return af
			},
		},
		{
			name:        "NavigateAuthCompletion returns error",
			statusCode:  http.StatusInternalServerError,
			pathQuery:   "/?code=xxx&state=yyy",
			redirectURL: nil,
			buildAuthflow: func(ctrl *gomock.Controller) authflow.GitHubAuthFlow {
				af := authflow.NewMockGitHubAuthFlow(ctrl)
				af.EXPECT().
					NavigateAuthCompletion(gomock.Any(), gomock.Eq("xxx"), gomock.Eq("yyy")).
					Times(1).
					Return(nil, fmt.Errorf("oops"))
				return af
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			w := &Web{
				githubAuthFlow: c.buildAuthflow(ctrl),
			}
			srv := httptest.NewServer(mw(w.handleGetAuthCallback()))
			defer srv.Close()

			req, err := http.NewRequest(http.MethodGet, srv.URL+c.pathQuery, nil)
			if err != nil {
				t.Fatal(err)
			}
			client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("resp=%#v", resp)
			if resp.StatusCode != c.statusCode {
				t.Errorf("status code expected=%d got=%d", c.statusCode, resp.StatusCode)
			}
			loc, err := resp.Location()
			if err != nil && err != http.ErrNoLocation {
				t.Fatal(err)
			}
			if fmt.Sprintf("%s", loc) != fmt.Sprintf("%s", c.redirectURL) {
				t.Errorf("location expected %q but got %q", c.redirectURL, loc)
			}
		})
	}
}

func mustURI(raw string) *url.URL {
	parsed, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return parsed
}

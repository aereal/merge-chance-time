package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aereal/merge-chance-time/app/adapter/githubapps"
	"github.com/aereal/merge-chance-time/logging"
	"github.com/aereal/merge-chance-time/usecase"
	"github.com/golang/mock/gomock"
	"github.com/google/go-github/v30/github"
	stackdriverlog "github.com/yfuruyama/stackdriver-request-context-log"
)

func TestCron(t *testing.T) {
	cfg := stackdriverlog.NewConfig("")
	cfg.ContextLogOut, cfg.RequestLogOut = ioutil.Discard, ioutil.Discard
	mw := logging.WithLogger(cfg)
	now := time.Now().Truncate(time.Second)

	cases := []struct {
		name         string
		reqBody      interface{}
		statusCode   int
		buildUsecase func(ctrl *gomock.Controller) usecase.Usecase
	}{
		{
			name: "ok",
			reqBody: &PubSubPayload{
				Subscription: "0xdeadbeaf",
				Message: &PubSubMessage{
					Data:        json.RawMessage("{}"),
					ID:          "xxx",
					PublishTime: PublishTime(now),
				},
			},
			statusCode: http.StatusNoContent,
			buildUsecase: func(ctrl *gomock.Controller) usecase.Usecase {
				uc := usecase.NewMockUsecase(ctrl)
				uc.EXPECT().UpdateChanceTime(gomock.Any(), gomock.Any(), eqTime(now)).AnyTimes()
				return uc
			},
		},
		{
			name:       "invalid",
			reqBody:    nil,
			statusCode: http.StatusUnprocessableEntity,
			buildUsecase: func(ctrl *gomock.Controller) usecase.Usecase {
				uc := usecase.NewMockUsecase(ctrl)
				return uc
			},
		},
		{
			name: "empty message",
			reqBody: &PubSubPayload{
				Subscription: "0xdeadbead",
			},
			statusCode: http.StatusUnprocessableEntity,
			buildUsecase: func(ctrl *gomock.Controller) usecase.Usecase {
				uc := usecase.NewMockUsecase(ctrl)
				return uc
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			buf := new(bytes.Buffer)
			json.NewEncoder(buf).Encode(c.reqBody)

			w := &Web{
				usecase: c.buildUsecase(ctrl),
			}
			srv := httptest.NewServer(mw(w.handleCron()))
			defer srv.Close()

			resp, err := http.Post(srv.URL, "application/json", strings.NewReader(buf.String()))
			if err != nil {
				t.Error(err)
			}
			t.Logf("resp=%#v", resp)
			if resp.StatusCode != c.statusCode {
				t.Errorf("status code expected=%d got=%d", c.statusCode, resp.StatusCode)
			}
		})
	}
}

func TestWebhook(t *testing.T) {
	cfg := stackdriverlog.NewConfig("")
	cfg.ContextLogOut, cfg.RequestLogOut = ioutil.Discard, ioutil.Discard
	mw := logging.WithLogger(cfg)

	cases := []struct {
		name           string
		eventType      string
		reqBody        interface{}
		statusCode     int
		buildUsecase   func(ctrl *gomock.Controller) usecase.Usecase
		buildGhAdapter func(ctrl *gomock.Controller) githubapps.GitHubAppsAdapter
	}{
		// installation
		{
			name:      "installed",
			eventType: "installation",
			reqBody: &github.InstallationEvent{
				Action: stringRef("created"),
				Repositories: []*github.Repository{
					{},
					{},
				},
			},
			statusCode: http.StatusNoContent,
			buildUsecase: func(ctrl *gomock.Controller) usecase.Usecase {
				uc := usecase.NewMockUsecase(ctrl)
				uc.EXPECT().OnInstallRepositories(gomock.Any(), gomock.Len(2)).Times(1)
				return uc
			},
		},
		{
			name:      "uninstalled",
			eventType: "installation",
			reqBody: &github.InstallationEvent{
				Action: stringRef("deleted"),
				Installation: &github.Installation{
					Account: &github.User{
						Login: stringRef("aereal"),
					},
				},
			},
			statusCode: http.StatusNoContent,
			buildUsecase: func(ctrl *gomock.Controller) usecase.Usecase {
				uc := usecase.NewMockUsecase(ctrl)
				uc.EXPECT().OnDeleteAppFromOwner(gomock.Any(), gomock.Eq("aereal")).Times(1)
				return uc
			},
		},
		{
			name:      "installation w/unhandled action",
			eventType: "installation",
			reqBody: &github.InstallationRepositoriesEvent{
				Action: stringRef("suspend"),
			},
			statusCode: http.StatusUnprocessableEntity,
			buildUsecase: func(ctrl *gomock.Controller) usecase.Usecase {
				uc := usecase.NewMockUsecase(ctrl)
				return uc
			},
		},

		// installation_repositories
		{
			name:      "installed on repository",
			eventType: "installation_repositories",
			reqBody: &github.InstallationRepositoriesEvent{
				Action:            stringRef("added"),
				RepositoriesAdded: []*github.Repository{{}, {}},
			},
			statusCode: http.StatusNoContent,
			buildUsecase: func(ctrl *gomock.Controller) usecase.Usecase {
				uc := usecase.NewMockUsecase(ctrl)
				uc.EXPECT().OnInstallRepositories(gomock.Any(), gomock.Len(2)).Times(1)
				return uc
			},
		},
		{
			name:      "uninstall on repository",
			eventType: "installation_repositories",
			reqBody: &github.InstallationRepositoriesEvent{
				Action:              stringRef("removed"),
				RepositoriesRemoved: []*github.Repository{{}, {}},
			},
			statusCode: http.StatusNoContent,
			buildUsecase: func(ctrl *gomock.Controller) usecase.Usecase {
				uc := usecase.NewMockUsecase(ctrl)
				uc.EXPECT().OnRemoveRepositories(gomock.Any(), gomock.Len(2)).Times(1)
				return uc
			},
		},
		{
			name:      "installation_repositories w/unhandled action",
			eventType: "installation_repositories",
			reqBody: &github.InstallationRepositoriesEvent{
				Action: stringRef("unknown"),
			},
			statusCode: http.StatusUnprocessableEntity,
			buildUsecase: func(ctrl *gomock.Controller) usecase.Usecase {
				uc := usecase.NewMockUsecase(ctrl)
				return uc
			},
		},

		// pull_request
		{
			name:      "pull_request opened",
			eventType: "pull_request",
			reqBody: &github.PullRequestEvent{
				Action: stringRef("opened"),
				Installation: &github.Installation{
					ID: int64ref(1234),
				},
				PullRequest: &github.PullRequest{},
			},
			statusCode: http.StatusNoContent,
			buildGhAdapter: func(ctrl *gomock.Controller) githubapps.GitHubAppsAdapter {
				a := githubapps.NewMockGitHubAppsAdapter(ctrl)
				a.EXPECT().NewInstallationClient(gomock.Eq(int64(1234))).Times(1)
				return a
			},
			buildUsecase: func(ctrl *gomock.Controller) usecase.Usecase {
				uc := usecase.NewMockUsecase(ctrl)
				uc.EXPECT().
					UpdatePullRequestCommitStatus(gomock.Any(), gomock.Any(), gomock.Eq(&github.PullRequest{})).
					Return(nil).
					Times(1)
				return uc
			},
		},
		{
			name:      "pull_request opened / config not found",
			eventType: "pull_request",
			reqBody: &github.PullRequestEvent{
				Action: stringRef("opened"),
				Installation: &github.Installation{
					ID: int64ref(1234),
				},
				PullRequest: &github.PullRequest{},
			},
			statusCode: http.StatusNotFound,
			buildGhAdapter: func(ctrl *gomock.Controller) githubapps.GitHubAppsAdapter {
				a := githubapps.NewMockGitHubAppsAdapter(ctrl)
				a.EXPECT().NewInstallationClient(gomock.Eq(int64(1234))).Times(1)
				return a
			},
			buildUsecase: func(ctrl *gomock.Controller) usecase.Usecase {
				uc := usecase.NewMockUsecase(ctrl)
				uc.EXPECT().
					UpdatePullRequestCommitStatus(gomock.Any(), gomock.Any(), gomock.Eq(&github.PullRequest{})).
					Return(usecase.ErrConfigNotFound).
					Times(1)
				return uc
			},
		},
		{
			name:      "pull_request opened / failed to update status",
			eventType: "pull_request",
			reqBody: &github.PullRequestEvent{
				Action: stringRef("opened"),
				Installation: &github.Installation{
					ID: int64ref(1234),
				},
				PullRequest: &github.PullRequest{},
			},
			statusCode: http.StatusInternalServerError,
			buildGhAdapter: func(ctrl *gomock.Controller) githubapps.GitHubAppsAdapter {
				a := githubapps.NewMockGitHubAppsAdapter(ctrl)
				a.EXPECT().NewInstallationClient(gomock.Eq(int64(1234))).Times(1)
				return a
			},
			buildUsecase: func(ctrl *gomock.Controller) usecase.Usecase {
				uc := usecase.NewMockUsecase(ctrl)
				uc.EXPECT().
					UpdatePullRequestCommitStatus(gomock.Any(), gomock.Any(), gomock.Eq(&github.PullRequest{})).
					Return(fmt.Errorf("oops")).
					Times(1)
				return uc
			},
		},
		{
			name:      "pull_request synchronize",
			eventType: "pull_request",
			reqBody: &github.PullRequestEvent{
				Action: stringRef("synchronize"),
				Installation: &github.Installation{
					ID: int64ref(1234),
				},
				PullRequest: &github.PullRequest{},
			},
			statusCode: http.StatusNoContent,
			buildGhAdapter: func(ctrl *gomock.Controller) githubapps.GitHubAppsAdapter {
				a := githubapps.NewMockGitHubAppsAdapter(ctrl)
				a.EXPECT().NewInstallationClient(gomock.Eq(int64(1234))).Times(1)
				return a
			},
			buildUsecase: func(ctrl *gomock.Controller) usecase.Usecase {
				uc := usecase.NewMockUsecase(ctrl)
				uc.EXPECT().
					UpdatePullRequestCommitStatus(gomock.Any(), gomock.Any(), gomock.Eq(&github.PullRequest{})).
					Return(nil).
					Times(1)
				return uc
			},
		},
		{
			name:      "pull_request closed",
			eventType: "pull_request",
			reqBody: &github.PullRequestEvent{
				Action: stringRef("closed"),
				Installation: &github.Installation{
					ID: int64ref(1234),
				},
				PullRequest: &github.PullRequest{},
			},
			statusCode: http.StatusNoContent,
			buildGhAdapter: func(ctrl *gomock.Controller) githubapps.GitHubAppsAdapter {
				a := githubapps.NewMockGitHubAppsAdapter(ctrl)
				return a
			},
			buildUsecase: func(ctrl *gomock.Controller) usecase.Usecase {
				uc := usecase.NewMockUsecase(ctrl)
				return uc
			},
		},

		// others
		{
			name:      "integration_installation",
			eventType: "integration_installation",
			reqBody: &github.InstallationRepositoriesEvent{
				Action: stringRef("created"),
			},
			statusCode: http.StatusNoContent,
			buildUsecase: func(ctrl *gomock.Controller) usecase.Usecase {
				uc := usecase.NewMockUsecase(ctrl)
				return uc
			},
		},
		{
			name:      "integration_installation_repositories",
			eventType: "integration_installation_repositories",
			reqBody: &github.InstallationRepositoriesEvent{
				Action: stringRef("added"),
			},
			statusCode: http.StatusNoContent,
			buildUsecase: func(ctrl *gomock.Controller) usecase.Usecase {
				uc := usecase.NewMockUsecase(ctrl)
				return uc
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			buf := new(bytes.Buffer)
			if err := json.NewEncoder(buf).Encode(c.reqBody); err != nil {
				t.Fatal(err)
			}

			w := &Web{
				usecase: c.buildUsecase(ctrl),
			}
			if c.buildGhAdapter != nil {
				w.ghAdapter = c.buildGhAdapter(ctrl)
			}
			srv := httptest.NewServer(mw(w.handleWebhook()))
			defer srv.Close()

			req, err := http.NewRequest(http.MethodPost, srv.URL, strings.NewReader(buf.String()))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("x-github-event", c.eventType)
			req.Header.Set("content-type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			b, _ := ioutil.ReadAll(resp.Body)
			t.Logf("resp=%#v body=%q", resp, string(b))
			if resp.StatusCode != c.statusCode {
				t.Errorf("status code expected=%d got=%d", c.statusCode, resp.StatusCode)
			}
		})
	}
}

type timeMatcher struct {
	expected time.Time
}

func (m timeMatcher) Matches(x interface{}) bool {
	t, ok := x.(time.Time)
	if !ok {
		return false
	}
	return t.Equal(m.expected)
}

func (m timeMatcher) String() string {
	return fmt.Sprintf("is equal to %s (%#v)", m.expected, m.expected)
}

func eqTime(expected time.Time) gomock.Matcher {
	return gomock.GotFormatterAdapter(
		gomock.GotFormatterFunc(func(i interface{}) string {
			return fmt.Sprintf("%s (%#v)", i, i)
		}),
		timeMatcher{expected},
	)
}

func stringRef(s string) *string {
	return &s
}

func int64ref(i int64) *int64 {
	return &i
}

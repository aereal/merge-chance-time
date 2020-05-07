package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/aereal/merge-chance-time/app/adapter/githubapi"
	"github.com/aereal/merge-chance-time/app/adapter/githubapps"
	"github.com/aereal/merge-chance-time/app/authz"
	"github.com/aereal/merge-chance-time/app/graph"
	"github.com/aereal/merge-chance-time/app/graph/generated"
	"github.com/aereal/merge-chance-time/domain/repo"
	"github.com/golang/mock/gomock"
	"github.com/google/go-github/v30/github"
)

func TestQuery(t *testing.T) {
	cases := []struct {
		name       string
		statusCode int
		expected   graphql.Response
		params     graphql.RawParams
		build      func(ctrl *gomock.Controller) *aggregate
	}{
		{
			name:       "ok",
			statusCode: http.StatusOK,
			params: graphql.RawParams{
				Query: "query($owner: String!, $name: String!) {repository(owner: $owner, name: $name){id}}",
				Variables: map[string]interface{}{
					"owner": "aereal",
					"name":  "example-repo",
				},
			},
			expected: graphql.Response{
				Data: json.RawMessage(`{"repository":{"id":1234}}`),
			},
			build: func(ctrl *gomock.Controller) *aggregate {
				a := authz.NewMockAuthorizer(ctrl)
				a.EXPECT().
					Middleware().
					AnyTimes().
					Return(func(next http.Handler) http.Handler {
						return next
					})
				a.EXPECT().GetCurrentClaims(gomock.Any()).Times(1).Return(&authz.AppClaims{AccessToken: "0xdeadbeaf"}, nil)

				ad := githubapps.NewMockGitHubAppsAdapter(ctrl)
				mockClient := githubapi.NewMockClient(ctrl)
				mockRepoSrv := githubapi.NewMockRepositoriesService(ctrl)
				mockRepoSrv.EXPECT().Get(gomock.Any(), gomock.Eq("aereal"), gomock.Eq("example-repo")).Times(1).Return(&github.Repository{
					ID:       github.Int64(1234),
					Name:     github.String("example-repo"),
					FullName: github.String("aereal/example-repo"),
					Owner: &github.User{
						Login: github.String("aereal"),
					},
				}, nil, nil)
				mockClient.EXPECT().Repositories().Times(1).Return(mockRepoSrv)
				ad.EXPECT().NewUserClient(gomock.Any(), gomock.Eq("0xdeadbeaf")).Times(1).Return(mockClient)

				r := repo.NewMockRepository(ctrl)

				aggr := &aggregate{
					authorizer: a,
					adapter:    ad,
					repo:       r,
				}
				return aggr
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			aggr := c.build(ctrl)
			es, err := aggr.executableSchema()
			if err != nil {
				t.Fatal(err)
			}
			w := &Web{
				authorizer: aggr.authorizer,
				es:         es,
			}
			srv := httptest.NewServer(w.handler())
			defer srv.Close()

			buf := new(bytes.Buffer)
			if err := json.NewEncoder(buf).Encode(c.params); err != nil {
				t.Fatal(err)
			}
			req, err := http.NewRequest(http.MethodPost, srv.URL+"/api/query", buf)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("content-type", "application/json")
			client := http.DefaultClient
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			var got graphql.Response
			if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
				t.Fatal(err)
			}
			t.Logf("response body=%#v response=%#v", got, resp)
			if resp.StatusCode != c.statusCode {
				t.Errorf("status code expected=%d got=%d", c.statusCode, resp.StatusCode)
			}
			for _, e := range got.Errors {
				t.Log(e)
			}
			t.Logf("got data=%s", string(got.Data))
			if !reflect.DeepEqual(got, c.expected) {
				t.Errorf("\nexpected=%#v\n     got=%#v", c.expected, got)
			}
		})
	}
}

type aggregate struct {
	authorizer authz.Authorizer
	adapter    githubapps.GitHubAppsAdapter
	repo       repo.Repository
}

func (a aggregate) executableSchema() (graphql.ExecutableSchema, error) {
	res, err := graph.New(a.authorizer, a.adapter, a.repo)
	if err != nil {
		return nil, err
	}
	return generated.NewExecutableSchema(generated.Config{Resolvers: res}), nil
}

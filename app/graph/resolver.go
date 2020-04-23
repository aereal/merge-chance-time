//go:generate go run github.com/99designs/gqlgen

package graph

import (
	"fmt"

	"github.com/aereal/merge-chance-time/app/adapter/githubapps"
	"github.com/aereal/merge-chance-time/app/authz"
)

func New(authorizer *authz.Authorizer, ghAdapter *githubapps.GitHubAppsAdapter) (*Resolver, error) {
	if authorizer == nil {
		return nil, fmt.Errorf("authorizer is nil")
	}
	if ghAdapter == nil {
		return nil, fmt.Errorf("ghAdapter is nil")
	}
	return &Resolver{
		authorizer: authorizer,
		ghAdapter:  ghAdapter,
	}, nil
}

type Resolver struct {
	authorizer *authz.Authorizer
	ghAdapter  *githubapps.GitHubAppsAdapter
}

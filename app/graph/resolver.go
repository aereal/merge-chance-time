//go:generate go run github.com/99designs/gqlgen

package graph

import (
	"fmt"

	"github.com/aereal/merge-chance-time/app/adapter/githubapps"
	"github.com/aereal/merge-chance-time/app/authz"
	"github.com/aereal/merge-chance-time/domain/repo"
)

func New(authorizer *authz.Authorizer, ghAdapter *githubapps.GitHubAppsAdapter, repo repo.Repository) (*Resolver, error) {
	if authorizer == nil {
		return nil, fmt.Errorf("authorizer is nil")
	}
	if ghAdapter == nil {
		return nil, fmt.Errorf("ghAdapter is nil")
	}
	if repo == nil {
		return nil, fmt.Errorf("repo is nil")
	}
	return &Resolver{
		authorizer: authorizer,
		ghAdapter:  ghAdapter,
		repo:       repo,
	}, nil
}

type Resolver struct {
	authorizer *authz.Authorizer
	ghAdapter  *githubapps.GitHubAppsAdapter
	repo       repo.Repository
}

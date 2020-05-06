//go:generate mockgen -package githubapi -destination api_mock.go . RepositoriesService,PullRequestService,AppsService,UsersService

package githubapi

import (
	"context"

	"github.com/google/go-github/v30/github"
)

func New(client *github.Client) Client {
	return &clientImpl{
		ghClient: client,
	}
}

type Client interface {
	Apps() AppsService
	PullRequests() PullRequestService
	Repositories() RepositoriesService
	Users() UsersService
}

type clientImpl struct {
	ghClient *github.Client
}

func (c *clientImpl) Apps() AppsService {
	return c.ghClient.Apps
}

func (c *clientImpl) PullRequests() PullRequestService {
	return c.ghClient.PullRequests
}

func (c *clientImpl) Repositories() RepositoriesService {
	return c.ghClient.Repositories
}

func (c *clientImpl) Users() UsersService {
	return c.ghClient.Users
}

type RepositoriesService interface {
	CreateStatus(ctx context.Context, owner, repo, ref string, status *github.RepoStatus) (*github.RepoStatus, *github.Response, error)
	Get(ctx context.Context, owner, repo string) (*github.Repository, *github.Response, error)
}

type PullRequestService interface {
	List(ctx context.Context, owner string, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error)
}

type AppsService interface {
	ListInstallations(ctx context.Context, opts *github.ListOptions) ([]*github.Installation, *github.Response, error)
	ListUserRepos(ctx context.Context, id int64, opts *github.ListOptions) ([]*github.Repository, *github.Response, error)
	ListUserInstallations(ctx context.Context, opts *github.ListOptions) ([]*github.Installation, *github.Response, error)
}

type UsersService interface {
	Get(ctx context.Context, user string) (*github.User, *github.Response, error)
}

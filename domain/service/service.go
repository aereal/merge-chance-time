package service

import (
	"context"
	"fmt"

	"github.com/aereal/merge-chance-time/app/adapter/githubapi"
	"github.com/google/go-github/v30/github"
)

var (
	ctxName = "merge-chance-time"
)

func New(ghClient githubapi.Client) (Service, error) {
	if ghClient == nil {
		return nil, fmt.Errorf("ghClient is nil")
	}
	return &serviceImpl{
		ghClient: ghClient,
	}, nil
}

type Service interface {
	ApprovePullRequest(ctx context.Context, pr *github.PullRequest) error
	PendingPullRequest(ctx context.Context, pr *github.PullRequest) error
}

type serviceImpl struct {
	ghClient githubapi.Client
}

func (s *serviceImpl) ApprovePullRequest(ctx context.Context, pr *github.PullRequest) error {
	return s.createCommitStatus(ctx, pr, "success")
}

func (s *serviceImpl) PendingPullRequest(ctx context.Context, pr *github.PullRequest) error {
	return s.createCommitStatus(ctx, pr, "pending")
}

func (s *serviceImpl) createCommitStatus(ctx context.Context, pr *github.PullRequest, state string) error {
	head := pr.GetHead()
	repo := head.GetRepo()
	desc := fmt.Sprintf("%s is open", ctxName)
	if state != "success" {
		desc = fmt.Sprintf("%s is pending", ctxName)
	}
	status := &github.RepoStatus{
		State:       &state,
		Context:     &ctxName,
		Description: &desc,
	}
	_, _, err := s.ghClient.Repositories().CreateStatus(ctx, repo.GetOwner().GetLogin(), repo.GetName(), head.GetSHA(), status)
	if err != nil {
		return fmt.Errorf("failed to create status on %s#%d: %w", repo.GetFullName(), pr.GetNumber(), err)
	}
	return nil
}

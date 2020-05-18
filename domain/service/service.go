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
	ApproveRepository(ctx context.Context, owner, name string) error
	ApprovePullRequest(ctx context.Context, pr *github.PullRequest) error
	PendingRepository(ctx context.Context, owner, name string) error
	PendingPullRequest(ctx context.Context, pr *github.PullRequest) error
}

type serviceImpl struct {
	ghClient githubapi.Client
}

func (s *serviceImpl) ApproveRepository(ctx context.Context, owner, name string) error {
	prs, _, err := s.ghClient.PullRequests().List(ctx, owner, name, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch pull requests on %s/%s: %w", owner, name, err)
	}
	for _, pr := range prs {
		if err := s.ApprovePullRequest(ctx, pr); err != nil {
			return fmt.Errorf("failed to approve pull request %s/%s#%d: %w", owner, name, pr.GetNumber(), err)
		}
	}
	return nil
}

func (s *serviceImpl) ApprovePullRequest(ctx context.Context, pr *github.PullRequest) error {
	return s.createCommitStatus(ctx, pr, "success")
}

func (s *serviceImpl) PendingRepository(ctx context.Context, owner, name string) error {
	prs, _, err := s.ghClient.PullRequests().List(ctx, owner, name, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch pull requests on %s/%s: %w", owner, name, err)
	}
	for _, pr := range prs {
		if err := s.PendingPullRequest(ctx, pr); err != nil {
			return fmt.Errorf("failed to pending pull request %s/%s#%d: %w", owner, name, pr.GetNumber(), err)
		}
	}
	return nil
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

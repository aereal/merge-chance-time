package service

import (
	"context"
	"fmt"

	"github.com/google/go-github/v30/github"
)

var (
	ctxName = "merge-chance-time"
)

func New() (*Service, error) {
	return &Service{}, nil
}

type Service struct{}

func (s *Service) ApprovePullRequest(ctx context.Context, client *github.Client, pr *github.PullRequest) error {
	return s.createCommitStatus(ctx, client, pr, "success")
}

func (s *Service) PendingPullRequest(ctx context.Context, client *github.Client, pr *github.PullRequest) error {
	return s.createCommitStatus(ctx, client, pr, "pending")
}

func (s *Service) createCommitStatus(ctx context.Context, client *github.Client, pr *github.PullRequest, state string) error {
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
	_, _, err := client.Repositories.CreateStatus(ctx, repo.GetOwner().GetLogin(), repo.GetName(), head.GetSHA(), status)
	if err != nil {
		return fmt.Errorf("failed to create status on %s#%d: %w", repo.GetFullName(), pr.GetNumber(), err)
	}
	return nil
}
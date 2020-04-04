package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/aereal/merge-chance-time/domain/model"
	"github.com/aereal/merge-chance-time/domain/repo"
	"github.com/google/go-github/v30/github"
)

var (
	ErrInvalidInput         = fmt.Errorf("invalid input")
	ErrInstallationNotFound = fmt.Errorf("repository installation not found")
)

func New(repo *repo.Repository) (*Usecase, error) {
	if repo == nil {
		return nil, fmt.Errorf("repo is nil")
	}
	return &Usecase{
		repo: repo,
	}, nil
}

type Usecase struct {
	repo *repo.Repository
}

func (u *Usecase) CreateRepositoryConfig(ctx context.Context, ghAppClient *github.Client, owner, name string, input io.Reader) error {
	var cfg model.RepositoryConfig
	if err := json.NewDecoder(input).Decode(&cfg); err != nil {
		return fmt.Errorf("failed to decode input as JSON: %w", ErrInvalidInput)
	}
	cfg.Owner = owner
	cfg.Name = name

	if err := cfg.Valid(); err != nil {
		return err
	}

	installation, _, err := ghAppClient.Apps.FindRepositoryInstallation(ctx, owner, name)
	if err != nil {
		return fmt.Errorf("failed to find repository installation: %w", err)
	}
	if installation == nil {
		return ErrInstallationNotFound
	}

	if err := u.repo.CreateRepositoryConfig(ctx, owner, name, &cfg); err != nil {
		return fmt.Errorf("failed to create repository config: %w", err)
	}

	return nil
}

type Notification struct {
	ReposToBeStarted []string
	ReposToBeStopped []string
}

func (u *Usecase) NotifyEvent(ctx context.Context, baseTime time.Time) (*Notification, error) {
	configs, err := u.repo.ListRepositoryConfigs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list repository config: %w", err)
	}

	notice := &Notification{ReposToBeStarted: []string{}, ReposToBeStopped: []string{}}
	for _, cfg := range configs {
		fullName := fmt.Sprintf("%s/%s", cfg.Owner, cfg.Name)
		if cfg.ShouldStartOn(baseTime) {
			notice.ReposToBeStarted = append(notice.ReposToBeStarted, fullName)
		}
		if cfg.ShouldStopOn(baseTime) {
			notice.ReposToBeStopped = append(notice.ReposToBeStopped, fullName)
		}
	}

	return notice, nil
}

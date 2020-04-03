package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

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

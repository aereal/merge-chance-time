package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aereal/merge-chance-time/app/adapter/githubapps"
	"github.com/aereal/merge-chance-time/domain/model"
	"github.com/aereal/merge-chance-time/domain/repo"
	"github.com/aereal/merge-chance-time/domain/service"
	"github.com/aereal/merge-chance-time/logging"
	"github.com/google/go-github/v30/github"
	"golang.org/x/sync/errgroup"
)

var (
	ErrInvalidInput         = fmt.Errorf("invalid input")
	ErrInstallationNotFound = fmt.Errorf("repository installation not found")
	ErrConfigNotFound       = fmt.Errorf("repository config not found")
)

func New(repo *repo.Repository) (Usecase, error) {
	if repo == nil {
		return nil, fmt.Errorf("repo is nil")
	}
	return &usecaseImpl{
		repo: repo,
	}, nil
}

type usecaseImpl struct {
	repo *repo.Repository
}

type Usecase interface {
	OnDeleteAppFromOwner(ctx context.Context, owner string) error
	OnRemoveRepositories(ctx context.Context, repos []*github.Repository) error
	OnInstallRepositories(ctx context.Context, repos []*github.Repository) error
	UpdateChanceTime(ctx context.Context, adapter *githubapps.GitHubAppsAdapter, baseTime time.Time) error
	PutRepositoryConfig(ctx context.Context, ghAppClient *github.Client, owner, name string, input io.Reader) error
	UpdatePullRequestCommitStatus(ctx context.Context, client *github.Client, pr *github.PullRequest) error
}

func (u *usecaseImpl) OnDeleteAppFromOwner(ctx context.Context, owner string) error {
	if err := u.repo.DeleteRepositoryConfigsByOwner(ctx, owner); err != nil {
		return err
	}
	return nil
}

func (u *usecaseImpl) OnRemoveRepositories(ctx context.Context, repos []*github.Repository) error {
	eg, ctx := errgroup.WithContext(ctx)
	for _, r := range repos {
		eg.Go(func() error {
			return u.onRemoveRepository(ctx, r)
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func (u *usecaseImpl) onRemoveRepository(ctx context.Context, removedRepo *github.Repository) error {
	parts := strings.Split(removedRepo.GetFullName(), "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid repo fullName")
	}
	return u.repo.DeleteRepositoryConfig(ctx, parts[0], parts[1])
}

func (u *usecaseImpl) OnInstallRepositories(ctx context.Context, repos []*github.Repository) error {
	eg, ctx := errgroup.WithContext(ctx)
	for _, r := range repos {
		eg.Go(func() error {
			return u.onInstallRepository(ctx, r)
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func (u *usecaseImpl) onInstallRepository(ctx context.Context, installedRepo *github.Repository) error {
	logger := logging.GetLogger(ctx)
	logger.Infof("install repository: %#v", installedRepo)
	parts := strings.Split(installedRepo.GetFullName(), "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid repo fullName")
	}
	return u.repo.PutRepositoryConfigs(ctx, []*model.RepositoryConfig{
		{
			Owner:          parts[0],
			Name:           parts[1],
			MergeAvailable: true,
		},
	})
}

func (u *usecaseImpl) PutRepositoryConfig(ctx context.Context, ghAppClient *github.Client, owner, name string, input io.Reader) error {
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

	if err := u.repo.PutRepositoryConfigs(ctx, []*model.RepositoryConfig{&cfg}); err != nil {
		return fmt.Errorf("failed to create repository config: %w", err)
	}

	return nil
}

func (u *usecaseImpl) UpdateChanceTime(ctx context.Context, adapter *githubapps.GitHubAppsAdapter, baseTime time.Time) error {
	logger := logging.GetLogger(ctx)
	installations, _, err := adapter.NewAppClient().Apps.ListInstallations(ctx, nil)
	if err != nil {
		return err
	}
	installationByOwner := map[string]*github.Installation{}
	for _, inst := range installations {
		owner := inst.GetAccount().GetLogin()
		installationByOwner[owner] = inst
	}

	configsByOwners, err := u.repo.ListConfigsByOwners(ctx)
	if err != nil {
		return fmt.Errorf("failed to list repository config: %w", err)
	}

	toBeUpdated := []*model.RepositoryConfig{}
	g, c := errgroup.WithContext(ctx)
	for _, configs := range configsByOwners {
		for _, cfg := range configs {
			config := cfg
			logger.Infof("owner=%s repo=%s", config.Owner, config.Name)
			install := installationByOwner[config.Owner]
			if install == nil {
				return fmt.Errorf("no installation found on %s", config.Owner)
			}
			installClient := adapter.NewInstallationClient(install.GetID())
			srv, err := service.New()
			if err != nil {
				return err
			}

			if config.ShouldStartOn(baseTime) {
				config.MergeAvailable = true
				toBeUpdated = append(toBeUpdated, config)

				g.Go(func() error {
					return updateCommitStatuses(c, installClient, install, config, srv, true)
				})
			}
			if config.ShouldStopOn(baseTime) {
				config.MergeAvailable = false
				toBeUpdated = append(toBeUpdated, config)

				g.Go(func() error {
					return updateCommitStatuses(c, installClient, install, config, srv, false)
				})
			}
		}
	}
	if len(toBeUpdated) > 0 {
		if err := u.repo.PutRepositoryConfigs(c, toBeUpdated); err != nil {
			return fmt.Errorf("failed to update config: %w", err)
		}
	}
	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to update commit status: %w", err)
	}

	return nil
}

func (u *usecaseImpl) UpdatePullRequestCommitStatus(ctx context.Context, client *github.Client, pr *github.PullRequest) error {
	targetRepo := pr.GetHead().GetRepo()
	config, err := u.repo.GetRepositoryConfig(ctx, targetRepo.GetOwner().GetLogin(), targetRepo.GetName())
	if err == repo.ErrNotFound {
		return ErrConfigNotFound
	}
	if err != nil {
		return err
	}
	if config == nil {
		return nil
	}

	srv, err := service.New()
	if err != nil {
		return err
	}
	if config.MergeAvailable {
		return srv.ApprovePullRequest(ctx, client, pr)
	}

	return srv.PendingPullRequest(ctx, client, pr)
}

func updateCommitStatuses(ctx context.Context, installClient *github.Client, install *github.Installation, cfg *model.RepositoryConfig, srv service.Service, approve bool) error {
	prs, _, err := installClient.PullRequests.List(ctx, cfg.Owner, cfg.Name, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch pull requests on %s/%s: %w", cfg.Owner, cfg.Name, err)
	}
	for _, pr := range prs {
		if approve {
			if err := srv.ApprovePullRequest(ctx, installClient, pr); err != nil {
				return err
			}
		} else {
			if err := srv.PendingPullRequest(ctx, installClient, pr); err != nil {
				return err
			}
		}
	}
	return nil
}

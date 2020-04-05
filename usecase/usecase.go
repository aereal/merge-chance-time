package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/aereal/merge-chance-time/app/adapter/githubapps"
	"github.com/aereal/merge-chance-time/domain/model"
	"github.com/aereal/merge-chance-time/domain/repo"
	"github.com/aereal/merge-chance-time/logging"
	"github.com/google/go-github/v30/github"
	"golang.org/x/sync/errgroup"
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

func (u *Usecase) PutRepositoryConfig(ctx context.Context, ghAppClient *github.Client, owner, name string, input io.Reader) error {
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

func (u *Usecase) UpdateChanceTime(ctx context.Context, adapter *githubapps.GitHubAppsAdapter, baseTime time.Time) error {
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

			if config.ShouldStartOn(baseTime) {
				config.MergeAvailable = true
				toBeUpdated = append(toBeUpdated, config)

				g.Go(func() error {
					return updateCommitStatuses(c, installClient, install, config, true)
				})
			}
			if config.ShouldStopOn(baseTime) {
				config.MergeAvailable = false
				toBeUpdated = append(toBeUpdated, config)

				g.Go(func() error {
					return updateCommitStatuses(c, installClient, install, config, false)
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

func updateCommitStatuses(ctx context.Context, installClient *github.Client, install *github.Installation, cfg *model.RepositoryConfig, approve bool) error {
	logger := logging.GetLogger(ctx)

	prs, _, err := installClient.PullRequests.List(ctx, cfg.Owner, cfg.Name, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch pull requests on %s/%s: %w", cfg.Owner, cfg.Name, err)
	}
	state := "success"
	ctxName := "merge-chance-time"
	desc := fmt.Sprintf("%s is open", ctxName)
	if !approve {
		state = "failure"
		desc = fmt.Sprintf("%s is closed", ctxName)
	}
	for _, pr := range prs {
		_, _, err := installClient.Repositories.CreateStatus(ctx, cfg.Owner, cfg.Name, pr.GetHead().GetSHA(), &github.RepoStatus{
			State:       &state,
			Context:     &ctxName,
			Description: &desc,
		})
		logger.Infof("create status on %s/%s#%d: error=%+v", cfg.Owner, cfg.Name, pr.GetNumber(), err)
		if err != nil {
			return fmt.Errorf("failed to create status on %s/%s#%0d: %w", cfg.Owner, cfg.Name, pr.GetNumber(), err)
		}
	}
	return nil
}

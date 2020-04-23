package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/aereal/merge-chance-time/app/graph/dto"
	"github.com/aereal/merge-chance-time/app/graph/generated"
	"github.com/aereal/merge-chance-time/domain/repo"
)

func (r *installationResolver) InstalledRepositories(ctx context.Context, obj *dto.Installation) ([]*dto.Repository, error) {
	claims, err := r.authorizer.GetCurrentClaims(ctx)
	if err != nil {
		return nil, err
	}
	client := r.ghAdapter.NewUserClient(ctx, claims.AccessToken)

	rs, _, err := client.Apps.ListUserRepos(ctx, obj.ID, nil)
	if err != nil {
		return nil, err
	}
	repos := make([]*dto.Repository, len(rs))
	for i, r := range rs {
		repos[i] = dto.NewRepositoryFromResponse(r)
	}
	return repos, nil
}

func (r *queryResolver) Visitor(ctx context.Context) (*dto.Visitor, error) {
	_, err := r.authorizer.GetCurrentClaims(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.Visitor{}, nil
}

func (r *queryResolver) Repository(ctx context.Context, owner string, name string) (*dto.Repository, error) {
	claims, err := r.authorizer.GetCurrentClaims(ctx)
	if err != nil {
		return nil, err
	}
	client := r.ghAdapter.NewUserClient(ctx, claims.AccessToken)
	ghRepo, _, err := client.Repositories.Get(ctx, owner, name)
	if err != nil {
		return nil, err
	}
	return dto.NewRepositoryFromResponse(ghRepo), nil
}

func (r *repositoryResolver) Config(ctx context.Context, obj *dto.Repository) (*dto.RepositoryConfig, error) {
	cfg, err := r.repo.GetRepositoryConfig(ctx, obj.Owner.GetLogin(), obj.Name)
	if err == repo.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &dto.RepositoryConfig{
		StartSchedule:  cfg.StartSchedule.String(),
		StopSchedule:   cfg.StopSchedule.String(),
		MergeAvailable: cfg.MergeAvailable,
	}, nil
}

func (r *visitorResolver) Login(ctx context.Context, obj *dto.Visitor) (string, error) {
	claims, err := r.authorizer.GetCurrentClaims(ctx)
	if err != nil {
		return "", err
	}
	client := r.ghAdapter.NewUserClient(ctx, claims.AccessToken)
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return "", err
	}
	return user.GetLogin(), nil
}

func (r *visitorResolver) Installations(ctx context.Context, obj *dto.Visitor) ([]*dto.Installation, error) {
	claims, err := r.authorizer.GetCurrentClaims(ctx)
	if err != nil {
		return nil, err
	}
	client := r.ghAdapter.NewUserClient(ctx, claims.AccessToken)
	installations, _, err := client.Apps.ListUserInstallations(ctx, nil)
	if err != nil {
		return nil, err
	}
	dtos := make([]*dto.Installation, len(installations))
	for i, inst := range installations {
		dtos[i] = &dto.Installation{
			ID: inst.GetID(),
		}
	}
	return dtos, nil
}

// Installation returns generated.InstallationResolver implementation.
func (r *Resolver) Installation() generated.InstallationResolver { return &installationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Repository returns generated.RepositoryResolver implementation.
func (r *Resolver) Repository() generated.RepositoryResolver { return &repositoryResolver{r} }

// Visitor returns generated.VisitorResolver implementation.
func (r *Resolver) Visitor() generated.VisitorResolver { return &visitorResolver{r} }

type installationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type repositoryResolver struct{ *Resolver }
type visitorResolver struct{ *Resolver }

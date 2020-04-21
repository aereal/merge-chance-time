package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/aereal/merge-chance-time/app/graph/dto"
	"github.com/aereal/merge-chance-time/app/graph/generated"
)

func (r *queryResolver) Visitor(ctx context.Context) (*dto.User, error) {
	_, err := r.authorizer.GetCurrentClaims(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.User{}, nil
}

func (r *userResolver) Login(ctx context.Context, obj *dto.User) (string, error) {
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

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type queryResolver struct{ *Resolver }
type userResolver struct{ *Resolver }

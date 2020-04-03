package repo

import (
	"context"
	"crypto/sha256"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/aereal/merge-chance-time/domain/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrNotFound = fmt.Errorf("not found")
)

func New(firestoreClient *firestore.Client) (*Repository, error) {
	if firestoreClient == nil {
		return nil, fmt.Errorf("firestoreClient is nil")
	}
	return &Repository{
		firestoreClient: firestoreClient,
	}, nil
}

type Repository struct {
	firestoreClient *firestore.Client
}

func (r *Repository) GetRepositoryConfig(ctx context.Context, owner, name string) (*model.RepositoryConfig, error) {
	key := keyOf(owner, name)
	snapshot, err := r.firestoreClient.Collection("RepositoryConfig").Doc(key).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RepositoryConfig: %w", err)
	}
	var cfg model.RepositoryConfig
	if err := snapshot.DataTo(&cfg); err != nil {
		return nil, fmt.Errorf("failed to convert fetched data to RepositoryConfig: %w", err)
	}
	return &cfg, nil
}

func keyOf(owner, name string) string {
	fullName := fmt.Sprintf("%s/%s", owner, name)
	sum := sha256.Sum256([]byte(fullName))
	return fmt.Sprintf("%x", sum)
}

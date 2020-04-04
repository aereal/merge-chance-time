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

func (r *Repository) CreateRepositoryConfig(ctx context.Context, owner, name string, config *model.RepositoryConfig) error {
	key := keyOf(owner, name)
	dto := &dtoRepositoryConfig{
		Owner:         config.Owner,
		Name:          config.Name,
		StartSchedule: config.StartSchedule.String(),
		StopSchedule:  config.StopSchedule.String(),
	}
	_, err := r.repositoryConfigs().Doc(key).Set(ctx, dto)
	return err
}

func (r *Repository) GetRepositoryConfig(ctx context.Context, owner, name string) (*model.RepositoryConfig, error) {
	key := keyOf(owner, name)
	snapshot, err := r.repositoryConfigs().Doc(key).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RepositoryConfig: %w", err)
	}
	var dto dtoRepositoryConfig
	if err := snapshot.DataTo(&dto); err != nil {
		return nil, fmt.Errorf("failed to convert fetched data to RepositoryConfig: %w", err)
	}
	return dto.ToModel()
}

func (r *Repository) repositoryConfigs() *firestore.CollectionRef {
	return r.firestoreClient.Collection("RepositoryConfig")
}

func keyOf(owner, name string) string {
	fullName := fmt.Sprintf("%s/%s", owner, name)
	sum := sha256.Sum256([]byte(fullName))
	return fmt.Sprintf("%x", sum)
}

type dtoRepositoryConfig struct {
	Owner         string
	Name          string
	StartSchedule string
	StopSchedule  string
}

func (d *dtoRepositoryConfig) ToModel() (*model.RepositoryConfig, error) {
	m, err := model.NewRepositoryConfig([]byte(d.StartSchedule), []byte(d.StopSchedule))
	if err != nil {
		return nil, err
	}
	m.Name = d.Name
	m.Owner = d.Owner
	return m, nil
}

package repo

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/aereal/merge-chance-time/domain/model"
	"google.golang.org/api/iterator"
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

func (r *Repository) DeleteRepositoryConfig(ctx context.Context, owner, name string) error {
	ref := r.firestoreClient.Collection("InstallationTarget").Doc(owner).Collection("Repository").Doc(name)
	_, err := ref.Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete repo %s/%s: %w", owner, name, err)
	}
	return nil
}

func (r *Repository) PutRepositoryConfigs(ctx context.Context, configs []*model.RepositoryConfig) error {
	dtos := []*dtoRepositoryConfig{}
	for _, config := range configs {
		dto := &dtoRepositoryConfig{
			Owner:          config.Owner,
			Name:           config.Name,
			StartSchedule:  config.StartSchedule.String(),
			StopSchedule:   config.StopSchedule.String(),
			MergeAvailable: config.MergeAvailable,
		}
		dtos = append(dtos, dto)
	}
	batch := r.firestoreClient.Batch()
	for _, dto := range dtos {
		ownerRef := r.firestoreClient.Collection("InstallationTarget").Doc(dto.Owner)
		repoRef := ownerRef.Collection("Repository").Doc(dto.Name)
		batch.Set(ownerRef, map[string]interface{}{})
		batch.Set(repoRef, dto)
	}
	_, err := batch.Commit(ctx)
	return err
}

func (r *Repository) GetRepositoryConfig(ctx context.Context, owner, name string) (*model.RepositoryConfig, error) {
	snapshot, err := r.firestoreClient.Collection("InstallationTarget").Doc(owner).Collection("Repository").Doc(name).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RepositoryConfig: %w", err)
	}
	return repoFrom(snapshot)
}

func (r *Repository) ListConfigsByOwners(ctx context.Context) (map[string][]*model.RepositoryConfig, error) {
	ownerIter := r.firestoreClient.Collection("InstallationTarget").Documents(ctx)
	configs := map[string][]*model.RepositoryConfig{}
	for {
		ownerSnapshot, err := ownerIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		repoIter := ownerSnapshot.Ref.Collection("Repository").Documents(ctx)
		cfgs, err := fetchRepoConfigs(ctx, repoIter)
		if err != nil {
			return nil, err
		}
		ownerName := ownerSnapshot.Ref.ID
		configs[ownerName] = cfgs
	}
	return configs, nil
}

func (r *Repository) ListRepositoryConfigs(ctx context.Context) ([]*model.RepositoryConfig, error) {
	ownerIter := r.firestoreClient.Collection("InstallationTarget").Documents(ctx)
	configs := []*model.RepositoryConfig{}
	for {
		snapshot, err := ownerIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		repoIter := snapshot.Ref.Collection("Repository").Documents(ctx)
		cfgs, err := fetchRepoConfigs(ctx, repoIter)
		if err != nil {
			return nil, err
		}
		configs = append(configs, cfgs...)
	}
	return configs, nil
}

func fetchRepoConfigs(ctx context.Context, iter *firestore.DocumentIterator) ([]*model.RepositoryConfig, error) {
	configs := []*model.RepositoryConfig{}
	for {
		snapshot, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		cfg, err := repoFrom(snapshot)
		if err != nil {
			return nil, err
		}
		configs = append(configs, cfg)
	}
	return configs, nil
}

func repoFrom(snapshot *firestore.DocumentSnapshot) (*model.RepositoryConfig, error) {
	var dto dtoRepositoryConfig
	if err := snapshot.DataTo(&dto); err != nil {
		return nil, err
	}
	m, err := dto.ToModel()
	if err != nil {
		return nil, fmt.Errorf("failed to convert DTO to model: %w", err)
	}
	return m, nil
}

type dtoRepositoryConfig struct {
	Owner          string
	Name           string
	StartSchedule  string
	StopSchedule   string
	MergeAvailable bool
}

func (d *dtoRepositoryConfig) ToModel() (*model.RepositoryConfig, error) {
	m, err := model.NewRepositoryConfig([]byte(d.StartSchedule), []byte(d.StopSchedule))
	if err != nil {
		return nil, err
	}
	m.Name = d.Name
	m.Owner = d.Owner
	m.MergeAvailable = d.MergeAvailable
	return m, nil
}

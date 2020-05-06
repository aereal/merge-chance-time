//go:generate mockgen -package repo -destination repo_mock.go . Repository

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

func New(firestoreClient *firestore.Client) (Repository, error) {
	if firestoreClient == nil {
		return nil, fmt.Errorf("firestoreClient is nil")
	}
	return &repoImpl{
		firestoreClient: firestoreClient,
	}, nil
}

type Repository interface {
	DeleteRepositoryConfig(ctx context.Context, owner, name string) error
	DeleteRepositoryConfigsByOwner(ctx context.Context, owner string) error
	PutRepositoryConfigs(ctx context.Context, configs []*model.RepositoryConfig) error
	GetRepositoryConfig(ctx context.Context, owner, name string) (*model.RepositoryConfig, error)
	ListConfigsByOwners(ctx context.Context) (map[string][]*model.RepositoryConfig, error)
}

type repoImpl struct {
	firestoreClient *firestore.Client
}

func (r *repoImpl) DeleteRepositoryConfig(ctx context.Context, owner, name string) error {
	ref := r.firestoreClient.Collection("InstallationTarget").Doc(owner).Collection("Repository").Doc(name)
	_, err := ref.Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete repo %s/%s: %w", owner, name, err)
	}
	return nil
}

func (r *repoImpl) DeleteRepositoryConfigsByOwner(ctx context.Context, owner string) error {
	_, err := r.firestoreClient.Collection("InstallationTarget").Doc(owner).Delete(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *repoImpl) PutRepositoryConfigs(ctx context.Context, configs []*model.RepositoryConfig) error {
	dtos := []*dtoRepositoryConfig{}
	for _, config := range configs {
		dto := &dtoRepositoryConfig{
			Owner:          config.Owner,
			Name:           config.Name,
			MergeAvailable: config.MergeAvailable,
			Schedules:      newDTOMergeChanceSchedulesFromModel(config.Schedules),
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

func (r *repoImpl) GetRepositoryConfig(ctx context.Context, owner, name string) (*model.RepositoryConfig, error) {
	snapshot, err := r.firestoreClient.Collection("InstallationTarget").Doc(owner).Collection("Repository").Doc(name).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RepositoryConfig: %w", err)
	}
	return repoFrom(snapshot)
}

func (r *repoImpl) ListConfigsByOwners(ctx context.Context) (map[string][]*model.RepositoryConfig, error) {
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
	Schedules      *dtoMergeChanceSchedules
	MergeAvailable bool
}

func (d *dtoRepositoryConfig) ToModel() (*model.RepositoryConfig, error) {
	m := &model.RepositoryConfig{Schedules: &model.MergeChanceSchedules{}}
	s, err := d.Schedules.toModel()
	if err != nil {
		return nil, err
	}
	m.Schedules = s
	m.Name = d.Name
	m.Owner = d.Owner
	m.MergeAvailable = d.MergeAvailable
	return m, nil
}

func newDTOMergeChanceSchedulesFromModel(s *model.MergeChanceSchedules) *dtoMergeChanceSchedules {
	dto := &dtoMergeChanceSchedules{}
	if s.Sunday != nil {
		dto.Sunday = &dtoMergeChanceSchedule{
			StartHour: s.Sunday.StartHour,
			StopHour:  s.Sunday.StopHour,
		}
	}
	if s.Monday != nil {
		dto.Monday = &dtoMergeChanceSchedule{
			StartHour: s.Monday.StartHour,
			StopHour:  s.Monday.StopHour,
		}
	}
	if s.Tuesday != nil {
		dto.Tuesday = &dtoMergeChanceSchedule{
			StartHour: s.Tuesday.StartHour,
			StopHour:  s.Tuesday.StopHour,
		}
	}
	if s.Wednesday != nil {
		dto.Wednesday = &dtoMergeChanceSchedule{
			StartHour: s.Wednesday.StartHour,
			StopHour:  s.Wednesday.StopHour,
		}
	}
	if s.Thursday != nil {
		dto.Thursday = &dtoMergeChanceSchedule{
			StartHour: s.Thursday.StartHour,
			StopHour:  s.Thursday.StopHour,
		}
	}
	if s.Saturday != nil {
		dto.Saturday = &dtoMergeChanceSchedule{
			StartHour: s.Saturday.StartHour,
			StopHour:  s.Saturday.StopHour,
		}
	}
	if s.Friday != nil {
		dto.Friday = &dtoMergeChanceSchedule{
			StartHour: s.Friday.StartHour,
			StopHour:  s.Friday.StopHour,
		}
	}
	return dto
}

type dtoMergeChanceSchedule struct {
	StartHour int
	StopHour  int
}

type dtoMergeChanceSchedules struct {
	Sunday    *dtoMergeChanceSchedule
	Monday    *dtoMergeChanceSchedule
	Tuesday   *dtoMergeChanceSchedule
	Wednesday *dtoMergeChanceSchedule
	Thursday  *dtoMergeChanceSchedule
	Friday    *dtoMergeChanceSchedule
	Saturday  *dtoMergeChanceSchedule
}

func (dto *dtoMergeChanceSchedules) toModel() (*model.MergeChanceSchedules, error) {
	if dto == nil {
		return &model.MergeChanceSchedules{}, nil
	}
	s := &model.MergeChanceSchedules{}
	if dto.Sunday != nil {
		s.Sunday = &model.MergeChanceSchedule{
			StartHour: dto.Sunday.StartHour,
			StopHour:  dto.Sunday.StopHour,
		}
	}
	if dto.Monday != nil {
		s.Monday = &model.MergeChanceSchedule{
			StartHour: dto.Monday.StartHour,
			StopHour:  dto.Monday.StopHour,
		}
	}
	if dto.Tuesday != nil {
		s.Tuesday = &model.MergeChanceSchedule{
			StartHour: dto.Tuesday.StartHour,
			StopHour:  dto.Tuesday.StopHour,
		}
	}
	if dto.Wednesday != nil {
		s.Wednesday = &model.MergeChanceSchedule{
			StartHour: dto.Wednesday.StartHour,
			StopHour:  dto.Wednesday.StopHour,
		}
	}
	if dto.Thursday != nil {
		s.Thursday = &model.MergeChanceSchedule{
			StartHour: dto.Thursday.StartHour,
			StopHour:  dto.Thursday.StopHour,
		}
	}
	if dto.Friday != nil {
		s.Friday = &model.MergeChanceSchedule{
			StartHour: dto.Friday.StartHour,
			StopHour:  dto.Friday.StopHour,
		}
	}
	if dto.Saturday != nil {
		s.Saturday = &model.MergeChanceSchedule{
			StartHour: dto.Saturday.StartHour,
			StopHour:  dto.Saturday.StopHour,
		}
	}
	return s, nil
}

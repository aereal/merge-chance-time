//go:generate mockgen -package usecase -destination usecase_mock.go . Usecase

package usecase

import (
	"context"
	"testing"

	"github.com/aereal/merge-chance-time/domain/model"
	"github.com/aereal/merge-chance-time/domain/repo"
	"github.com/aereal/merge-chance-time/logging"
	"github.com/golang/mock/gomock"
	"github.com/google/go-github/v30/github"
)

func Test_usecaseImpl_onInstallRepository(t *testing.T) {
	type fields struct {
		repo func(ctrl *gomock.Controller) repo.Repository
	}
	type args struct {
		installedRepo *github.Repository
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "OK",
			fields: fields{
				repo: func(ctrl *gomock.Controller) repo.Repository {
					r := repo.NewMockRepository(ctrl)
					r.EXPECT().
						PutRepositoryConfigs(gomock.Any(), gomock.Eq([]*model.RepositoryConfig{
							{
								Owner:          "aereal",
								Name:           "example-repo",
								MergeAvailable: true,
								Schedules: &model.MergeChanceSchedules{
									Sunday:    nil,
									Monday:    model.WholeDay,
									Tuesday:   model.WholeDay,
									Wednesday: model.WholeDay,
									Thursday:  model.WholeDay,
									Friday:    model.WholeDay,
									Saturday:  nil,
								},
							},
						})).
						Return(nil).
						Times(1)
					return r
				},
			},
			args: args{
				installedRepo: &github.Repository{
					Name:     github.String("example-repo"),
					FullName: github.String("aereal/example-repo"),
					Owner: &github.User{
						Login: github.String("aereal"),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			u := &usecaseImpl{
				repo: tt.fields.repo(ctrl),
			}
			ctx := logging.SetNilLogger(context.Background())
			if err := u.onInstallRepository(ctx, tt.args.installedRepo); (err != nil) != tt.wantErr {
				t.Errorf("usecaseImpl.onInstallRepository() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

package dto

import (
	"github.com/aereal/merge-chance-time/domain/model"
	"github.com/google/go-github/v30/github"
)

type Visitor struct{}

type RepositoryOwner interface {
	IsRepositoryOwner()
	GetLogin() string
}

var (
	_ RepositoryOwner = &User{}
	_ RepositoryOwner = &Organization{}
)

type User struct {
	Login string `json:"login"`
}

func (u *User) GetLogin() string { return u.Login }

func (User) IsRepositoryOwner() {}

type Organization struct {
	Login string `json:"login"`
}

func (o *Organization) GetLogin() string { return o.Login }

func (Organization) IsRepositoryOwner() {}

type Installation struct {
	ID int64 `json:"id"`
}

type Repository struct {
	ID       int64           `json:"id"`
	Name     string          `json:"name"`
	FullName string          `json:"fullName"`
	Owner    RepositoryOwner `json:"owner"`
}

func NewRepositoryFromResponse(r *github.Repository) *Repository {
	var owner RepositoryOwner
	switch r.Owner.GetType() {
	case "User":
		owner = &User{
			Login: r.Owner.GetLogin(),
		}
	case "Organization":
		owner = &Organization{
			Login: r.Owner.GetLogin(),
		}
	}
	return &Repository{
		ID:       r.GetID(),
		Name:     r.GetName(),
		FullName: r.GetFullName(),
		Owner:    owner,
	}
}

func NewMergeChanceSchedules(m *model.MergeChanceSchedules) *MergeChanceSchedules {
	d := &MergeChanceSchedules{}
	if m.Sunday != nil {
		d.Sunday = &MergeChanceSchedule{
			StartHour: m.Sunday.StartHour,
			StopHour:  m.Sunday.StopHour,
		}
	}
	if m.Monday != nil {
		d.Monday = &MergeChanceSchedule{
			StartHour: m.Monday.StartHour,
			StopHour:  m.Monday.StopHour,
		}
	}
	if m.Tuesday != nil {
		d.Tuesday = &MergeChanceSchedule{
			StartHour: m.Tuesday.StartHour,
			StopHour:  m.Tuesday.StopHour,
		}
	}
	if m.Wednesday != nil {
		d.Wednesday = &MergeChanceSchedule{
			StartHour: m.Wednesday.StartHour,
			StopHour:  m.Wednesday.StopHour,
		}
	}
	if m.Thursday != nil {
		d.Thursday = &MergeChanceSchedule{
			StartHour: m.Thursday.StartHour,
			StopHour:  m.Thursday.StopHour,
		}
	}
	if m.Friday != nil {
		d.Friday = &MergeChanceSchedule{
			StartHour: m.Friday.StartHour,
			StopHour:  m.Friday.StopHour,
		}
	}
	if m.Saturday != nil {
		d.Saturday = &MergeChanceSchedule{
			StartHour: m.Saturday.StartHour,
			StopHour:  m.Saturday.StopHour,
		}
	}
	return d
}

func (d *MergeChanceSchedulesToUpdate) ToModel() *model.MergeChanceSchedules {
	m := &model.MergeChanceSchedules{}
	if d.Sunday != nil {
		m.Sunday = &model.MergeChanceSchedule{
			StartHour: d.Sunday.StartHour,
			StopHour:  d.Sunday.StopHour,
		}
	}
	if d.Monday != nil {
		m.Monday = &model.MergeChanceSchedule{
			StartHour: d.Monday.StartHour,
			StopHour:  d.Monday.StopHour,
		}
	}
	if d.Tuesday != nil {
		m.Tuesday = &model.MergeChanceSchedule{
			StartHour: d.Tuesday.StartHour,
			StopHour:  d.Tuesday.StopHour,
		}
	}
	if d.Wednesday != nil {
		m.Wednesday = &model.MergeChanceSchedule{
			StartHour: d.Wednesday.StartHour,
			StopHour:  d.Wednesday.StopHour,
		}
	}
	if d.Thursday != nil {
		m.Thursday = &model.MergeChanceSchedule{
			StartHour: d.Thursday.StartHour,
			StopHour:  d.Thursday.StopHour,
		}
	}
	if d.Friday != nil {
		m.Friday = &model.MergeChanceSchedule{
			StartHour: d.Friday.StartHour,
			StopHour:  d.Friday.StopHour,
		}
	}
	if d.Saturday != nil {
		m.Saturday = &model.MergeChanceSchedule{
			StartHour: d.Saturday.StartHour,
			StopHour:  d.Saturday.StopHour,
		}
	}
	return m
}

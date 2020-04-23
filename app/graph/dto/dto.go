package dto

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

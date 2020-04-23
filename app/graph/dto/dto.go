package dto

type Visitor struct{}

type User struct {
	Login string `json:"login"`
}

func (u *User) IsRepositoryOwner() {}

type Installation struct {
	ID int64 `json:"id"`
}

type Repository struct {
	ID       int64           `json:"id"`
	Name     string          `json:"name"`
	FullName string          `json:"fullName"`
	Owner    RepositoryOwner `json:"owner"`
}

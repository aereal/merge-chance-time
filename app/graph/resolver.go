//go:generate go run github.com/99designs/gqlgen

package graph

func New() (*Resolver, error) {
	return &Resolver{}, nil
}

type Resolver struct{}

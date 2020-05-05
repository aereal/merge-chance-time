package api

import (
	"fmt"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/aereal/gqlgen-tracer-opencensus/tracer"
	"github.com/aereal/merge-chance-time/app/authz"
	"github.com/dimfeld/httptreemux/v5"
)

func New(authorizer authz.Authorizer, es graphql.ExecutableSchema) (*Web, error) {
	if es == nil {
		return nil, fmt.Errorf("graphql.ExecutableSchema is nil")
	}
	return &Web{
		authorizer: authorizer,
		es:         es,
	}, nil
}

type Web struct {
	authorizer authz.Authorizer
	es         graphql.ExecutableSchema
}

func (s *Web) Routes() func(router *httptreemux.TreeMux) {
	srv := s.newHandler()
	return func(router *httptreemux.TreeMux) {
		group := router.UsingContext().NewContextGroup("/api")
		group.UseHandler(s.authorizer.Middleware())
		group.Handler(http.MethodOptions, "/query", srv)
		group.Handler(http.MethodPost, "/query", srv)
	}
}

func (s *Web) newHandler() *handler.Server {
	srv := handler.New(s.es)
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.POST{})
	srv.Use(extension.Introspection{})
	srv.Use(tracer.Tracer{})
	return srv
}

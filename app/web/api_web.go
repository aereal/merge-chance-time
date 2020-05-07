package web

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/aereal/gqlgen-tracer-opencensus/tracer"
)

func (s *Web) newHandler() *handler.Server {
	srv := handler.New(s.es)
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.POST{})
	srv.Use(extension.Introspection{})
	srv.Use(tracer.Tracer{})
	return srv
}

package main

import (
	"context"
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/firestore"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/aereal/merge-chance-time/admin/web/api"
	"github.com/aereal/merge-chance-time/admin/web/auth"
	"github.com/aereal/merge-chance-time/app/adapter/githubapps"
	"github.com/aereal/merge-chance-time/app/authz"
	"github.com/aereal/merge-chance-time/app/config"
	"github.com/aereal/merge-chance-time/app/graph"
	"github.com/aereal/merge-chance-time/app/graph/generated"
	"github.com/aereal/merge-chance-time/app/web"
	"github.com/aereal/merge-chance-time/app/web/ghapps"
	"github.com/aereal/merge-chance-time/authflow"
	"github.com/aereal/merge-chance-time/domain/repo"
	"github.com/aereal/merge-chance-time/jwtissuer"
	"github.com/aereal/merge-chance-time/usecase"
	"github.com/dgrijalva/jwt-go"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

var (
	onGAE bool
)

func main() {
	if err := run(); err != nil {
		log.Printf("error: %#v", err)
		os.Exit(1)
	}
}

func init() {
	onGAE = os.Getenv("GAE_ENV") != ""
}

func run() error {
	cfg, err := config.NewFromEnvironment()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	httpClient := http.DefaultClient
	if httpClient.Transport == nil {
		httpClient.Transport = http.DefaultTransport
	}
	if onGAE {
		exporter, err := stackdriver.NewExporter(stackdriver.Options{Context: ctx})
		if err != nil {
			return err
		}
		defer exporter.Flush()
		trace.RegisterExporter(exporter)

		if err := view.Register(ochttp.ClientSentBytesDistribution, ochttp.ClientReceivedBytesDistribution, ochttp.ClientLatencyView, ochttp.ClientCompletedCount, ochttp.ClientRoundtripLatencyDistribution); err != nil {
			return err
		}
		httpClient.Transport = &ochttp.Transport{}
	}

	githubAppPrivateKey, err := parseRSAPrivateKeyFile("./github-app.private-key.pem")
	if err != nil {
		return err
	}

	tokenPrivateKey, err := parseRSAPrivateKeyFile("./keys/private.pem")
	if err != nil {
		return err
	}

	issuer, err := jwtissuer.NewIssuer(tokenPrivateKey)
	if err != nil {
		return err
	}

	ghAdapter := githubapps.New(cfg.GitHubAppConfig.ID, githubAppPrivateKey, httpClient)

	authorizer, err := authz.New(issuer)
	if err != nil {
		return err
	}

	ghAuthFlow, err := authflow.NewGitHubAuthFlow(cfg.GitHubAppConfig, issuer, httpClient, authorizer)
	if err != nil {
		return err
	}

	fsClient, err := firestore.NewClient(ctx, cfg.GCPProjectID)
	if err != nil {
		return err
	}

	r, err := repo.New(fsClient)
	if err != nil {
		return err
	}

	uc, err := usecase.New(r)
	if err != nil {
		return err
	}

	ga, err := ghapps.New(cfg.GitHubAppConfig, ghAdapter, uc)
	if err != nil {
		return err
	}

	resolver, err := graph.New(authorizer, ghAdapter)
	if err != nil {
		return err
	}
	es := generated.NewExecutableSchema(generated.Config{Resolvers: resolver})

	a, err := api.New(ghAdapter, authorizer, es)
	if err != nil {
		return err
	}

	authWeb, err := auth.New(ghAuthFlow)
	if err != nil {
		return err
	}

	w := web.New(onGAE, cfg, ga.Routes(), a.Routes(), authWeb.Routes())
	server := w.Server(cfg.ListenPort)
	go graceful(ctx, server, 5*time.Second)

	log.Printf("starting server; accepting request on %s", server.Addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("cannot start server: %w", err)
	}

	return nil
}

func parseRSAPrivateKeyFile(fileName string) (*rsa.PrivateKey, error) {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %s: %w", fileName, err)
	}
	return jwt.ParseRSAPrivateKeyFromPEM(content)
}

func graceful(parent context.Context, server *http.Server, timeout time.Duration) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	sig := <-sigChan
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	log.Printf("shutting down server signal=%q", sig)
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("failed to shutdown: %s", err)
	}
}

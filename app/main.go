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
	"strconv"
	"syscall"
	"time"

	"cloud.google.com/go/firestore"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/aereal/merge-chance-time/app/adapter/githubapps"
	"github.com/aereal/merge-chance-time/app/web"
	"github.com/aereal/merge-chance-time/domain/repo"
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
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
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

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		return fmt.Errorf("GOOGLE_CLOUD_PROJECT must be defined")
	}

	githubAppPrivateKey, err := parseRSAPrivateKeyFile("./github-app.private-key.pem")
	if err != nil {
		return err
	}

	githubAppID, err := strconv.Atoi(os.Getenv("GITHUB_APP_IDENTIFIER"))
	if err != nil {
		return fmt.Errorf("GITHUB_APP_IDENTIFIER must be valid int")
	}

	githubWebhookSecret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	if githubWebhookSecret == "" {
		log.Printf("warning: GITHUB_WEBHOOK_SECRET is empty")
	}

	ghAdapter := githubapps.New(int64(githubAppID), githubAppPrivateKey, httpClient.Transport)

	fsClient, err := firestore.NewClient(ctx, projectID)
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

	w := web.New(onGAE, projectID, ghAdapter, []byte(githubWebhookSecret), r, uc)
	server := w.Server(port)
	go graceful(ctx, server, 5*time.Second)

	log.Printf("starting server; accepting request on %s", server.Addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("cannot start server: %w", err)
	}

	return nil
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

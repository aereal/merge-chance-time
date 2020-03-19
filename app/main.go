package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/aereal/merge-chance-time/app/web"
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

	w := web.New(onGAE, projectID)
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

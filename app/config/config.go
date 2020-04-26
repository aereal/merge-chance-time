package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
)

var (
	keyPort          = "PORT"
	keyGCPProjectID  = "GOOGLE_CLOUD_PROJECT"
	keyAppID         = "GH_APP_IDENTIFIER"
	keyWebhookSecret = "GH_WEBHOOK_SECRET"
	keyClientID      = "GH_APP_CLIENT_ID"
	keyClientSecret  = "GH_APP_CLIENT_SECRET"
	keyAdminOrigin   = "ADMIN_ORIGIN"
)

func NewFromEnvironment() (*Config, error) {
	cfg := &Config{GitHubAppConfig: &GitHubAppConfig{}}
	envs := getEnvs(keyPort, keyGCPProjectID, keyAppID, keyWebhookSecret, keyClientID, keyClientSecret, keyAdminOrigin)

	cfg.ListenPort = envs[keyPort]
	if cfg.ListenPort == "" {
		cfg.ListenPort = "8000"
	}

	cfg.GCPProjectID = envs[keyGCPProjectID]
	if cfg.GCPProjectID == "" {
		return nil, fmt.Errorf("GOOGLE_CLOUD_PROJECT must be defined")
	}

	rawURL := envs[keyAdminOrigin]
	if rawURL == "" {
		return nil, fmt.Errorf("%s must be defined", keyAdminOrigin)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	parsed.Path = ""
	parsed.RawPath = ""
	cfg.AdminOrigin = parsed

	cfg.GitHubAppConfig.WebhookSecret = []byte(envs[keyWebhookSecret])
	cfg.GitHubAppConfig.ClientID = envs[keyClientID]
	if cfg.GitHubAppConfig.ClientID == "" {
		return nil, fmt.Errorf("GH_APP_CLIENT_ID must be defined")
	}
	cfg.GitHubAppConfig.ClientSecret = envs[keyClientSecret]
	if cfg.GitHubAppConfig.ClientSecret == "" {
		return nil, fmt.Errorf("GH_APP_CLIENT_SECRET must be defined")
	}
	appID, err := strconv.Atoi(envs[keyAppID])
	if err != nil {
		return nil, fmt.Errorf("%s is invalid: %w", keyAppID, err)
	}
	cfg.GitHubAppConfig.ID = int64(appID)

	return cfg, nil
}

type Config struct {
	ListenPort      string
	GCPProjectID    string
	GitHubAppConfig *GitHubAppConfig
	AdminOrigin     *url.URL
}

type GitHubAppConfig struct {
	ID            int64
	WebhookSecret []byte
	ClientID      string
	ClientSecret  string
}

func getEnvs(names ...string) map[string]string {
	envs := map[string]string{}
	for _, name := range names {
		envs[name] = os.Getenv(name)
	}
	return envs
}

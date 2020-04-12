package config

import (
	"fmt"
	"os"
	"strconv"
)

var (
	keyPort          = "PORT"
	keyGCPProjectID  = "GOOGLE_CLOUD_PROJECT"
	keyAppID         = "GITHUB_APP_IDENTIFIER"
	keyWebhookSecret = "GITHUB_WEBHOOK_SECRET"
)

func NewFromEnvironment() (*Config, error) {
	cfg := &Config{GitHubAppConfig: &GitHubAppConfig{}}
	envs := getEnvs(keyPort, keyGCPProjectID, keyAppID, keyWebhookSecret)

	cfg.ListenPort = envs[keyPort]
	if cfg.ListenPort == "" {
		cfg.ListenPort = "8000"
	}

	cfg.GCPProjectID = envs[keyGCPProjectID]
	if cfg.GCPProjectID == "" {
		return nil, fmt.Errorf("GOOGLE_CLOUD_PROJECT must be defined")
	}

	cfg.GitHubAppConfig.WebhookSecret = []byte(envs[keyWebhookSecret])
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
}

type GitHubAppConfig struct {
	ID            int64
	WebhookSecret []byte
}

func getEnvs(names ...string) map[string]string {
	envs := map[string]string{}
	for _, name := range names {
		envs[name] = os.Getenv(name)
	}
	return envs
}

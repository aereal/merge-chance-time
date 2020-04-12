package githubapps

import (
	"context"
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v30/github"
)

func New(appID int64, clientID, clientSecret string, privKey *rsa.PrivateKey, httpClient *http.Client) *GitHubAppsAdapter {
	return &GitHubAppsAdapter{
		appID:        appID,
		clientID:     clientID,
		clientSecret: clientSecret,
		privKey:      privKey,
		httpClient:   httpClient,
	}
}

type GitHubAppsAdapter struct {
	appID        int64
	clientID     string
	clientSecret string
	privKey      *rsa.PrivateKey
	httpClient   *http.Client
}

func (a *GitHubAppsAdapter) appTransport() *ghinstallation.AppsTransport {
	return ghinstallation.NewAppsTransportFromPrivateKey(a.httpClient.Transport, a.appID, a.privKey)
}

func (a *GitHubAppsAdapter) NewAppClient() *github.Client {
	return github.NewClient(&http.Client{Transport: a.appTransport()})
}

func (a *GitHubAppsAdapter) NewInstallationClient(installID int64) *github.Client {
	return github.NewClient(&http.Client{Transport: ghinstallation.NewFromAppsTransport(a.appTransport(), installID)})
}

func (a *GitHubAppsAdapter) CreateUserAccessToken(ctx context.Context, code, state string) (string, error) {
	if code == "" {
		return "", fmt.Errorf("code is empty")
	}

	params := url.Values{}
	params.Set("client_id", a.clientID)
	params.Set("client_secret", a.clientSecret)
	params.Set("code", code)
	params.Set("state", state)

	authReq, err := http.NewRequest(http.MethodPost, "https://github.com/login/oauth/access_token", strings.NewReader(params.Encode()))
	if err != nil {
		return "", err
	}
	resp, err := a.httpClient.Do(authReq.WithContext(ctx))
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	payload, err := url.ParseQuery(string(body))
	if err != nil {
		return "", fmt.Errorf("response body is invalid: %w", err)
	}
	token := payload.Get("access_token")
	if token == "" {
		return "", fmt.Errorf("response body contains no access_token")
	}

	return token, nil
}

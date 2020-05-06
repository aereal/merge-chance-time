//go:generate mockgen -package githubapps -destination adapter_mock.go . GitHubAppsAdapter

package githubapps

import (
	"context"
	"crypto/rsa"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v30/github"
	"golang.org/x/oauth2"
)

func New(appID int64, privKey *rsa.PrivateKey, httpClient *http.Client) GitHubAppsAdapter {
	return &ghAdapterImpl{
		appID:      appID,
		privKey:    privKey,
		httpClient: httpClient,
	}
}

type GitHubAppsAdapter interface {
	NewAppClient() *github.Client
	NewInstallationClient(installID int64) *github.Client
	NewUserClient(ctx context.Context, accessToken string) *github.Client
}

type ghAdapterImpl struct {
	appID      int64
	privKey    *rsa.PrivateKey
	httpClient *http.Client
}

func (a *ghAdapterImpl) appTransport() *ghinstallation.AppsTransport {
	return ghinstallation.NewAppsTransportFromPrivateKey(a.httpClient.Transport, a.appID, a.privKey)
}

func (a *ghAdapterImpl) NewAppClient() *github.Client {
	return github.NewClient(&http.Client{Transport: a.appTransport()})
}

func (a *ghAdapterImpl) NewInstallationClient(installID int64) *github.Client {
	return github.NewClient(&http.Client{Transport: ghinstallation.NewFromAppsTransport(a.appTransport(), installID)})
}

func (a *ghAdapterImpl) NewUserClient(ctx context.Context, accessToken string) *github.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	client := oauth2.NewClient(context.WithValue(ctx, oauth2.HTTPClient, a.httpClient), ts)
	return github.NewClient(client)
}

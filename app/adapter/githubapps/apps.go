package githubapps

import (
	"crypto/rsa"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v30/github"
)

func New(appID int64, privKey *rsa.PrivateKey, rt http.RoundTripper) *GitHubAppsAdapter {
	return &GitHubAppsAdapter{
		appID:   appID,
		privKey: privKey,
		rt:      rt,
	}
}

type GitHubAppsAdapter struct {
	appID   int64
	privKey *rsa.PrivateKey
	rt      http.RoundTripper
}

func (a *GitHubAppsAdapter) appTransport() *ghinstallation.AppsTransport {
	return ghinstallation.NewAppsTransportFromPrivateKey(a.rt, a.appID, a.privKey)
}

func (a *GitHubAppsAdapter) NewAppClient() *github.Client {
	return github.NewClient(&http.Client{Transport: a.appTransport()})
}

func (a *GitHubAppsAdapter) NewInstallationClient(installID int64) *github.Client {
	return github.NewClient(&http.Client{Transport: ghinstallation.NewFromAppsTransport(a.appTransport(), installID)})
}

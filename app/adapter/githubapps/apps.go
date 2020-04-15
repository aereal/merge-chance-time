package githubapps

import (
	"context"
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/aereal/merge-chance-time/jwtissuer"
	"github.com/aereal/merge-chance-time/logging"
	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v30/github"
	"golang.org/x/oauth2"
)

func New(appID int64, clientID, clientSecret string, privKey *rsa.PrivateKey, httpClient *http.Client, issuer *jwtissuer.Issuer) *GitHubAppsAdapter {
	return &GitHubAppsAdapter{
		appID:        appID,
		clientID:     clientID,
		clientSecret: clientSecret,
		privKey:      privKey,
		httpClient:   httpClient,
		issuer:       issuer,
	}
}

type GitHubAppsAdapter struct {
	appID        int64
	clientID     string
	clientSecret string
	privKey      *rsa.PrivateKey
	httpClient   *http.Client
	issuer       *jwtissuer.Issuer
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

func (a *GitHubAppsAdapter) NewUserClient(ctx context.Context, accessToken string) *github.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	client := oauth2.NewClient(context.WithValue(ctx, oauth2.HTTPClient, a.httpClient), ts)
	return github.NewClient(client)
}

func (a *GitHubAppsAdapter) NewAuthorizeURL(ctx context.Context) string {
	params := url.Values{}
	params.Set("client_id", a.clientID)
	params.Set("redirect_uri", "http://localhost:8000/auth/callback") // TODO: built from request URL
	state, err := a.generateState()
	if err != nil {
		logger := logging.GetLogger(ctx)
		logger.Warnf("failed to generate authorize state (but skip): %+v", err)
	}
	if state != "" {
		params.Set("state", state)
	}
	base := "https://github.com/login/oauth/authorize"
	return fmt.Sprintf("%s?%s", base, params.Encode())
}

func (a *GitHubAppsAdapter) generateState() (string, error) {
	stdClaims := jwtissuer.NewStandardClaims()
	token, err := a.issuer.Signed(stdClaims)
	if err != nil {
		return "", err
	}
	return token, nil
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

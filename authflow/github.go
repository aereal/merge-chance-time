package authflow

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/aereal/merge-chance-time/app/authz"
	"github.com/aereal/merge-chance-time/app/config"
	"github.com/aereal/merge-chance-time/jwtissuer"
	"github.com/aereal/merge-chance-time/logging"
)

func NewGitHubAuthFlow(appConfig *config.GitHubAppConfig, issuer *jwtissuer.Issuer, httpClient *http.Client, authorizer *authz.Authorizer) (*GitHubAuthFlow, error) {
	if appConfig == nil {
		return nil, fmt.Errorf("appConfig is nil")
	}
	if issuer == nil {
		return nil, fmt.Errorf("issuer is nil")
	}
	if httpClient == nil {
		return nil, fmt.Errorf("httpClient is nil")
	}
	if authorizer == nil {
		return nil, fmt.Errorf("authroizer is nil")
	}
	return &GitHubAuthFlow{
		clientID:     appConfig.ClientID,
		clientSecret: appConfig.ClientSecret,
		issuer:       issuer,
		httpClient:   httpClient,
		authorizer:   authorizer,
	}, nil
}

type GitHubAuthFlow struct {
	clientID     string
	clientSecret string
	issuer       *jwtissuer.Issuer
	httpClient   *http.Client
	authorizer   *authz.Authorizer
}

func (f *GitHubAuthFlow) IssueEncryptedToken(ctx context.Context, code, state string) (string, error) {
	accessToken, err := f.createUserAccessToken(ctx, code, state)
	if err != nil {
		return "", err
	}
	crypted, err := f.authorizer.IssueAuthenticationToken(&authz.AppClaims{AccessToken: accessToken})
	if err != nil {
		return "", err
	}
	return crypted, nil
}

func (f *GitHubAuthFlow) createUserAccessToken(ctx context.Context, code, state string) (string, error) {
	if code == "" {
		return "", fmt.Errorf("code is empty")
	}

	params := url.Values{}
	params.Set("client_id", f.clientID)
	params.Set("client_secret", f.clientSecret)
	params.Set("code", code)
	params.Set("state", state)

	authReq, err := http.NewRequest(http.MethodPost, "https://github.com/login/oauth/access_token", strings.NewReader(params.Encode()))
	if err != nil {
		return "", err
	}
	resp, err := f.httpClient.Do(authReq.WithContext(ctx))
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

func (f *GitHubAuthFlow) NewAuthorizeURL(ctx context.Context) string {
	params := url.Values{}
	params.Set("client_id", f.clientID)
	params.Set("redirect_uri", "http://localhost:8000/auth/callback") // TODO: built from request URL
	state, err := f.generateState()
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

func (f *GitHubAuthFlow) generateState() (string, error) {
	stdClaims := jwtissuer.NewStandardClaims()
	token, err := f.issuer.Signed(stdClaims)
	if err != nil {
		return "", err
	}
	return token, nil
}

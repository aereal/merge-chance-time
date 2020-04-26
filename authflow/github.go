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
	"gopkg.in/square/go-jose.v2/jwt"
)

type State struct {
	InitiatorURL string `json:"initiator_url"`
}

type StateClaims struct {
	jwt.Claims
	State
}

func NewGitHubAuthFlow(cfg *config.Config, issuer *jwtissuer.Issuer, httpClient *http.Client, authorizer *authz.Authorizer) (*GitHubAuthFlow, error) {
	if cfg == nil {
		return nil, fmt.Errorf("appConfig is nil")
	}
	parsed, _ := url.Parse(cfg.AdminOrigin.String())
	parsed.Path = "/auth/callback"
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
		clientID:            cfg.GitHubAppConfig.ClientID,
		clientSecret:        cfg.GitHubAppConfig.ClientSecret,
		issuer:              issuer,
		httpClient:          httpClient,
		authorizer:          authorizer,
		defaultInitiatorURL: parsed,
	}, nil
}

type GitHubAuthFlow struct {
	clientID            string
	clientSecret        string
	issuer              *jwtissuer.Issuer
	httpClient          *http.Client
	authorizer          *authz.Authorizer
	defaultInitiatorURL *url.URL
}

func (f *GitHubAuthFlow) NavigateAuthCompletion(ctx context.Context, code, state string) (*url.URL, error) {
	accessToken, err := f.createUserAccessToken(ctx, code, state)
	if err != nil {
		return nil, fmt.Errorf("cannot create user access token: %w", err)
	}
	initiatorURL, err := f.determineInitiatorURL(state)
	crypted, err := f.authorizer.IssueAuthenticationToken(&authz.AppClaims{AccessToken: accessToken})
	if err != nil {
		return nil, fmt.Errorf("cannot issue token: %w", err)
	}

	params := initiatorURL.Query()
	params.Set("accessToken", crypted)
	initiatorURL.RawQuery = params.Encode()
	return initiatorURL, nil
}

func (f *GitHubAuthFlow) determineInitiatorURL(state string) (*url.URL, error) {
	if state == "" {
		return f.defaultInitiatorURL, nil
	}
	var claims StateClaims
	if err := f.issuer.ParseSigned(state, &claims); err != nil {
		return nil, fmt.Errorf("cannot parse token: %w", err)
	}
	initiatorURL, err := url.Parse(claims.InitiatorURL)
	if err != nil {
		return nil, fmt.Errorf("initiatorURL is invalid: %w", err)
	}
	return initiatorURL, nil
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

func (f *GitHubAuthFlow) NewAuthorizeURL(ctx context.Context, appOrigin string, initiatorURL string) (string, error) {
	params := url.Values{}
	params.Set("client_id", f.clientID)
	params.Set("redirect_uri", fmt.Sprintf("%s/auth/callback", appOrigin))
	state, err := f.generateState(initiatorURL)
	if err != nil {
		return "", fmt.Errorf("failed to generate authorize state: %w", err)
	}
	params.Set("state", state)
	base := "https://github.com/login/oauth/authorize"
	return fmt.Sprintf("%s?%s", base, params.Encode()), nil
}

func (f *GitHubAuthFlow) generateState(initiatorURL string) (string, error) {
	stdClaims := jwtissuer.NewStandardClaims()
	claims := StateClaims{
		stdClaims,
		State{initiatorURL},
	}
	token, err := f.issuer.Signed(claims)
	if err != nil {
		return "", err
	}
	return token, nil
}

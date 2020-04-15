package authflow

import (
	"context"
	"fmt"
	"net/url"

	"github.com/aereal/merge-chance-time/app/config"
	"github.com/aereal/merge-chance-time/jwtissuer"
	"github.com/aereal/merge-chance-time/logging"
)

func NewGitHubAuthFlow(appConfig *config.GitHubAppConfig, issuer *jwtissuer.Issuer) (*GitHubAuthFlow, error) {
	if appConfig == nil {
		return nil, fmt.Errorf("appConfig is nil")
	}
	if issuer == nil {
		return nil, fmt.Errorf("issuer is nil")
	}
	return &GitHubAuthFlow{
		clientID: appConfig.ClientID,
		issuer:   issuer,
	}, nil
}

type GitHubAuthFlow struct {
	clientID string
	issuer   *jwtissuer.Issuer
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

package authz

import (
	"fmt"

	"github.com/aereal/merge-chance-time/jwtissuer"
	"gopkg.in/square/go-jose.v2/jwt"
)

func New(issuer *jwtissuer.Issuer) (*Authorizer, error) {
	if issuer == nil {
		return nil, fmt.Errorf("issuer is nil")
	}
	return &Authorizer{issuer}, nil
}

type Authorizer struct {
	issuer *jwtissuer.Issuer
}

type AppClaims struct {
	AccessToken string
}

type Claims struct {
	jwt.Claims
	*AppClaims
}

var _ jwtissuer.ValidatableClaims = Claims{}

func (a *Authorizer) AuthenticateWithToken(token string) (*AppClaims, error) {
	var out Claims
	if err := a.issuer.ParseSignedAndEncrypted(token, &out); err != nil {
		return nil, err
	}

	return out.AppClaims, nil
}

func (a *Authorizer) IssueAuthenticationToken(appClaims *AppClaims) (string, error) {
	stdClaims := jwtissuer.NewStandardClaims()
	claims := Claims{
		stdClaims,
		appClaims,
	}
	token, err := a.issuer.SignedAndEncrypted(claims)
	if err != nil {
		return "", err
	}
	return token, nil
}

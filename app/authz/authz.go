package authz

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/aereal/merge-chance-time/jwtissuer"
	"gopkg.in/square/go-jose.v2/jwt"
)

func New(issuer jwtissuer.Issuer) (Authorizer, error) {
	if issuer == nil {
		return nil, fmt.Errorf("issuer is nil")
	}
	return &authorizerImpl{issuer}, nil
}

type Authorizer interface {
	GetCurrentClaims(ctx context.Context) (*AppClaims, error)
	Authenticate(ctx context.Context, token string) (context.Context, error)
	AuthenticateWithToken(token string) (*AppClaims, error)
	Middleware() func(next http.Handler) http.Handler
	IssueAuthenticationToken(appClaims *AppClaims) (string, error)
}

type authorizerImpl struct {
	issuer jwtissuer.Issuer
}

type AppClaims struct {
	AccessToken string
}

type Claims struct {
	jwt.Claims
	*AppClaims
}

var _ jwtissuer.ValidatableClaims = Claims{}

type keyType struct{}

var ctxKeyAppClaims = &keyType{}

func (a *authorizerImpl) Authenticate(ctx context.Context, token string) (context.Context, error) {
	claims, err := a.AuthenticateWithToken(token)
	if err != nil {
		return ctx, err
	}

	return context.WithValue(ctx, ctxKeyAppClaims, claims), nil
}

func (a *authorizerImpl) GetCurrentClaims(ctx context.Context) (*AppClaims, error) {
	if claims, ok := ctx.Value(ctxKeyAppClaims).(*AppClaims); ok {
		return claims, nil
	}
	return nil, fmt.Errorf("not authenticated")
}

func (a *authorizerImpl) Middleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "application/json")
			token := strings.Replace(r.Header.Get("authorization"), "Bearer ", "", 1)
			ctx, _ := a.Authenticate(r.Context(), token)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (a *authorizerImpl) AuthenticateWithToken(token string) (*AppClaims, error) {
	var out Claims
	if err := a.issuer.ParseSignedAndEncrypted(token, &out); err != nil {
		return nil, err
	}

	return out.AppClaims, nil
}

func (a *authorizerImpl) IssueAuthenticationToken(appClaims *AppClaims) (string, error) {
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

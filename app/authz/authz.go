package authz

import (
	"crypto/rsa"
	"fmt"
	"time"

	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

func New(privateKey *rsa.PrivateKey) (*Authorizer, error) {
	if privateKey == nil {
		return nil, fmt.Errorf("privateKey is nil")
	}
	return &Authorizer{privateKey}, nil
}

type Authorizer struct {
	privateKey *rsa.PrivateKey
}

type AppClaims struct {
	AccessToken string
}

type Claims struct {
	jwt.Claims
	*AppClaims
}

func (a *Authorizer) AuthenticateWithToken(token string) (*AppClaims, error) {
	t, err := jwt.ParseSignedAndEncrypted(token)
	if err != nil {
		return nil, fmt.Errorf("cannot parse token: %w", err)
	}
	nested, err := t.Decrypt(a.privateKey)
	if err != nil {
		return nil, fmt.Errorf("cannot decrypt: %w", err)
	}

	var out Claims
	if err := nested.Claims(a.privateKey.Public(), &out); err != nil {
		return nil, fmt.Errorf("cannot decode token: %w", err)
	}

	if err := out.ValidateWithLeeway(jwt.Expected{
		Audience: jwt.Audience{"mergechancetime.app"},
		Time:     time.Now(),
	}, 0); err != nil {
		return nil, fmt.Errorf("token is invalid: %w", err)
	}

	return out.AppClaims, nil
}

func (a *Authorizer) IssueAuthenticationToken(appClaims *AppClaims) (string, error) {
	enc, err := a.newEncrypter()
	if err != nil {
		return "", err
	}
	signer, err := a.newSigner()
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := Claims{
		jwt.Claims{
			Issuer:    "mergechancetime.app",
			Audience:  jwt.Audience{"mergechancetime.app"},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Expiry:    jwt.NewNumericDate(now.Add(time.Hour * 12)),
		},
		appClaims,
	}
	token, err := jwt.SignedAndEncrypted(signer, enc).Claims(claims).CompactSerialize()
	if err != nil {
		return "", err
	}
	return token, nil
}

func (a *Authorizer) newEncrypter() (jose.Encrypter, error) {
	recp := jose.Recipient{
		Algorithm: jose.RSA1_5,
		Key:       a.privateKey.Public(),
	}
	opts := (&jose.EncrypterOptions{}).WithContentType("JWT").WithType("JWT")
	return jose.NewEncrypter(jose.A256CBC_HS512, recp, opts)
}

func (a *Authorizer) newSigner() (jose.Signer, error) {
	signingKey := jose.SigningKey{
		Algorithm: jose.RS256,
		Key:       a.privateKey,
	}
	return jose.NewSigner(signingKey, (&jose.SignerOptions{}).WithType("JWT"))
}

package jwtissuer

import (
	"crypto/rsa"
	"fmt"
	"time"

	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

var (
	tokenLifetime = time.Hour * 24 * 2
)

type ValidatableClaims interface {
	Validate(e jwt.Expected) error
	ValidateWithLeeway(e jwt.Expected, leeway time.Duration) error
}

func NewIssuer(privateKey *rsa.PrivateKey) (Issuer, error) {
	if privateKey == nil {
		return nil, fmt.Errorf("privateKey is nil")
	}
	return &issuerImpl{
		privateKey: privateKey,
	}, nil
}

type issuerImpl struct {
	privateKey *rsa.PrivateKey
}

type Issuer interface {
	SignedAndEncrypted(claims interface{}) (string, error)
	Signed(claims interface{}) (string, error)
	ParseSignedAndEncrypted(token string, claims ValidatableClaims) error
	ParseSigned(token string, claims ValidatableClaims) error
}

func (i *issuerImpl) newSigner() (jose.Signer, error) {
	signingKey := jose.SigningKey{
		Algorithm: jose.RS256,
		Key:       i.privateKey,
	}
	return jose.NewSigner(signingKey, (&jose.SignerOptions{}).WithType("JWT"))
}

func (i *issuerImpl) newEncrypter() (jose.Encrypter, error) {
	recp := jose.Recipient{
		Algorithm: jose.RSA1_5,
		Key:       i.privateKey.Public(),
	}
	opts := (&jose.EncrypterOptions{}).WithContentType("JWT").WithType("JWT")
	return jose.NewEncrypter(jose.A256CBC_HS512, recp, opts)
}

func (i *issuerImpl) SignedAndEncrypted(claims interface{}) (string, error) {
	signer, err := i.newSigner()
	if err != nil {
		return "", err
	}
	encrypter, err := i.newEncrypter()
	if err != nil {
		return "", err
	}
	builder := jwt.SignedAndEncrypted(signer, encrypter)
	return builder.Claims(claims).CompactSerialize()
}

func (i *issuerImpl) Signed(claims interface{}) (string, error) {
	signer, err := i.newSigner()
	if err != nil {
		return "", err
	}
	return jwt.Signed(signer).Claims(claims).CompactSerialize()
}

func (i *issuerImpl) ParseSignedAndEncrypted(token string, claims ValidatableClaims) error {
	t, err := jwt.ParseSignedAndEncrypted(token)
	if err != nil {
		return err
	}
	nested, err := t.Decrypt(i.privateKey)
	if err != nil {
		return err
	}
	if err := nested.Claims(i.privateKey.Public(), &claims); err != nil {
		return err
	}
	if err := validateClaims(claims); err != nil {
		return err
	}
	return nil
}

func (i *issuerImpl) ParseSigned(token string, claims ValidatableClaims) error {
	t, err := jwt.ParseSigned(token)
	if err != nil {
		return err
	}
	if err := t.Claims(i.privateKey.Public(), &claims); err != nil {
		return err
	}
	if err := validateClaims(claims); err != nil {
		return err
	}
	return nil
}

func validateClaims(claims ValidatableClaims) error {
	expected := jwt.Expected{
		Audience: jwt.Audience{"mergechancetime.app"},
		Time:     time.Now(),
	}
	return claims.Validate(expected)
}

func NewStandardClaims() jwt.Claims {
	now := time.Now()
	return jwt.Claims{
		Issuer:    "mergechancetime.app",
		Audience:  jwt.Audience{"mergechancetime.app"},
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Expiry:    jwt.NewNumericDate(now.Add(tokenLifetime)),
	}
}

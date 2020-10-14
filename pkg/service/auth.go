package service

import (
	"context"
	"crypto/rsa"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/marcosQuesada/githubTop/pkg/log"
	"io/ioutil"
	"path/filepath"
	"time"
)

const (
	TokenName = "AccessToken"
)

// ErrUnauthorized happens on not authorized
var ErrUnauthorized = errors.New("forbidden access")

// AuthService takes care on authentication features
type AuthService interface {
	// Authorize user credentials and generate token on success
	Authorize(ctx context.Context, user, pass string) (token string, err error)

	// IsValidToken does token validation
	IsValidToken(ctx context.Context, token string) (bool, error)
}

// Authorizer delegates credential validation
type Authorizer interface {
	// ValidCredentials validates user and pass access
	ValidCredentials(user, pass string) (bool, error)
}

type defaultAuthService struct {
	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
	auth      Authorizer
	ttl       time.Duration
	appName   string
}

// NewAuth instantiates auth service
func NewAuth(auth Authorizer, pubKey, privKey string, exp time.Duration, appName string) *defaultAuthService {
	verifyKey, signKey := loadKeys(pubKey, privKey)

	return &defaultAuthService{
		verifyKey: verifyKey,
		signKey:   signKey,
		auth:      auth,
		ttl:       exp,
		appName:   appName,
	}
}

// Authorize generates a token on valid credentials access
func (s *defaultAuthService) Authorize(_ context.Context, user, pass string) (string, error) {

	v, err := s.auth.ValidCredentials(user, pass)
	if err != nil {
		log.Errorf("Unexpected error on credentials validation, error: %s", err)

		return "", ErrUnauthorized
	}

	if !v {
		return "", ErrUnauthorized
	}

	// define claims
	claims := jwt.MapClaims{
		"AccessToken": "level1",
		"CustomUserInfo": struct {
			Name    string
			AppName string
		}{user, s.appName},
		"exp": time.Now().Add(s.ttl).Unix(),
	}

	// create RS256 signer
	t := jwt.NewWithClaims(jwt.GetSigningMethod("RS256"), claims)

	// sign token using private key
	tokenKey, err := t.SignedString(s.signKey)
	if err != nil {
		log.Errorf("Unexpected error on Token Signing, error: %s", err)

		return "", ErrUnauthorized
	}

	return tokenKey, nil
}

// IsValidToken validates jwt token
func (s *defaultAuthService) IsValidToken(_ context.Context, tokenKey string) (bool, error) {
	if tokenKey == "" {
		log.Error("Token is Empty")

		return false, ErrUnauthorized
	}

	// we use private keys to sign tokens, so use public counter part to verify
	token, err := jwt.Parse(tokenKey, func(token *jwt.Token) (interface{}, error) {
		return s.verifyKey, nil
	})

	if err != nil {
		switch e := err.(type) {
		case *jwt.ValidationError:
			switch e.Errors {
			case jwt.ValidationErrorExpired:
				log.Error("Token Expired, get a new one")

			default:
				log.Errorf("Validation error, error: %s", err.Error())
			}

		default:
			log.Errorf("Unexpected error, error: %s", err.Error())

		}

		return false, ErrUnauthorized
	}

	if !token.Valid {
		log.Error("Token is Invalid")

		return false, ErrUnauthorized
	}

	log.Infof("Restricted access Enabled: %v", token)

	return true, nil

}

// loadKeys read public and private keys from file
func loadKeys(pubKey, privKey string) (verifyKey *rsa.PublicKey, signKey *rsa.PrivateKey) {
	absPath, _ := filepath.Abs(pubKey)
	signBytes, err := ioutil.ReadFile(absPath)
	if err != nil {
		log.Fatalf("Error reading Private Key file, err: %s", err.Error())
	}

	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		log.Fatalf("Error parsing Private Key file, err: %s", err.Error())
	}

	absPath, _ = filepath.Abs(privKey)
	verifyBytes, err := ioutil.ReadFile(absPath)
	if err != nil {
		log.Fatalf("Error reading Public Key file, err: %s", err.Error())
	}

	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		log.Fatalf("Error parsing Public Key file, err: %s", err.Error())
	}

	return
}

type staticAuthorizer struct {
	user, pass string
}

func NewDefaultStaticAuthorizer() *staticAuthorizer {
	return &staticAuthorizer{
		user: "test",
		pass: "known",
	}
}

func (s *staticAuthorizer) ValidCredentials(user, pass string) (bool, error) {
	return user == s.user && pass == s.pass, nil
}

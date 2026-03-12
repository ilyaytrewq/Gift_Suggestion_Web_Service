package id_generators

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/domain/authsession"
	"github.com/pkg/errors"
)

type AuthSessionIDGenerator struct{}

// var _ auth.AuthSessionIDGenerator = &AuthSessionIDGenerator{}
func (ag *AuthSessionIDGenerator) NewAuthSessionID() (*authsession.AuthSessionID, error) {
	s, err := randomB64URL(32)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate auth session id")
	}
	id := authsession.AuthSessionID(s)
	return &id, nil
}

func (ag *AuthSessionIDGenerator) NewCSRFSecret() (*authsession.CSRFSecret, error) {
	s, err := randomB64URL(32)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate csrf secret")
	}
	id := authsession.CSRFSecret(s)
	return &id, nil
}

func randomB64URL(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

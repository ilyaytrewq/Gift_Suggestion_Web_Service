package authjwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	authusecase "github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/modules/auth/usecase"
	"github.com/ilyaytrewq/Gift_Suggestion_Web_Service/internal/platform/config"
)

var ErrInvalidAccessToken = errors.New("invalid access token")

type Manager struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration
}

type claims struct {
	Role      string `json:"role"`
	SessionID string `json:"sid"`
	jwt.RegisteredClaims
}

func NewManager(cfg config.AuthConfig) *Manager {
	return &Manager{
		secret:   []byte(cfg.JWTSecret),
		issuer:   cfg.JWTIssuer,
		audience: cfg.JWTAudience,
		ttl:      cfg.AccessTokenTTL,
	}
}

func (m *Manager) IssueToken(actor authusecase.Actor, issuedAt time.Time) (string, error) {
	now := issuedAt.UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		Role:      actor.Role,
		SessionID: actor.SessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   actor.UserID,
			Issuer:    m.issuer,
			Audience:  []string{m.audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	})

	return token.SignedString(m.secret)
}

func (m *Manager) ParseToken(rawToken string) (authusecase.Actor, error) {
	token, err := jwt.ParseWithClaims(rawToken, &claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidAccessToken
		}

		return m.secret, nil
	}, jwt.WithIssuer(m.issuer), jwt.WithAudience(m.audience))
	if err != nil {
		return authusecase.Actor{}, err
	}

	tokenClaims, ok := token.Claims.(*claims)
	if !ok || !token.Valid {
		return authusecase.Actor{}, ErrInvalidAccessToken
	}

	return authusecase.Actor{
		UserID:    tokenClaims.Subject,
		SessionID: tokenClaims.SessionID,
		Role:      tokenClaims.Role,
	}, nil
}

func (m *Manager) TokenTTL() time.Duration {
	return m.ttl
}

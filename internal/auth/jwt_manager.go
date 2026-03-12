package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalidToken = errors.New("invalid token")

type JWTManager struct {
	secret []byte
	ttl    time.Duration
}

type ParsedToken struct {
	Subject   string
	ExpiresAt time.Time
}

func NewJWTManager(secret string, ttl time.Duration) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

func (m *JWTManager) Generate(userUUID string) (string, error) {
	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Subject:   userUUID,
		ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		IssuedAt:  jwt.NewNumericDate(now),
		ID:        uuid.NewString(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *JWTManager) Parse(tokenString string) (string, error) {
	parsed, err := m.ParseDetailed(tokenString)
	if err != nil {
		return "", err
	}
	return parsed.Subject, nil
}

func (m *JWTManager) ParseDetailed(tokenString string) (ParsedToken, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil {
		return ParsedToken{}, ErrInvalidToken
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid || claims.Subject == "" || claims.ExpiresAt == nil {
		return ParsedToken{}, ErrInvalidToken
	}

	return ParsedToken{
		Subject:   claims.Subject,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}

package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kxs888/kxs/backend/internal/errcode"
)

const issuer = "cga-api"

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func IssueAccess(secret string, ttl time.Duration, userID uuid.UUID, username string, now time.Time) (token string, expiresIn int, err error) {
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	exp := now.Add(ttl)
	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString([]byte(secret))
	if err != nil {
		return "", 0, fmt.Errorf("sign jwt: %w", err)
	}
	return s, int(ttl.Seconds()), nil
}

func ParseAccess(secret, token string, now time.Time) (*Claims, error) {
	if token == "" {
		return nil, errcode.Unauthorized("missing access token")
	}
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(issuer),
		jwt.WithLeeway(5*time.Second),
		jwt.WithTimeFunc(func() time.Time { return now }),
	)
	var claims Claims
	_, err := parser.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) || strings.Contains(err.Error(), "token is expired") {
			return nil, errcode.TokenExpired()
		}
		return nil, errcode.TokenInvalid()
	}
	if claims.Subject == "" {
		return nil, errcode.TokenInvalid()
	}
	return &claims, nil
}

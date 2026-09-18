package authn

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const Issuer = "sso"

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	UID     int64  `json:"uid"`
	Email   string `json:"email"`
	AppID   int64  `json:"app_id"`
	IsAdmin bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

func Parse(tokenString, secret, audience string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(secret), nil
		},
		jwt.WithIssuer(Issuer),
		jwt.WithAudience(audience),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidToken, err.Error())
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	if claims.UID <= 0 || claims.AppID <= 0 {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func newToken(uid int64, email string, appID int64, secret, audience string, ttl time.Duration) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UID:   uid,
		Email: email,
		AppID: appID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    Issuer,
			Audience:  jwt.ClaimStrings{audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	})
	return token.SignedString([]byte(secret))
}

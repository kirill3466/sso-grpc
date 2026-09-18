package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"sso/internal/models"
)

const Issuer = "sso"

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	UID   int64  `json:"uid"`
	Email string `json:"email"`
	AppID int64  `json:"app_id"`
	jwt.RegisteredClaims
}

func NewToken(user models.User, app models.App, ttl time.Duration) (string, error) {
	now := time.Now()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UID:   user.ID,
		Email: user.Email,
		AppID: app.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    Issuer,
			Audience:  jwt.ClaimStrings{app.Name},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	})

	signed, err := token.SignedString([]byte(app.Secret))
	if err != nil {
		return "", fmt.Errorf("jwt.NewToken: %w", err)
	}

	return signed, nil
}

func PeekAppID(tokenString string) (int64, error) {
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return 0, fmt.Errorf("%w: %s", ErrInvalidToken, err.Error())
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || claims.AppID <= 0 {
		return 0, ErrInvalidToken
	}

	return claims.AppID, nil
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

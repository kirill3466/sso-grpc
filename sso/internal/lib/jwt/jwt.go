package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"sso/internal/domain/models"
)

func NewToken(user models.User, app models.App, ttl time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid":    user.ID,
		"email":  user.Email,
		"app_id": app.ID,
		"exp":    time.Now().Add(ttl).Unix(),
	})

	signed, err := token.SignedString([]byte(app.Secret))
	if err != nil {
		return "", fmt.Errorf("jwt.NewToken: %w", err)
	}

	return signed, nil
}

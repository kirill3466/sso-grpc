package auth

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Auth struct {
	log      *slog.Logger
	tokenTTL time.Duration
}

func New(log *slog.Logger, tokenTTL time.Duration) *Auth {
	return &Auth{
		log:      log,
		tokenTTL: tokenTTL,
	}
}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int,
) (string, error) {
	const op = "auth.Login"

	return "", fmt.Errorf("%s: not implemented", op)
}

func (a *Auth) RegisterNewUser(
	ctx context.Context,
	email string,
	password string,
) (int64, error) {
	const op = "auth.RegisterNewUser"

	return 0, fmt.Errorf("%s: not implemented", op)
}

func (a *Auth) IsAdmin(
	ctx context.Context,
	userID int64,
) (bool, error) {
	const op = "auth.IsAdmin"

	return false, fmt.Errorf("%s: not implemented", op)
}

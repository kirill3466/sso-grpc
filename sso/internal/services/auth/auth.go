package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"

	"sso/internal/domain/models"
	sl "sso/internal/lib"
	"sso/internal/lib/jwt"
)

type Auth struct {
	log            *slog.Logger
	tokenTTL       time.Duration
	userRepository UserRepository
	appRepository  AppRepository
}

type UserRepository interface {
	CreateUser(
		ctx context.Context,
		email string,
		passwordHash string,
		isAdmin bool,
	) (userID int64, err error)
	GetUserByEmail(
		ctx context.Context,
		email string,
	) (user *models.User, err error)
	IsAdmin(
		ctx context.Context,
		userID int64,
	) (isAdmin bool, err error)
}

type AppRepository interface {
	CreateApp(
		ctx context.Context,
		name string,
		secret string,
	) (appID int64, err error)
	GetAppByID(
		ctx context.Context,
		appID int64,
	) (app *models.App, err error)
}

func New(
	log *slog.Logger,
	tokenTTL time.Duration,
	userRepository UserRepository,
	appRepository AppRepository,
) *Auth {
	return &Auth{
		log:            log,
		tokenTTL:       tokenTTL,
		userRepository: userRepository,
		appRepository:  appRepository,
	}
}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int,
) (string, error) {
	const op = "auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
		slog.Int("app_id", appID),
	)

	log.Info("logging in user")

	user, err := a.userRepository.GetUserByEmail(ctx, email)

	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			log.Warn("user not found", sl.Err(err))
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}
		log.Error("failed to get user by email", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		log.Warn("invalid credentials", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appRepository.GetAppByID(ctx, int64(appID))
	if err != nil {
		if errors.Is(err, ErrAppNotFound) {
			return "", fmt.Errorf("%s: %w", op, ErrAppNotFound)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	token, err := jwt.NewToken(*user, *app, a.tokenTTL)
	if err != nil {
		log.Error("failed to generate token", sl.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (a *Auth) RegisterNewUser(
	ctx context.Context,
	email string,
	password string,
) (int64, error) {
	const op = "auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("registering new user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", "error", err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	userID, err := a.userRepository.CreateUser(ctx, email, string(passHash), false)
	if err != nil {
		log.Error("failed to create user", "error", err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user created", "user_id", userID)

	return userID, nil
}

func (a *Auth) IsAdmin(
	ctx context.Context,
	userID int64,
) (bool, error) {
	const op = "auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.Int64("user_id", userID),
	)

	log.Info("checking if user is admin")

	isAdmin, err := a.userRepository.IsAdmin(ctx, userID)
	if err != nil {
		log.Error("failed to check if user is admin", sl.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user is admin", "is_admin", isAdmin)

	return isAdmin, nil
}

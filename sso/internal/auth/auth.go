package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"

	"sso/internal/jwt"
	"sso/internal/models"
	"sso/internal/slogx"
	"sso/internal/storage"
)

var dummyPasswordHash []byte

func init() {
	hash, err := bcrypt.GenerateFromPassword([]byte("timing-dummy"), bcrypt.DefaultCost)
	if err != nil {
		panic("auth: dummy password hash: " + err.Error())
	}
	dummyPasswordHash = hash
}

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
		slog.Int("app_id", appID),
	)

	log.Debug("logging in user")

	user, err := a.userRepository.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		log.Error("failed to get user by email", slogx.Err(err))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	hash := dummyPasswordHash
	if user != nil {
		hash = []byte(user.PasswordHash)
	}

	if err := bcrypt.CompareHashAndPassword(hash, []byte(password)); err != nil || user == nil {
		log.Warn("invalid credentials")
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appRepository.GetAppByID(ctx, int64(appID))
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return "", fmt.Errorf("%s: %w", op, ErrAppNotFound)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	token, err := jwt.NewToken(*user, *app, a.tokenTTL)
	if err != nil {
		log.Error("failed to generate token", slogx.Err(err))
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

	log := a.log.With(slog.String("op", op))

	log.Debug("registering new user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", slogx.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	userID, err := a.userRepository.CreateUser(ctx, email, string(passHash), false)
	if err != nil {
		if errors.Is(err, storage.ErrAlreadyExists) {
			log.Warn("user already exists")
			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		log.Error("failed to create user", slogx.Err(err))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user created", slog.Int64("user_id", userID))

	return userID, nil
}

func (a *Auth) IsAdmin(
	ctx context.Context,
	accessToken string,
	userID int64,
) (bool, error) {
	const op = "auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.Int64("user_id", userID),
	)

	if accessToken == "" {
		return false, fmt.Errorf("%s: %w", op, ErrUnauthenticated)
	}

	appID, err := jwt.PeekAppID(accessToken)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, ErrUnauthenticated)
	}

	app, err := a.appRepository.GetAppByID(ctx, appID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return false, fmt.Errorf("%s: %w", op, ErrUnauthenticated)
		}
		log.Error("failed to get app", slogx.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}

	claims, err := jwt.Parse(accessToken, app.Secret, app.Name)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, ErrUnauthenticated)
	}

	if claims.UID != userID {
		return false, fmt.Errorf("%s: %w", op, ErrAccessDenied)
	}

	isAdmin, err := a.userRepository.IsAdmin(ctx, claims.UID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return false, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		log.Error("failed to check if user is admin", slogx.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Debug("user admin check done", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}

type TokenInfo struct {
	UserID  int64
	Email   string
	AppID   int64
	IsAdmin bool
}

func (a *Auth) Validate(ctx context.Context, accessToken string) (TokenInfo, error) {
	const op = "auth.Validate"

	if accessToken == "" {
		return TokenInfo{}, fmt.Errorf("%s: %w", op, ErrUnauthenticated)
	}

	appID, err := jwt.PeekAppID(accessToken)
	if err != nil {
		return TokenInfo{}, fmt.Errorf("%s: %w", op, ErrUnauthenticated)
	}

	app, err := a.appRepository.GetAppByID(ctx, appID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return TokenInfo{}, fmt.Errorf("%s: %w", op, ErrUnauthenticated)
		}
		a.log.Error("failed to get app", slog.String("op", op), slogx.Err(err))
		return TokenInfo{}, fmt.Errorf("%s: %w", op, err)
	}

	claims, err := jwt.Parse(accessToken, app.Secret, app.Name)
	if err != nil {
		return TokenInfo{}, fmt.Errorf("%s: %w", op, ErrUnauthenticated)
	}

	isAdmin, err := a.userRepository.IsAdmin(ctx, claims.UID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return TokenInfo{}, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		a.log.Error("failed to check if user is admin", slog.String("op", op), slogx.Err(err))
		return TokenInfo{}, fmt.Errorf("%s: %w", op, err)
	}

	return TokenInfo{
		UserID:  claims.UID,
		Email:   claims.Email,
		AppID:   claims.AppID,
		IsAdmin: isAdmin,
	}, nil
}

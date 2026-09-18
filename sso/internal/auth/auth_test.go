package auth_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	authservice "sso/internal/auth"
	"sso/internal/jwt"
	"sso/internal/models"
	"sso/internal/storage"
)

type fakeUsers struct {
	user       *models.User
	getErr     error
	createID   int64
	createErr  error
	isAdmin    bool
	isAdminErr error
}

func (f fakeUsers) CreateUser(context.Context, string, string, bool) (int64, error) {
	return f.createID, f.createErr
}

func (f fakeUsers) GetUserByEmail(context.Context, string) (*models.User, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.user, nil
}

func (f fakeUsers) IsAdmin(context.Context, int64) (bool, error) {
	return f.isAdmin, f.isAdminErr
}

type fakeApps struct {
	app *models.App
	err error
}

func (f fakeApps) CreateApp(context.Context, string, string) (int64, error) {
	return 0, nil
}

func (f fakeApps) GetAppByID(context.Context, int64) (*models.App, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.app, nil
}

func testAuth(users authservice.UserRepository, apps authservice.AppRepository) *authservice.Auth {
	return authservice.New(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		time.Hour,
		users,
		apps,
	)
}

func TestLogin_UnknownUser(t *testing.T) {
	t.Parallel()

	_, err := testAuth(fakeUsers{getErr: storage.ErrNotFound}, fakeApps{}).Login(
		context.Background(), "nope@example.com", "password", 1,
	)
	if !errors.Is(err, authservice.ErrInvalidCredentials) {
		t.Fatalf("got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	t.Parallel()

	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}

	user := &models.User{ID: 3, Email: "user@example.com", PasswordHash: string(hash)}
	app := &models.App{ID: 1, Name: "test", Secret: "app-secret"}

	token, err := testAuth(fakeUsers{user: user}, fakeApps{app: app}).Login(
		context.Background(), user.Email, "secret", 1,
	)
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("empty token")
	}

	claims, err := jwt.Parse(token, app.Secret, app.Name)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UID != user.ID {
		t.Fatalf("uid = %d", claims.UID)
	}
}

func TestRegister_AlreadyExists(t *testing.T) {
	t.Parallel()

	_, err := testAuth(
		fakeUsers{createErr: storage.ErrAlreadyExists},
		fakeApps{},
	).RegisterNewUser(context.Background(), "user@example.com", "long-pass")
	if !errors.Is(err, authservice.ErrUserExists) {
		t.Fatalf("got %v", err)
	}
}

func TestIsAdmin_RequiresToken(t *testing.T) {
	t.Parallel()

	_, err := testAuth(fakeUsers{isAdmin: true}, fakeApps{}).IsAdmin(context.Background(), "", 1)
	if !errors.Is(err, authservice.ErrUnauthenticated) {
		t.Fatalf("got %v", err)
	}
}

func TestIsAdmin_RejectsOtherUser(t *testing.T) {
	t.Parallel()

	app := models.App{ID: 1, Name: "test", Secret: "app-secret"}
	token, err := jwt.NewToken(models.User{ID: 1, Email: "a@b.c"}, app, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	_, err = testAuth(fakeUsers{isAdmin: true}, fakeApps{app: &app}).IsAdmin(
		context.Background(), token, 99,
	)
	if !errors.Is(err, authservice.ErrAccessDenied) {
		t.Fatalf("got %v", err)
	}
}

func TestIsAdmin_Success(t *testing.T) {
	t.Parallel()

	app := models.App{ID: 1, Name: "test", Secret: "app-secret"}
	token, err := jwt.NewToken(models.User{ID: 4, Email: "a@b.c"}, app, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	ok, err := testAuth(fakeUsers{isAdmin: true}, fakeApps{app: &app}).IsAdmin(
		context.Background(), token, 4,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("want admin")
	}
}

package auth

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	ssov1 "github.com/kirill3466/protos/gen/go/sso"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	authservice "sso/internal/auth"
)

type stubAuth struct {
	loginErr   error
	token      string
	isAdmin    bool
	isAdminErr error
}

func (s stubAuth) Login(context.Context, string, string, int) (string, error) {
	return s.token, s.loginErr
}

func (s stubAuth) RegisterNewUser(context.Context, string, string) (int64, error) {
	return 0, nil
}

func (s stubAuth) IsAdmin(context.Context, string, int64) (bool, error) {
	return s.isAdmin, s.isAdminErr
}

func newTestServer(auth Auth) *serverAPI {
	return &serverAPI{
		log:  slog.New(slog.NewTextHandler(io.Discard, nil)),
		auth: auth,
	}
}

func TestLogin_MapsDomainErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		code codes.Code
	}{
		{name: "invalid credentials", err: authservice.ErrInvalidCredentials, code: codes.Unauthenticated},
		{name: "app not found", err: authservice.ErrAppNotFound, code: codes.NotFound},
		{name: "deadline", err: context.DeadlineExceeded, code: codes.DeadlineExceeded},
		{name: "canceled", err: context.Canceled, code: codes.Canceled},
		{name: "unknown", err: errors.New("sql: boom"), code: codes.Internal},
	}

	req := &ssov1.LoginRequest{Email: "user@example.com", Password: "secret", AppId: 1}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := newTestServer(stubAuth{loginErr: tt.err}).Login(context.Background(), req)
			if status.Code(err) != tt.code {
				t.Fatalf("got %v, want %s", err, tt.code)
			}
			if tt.code == codes.Internal && status.Convert(err).Message() != "internal error" {
				t.Fatalf("internal error must not leak details: %v", err)
			}
		})
	}
}

func TestLogin_Success(t *testing.T) {
	t.Parallel()

	resp, err := newTestServer(stubAuth{token: "jwt-token"}).Login(context.Background(), &ssov1.LoginRequest{
		Email:    "user@example.com",
		Password: "secret",
		AppId:    1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.GetToken() != "jwt-token" {
		t.Fatalf("token = %q", resp.GetToken())
	}
}

func TestIsAdmin_RequiresBearer(t *testing.T) {
	t.Parallel()

	_, err := newTestServer(stubAuth{isAdmin: true}).IsAdmin(
		context.Background(),
		&ssov1.IsAdminRequest{UserId: 1},
	)
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("got %v, want Unauthenticated", err)
	}
}

func TestIsAdmin_Success(t *testing.T) {
	t.Parallel()

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"authorization", "Bearer test-token",
	))

	resp, err := newTestServer(stubAuth{isAdmin: true}).IsAdmin(ctx, &ssov1.IsAdminRequest{UserId: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.GetIsAdmin() {
		t.Fatal("want is_admin true")
	}
}

func TestIsAdmin_MapsAccessDenied(t *testing.T) {
	t.Parallel()

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"authorization", "Bearer test-token",
	))

	_, err := newTestServer(stubAuth{isAdminErr: authservice.ErrAccessDenied}).IsAdmin(
		ctx,
		&ssov1.IsAdminRequest{UserId: 2},
	)
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("got %v, want PermissionDenied", err)
	}
}

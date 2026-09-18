package auth

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	ssov1 "github.com/kirill3466/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	authservice "sso/internal/auth"
	"sso/internal/slogx"
)

type Auth interface {
	Login(ctx context.Context, email string, password string, appID int) (token string, err error)
	RegisterNewUser(ctx context.Context, email string, password string) (userID int64, err error)
	IsAdmin(ctx context.Context, accessToken string, userID int64) (isAdmin bool, err error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	log  *slog.Logger
	auth Auth
}

func Register(gRPC *grpc.Server, log *slog.Logger, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{
		log:  log,
		auth: auth,
	})
}

func (s *serverAPI) Login(
	ctx context.Context,
	req *ssov1.LoginRequest,
) (*ssov1.LoginResponse, error) {
	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		return nil, s.grpcError("Login", err)
	}

	return &ssov1.LoginResponse{Token: token}, nil
}

func (s *serverAPI) Register(
	ctx context.Context,
	req *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {
	userID, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, s.grpcError("Register", err)
	}

	return &ssov1.RegisterResponse{UserId: userID}, nil
}

func (s *serverAPI) IsAdmin(
	ctx context.Context,
	req *ssov1.IsAdminRequest,
) (*ssov1.IsAdminResponse, error) {
	token, err := bearerFrom(ctx)
	if err != nil {
		return nil, s.grpcError("IsAdmin", err)
	}

	isAdmin, err := s.auth.IsAdmin(ctx, token, req.GetUserId())
	if err != nil {
		return nil, s.grpcError("IsAdmin", err)
	}

	return &ssov1.IsAdminResponse{IsAdmin: isAdmin}, nil
}

func bearerFrom(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", authservice.ErrUnauthenticated
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", authservice.ErrUnauthenticated
	}

	scheme, token, found := strings.Cut(values[0], " ")
	if !found || !strings.EqualFold(scheme, "bearer") || token == "" {
		return "", authservice.ErrUnauthenticated
	}

	return token, nil
}

func (s *serverAPI) grpcError(method string, err error) error {
	switch {
	case errors.Is(err, authservice.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case errors.Is(err, authservice.ErrUnauthenticated):
		return status.Error(codes.Unauthenticated, "unauthenticated")
	case errors.Is(err, authservice.ErrAccessDenied):
		return status.Error(codes.PermissionDenied, "permission denied")
	case errors.Is(err, authservice.ErrUserExists):
		return status.Error(codes.AlreadyExists, "user already exists")
	case errors.Is(err, authservice.ErrUserNotFound):
		return status.Error(codes.NotFound, "user not found")
	case errors.Is(err, authservice.ErrAppNotFound):
		return status.Error(codes.NotFound, "app not found")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "deadline exceeded")
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "canceled")
	default:
		s.log.Error("handler failed", slog.String("method", method), slogx.Err(err))
		return status.Error(codes.Internal, "internal error")
	}
}

package authn

import (
	"context"
	"errors"

	rtdbv1 "github.com/kirill3466/protos/gen/go/rtdb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Validator interface {
	Validate(ctx context.Context, token string) (*Claims, error)
}

type Config struct {
	Validator Validator
	Write     WriteAccess
}

func UnaryServerInterceptor(cfg Config) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		claims, err := authenticate(ctx, cfg.Validator)
		if err != nil {
			return nil, grpcAuthError(err)
		}
		if info.FullMethod == rtdbv1.RTDB_SetTag_FullMethodName && !canWrite(cfg.Write, claims) {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}

		return handler(WithClaims(ctx, claims), req)
	}
}

func StreamServerInterceptor(cfg Config) grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		_ *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		claims, err := authenticate(ss.Context(), cfg.Validator)
		if err != nil {
			return grpcAuthError(err)
		}

		return handler(srv, &contextStream{
			ServerStream: ss,
			ctx:          WithClaims(ss.Context(), claims),
		})
	}
}

func authenticate(ctx context.Context, v Validator) (*Claims, error) {
	token, err := BearerFrom(ctx)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, ErrUnauthenticated
	}

	claims, err := v.Validate(ctx, token)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func canWrite(access WriteAccess, claims *Claims) bool {
	return claims.IsAdmin || access.Allows(claims)
}

func grpcAuthError(err error) error {
	switch {
	case errors.Is(err, ErrUnauthenticated), errors.Is(err, ErrInvalidToken):
		return status.Error(codes.Unauthenticated, "unauthenticated")
	default:
		return status.Error(codes.Unavailable, "auth service unavailable")
	}
}

type contextStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *contextStream) Context() context.Context {
	return s.ctx
}

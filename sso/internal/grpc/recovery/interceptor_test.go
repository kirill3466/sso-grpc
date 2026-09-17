package recovery_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grprecovery "sso/internal/grpc/recovery"
)

func TestUnaryInterceptorRecoversPanic(t *testing.T) {
	t.Parallel()

	interceptor := grprecovery.UnaryServerInterceptor(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)

	_, err := interceptor(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/sso.v1.Auth/Login"},
		func(context.Context, any) (any, error) {
			panic("boom")
		},
	)
	if status.Code(err) != codes.Internal {
		t.Fatalf("want Internal, got %v", err)
	}
	if status.Convert(err).Message() != "internal error" {
		t.Fatalf("panic must not leak: %v", err)
	}
}

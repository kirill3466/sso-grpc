package timeout_test

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"

	grptimeout "sso/internal/grpc/interceptors/timeout"
)

func TestUnaryInterceptorSetsDeadline(t *testing.T) {
	t.Parallel()

	interceptor := grptimeout.UnaryServerInterceptor(50 * time.Millisecond)

	_, err := interceptor(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/test"},
		func(ctx context.Context, _ any) (any, error) {
			deadline, ok := ctx.Deadline()
			if !ok {
				t.Fatal("deadline must be set")
			}
			if time.Until(deadline) > 50*time.Millisecond {
				t.Fatalf("deadline too far: %v", time.Until(deadline))
			}
			return "ok", nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
}

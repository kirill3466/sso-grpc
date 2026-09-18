package timeout

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

func UnaryServerInterceptor(d time.Duration) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if d <= 0 {
			return handler(ctx, req)
		}

		ctx, cancel := context.WithTimeout(ctx, d)
		defer cancel()

		return handler(ctx, req)
	}
}

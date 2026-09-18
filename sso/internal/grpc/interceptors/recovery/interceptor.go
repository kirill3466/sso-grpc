package recovery

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryServerInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}

			log.Error(
				"panic recovered",
				slog.Any("panic", rec),
				slog.String("method", info.FullMethod),
			)
			err = status.Error(codes.Internal, "internal error")
		}()

		return handler(ctx, req)
	}
}

package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	authgrpc "sso/internal/grpc/auth"
	grprecovery "sso/internal/grpc/interceptors/recovery"
	grptimeout "sso/internal/grpc/interceptors/timeout"
	grpvalidate "sso/internal/grpc/interceptors/validate"

	"buf.build/go/protovalidate"
	gogrpc "google.golang.org/grpc"
)

type App struct {
	gRPCServer *gogrpc.Server
	log        *slog.Logger
	port       int
}

func New(
	log *slog.Logger,
	authService authgrpc.Auth,
	port int,
	timeout time.Duration,
) *App {
	validator, err := protovalidate.New()
	if err != nil {
		panic(fmt.Errorf("protovalidate: %w", err))
	}

	gRPCServer := gogrpc.NewServer(
		gogrpc.ChainUnaryInterceptor(
			grprecovery.UnaryServerInterceptor(log),
			grptimeout.UnaryServerInterceptor(timeout),
			grpvalidate.UnaryServerInterceptor(validator),
		),
	)

	authgrpc.Register(gRPCServer, log, authService)

	return &App{
		gRPCServer: gRPCServer,
		log:        log,
		port:       port,
	}
}

func (a *App) Run() error {
	const op = "grpc.Run"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("grpc server is running", "address", listener.Addr().String())

	if err := a.gRPCServer.Serve(listener); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop(ctx context.Context) error {
	const op = "grpc.Stop"

	a.log.With(slog.String("op", op)).
		Info("stopping gRPC server", slog.Int("port", a.port))

	stopped := make(chan struct{})
	go func() {
		a.gRPCServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-ctx.Done():
		a.gRPCServer.Stop()
		<-stopped
		return fmt.Errorf("%s: %w", op, ctx.Err())
	}

	a.log.With(slog.String("op", op)).
		Info("gRPC server stopped", slog.Int("port", a.port))

	return nil
}

package grpc_app

import (
	"fmt"
	"log/slog"
	"net"
	authgrpc "sso/internal/grpc/auth"
	grpvalidate "sso/internal/grpc/validate"

	"buf.build/go/protovalidate"
	"google.golang.org/grpc"
)

type App struct {
	gRPCServer *grpc.Server
	log        *slog.Logger
	port       int
}

func New(
	log *slog.Logger,
	authService authgrpc.Auth,
	port int,
) *App {
	validator, err := protovalidate.New()
	if err != nil {
		panic(fmt.Errorf("protovalidate: %w", err))
	}

	gRPCServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpvalidate.UnaryServerInterceptor(validator)),
	)

	authgrpc.Register(gRPCServer, log, authService)

	return &App{
		gRPCServer: gRPCServer,
		log:        log,
		port:       port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "app.grpc_app.Run"

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

func (a *App) Stop() error {
	const op = "app.grpc_app.Stop"

	a.log.With(slog.String("op", op)).
		Info("stopping gRPC server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()

	a.log.With(slog.String("op", op)).
		Info("gRPC server stopped", slog.Int("port", a.port))

	return nil
}

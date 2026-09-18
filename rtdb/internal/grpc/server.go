package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"rtdb/internal/authn"
	"rtdb/internal/config"
	grprecovery "rtdb/internal/grpc/interceptors/recovery"
	grptimeout "rtdb/internal/grpc/interceptors/timeout"
	grpvalidate "rtdb/internal/grpc/interceptors/validate"
	rtdbgrpc "rtdb/internal/grpc/rtdb"

	"buf.build/go/protovalidate"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type App struct {
	gRPCServer *gogrpc.Server
	log        *slog.Logger
	port       int
}

func New(
	log *slog.Logger,
	rtdbService rtdbgrpc.RTDB,
	cfg *config.Config,
	tokenValidator authn.Validator,
) *App {
	pv, err := protovalidate.New()
	if err != nil {
		panic(fmt.Errorf("protovalidate: %w", err))
	}

	authnCfg := authn.Config{
		Validator: tokenValidator,
		Write:     authn.NewWriteAccess(cfg.WriteUserIDs, cfg.WriteEmails),
	}

	gRPCServer := gogrpc.NewServer(
		gogrpc.ChainUnaryInterceptor(
			grprecovery.UnaryServerInterceptor(log),
			grptimeout.UnaryServerInterceptor(cfg.GRPC.Timeout),
			authn.UnaryServerInterceptor(authnCfg),
			grpvalidate.UnaryServerInterceptor(pv),
		),
		gogrpc.ChainStreamInterceptor(
			authn.StreamServerInterceptor(authnCfg),
			grpvalidate.StreamServerInterceptor(pv),
		),
	)

	rtdbgrpc.Register(gRPCServer, log, rtdbService)
	reflection.Register(gRPCServer)

	return &App{
		gRPCServer: gRPCServer,
		log:        log,
		port:       cfg.GRPC.Port,
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

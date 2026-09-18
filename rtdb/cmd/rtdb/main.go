package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"rtdb/internal/app"
	"rtdb/internal/config"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info(
		"starting rtdb server",
		slog.String("env", cfg.Env),
		slog.Int("grpc_port", cfg.GRPC.Port),
	)

	application, err := app.New(ctx, log, cfg)
	if err != nil {
		log.Error("failed to create application", slog.String("error", err.Error()))
		os.Exit(1)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- application.GRPCSrv.Run()
	}()

	select {
	case <-ctx.Done():
		log.Info("shutting down rtdb server")
	case err := <-errCh:
		if err != nil {
			log.Error("grpc server failed", slog.String("error", err.Error()))
		}
	}

	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := application.Stop(shutdownCtx); err != nil {
		log.Error("shutdown failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log.Info("rtdb server stopped")
}

func setupLogger(env string) *slog.Logger {
	switch env {
	case envLocal:
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	case envDev:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	case envProd:
		fallthrough
	default:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}
}

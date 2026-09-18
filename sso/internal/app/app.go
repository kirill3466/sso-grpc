package app

import (
	"context"
	"errors"
	"log/slog"

	grpc_app "sso/internal/app/grpc"
	"sso/internal/config"
	authservice "sso/internal/services/auth"
	"sso/internal/storage/sqlite"
)

type App struct {
	GRPCSrv *grpc_app.App
	storage *sqlite.Storage
}

func New(ctx context.Context, log *slog.Logger, cfg *config.Config) (*App, error) {
	storage, err := sqlite.New(ctx, cfg.StoragePath)
	if err != nil {
		return nil, err
	}

	authService := authservice.New(log, cfg.TokenTTL, storage, storage)
	grpcApp := grpc_app.New(log, authService, cfg.GRPC.Port, cfg.GRPC.Timeout)

	return &App{GRPCSrv: grpcApp, storage: storage}, nil
}

func (a *App) Stop(ctx context.Context) error {
	return errors.Join(a.GRPCSrv.Stop(ctx), a.storage.Close())
}

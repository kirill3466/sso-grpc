package app

import (
	"context"
	"errors"
	"log/slog"

	"sso/internal/auth"
	"sso/internal/config"
	ssogrpc "sso/internal/grpc"
	"sso/internal/storage/sqlite"
)

type App struct {
	GRPCSrv *ssogrpc.App
	storage *sqlite.Storage
}

func New(ctx context.Context, log *slog.Logger, cfg *config.Config) (*App, error) {
	storage, err := sqlite.New(ctx, cfg.StoragePath)
	if err != nil {
		return nil, err
	}

	authService := auth.New(log, cfg.TokenTTL, storage, storage)
	grpcApp := ssogrpc.New(log, authService, cfg.GRPC.Port, cfg.GRPC.Timeout)

	return &App{GRPCSrv: grpcApp, storage: storage}, nil
}

func (a *App) Stop(ctx context.Context) error {
	return errors.Join(a.GRPCSrv.Stop(ctx), a.storage.Close())
}

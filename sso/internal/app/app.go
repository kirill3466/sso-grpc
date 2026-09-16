package app

import (
	"log/slog"
	"time"

	grpc_app "sso/internal/app/grpc"
	authservice "sso/internal/services/auth"
)

type App struct {
	GRPCSrv *grpc_app.App
}

func New(
	log *slog.Logger,
	port int,
	storagePath string,
	tokenTTL time.Duration,
) *App {
	_ = storagePath

	authService := authservice.New(log, tokenTTL, nil, nil)

	grpcApp := grpc_app.New(log, authService, port)

	return &App{
		GRPCSrv: grpcApp,
	}
}

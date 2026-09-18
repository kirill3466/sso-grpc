package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	ssov1 "github.com/kirill3466/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"rtdb/internal/authn"
	"rtdb/internal/catalog"
	"rtdb/internal/config"
	rtdbgrpc "rtdb/internal/grpc"
	"rtdb/internal/rtdb"
	"rtdb/internal/storage/memstore"
)

type App struct {
	GRPCSrv *rtdbgrpc.App
	ssoConn *grpc.ClientConn
}

func New(ctx context.Context, log *slog.Logger, cfg *config.Config) (*App, error) {
	store, err := memstore.New(ctx, catalog.Tags(), cfg.StaleAfter)
	if err != nil {
		return nil, err
	}

	if cfg.Simulate {
		go store.RunSimulator(ctx, cfg.SimulateEvery)
	}

	ssoConn, err := grpc.NewClient(
		cfg.SSO.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("dial sso: %w", err)
	}

	validator := authn.NewCachedValidator(
		authn.NewSSOValidator(ssov1.NewAuthClient(ssoConn)),
		5*time.Second,
	)

	service := rtdb.New(log, store)
	grpcApp := rtdbgrpc.New(log, service, cfg, validator)

	return &App{GRPCSrv: grpcApp, ssoConn: ssoConn}, nil
}

func (a *App) Stop(ctx context.Context) error {
	return errors.Join(a.GRPCSrv.Stop(ctx), a.ssoConn.Close())
}

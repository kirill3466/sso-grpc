package sqlite

import (
	"context"
	"fmt"

	"sso/internal/models"
)

const (
	queryCreateApp = `
		INSERT INTO apps (name, secret)
		VALUES (?, ?)`

	queryGetAppByID = `
		SELECT id, name, secret
		FROM apps
		WHERE id = ?
		LIMIT 1`
)

func (s *Storage) CreateApp(ctx context.Context, name string, secret string) (int64, error) {
	const op = "storage.sqlite.CreateApp"

	ctx, cancel := withQueryTimeout(ctx)
	defer cancel()

	res, err := s.createAppStmt.ExecContext(ctx, name, secret)
	if err != nil {
		return 0, mapSQLError(op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: last insert id: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetAppByID(ctx context.Context, appID int64) (*models.App, error) {
	const op = "storage.sqlite.GetAppByID"

	ctx, cancel := withQueryTimeout(ctx)
	defer cancel()

	var app models.App
	err := s.getAppByIDStmt.QueryRowContext(ctx, appID).Scan(
		&app.ID,
		&app.Name,
		&app.Secret,
	)
	if err != nil {
		return nil, mapSQLError(op, err)
	}

	return &app, nil
}

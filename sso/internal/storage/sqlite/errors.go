package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	libsqlite "modernc.org/sqlite"

	"sso/internal/storage"
)

const (
	sqliteConstraintUnique     = 2067
	sqliteConstraintPrimaryKey = 1555
)

func mapSQLError(op string, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, sql.ErrNoRows):
		return fmt.Errorf("%s: %w", op, storage.ErrNotFound)
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return fmt.Errorf("%s: %w", op, err)
	case isUniqueConstraint(err):
		return fmt.Errorf("%s: %w", op, storage.ErrAlreadyExists)
	default:
		return fmt.Errorf("%s: %w", op, err)
	}
}

func isUniqueConstraint(err error) bool {
	var sqliteErr *libsqlite.Error
	if errors.As(err, &sqliteErr) {
		switch sqliteErr.Code() {
		case sqliteConstraintUnique, sqliteConstraintPrimaryKey:
			return true
		}
	}

	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}

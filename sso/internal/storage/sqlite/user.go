package sqlite

import (
	"context"
	"fmt"

	"sso/internal/models"
)

const (
	queryCreateUser = `
		INSERT INTO users (email, pass_hash, is_admin)
		VALUES (?, ?, ?)`

	queryGetUserByEmail = `
		SELECT id, email, pass_hash, is_admin
		FROM users
		WHERE email = ?
		LIMIT 1`

	queryIsAdmin = `
		SELECT is_admin
		FROM users
		WHERE id = ?
		LIMIT 1`
)

func (s *Storage) CreateUser(
	ctx context.Context,
	email string,
	passwordHash string,
	isAdmin bool,
) (int64, error) {
	const op = "storage.sqlite.CreateUser"

	ctx, cancel := withQueryTimeout(ctx)
	defer cancel()

	res, err := s.createUserStmt.ExecContext(ctx, email, passwordHash, isAdmin)
	if err != nil {
		return 0, mapSQLError(op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: last insert id: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	const op = "storage.sqlite.GetUserByEmail"

	ctx, cancel := withQueryTimeout(ctx)
	defer cancel()

	var (
		user     models.User
		passHash []byte
	)

	err := s.getUserByEmailStmt.QueryRowContext(ctx, email).Scan(
		&user.ID,
		&user.Email,
		&passHash,
		&user.IsAdmin,
	)
	if err != nil {
		return nil, mapSQLError(op, err)
	}

	user.PasswordHash = string(passHash)
	return &user, nil
}

func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "storage.sqlite.IsAdmin"

	ctx, cancel := withQueryTimeout(ctx)
	defer cancel()

	var isAdmin bool
	err := s.isAdminStmt.QueryRowContext(ctx, userID).Scan(&isAdmin)
	if err != nil {
		return false, mapSQLError(op, err)
	}

	return isAdmin, nil
}

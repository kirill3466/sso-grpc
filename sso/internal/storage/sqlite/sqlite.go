package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

const queryTimeout = time.Second

type Storage struct {
	db *sql.DB

	createUserStmt     *sql.Stmt
	getUserByEmailStmt *sql.Stmt
	isAdminStmt        *sql.Stmt
	createAppStmt      *sql.Stmt
	getAppByIDStmt     *sql.Stmt
}

func New(ctx context.Context, path string) (*Storage, error) {
	const op = "storage.sqlite.New"

	dsn := fmt.Sprintf(
		"file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=synchronous(NORMAL)",
		path,
	)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: open: %w", op, err)
	}

	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(0)

	s := &Storage{db: db}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = s.Close()
		return nil, fmt.Errorf("%s: ping: %w", op, err)
	}

	if err := s.prepare(pingCtx); err != nil {
		_ = s.Close()
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return s, nil
}

func (s *Storage) prepare(ctx context.Context) error {
	var err error

	s.createUserStmt, err = s.db.PrepareContext(ctx, queryCreateUser)
	if err != nil {
		return fmt.Errorf("prepare createUser: %w", err)
	}
	s.getUserByEmailStmt, err = s.db.PrepareContext(ctx, queryGetUserByEmail)
	if err != nil {
		return fmt.Errorf("prepare getUserByEmail: %w", err)
	}
	s.isAdminStmt, err = s.db.PrepareContext(ctx, queryIsAdmin)
	if err != nil {
		return fmt.Errorf("prepare isAdmin: %w", err)
	}
	s.createAppStmt, err = s.db.PrepareContext(ctx, queryCreateApp)
	if err != nil {
		return fmt.Errorf("prepare createApp: %w", err)
	}
	s.getAppByIDStmt, err = s.db.PrepareContext(ctx, queryGetAppByID)
	if err != nil {
		return fmt.Errorf("prepare getAppByID: %w", err)
	}

	return nil
}

func (s *Storage) Ping(ctx context.Context) error {
	const op = "storage.sqlite.Ping"

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) Close() error {
	const op = "storage.sqlite.Close"

	err := errors.Join(
		closeStmt(s.createUserStmt),
		closeStmt(s.getUserByEmailStmt),
		closeStmt(s.isAdminStmt),
		closeStmt(s.createAppStmt),
		closeStmt(s.getAppByIDStmt),
		s.db.Close(),
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func closeStmt(stmt *sql.Stmt) error {
	if stmt == nil {
		return nil
	}
	return stmt.Close()
}

func withQueryTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, queryTimeout)
}

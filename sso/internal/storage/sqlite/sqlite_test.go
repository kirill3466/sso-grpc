package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"sso/internal/storage"
)

func testStorage(t *testing.T) *Storage {
	t.Helper()

	path := filepath.Join(t.TempDir(), "sso.db")
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			pass_hash BLOB NOT NULL,
			is_admin BOOLEAN NOT NULL DEFAULT FALSE
		);
		CREATE TABLE apps (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			secret TEXT NOT NULL UNIQUE
		);
	`)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	st, err := New(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })

	return st
}

func TestCreateAndGetUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	st := testStorage(t)

	id, err := st.CreateUser(ctx, "user@example.com", "hash", true)
	if err != nil {
		t.Fatal(err)
	}
	if id == 0 {
		t.Fatal("empty id")
	}

	user, err := st.GetUserByEmail(ctx, "user@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if user.ID != id || user.Email != "user@example.com" || !user.IsAdmin {
		t.Fatalf("user = %+v", user)
	}

	ok, err := st.IsAdmin(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("want admin")
	}
}

func TestCreateUserDuplicate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	st := testStorage(t)

	if _, err := st.CreateUser(ctx, "user@example.com", "hash", false); err != nil {
		t.Fatal(err)
	}

	_, err := st.CreateUser(ctx, "user@example.com", "hash", false)
	if !errors.Is(err, storage.ErrAlreadyExists) {
		t.Fatalf("got %v", err)
	}
}

func TestGetUserNotFound(t *testing.T) {
	t.Parallel()

	_, err := testStorage(t).GetUserByEmail(context.Background(), "missing@example.com")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestCreateAndGetApp(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	st := testStorage(t)

	id, err := st.CreateApp(ctx, "test", "secret")
	if err != nil {
		t.Fatal(err)
	}

	app, err := st.GetAppByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if app.Name != "test" || app.Secret != "secret" {
		t.Fatalf("app = %+v", app)
	}
}

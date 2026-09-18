package jwt_test

import (
	"testing"
	"time"

	"sso/internal/jwt"
	"sso/internal/models"
)

func TestNewTokenParseRoundTrip(t *testing.T) {
	t.Parallel()

	user := models.User{ID: 7, Email: "user@example.com"}
	app := models.App{ID: 1, Name: "test", Secret: "app-secret"}

	raw, err := jwt.NewToken(user, app, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	appID, err := jwt.PeekAppID(raw)
	if err != nil {
		t.Fatal(err)
	}
	if appID != app.ID {
		t.Fatalf("app_id = %d", appID)
	}

	claims, err := jwt.Parse(raw, app.Secret, app.Name)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UID != user.ID || claims.Email != user.Email || claims.AppID != app.ID {
		t.Fatalf("claims = %+v", claims)
	}
	if claims.Issuer != jwt.Issuer {
		t.Fatalf("iss = %q", claims.Issuer)
	}
}

func TestParseRejectsWrongSecret(t *testing.T) {
	t.Parallel()

	user := models.User{ID: 1, Email: "user@example.com"}
	app := models.App{ID: 1, Name: "test", Secret: "right"}

	raw, err := jwt.NewToken(user, app, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := jwt.Parse(raw, "wrong", app.Name); err == nil {
		t.Fatal("want invalid token")
	}
}

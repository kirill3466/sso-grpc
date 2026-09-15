package validate_test

import (
	"context"
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	ssov1 "github.com/kirill3466/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpvalidate "sso/internal/grpc/validate"
)

func TestLoginRequestRules(t *testing.T) {
	t.Parallel()

	err := protovalidate.Validate(&ssov1.LoginRequest{})
	if err == nil {
		t.Fatal("empty login must fail validation")
	}

	err = protovalidate.Validate(&ssov1.LoginRequest{
		Email:    "not-an-email",
		Password: "secret",
		AppId:    1,
	})
	if err == nil {
		t.Fatal("invalid email must fail validation")
	}

	err = protovalidate.Validate(&ssov1.LoginRequest{
		Email:    "user@example.com",
		Password: "secret",
		AppId:    1,
	})
	if err != nil {
		t.Fatalf("valid login must pass: %v", err)
	}
}

func TestRegisterRequestPasswordMinLen(t *testing.T) {
	t.Parallel()

	err := protovalidate.Validate(&ssov1.RegisterRequest{
		Email:    "user@example.com",
		Password: "short",
	})
	if err == nil {
		t.Fatal("short password must fail validation")
	}

	err = protovalidate.Validate(&ssov1.RegisterRequest{
		Email:    "user@example.com",
		Password: "long-enough",
	})
	if err != nil {
		t.Fatalf("valid register must pass: %v", err)
	}
}

func TestUnaryInterceptorRejectsInvalidRequest(t *testing.T) {
	t.Parallel()

	validator, err := protovalidate.New()
	if err != nil {
		t.Fatal(err)
	}

	interceptor := grpvalidate.UnaryServerInterceptor(validator)
	handler := func(context.Context, any) (any, error) {
		t.Fatal("handler must not run for invalid request")
		return nil, nil
	}

	_, err = interceptor(
		context.Background(),
		&ssov1.LoginRequest{Email: "bad", Password: "x"},
		&grpc.UnaryServerInfo{FullMethod: "/sso.v1.Auth/Login"},
		handler,
	)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("want InvalidArgument, got %v", err)
	}
	if !strings.Contains(err.Error(), "email") {
		t.Fatalf("error should mention email, got %v", err)
	}
}

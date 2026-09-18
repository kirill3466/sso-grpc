package authn

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc/metadata"
)

var ErrUnauthenticated = errors.New("unauthenticated")

func BearerFrom(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", ErrUnauthenticated
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", ErrUnauthenticated
	}

	scheme, token, found := strings.Cut(values[0], " ")
	if !found || !strings.EqualFold(scheme, "bearer") || token == "" {
		return "", ErrUnauthenticated
	}

	return token, nil
}

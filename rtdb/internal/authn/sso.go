package authn

import (
	"context"
	"sync"
	"time"

	ssov1 "github.com/kirill3466/protos/gen/go/sso"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type SSOValidator struct {
	client ssov1.AuthClient
}

func NewSSOValidator(client ssov1.AuthClient) *SSOValidator {
	return &SSOValidator{client: client}
}

func (v *SSOValidator) Validate(ctx context.Context, token string) (*Claims, error) {
	outCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs(
		"authorization", "Bearer "+token,
	))

	resp, err := v.client.Validate(outCtx, &ssov1.ValidateRequest{})
	if err != nil {
		if status.Code(err) == codes.Unauthenticated {
			return nil, ErrUnauthenticated
		}
		return nil, err
	}

	return &Claims{
		UID:     resp.GetUserId(),
		Email:   resp.GetEmail(),
		AppID:   resp.GetAppId(),
		IsAdmin: resp.GetIsAdmin(),
	}, nil
}

type cached struct {
	claims    *Claims
	expiresAt time.Time
}

type CachedValidator struct {
	inner Validator
	ttl   time.Duration
	mu    sync.Mutex
	cache map[string]cached
}

func NewCachedValidator(inner Validator, ttl time.Duration) *CachedValidator {
	if ttl <= 0 {
		ttl = 5 * time.Second
	}
	return &CachedValidator{
		inner: inner,
		ttl:   ttl,
		cache: make(map[string]cached),
	}
}

func (v *CachedValidator) Validate(ctx context.Context, token string) (*Claims, error) {
	now := time.Now()

	v.mu.Lock()
	item, ok := v.cache[token]
	v.mu.Unlock()
	if ok && now.Before(item.expiresAt) {
		return item.claims, nil
	}

	claims, err := v.inner.Validate(ctx, token)
	if err != nil {
		return nil, err
	}

	v.mu.Lock()
	v.cache[token] = cached{claims: claims, expiresAt: now.Add(v.ttl)}
	v.mu.Unlock()

	return claims, nil
}

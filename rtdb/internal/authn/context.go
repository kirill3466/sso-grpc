package authn

import "context"

type ctxKey struct{}

func WithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, ctxKey{}, claims)
}

func ClaimsFrom(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(ctxKey{}).(*Claims)
	return claims, ok && claims != nil
}

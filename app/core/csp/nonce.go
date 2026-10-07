package csp

import "context"

type nonceKey struct{}

func WithNonce(ctx context.Context, nonce string) context.Context {
	return context.WithValue(ctx, nonceKey{}, nonce)
}

func Nonce(ctx context.Context) string {
	nonce, _ := ctx.Value(nonceKey{}).(string)

	return nonce
}

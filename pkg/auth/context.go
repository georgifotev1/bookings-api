package auth

import "context"

type contextKey string

const claimsKey contextKey = "claims"
const customerClaimsKey contextKey = "customer_claims"

func ContextWithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

func ClaimsFromContext(ctx context.Context) *Claims {
	claims, _ := ctx.Value(claimsKey).(*Claims)
	return claims
}

func ContextWithCustomerClaims(ctx context.Context, claims *CustomerClaims) context.Context {
	return context.WithValue(ctx, customerClaimsKey, claims)
}

func CustomerClaimsFromContext(ctx context.Context) *CustomerClaims {
	claims, _ := ctx.Value(customerClaimsKey).(*CustomerClaims)
	return claims
}

package domain

import "context"

type tenantContextKey string

const tenantKey tenantContextKey = "tenant"

func ContextWithTenant(ctx context.Context, tenant *Tenant) context.Context {
	return context.WithValue(ctx, tenantKey, tenant)
}

func TenantFromContext(ctx context.Context) *Tenant {
	tenant, _ := ctx.Value(tenantKey).(*Tenant)
	return tenant
}

func TenantIDFromContext(ctx context.Context) string {
	tenant := TenantFromContext(ctx)
	if tenant == nil {
		return ""
	}
	return tenant.ID
}

package middleware

import (
	"net/http"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/georgifotev1/nuvelaone-api/internal/repository"
	"github.com/georgifotev1/nuvelaone-api/pkg/auth"
	"github.com/georgifotev1/nuvelaone-api/pkg/jsonutil"
	"github.com/go-chi/chi/v5"
)

func TenantResolver(tenantRepo repository.TenantRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slug := chi.URLParam(r, "slug")
			if slug == "" {
				jsonutil.WriteError(w, http.StatusBadRequest, "missing tenant slug")
				return
			}
			tenant, err := tenantRepo.GetBySlug(r.Context(), slug)
			if err != nil {
				jsonutil.WriteError(w, http.StatusNotFound, "tenant not found")
				return
			}
			ctx := domain.ContextWithTenant(r.Context(), tenant)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CustomerJWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				jsonutil.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			claims, err := auth.ParseCustomerToken(token, secret)
			if err != nil {
				jsonutil.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			tenant := domain.TenantFromContext(r.Context())
			if tenant == nil || claims.TenantID != tenant.ID {
				jsonutil.WriteError(w, http.StatusForbidden, "forbidden")
				return
			}
			ctx := auth.ContextWithCustomerClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

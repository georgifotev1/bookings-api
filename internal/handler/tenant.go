package handler

import (
	"net/http"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/georgifotev1/nuvelaone-api/internal/service"
	"github.com/georgifotev1/nuvelaone-api/pkg/auth"
	"github.com/georgifotev1/nuvelaone-api/pkg/jsonutil"
)

type TenantHandler struct {
	svc service.TenantService
}

func NewTenantHandler(svc service.TenantService) *TenantHandler {
	return &TenantHandler{svc: svc}
}

type TenantResponse struct {
	Tenant       *domain.Tenant        `json:"tenant"`
	WorkingHours []domain.WorkingHours `json:"working_hours"`
}

// GetMyTenant godoc
//
//	@Summary		Get current tenant
//	@Description	Get the authenticated user's tenant details
//	@Tags			tenants
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	TenantResponse
//	@Failure		401	{object}	jsonutil.ErrorResponse
//	@Failure		404	{object}	jsonutil.ErrorResponse
//	@Router			/tenants/me [get]
func (h *TenantHandler) GetMyTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.ClaimsFromContext(ctx)

	tenant, hours, err := h.svc.GetMyTenant(ctx, claims.TenantID)
	if err != nil {
		writeError(w, err)
		return
	}

	jsonutil.Write(w, http.StatusOK, jsonutil.NewResponse(TenantResponse{
		Tenant:       tenant,
		WorkingHours: hours,
	}))
}

// Update godoc
//
//	@Summary		Update tenant
//	@Description	Update tenant details (admin/owner only)
//	@Tags			tenants
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		domain.UpdateTenantRequest	true	"Update request"
//	@Success		200	{object}	domain.Tenant
//	@Failure		400	{object}	jsonutil.ErrorResponse
//	@Failure		401	{object}	jsonutil.ErrorResponse
//	@Failure		403	{object}	jsonutil.ErrorResponse
//	@Failure		404	{object}	jsonutil.ErrorResponse
//	@Router			/tenants [put]
func (h *TenantHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.ClaimsFromContext(ctx)

	if claims.Role != domain.RoleOwner && claims.Role != domain.RoleAdmin {
		jsonutil.WriteError(w, http.StatusForbidden, "forbidden")
		return
	}

	var req domain.UpdateTenantRequest
	if err := jsonutil.Read(r, &req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenant, err := h.svc.Update(ctx, claims.TenantID, req)
	if err != nil {
		writeError(w, err)
		return
	}

	jsonutil.Write(w, http.StatusOK, jsonutil.NewResponse(tenant))
}

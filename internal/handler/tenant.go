package handler

import (
	"net/http"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/georgifotev1/nuvelaone-api/internal/service"
	"github.com/georgifotev1/nuvelaone-api/pkg/auth"
	"github.com/georgifotev1/nuvelaone-api/pkg/jsonutil"
	"go.uber.org/zap"
)

type TenantHandler struct {
	svc    service.TenantService
	logger *zap.SugaredLogger
}

func NewTenantHandler(svc service.TenantService, logger *zap.SugaredLogger) *TenantHandler {
	return &TenantHandler{svc: svc, logger: logger}
}

type TenantResponse struct {
	Tenant       *domain.Tenant        `json:"tenant"`
	WorkingHours []domain.WorkingHours `json:"working_hours"`
}

func (h *TenantHandler) GetMyTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.ClaimsFromContext(ctx)

	tenant, hours, err := h.svc.GetMyTenant(ctx, claims.TenantID)
	if err != nil {
		handleError(w, err, h.logger)
		return
	}

	jsonutil.Write(w, http.StatusOK, jsonutil.NewResponse(TenantResponse{
		Tenant:       tenant,
		WorkingHours: hours,
	}))
}

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
		handleError(w, err, h.logger)
		return
	}

	jsonutil.Write(w, http.StatusOK, jsonutil.NewResponse(tenant))
}

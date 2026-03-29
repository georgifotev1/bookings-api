package handler

import (
	"net/http"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/georgifotev1/nuvelaone-api/internal/service"
	"github.com/georgifotev1/nuvelaone-api/pkg/auth"
	"github.com/georgifotev1/nuvelaone-api/pkg/jsonutil"
	"github.com/georgifotev1/nuvelaone-api/pkg/validator"
	"github.com/go-chi/chi/v5"
)

type InvitationHandler struct {
	svc service.InvitationService
}

func NewInvitationHandler(svc service.InvitationService) *InvitationHandler {
	return &InvitationHandler{svc: svc}
}

// Create godoc
//
//	@Summary		Create invitation
//	@Description	Create a new invitation to join the tenant (Pro tier only)
//	@Tags			invitations
//	@Accept			json
//	@Produce		json
//	@Param			request	body		domain.CreateInvitationRequest	true	"Invitation request"
//	@Success		201		{object}	domain.UserInvitation
//	@Failure		400		{object}	jsonutil.ErrorResponse
//	@Failure		403		{object}	jsonutil.ErrorResponse
//	@Security		BearerAuth
//	@Router			/invitations [post]
func (h *InvitationHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.ClaimsFromContext(ctx)

	var req domain.CreateInvitationRequest
	if err := jsonutil.Read(r, &req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	invitation, err := h.svc.CreateInvitation(ctx, claims.UserID, claims.TenantID, req)
	if err != nil {
		writeError(w, err)
		return
	}

	jsonutil.Write(w, http.StatusCreated, jsonutil.NewResponse(invitation))
}

// List godoc
//
//	@Summary		List invitations
//	@Description	List all pending invitations for the current tenant
//	@Tags			invitations
//	@Produce		json
//	@Success		200		{array}		domain.UserInvitation
//	@Failure		401		{object}	jsonutil.ErrorResponse
//	@Security		BearerAuth
//	@Router			/invitations [get]
func (h *InvitationHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.ClaimsFromContext(ctx)

	invitations, err := h.svc.ListInvitations(ctx, claims.TenantID)
	if err != nil {
		writeError(w, err)
		return
	}

	jsonutil.Write(w, http.StatusOK, jsonutil.NewResponse(invitations))
}

// Revoke godoc
//
//	@Summary		Revoke invitation
//	@Description	Revoke (delete) a pending invitation
//	@Tags			invitations
//	@Produce		json
//	@Param			id	path		string	true	"Invitation ID"
//	@Success		204		"No Content"
//	@Failure		401		{object}	jsonutil.ErrorResponse
//	@Failure		404		{object}	jsonutil.ErrorResponse
//	@Security		BearerAuth
//	@Router			/invitations/{id} [delete]
func (h *InvitationHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.ClaimsFromContext(ctx)
	id := chi.URLParam(r, "id")

	if err := h.svc.RevokeInvitation(ctx, id, claims.TenantID); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Resend godoc
//
//	@Summary		Resend invitation
//	@Description	Resend an existing invitation with a new token (Pro tier only)
//	@Tags			invitations
//	@Produce		json
//	@Param			id	path		string	true	"Invitation ID"
//	@Success		204		"No Content"
//	@Failure		400		{object}	jsonutil.ErrorResponse
//	@Failure		403		{object}	jsonutil.ErrorResponse
//	@Security		BearerAuth
//	@Router			/invitations/{id}/resend [post]
func (h *InvitationHandler) Resend(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.ClaimsFromContext(ctx)
	id := chi.URLParam(r, "id")

	if err := h.svc.ResendInvitation(ctx, id, claims.UserID); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Accept godoc
//
//	@Summary		Accept invitation
//	@Description	Accept an invitation and create an account
//	@Tags			invitations
//	@Accept			json
//	@Produce		json
//	@Param			request	body		domain.AcceptInvitationRequest	true	"Accept invitation request"
//	@Success		201		{object}	domain.User
//	@Failure		400		{object}	jsonutil.ErrorResponse
//	@Failure		410		{object}	jsonutil.ErrorResponse
//	@Router			/invitations/accept [post]
func (h *InvitationHandler) Accept(w http.ResponseWriter, r *http.Request) {
	var req domain.AcceptInvitationRequest
	if err := jsonutil.Read(r, &req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.svc.AcceptInvitation(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}

	jsonutil.Write(w, http.StatusCreated, jsonutil.NewResponse(user))
}

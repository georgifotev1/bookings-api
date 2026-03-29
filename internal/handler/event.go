package handler

import (
	"net/http"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	svc "github.com/georgifotev1/nuvelaone-api/internal/service"
	"github.com/georgifotev1/nuvelaone-api/pkg/auth"
	"github.com/georgifotev1/nuvelaone-api/pkg/jsonutil"
	"github.com/georgifotev1/nuvelaone-api/pkg/validator"
	"github.com/go-chi/chi/v5"
)

type EventHandler struct {
	svc svc.EventService
}

func NewEventHandler(svc svc.EventService) *EventHandler {
	return &EventHandler{svc: svc}
}

// Create godoc
//
//	@Summary		Create event
//	@Description	Create a new event
//	@Tags			events
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		domain.EventRequest	true	"Create request"
//	@Success		201	{object}	domain.Event
//	@Failure		400	{object}	jsonutil.ErrorResponse
//	@Failure		401	{object}	jsonutil.ErrorResponse
//	@Failure		409	{object}	jsonutil.ErrorResponse
//	@Router			/events [post]
func (h *EventHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.ClaimsFromContext(ctx)

	var req domain.EventRequest
	if err := jsonutil.Read(r, &req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	event, err := h.svc.Create(ctx, claims.TenantID, req)
	if err != nil {
		writeError(w, err)
		return
	}

	jsonutil.Write(w, http.StatusCreated, jsonutil.NewResponse(event))
}

// Update godoc
//
//	@Summary		Update event
//	@Description	Update an existing event
//	@Tags			events
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Event ID"
//	@Param			request	body		domain.EventUpdateRequest	true	"Update request"
//	@Success		200	{object}	domain.Event
//	@Failure		400	{object}	jsonutil.ErrorResponse
//	@Failure		401	{object}	jsonutil.ErrorResponse
//	@Failure		409	{object}	jsonutil.ErrorResponse
//	@Router			/events/{id} [put]
func (h *EventHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := auth.ClaimsFromContext(ctx)

	id := chi.URLParam(r, "id")
	if id == "" {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req domain.EventUpdateRequest
	if err := jsonutil.Read(r, &req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	event, err := h.svc.Update(ctx, claims.TenantID, id, req)
	if err != nil {
		writeError(w, err)
		return
	}

	jsonutil.Write(w, http.StatusOK, jsonutil.NewResponse(event))
}

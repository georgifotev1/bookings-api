package handler

import (
	"net/http"
	"time"

	"github.com/georgifotev1/nuvelaone-api/internal/domain"
	"github.com/georgifotev1/nuvelaone-api/internal/service"
	"github.com/georgifotev1/nuvelaone-api/pkg/jsonutil"
	"github.com/georgifotev1/nuvelaone-api/pkg/validator"
	"github.com/go-chi/chi/v5"
)

type CustomerAuthHandler struct {
	svc             service.CustomerAuthService
	refreshTokenTTL time.Duration
}

func NewCustomerAuthHandler(svc service.CustomerAuthService, refreshTokenTTL time.Duration) *CustomerAuthHandler {
	return &CustomerAuthHandler{
		svc: svc, refreshTokenTTL: refreshTokenTTL,
	}
}

func (h *CustomerAuthHandler) Routes(r chi.Router) {
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.Refresh)
	r.Post("/logout", h.Logout)
}

// Register godoc
//
//	@Summary		Register customer
//	@Description	Register a new customer for the tenant
//	@Tags			public
//	@Accept			json
//	@Produce		json
//	@Param			slug		path		string							true	"Tenant slug"
//	@Param			request	body		domain.CustomerRegisterRequest	true	"Register request"
//	@Success		201	{object}	jsonutil.Response[map[string]string]
//	@Router			/p/{slug}/auth/register [post]
func (h *CustomerAuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	tenant := domain.TenantFromContext(r.Context())
	if tenant == nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "tenant not found")
		return
	}

	var req domain.CustomerRegisterRequest
	if err := jsonutil.Read(r, &req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	customer, err := h.svc.Register(r.Context(), tenant.ID, req)
	if err != nil {
		writeError(w, err)
		return
	}
	jsonutil.Write(w, http.StatusCreated, jsonutil.NewResponse(map[string]string{
		"id":    customer.ID,
		"name":  customer.Name,
		"email": *customer.Email,
	}))
}

// Login godoc
//
//	@Summary		Login customer
//	@Description	Login a customer for the tenant
//	@Tags			public
//	@Accept			json
//	@Produce		json
//	@Param			slug		path		string						true	"Tenant slug"
//	@Param			request	body		domain.CustomerLoginRequest	true	"Login request"
//	@Success		200	{object}	jsonutil.Response[map[string]string]
//	@Router			/p/{slug}/auth/login [post]
func (h *CustomerAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	tenant := domain.TenantFromContext(r.Context())
	if tenant == nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "tenant not found")
		return
	}

	var req domain.CustomerLoginRequest
	if err := jsonutil.Read(r, &req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	pair, err := h.svc.Login(r.Context(), tenant.ID, req)
	if err != nil {
		writeError(w, err)
		return
	}

	h.setRefreshCookie(w, pair.RefreshToken, int(h.refreshTokenTTL.Seconds()))
	jsonutil.Write(w, http.StatusOK, jsonutil.NewResponse(map[string]string{
		"access_token": pair.AccessToken,
	}))
}

// Refresh godoc
//
//	@Summary		Refresh customer token
//	@Description	Refresh customer access token
//	@Tags			public
//	@Produce		json
//	@Param			slug	path		string	true	"Tenant slug"
//	@Success		200	{object}	jsonutil.Response[map[string]string]
//	@Router			/p/{slug}/auth/refresh [post]
func (h *CustomerAuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("customer_refresh_token")
	if err != nil {
		jsonutil.WriteError(w, http.StatusUnauthorized, "missing refresh token")
		return
	}

	pair, err := h.svc.Refresh(r.Context(), cookie.Value)
	if err != nil {
		writeError(w, err)
		return
	}

	h.setRefreshCookie(w, pair.RefreshToken, int(h.refreshTokenTTL.Seconds()))

	jsonutil.Write(w, http.StatusOK, jsonutil.NewResponse(map[string]string{
		"access_token": pair.AccessToken,
	}))
}

// Logout godoc
//
//	@Summary		Logout customer
//	@Description	Logout customer and revoke refresh token
//	@Tags			public
//	@Produce		json
//	@Param			slug	path		string	true	"Tenant slug"
//	@Success		204
//	@Router			/p/{slug}/auth/logout [post]
func (h *CustomerAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("customer_refresh_token")
	if err != nil {
		jsonutil.WriteError(w, http.StatusBadRequest, "missing refresh token")
		return
	}

	if err := h.svc.Logout(r.Context(), cookie.Value); err != nil {
		writeError(w, err)
		return
	}

	h.setRefreshCookie(w, "", -1)

	w.WriteHeader(http.StatusNoContent)
}

func (h *CustomerAuthHandler) setRefreshCookie(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     "customer_refresh_token",
		Value:    value,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   maxAge,
	})
}

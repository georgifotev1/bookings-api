package handler

import (
	"errors"
	"log/slog"
	"net/http"

	apperr "github.com/georgifotev1/nuvelaone-api/internal/errors"
	"github.com/georgifotev1/nuvelaone-api/pkg/jsonutil"
)

func writeError(w http.ResponseWriter, err error) {
	var appErr *apperr.AppError

	if !errors.As(err, &appErr) {
		slog.Error("unknown error", "error", err)
		jsonutil.WriteError(w, http.StatusInternalServerError, "an unexpected error occurred")
		return
	}

	status := http.StatusInternalServerError
	switch {
	case errors.Is(appErr.Err, apperr.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(appErr.Err, apperr.ErrConflict):
		status = http.StatusConflict
	case errors.Is(appErr.Err, apperr.ErrValidation):
		status = http.StatusBadRequest
	case errors.Is(appErr.Err, apperr.ErrUnauthorized):
		status = http.StatusUnauthorized
	case errors.Is(appErr.Err, apperr.ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(appErr.Err, apperr.ErrInternal):
		slog.Error("internal error", "error", appErr.Cause())
		status = http.StatusInternalServerError
	}

	jsonutil.Write(w, status, jsonutil.ErrorResponse{Error: appErr.Message})
}

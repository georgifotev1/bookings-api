package handler

import (
	"errors"
	"net/http"

	"github.com/georgifotev1/nuvelaone-api/internal/service"
	"github.com/georgifotev1/nuvelaone-api/pkg/jsonutil"
	"go.uber.org/zap"
)

func handleError(w http.ResponseWriter, err error, logger *zap.SugaredLogger) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		jsonutil.WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrConflict):
		jsonutil.WriteError(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrInvalidCredentials):
		jsonutil.WriteError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, service.ErrInvalidToken):
		jsonutil.WriteError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, service.ErrForbidden):
		jsonutil.WriteError(w, http.StatusForbidden, err.Error())
	default:
		logger.Errorw("internal server error", "error", err)
		jsonutil.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}

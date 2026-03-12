package handlers

import (
	"context"
	"errors"
	"net/http"

	"robot-hand-server/internal/httpapi/httputil"
	"robot-hand-server/internal/httpapi/middleware"
	"robot-hand-server/internal/service"
)

type ProfileUseCase interface {
	GetProfile(ctx context.Context, userUUID string) (service.ProfileResponse, error)
	UpdateProfile(ctx context.Context, userUUID string, req service.UpdateProfileRequest) (service.ProfileResponse, error)
	DeleteProfile(ctx context.Context, userUUID string) error
}

type ProfileHandler struct {
	profileService ProfileUseCase
}

func NewProfileHandler(profileService ProfileUseCase) *ProfileHandler {
	return &ProfileHandler{profileService: profileService}
}

func (h *ProfileHandler) Profile(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := middleware.UserUUIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	switch r.Method {
	case http.MethodGet:
		resp, err := h.profileService.GetProfile(r.Context(), userUUID)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrNotFound):
				httputil.WriteError(w, http.StatusNotFound, "user not found")
			default:
				httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			}
			return
		}

		httputil.WriteJSON(w, http.StatusOK, resp)
	case http.MethodPut:
		var req struct {
			Email           string `json:"email"`
			Name            string `json:"name"`
			Surname         string `json:"surname"`
			CurrentPassword string `json:"current_password"`
			NewPassword     string `json:"new_password"`
		}

		if err := httputil.DecodeJSONBody(r, &req); err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		resp, err := h.profileService.UpdateProfile(r.Context(), userUUID, service.UpdateProfileRequest{
			Email:           req.Email,
			Name:            req.Name,
			Surname:         req.Surname,
			CurrentPassword: req.CurrentPassword,
			NewPassword:     req.NewPassword,
		})
		if err != nil {
			switch {
			case errors.Is(err, service.ErrInvalidInput):
				httputil.WriteError(w, http.StatusBadRequest, "invalid data")
			case errors.Is(err, service.ErrConflict):
				httputil.WriteError(w, http.StatusConflict, "email already exists")
			case errors.Is(err, service.ErrUnauthorized):
				httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
			case errors.Is(err, service.ErrNotFound):
				httputil.WriteError(w, http.StatusNotFound, "user not found")
			default:
				httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			}
			return
		}

		httputil.WriteJSON(w, http.StatusOK, resp)

	case http.MethodDelete:
		err := h.profileService.DeleteProfile(r.Context(), userUUID)
		if err != nil {
			if errors.Is(err, service.ErrNotFound) {
				httputil.WriteError(w, http.StatusNotFound, "user not found")
			} else {
				httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			}
			return
		}

		httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "account deleted"})
	
	default:
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

package handlers

import (
	"context"
	"errors"
	"net/http"

	"robot-hand-server/internal/httpapi/httputil"
	"robot-hand-server/internal/httpapi/middleware"
	"robot-hand-server/internal/service"
)

type AuthUseCase interface {
	Register(ctx context.Context, req service.RegisterRequest) (service.AuthResponse, error)
	Login(ctx context.Context, req service.LoginRequest) (service.AuthResponse, error)
	ForgotPassword(ctx context.Context, req service.ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, req service.ResetPasswordRequest) error
	Logout(ctx context.Context, token string) error
}

type AuthHandler struct {
	authService AuthUseCase
}

func NewAuthHandler(authService AuthUseCase) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Surname  string `json:"surname"`
	}

	if err := httputil.DecodeJSONBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.authService.Register(r.Context(), service.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
		Surname:  req.Surname,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput), errors.Is(err, service.ErrConflict):
			httputil.WriteError(w, http.StatusBadRequest, "invalid data")
		default:
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := httputil.DecodeJSONBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.authService.Login(r.Context(), service.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUnauthorized):
			httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		default:
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Email string `json:"email"`
	}

	if err := httputil.DecodeJSONBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.authService.ForgotPassword(r.Context(), service.ForgotPasswordRequest{Email: req.Email})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			httputil.WriteError(w, http.StatusNotFound, "email not found")
		default:
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "success"})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	token, ok := middleware.TokenFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.authService.Logout(r.Context(), token); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidToken):
			httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		default:
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "success"})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Email       string `json:"email"`
		Code        string `json:"code"`
		NewPassword string `json:"new_password"`
	}

	if err := httputil.DecodeJSONBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.authService.ResetPassword(r.Context(), service.ResetPasswordRequest{
		Email:       req.Email,
		Code:        req.Code,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput), errors.Is(err, service.ErrInvalidCode):
			httputil.WriteError(w, http.StatusBadRequest, "invalid data")
		case errors.Is(err, service.ErrNotFound):
			httputil.WriteError(w, http.StatusNotFound, "email not found")
		default:
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "password changed"})
}

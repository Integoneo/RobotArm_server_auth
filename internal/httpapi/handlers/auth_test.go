package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"robot-hand-server/internal/httpapi/middleware"
	"robot-hand-server/internal/service"
)

type authServiceStub struct {
	logoutToken string
}

func (s *authServiceStub) Register(context.Context, service.RegisterRequest) (service.AuthResponse, error) {
	return service.AuthResponse{}, nil
}

func (s *authServiceStub) Login(context.Context, service.LoginRequest) (service.AuthResponse, error) {
	return service.AuthResponse{}, nil
}

func (s *authServiceStub) ForgotPassword(context.Context, service.ForgotPasswordRequest) error {
	return nil
}

func (s *authServiceStub) ResetPassword(context.Context, service.ResetPasswordRequest) error {
	return nil
}

func (s *authServiceStub) Logout(_ context.Context, token string) error {
	s.logoutToken = token
	return nil
}

type tokenAuthenticatorStub struct{}

func (tokenAuthenticatorStub) AuthenticateToken(context.Context, string) (string, error) {
	return "user-uuid", nil
}

func TestAuthHandlerLogoutAcceptsEmptyBody(t *testing.T) {
	authService := &authServiceStub{}
	authHandler := NewAuthHandler(authService)
	authMiddleware := middleware.NewAuthMiddleware(tokenAuthenticatorStub{})
	handler := authMiddleware.RequireBearer(http.HandlerFunc(authHandler.Logout))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if authService.logoutToken != "test-token" {
		t.Fatalf("expected logout token to be passed to service, got %q", authService.logoutToken)
	}
}

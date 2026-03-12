package middleware

import (
	"context"
	"net/http"
	"strings"

	"robot-hand-server/internal/httpapi/httputil"
)

type contextKey string

const (
	contextUserUUIDKey contextKey = "user_uuid"
	contextTokenKey    contextKey = "token"
)

type TokenAuthenticator interface {
	AuthenticateToken(ctx context.Context, token string) (string, error)
}

type AuthMiddleware struct {
	authenticator TokenAuthenticator
}

func NewAuthMiddleware(authenticator TokenAuthenticator) *AuthMiddleware {
	return &AuthMiddleware{authenticator: authenticator}
}

func (m *AuthMiddleware) RequireBearer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := extractBearerToken(r.Header.Get("Authorization"))
		if !ok {
			httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		userUUID, err := m.authenticator.AuthenticateToken(r.Context(), token)
		if err != nil {
			httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		ctx := context.WithValue(r.Context(), contextUserUUIDKey, userUUID)
		ctx = context.WithValue(ctx, contextTokenKey, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserUUIDFromContext(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(contextUserUUIDKey).(string)
	return value, ok
}

func TokenFromContext(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(contextTokenKey).(string)
	return value, ok
}

func extractBearerToken(header string) (string, bool) {
	header = strings.TrimSpace(header)
	if header == "" {
		return "", false
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", false
	}

	return token, true
}

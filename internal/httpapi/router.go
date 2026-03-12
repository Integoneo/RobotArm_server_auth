package httpapi

import (
	"net/http"

	"robot-hand-server/internal/httpapi/handlers"
	"robot-hand-server/internal/httpapi/middleware"
)

func NewRouter(
	authHandler *handlers.AuthHandler,
	supportHandler *handlers.SupportHandler,
	profileHandler *handlers.ProfileHandler,
	syncHandler *handlers.SyncHandler, // <-- ДОБАВИТЬ ЭТУ СТРОКУ
	authMiddleware *middleware.AuthMiddleware,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/auth/register", authHandler.Register)
	mux.HandleFunc("/api/auth/login", authHandler.Login)
	mux.HandleFunc("/api/auth/forgot-password", authHandler.ForgotPassword)
	mux.HandleFunc("/api/auth/reset-password", authHandler.ResetPassword)
	mux.Handle("/api/auth/logout", authMiddleware.RequireBearer(http.HandlerFunc(authHandler.Logout)))
	mux.Handle("/api/support/send_email", authMiddleware.RequireBearer(http.HandlerFunc(supportHandler.SendEmail)))
	mux.Handle("/api/user/profile", authMiddleware.RequireBearer(http.HandlerFunc(profileHandler.Profile)))
	mux.Handle("/api/sync", authMiddleware.RequireBearer(http.HandlerFunc(syncHandler.Sync)))
	return mux
}

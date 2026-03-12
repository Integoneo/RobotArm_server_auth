package handlers
import (
	"context"
	"net/http"
	"log"
	"robot-hand-server/internal/httpapi/httputil"
	"robot-hand-server/internal/httpapi/middleware"
	"robot-hand-server/internal/service"
)

type SyncUseCase interface {
	Sync(ctx context.Context, userUUID string, req service.SyncRequest) (service.SyncResponse, error)
}

type SyncHandler struct {
	syncService SyncUseCase
}

func NewSyncHandler(syncService SyncUseCase) *SyncHandler {
	return &SyncHandler{syncService: syncService}
}

func (h *SyncHandler) Sync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userUUID, ok := middleware.UserUUIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req service.SyncRequest
	if err := httputil.DecodeJSONBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.syncService.Sync(r.Context(), userUUID, req)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	log.Printf("[SYNC] Юзер: %s | Прислал пресетов: %d | Отдаем пресетов: %d\n", userUUID, len(req.LocalPresets), len(resp.MergedPresets))


	httputil.WriteJSON(w, http.StatusOK, resp)
}
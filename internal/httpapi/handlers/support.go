package handlers

import (
	"context"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	"robot-hand-server/internal/httpapi/httputil"
	"robot-hand-server/internal/httpapi/middleware"
	"robot-hand-server/internal/service"
)

const (
	maxMultipartMemory = 12 << 20 // 12MB
	maxImageReadBytes  = 6 << 20  // 6MB upper bound at handler level
)

type SupportUseCase interface {
	CreateTicket(ctx context.Context, req service.CreateSupportTicketRequest) error
}

type SupportHandler struct {
	supportService SupportUseCase
}

func NewSupportHandler(supportService SupportUseCase) *SupportHandler {
	return &SupportHandler{supportService: supportService}
}

func (h *SupportHandler) SendEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	contentType := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || !strings.EqualFold(mediaType, "multipart/form-data") {
		httputil.WriteError(w, http.StatusBadRequest, "invalid form")
		return
	}

	if err := r.ParseMultipartForm(maxMultipartMemory); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid form")
		return
	}

	userUUID, ok := middleware.UserUUIDFromContext(r.Context())
	if !ok {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	req := service.CreateSupportTicketRequest{
		AuthUserUUID: userUUID,
		UUID:         r.FormValue("uuid"),
		Email:        r.FormValue("email"),
		Header:       r.FormValue("header"),
		Text:         r.FormValue("text"),
	}

	file, fileHeader, err := r.FormFile("image")
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			httputil.WriteError(w, http.StatusBadRequest, "invalid form")
			return
		}
	} else {
		defer file.Close()

		content, readErr := io.ReadAll(io.LimitReader(file, maxImageReadBytes))
		if readErr != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid image")
			return
		}

		req.Image = &service.SupportImage{
			Filename: fileHeader.Filename,
			Content:  content,
		}
	}

	err = h.supportService.CreateTicket(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput),
			errors.Is(err, service.ErrUserMismatch),
			errors.Is(err, service.ErrUnsupportedImage),
			errors.Is(err, service.ErrImageTooLarge):
			httputil.WriteError(w, http.StatusBadRequest, "invalid form")
		default:
			httputil.WriteError(w, http.StatusInternalServerError, "support service error")
		}
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "ticket created"})
}

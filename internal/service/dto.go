package service
import "robot-hand-server/internal/domain"

type RegisterRequest struct {
	Email    string
	Password string
	Name     string
	Surname  string
}

type LoginRequest struct {
	Email    string
	Password string
}

type AuthResponse struct {
  Token   string `json:"token"`
  UUID    string `json:"uuid"`
  Name    string `json:"name"`
  Surname string `json:"surname"`
}

type ForgotPasswordRequest struct {
	Email string
}

type ResetPasswordRequest struct {
	Email       string
	Code        string
	NewPassword string
}

type ProfileResponse struct {
	UUID    string `json:"uuid"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
}

type UpdateProfileRequest struct {
	Email           string
	Name            string
	Surname         string
	CurrentPassword string
	NewPassword     string
}

type CreateSupportTicketRequest struct {
	AuthUserUUID string
	UUID         string
	Email        string
	Header       string
	Text         string
	Image        *SupportImage
}

type SupportImage struct {
	Filename string
	Content  []byte
}


type SyncRequest struct {
	LocalPresets []domain.Preset      `json:"localPresets"`
	LocalHistory []domain.HistoryItem `json:"localHistory"`
	DeletedPresetIDs []string         `json:"deletedPresetIDs"` // <-- ДОБАВИТЬ ЭТО
}

type SyncResponse struct {
	MergedPresets []domain.Preset      `json:"mergedPresets"`
	MergedHistory []domain.HistoryItem `json:"mergedHistory"`
}

package repository

import (
	"context"
	"time"

	"robot-hand-server/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByUUID(ctx context.Context, uuid string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	UpdatePasswordHashByEmail(ctx context.Context, email, passwordHash string) error
	Delete(ctx context.Context, uuid string) error
}

type RevokedTokenRepository interface {
	Revoke(ctx context.Context, token string, expiresAt time.Time) error
	IsRevoked(ctx context.Context, token string) (bool, error)
}

type SupportTicketRepository interface {
	Create(ctx context.Context, ticket *domain.SupportTicket) error
}

type PasswordResetCodeRepository interface {
	Save(ctx context.Context, email, codeHash string, expiresAt time.Time) error
	Consume(ctx context.Context, email, codeHash string, now time.Time) (bool, error)
}

type SyncRepository interface {
	GetSyncState(ctx context.Context, userUUID string) ([]domain.Preset, []domain.HistoryItem, error)
	SaveSyncState(ctx context.Context, userUUID string, presets []domain.Preset, history []domain.HistoryItem) error
	DeleteSyncState(ctx context.Context, userUUID string) error
}

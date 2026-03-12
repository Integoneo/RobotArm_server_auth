package repository

import (
	"context"
	"database/sql"

	"robot-hand-server/internal/domain"
)

type PostgresSupportTicketRepository struct {
	db *sql.DB
}

func NewPostgresSupportTicketRepository(db *sql.DB) *PostgresSupportTicketRepository {
	return &PostgresSupportTicketRepository{db: db}
}

func (r *PostgresSupportTicketRepository) Create(ctx context.Context, ticket *domain.SupportTicket) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO support_tickets (id, user_uuid, email, header, text, image_path, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, ticket.ID, ticket.UserUUID, ticket.Email, ticket.Header, ticket.Text, ticket.ImagePath, ticket.CreatedAt)
	return err
}

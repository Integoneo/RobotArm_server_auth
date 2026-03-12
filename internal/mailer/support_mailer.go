package mailer

import (
	"context"
	"log"

	"robot-hand-server/internal/domain"
)

type SupportMailer interface {
	SendSupportTicket(ctx context.Context, ticket domain.SupportTicket) error
}

type LogSupportMailer struct{}

func NewLogSupportMailer() *LogSupportMailer {
	return &LogSupportMailer{}
}

func (m *LogSupportMailer) SendSupportTicket(_ context.Context, ticket domain.SupportTicket) error {
	log.Printf("support ticket: id=%s user_uuid=%s email=%s header=%q image=%q", ticket.ID, ticket.UserUUID, ticket.Email, ticket.Header, ticket.ImagePath)
	return nil
}

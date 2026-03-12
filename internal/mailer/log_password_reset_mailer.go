package mailer

import (
	"context"
	"log"
)

type LogPasswordResetMailer struct{}

func NewLogPasswordResetMailer() *LogPasswordResetMailer {
	return &LogPasswordResetMailer{}
}

func (m *LogPasswordResetMailer) SendPasswordResetCode(_ context.Context, email, code string) error {
	log.Printf("password reset code generated for %s: %s", email, code)
	return nil
}

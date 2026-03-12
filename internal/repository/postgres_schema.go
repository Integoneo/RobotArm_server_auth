package repository

import (
	"context"
	"database/sql"
)

func EnsureSchema(ctx context.Context, db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			uuid UUID PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			name TEXT NOT NULL,
			surname TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS revoked_tokens (
			token TEXT PRIMARY KEY,
			expires_at TIMESTAMPTZ NOT NULL,
			revoked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS support_tickets (
			id UUID PRIMARY KEY,
			user_uuid UUID NOT NULL REFERENCES users(uuid),
			email TEXT NOT NULL,
			header TEXT NOT NULL,
			text TEXT NOT NULL,
			image_path TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS password_reset_codes (
			id BIGSERIAL PRIMARY KEY,
			email TEXT NOT NULL,
			code_hash TEXT NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			used_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`CREATE INDEX IF NOT EXISTS idx_password_reset_codes_email_created_at
		 ON password_reset_codes(email, created_at DESC);`,
	}

	for _, query := range queries {
		if _, err := db.ExecContext(ctx, query); err != nil {
			return err
		}
	}

	return nil
}

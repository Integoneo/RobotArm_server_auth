package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type PostgresRevokedTokenRepository struct {
	db *sql.DB
}

func NewPostgresRevokedTokenRepository(db *sql.DB) *PostgresRevokedTokenRepository {
	return &PostgresRevokedTokenRepository{db: db}
}

func (r *PostgresRevokedTokenRepository) Revoke(ctx context.Context, token string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO revoked_tokens (token, expires_at)
		VALUES ($1, $2)
		ON CONFLICT (token) DO NOTHING
	`, token, expiresAt)
	return err
}

func (r *PostgresRevokedTokenRepository) IsRevoked(ctx context.Context, token string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM revoked_tokens
			WHERE token = $1
		)
	`, token).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return exists, nil
}

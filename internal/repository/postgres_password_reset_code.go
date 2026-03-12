package repository

import (
	"context"
	"database/sql"
	"time"
)

type PostgresPasswordResetCodeRepository struct {
	db *sql.DB
}

func NewPostgresPasswordResetCodeRepository(db *sql.DB) *PostgresPasswordResetCodeRepository {
	return &PostgresPasswordResetCodeRepository{db: db}
}

func (r *PostgresPasswordResetCodeRepository) Save(ctx context.Context, email, codeHash string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO password_reset_codes (email, code_hash, expires_at)
		VALUES ($1, $2, $3)
	`, email, codeHash, expiresAt)
	return err
}

func (r *PostgresPasswordResetCodeRepository) Consume(ctx context.Context, email, codeHash string, now time.Time) (bool, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()

	var id int64
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM password_reset_codes
		WHERE email = $1
		  AND code_hash = $2
		  AND used_at IS NULL
		  AND expires_at > $3
		ORDER BY created_at DESC
		LIMIT 1
		FOR UPDATE
	`, email, codeHash, now).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE password_reset_codes
		SET used_at = $1
		WHERE id = $2
	`, now, id); err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}

	return true, nil
}

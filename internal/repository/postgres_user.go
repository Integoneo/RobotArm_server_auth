package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"robot-hand-server/internal/domain"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	normalizedEmail := strings.ToLower(strings.TrimSpace(user.Email))

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (uuid, email, password_hash, name, surname)
		VALUES ($1, $2, $3, $4, $5)
	`, user.UUID, normalizedEmail, user.PasswordHash, user.Name, user.Surname)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrUserExists
		}
		return err
	}

	return nil
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	var user domain.User
	err := r.db.QueryRowContext(ctx, `
		SELECT uuid, email, password_hash, name, surname
		FROM users
		WHERE email = $1
	`, normalizedEmail).Scan(&user.UUID, &user.Email, &user.PasswordHash, &user.Name, &user.Surname)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *PostgresUserRepository) FindByUUID(ctx context.Context, uuid string) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRowContext(ctx, `
		SELECT uuid, email, password_hash, name, surname
		FROM users
		WHERE uuid = $1
	`, uuid).Scan(&user.UUID, &user.Email, &user.PasswordHash, &user.Name, &user.Surname)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *domain.User) error {
	normalizedEmail := strings.ToLower(strings.TrimSpace(user.Email))

	result, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET email = $1, password_hash = $2, name = $3, surname = $4
		WHERE uuid = $5
	`, normalizedEmail, user.PasswordHash, user.Name, user.Surname, user.UUID)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrUserExists
		}
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *PostgresUserRepository) UpdatePasswordHashByEmail(ctx context.Context, email, passwordHash string) error {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	result, err := r.db.ExecContext(ctx, `
		UPDATE users SET password_hash = $1 WHERE email = $2
	`, passwordHash, normalizedEmail)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}


func (r *PostgresUserRepository) Delete(ctx context.Context, uuid string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE uuid = $1`, uuid)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

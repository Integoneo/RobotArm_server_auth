package repository

import (
	"context"
	"strings"
	"sync"

	"robot-hand-server/internal/domain"
)

type InMemoryUserRepository struct {
	mu         sync.RWMutex
	byUUID     map[string]*domain.User
	uuidByMail map[string]string
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		byUUID:     make(map[string]*domain.User),
		uuidByMail: make(map[string]string),
	}
}

func (r *InMemoryUserRepository) Create(_ context.Context, user *domain.User) error {
	normalizedEmail := strings.ToLower(strings.TrimSpace(user.Email))

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.uuidByMail[normalizedEmail]; exists {
		return ErrUserExists
	}

	copyUser := *user
	copyUser.Email = normalizedEmail
	r.byUUID[copyUser.UUID] = &copyUser
	r.uuidByMail[normalizedEmail] = copyUser.UUID

	return nil
}

func (r *InMemoryUserRepository) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.uuidByMail[normalizedEmail]
	if !exists {
		return nil, ErrUserNotFound
	}

	user, exists := r.byUUID[id]
	if !exists {
		return nil, ErrUserNotFound
	}

	copyUser := *user
	return &copyUser, nil
}

func (r *InMemoryUserRepository) FindByUUID(_ context.Context, uuid string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.byUUID[uuid]
	if !exists {
		return nil, ErrUserNotFound
	}

	copyUser := *user
	return &copyUser, nil
}

func (r *InMemoryUserRepository) Update(_ context.Context, user *domain.User) error {
	normalizedEmail := strings.ToLower(strings.TrimSpace(user.Email))

	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.byUUID[user.UUID]
	if !exists {
		return ErrUserNotFound
	}

	if ownerUUID, exists := r.uuidByMail[normalizedEmail]; exists && ownerUUID != user.UUID {
		return ErrUserExists
	}

	delete(r.uuidByMail, strings.ToLower(strings.TrimSpace(existing.Email)))

	copyUser := *user
	copyUser.Email = normalizedEmail
	r.byUUID[user.UUID] = &copyUser
	r.uuidByMail[normalizedEmail] = user.UUID
	return nil
}

func (r *InMemoryUserRepository) UpdatePasswordHashByEmail(_ context.Context, email, passwordHash string) error {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	r.mu.Lock()
	defer r.mu.Unlock()

	id, exists := r.uuidByMail[normalizedEmail]
	if !exists {
		return ErrUserNotFound
	}

	user, exists := r.byUUID[id]
	if !exists {
		return ErrUserNotFound
	}

	user.PasswordHash = passwordHash
	return nil
}


func (r *InMemoryUserRepository) Delete(_ context.Context, uuid string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.byUUID[uuid]
	if !exists {
		return ErrUserNotFound
	}

	// Удаляем из обоих мапов, чтобы email снова стал свободен
	delete(r.uuidByMail, strings.ToLower(strings.TrimSpace(user.Email)))
	delete(r.byUUID, uuid)
	return nil
}

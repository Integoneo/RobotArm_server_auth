package repository

import (
	"context"
	"sync"
	"time"
)

type InMemoryPasswordResetCodeRepository struct {
	mu    sync.Mutex
	items []passwordResetItem
}

type passwordResetItem struct {
	email     string
	codeHash  string
	expiresAt time.Time
	used      bool
}

func NewInMemoryPasswordResetCodeRepository() *InMemoryPasswordResetCodeRepository {
	return &InMemoryPasswordResetCodeRepository{items: make([]passwordResetItem, 0)}
}

func (r *InMemoryPasswordResetCodeRepository) Save(_ context.Context, email, codeHash string, expiresAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.items = append(r.items, passwordResetItem{
		email:     email,
		codeHash:  codeHash,
		expiresAt: expiresAt,
	})
	return nil
}

func (r *InMemoryPasswordResetCodeRepository) Consume(_ context.Context, email, codeHash string, now time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := len(r.items) - 1; i >= 0; i-- {
		item := &r.items[i]
		if item.email != email || item.used || item.expiresAt.Before(now) {
			continue
		}
		if item.codeHash == codeHash {
			item.used = true
			return true, nil
		}
	}

	return false, nil
}

package repository

import (
	"context"
	"sync"
	"time"
)

type InMemoryRevokedTokenRepository struct {
	mu     sync.RWMutex
	tokens map[string]struct{}
}

func NewInMemoryRevokedTokenRepository() *InMemoryRevokedTokenRepository {
	return &InMemoryRevokedTokenRepository{
		tokens: make(map[string]struct{}),
	}
}

func (r *InMemoryRevokedTokenRepository) Revoke(_ context.Context, token string, _ time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokens[token] = struct{}{}
	return nil
}

func (r *InMemoryRevokedTokenRepository) IsRevoked(_ context.Context, token string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.tokens[token]
	return exists, nil
}

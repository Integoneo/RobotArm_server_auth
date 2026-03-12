package repository

import (
	"context"
	"sync"

	"robot-hand-server/internal/domain"
)

type userSyncState struct {
	presets []domain.Preset
	history []domain.HistoryItem
}

type InMemorySyncRepository struct {
	mu    sync.RWMutex
	state map[string]userSyncState
}

func NewInMemorySyncRepository() *InMemorySyncRepository {
	return &InMemorySyncRepository{
		state: make(map[string]userSyncState),
	}
}

func (r *InMemorySyncRepository) GetSyncState(_ context.Context, userUUID string) ([]domain.Preset, []domain.HistoryItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, exists := r.state[userUUID]
	if !exists {
		// Возвращаем пустые массивы, а не nil, чтобы в JSON был [] а не null
		return []domain.Preset{}, []domain.HistoryItem{}, nil
	}
	return s.presets, s.history, nil
}

func (r *InMemorySyncRepository) SaveSyncState(_ context.Context, userUUID string, presets []domain.Preset, history []domain.HistoryItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.state[userUUID] = userSyncState{
		presets: presets,
		history: history,
	}
	return nil
}



func (r *InMemorySyncRepository) DeleteSyncState(_ context.Context, userUUID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.state, userUUID) // Удаляем все пресеты и историю юзера
	return nil
}
package service

import (
	"context"

	"robot-hand-server/internal/domain"
	"robot-hand-server/internal/repository"
)

type SyncService struct {
	syncRepo repository.SyncRepository
}

func NewSyncService(syncRepo repository.SyncRepository) *SyncService {
	return &SyncService{syncRepo: syncRepo}
}

func (s *SyncService) Sync(ctx context.Context, userUUID string, req SyncRequest) (SyncResponse, error) {
	serverPresets, serverHistory, err := s.syncRepo.GetSyncState(ctx, userUUID)
	if err != nil {
		return SyncResponse{}, err
	}

	// === 1. ВЫКИДЫВАЕМ ИЗ ПАМЯТИ ТО, ЧТО УДАЛИЛ ЮЗЕР ===
	deletedMap := make(map[string]bool)
	for _, id := range req.DeletedPresetIDs {
		deletedMap[id] = true
	}

	var filteredServerPresets []domain.Preset
	for _, p := range serverPresets {
		if !deletedMap[p.ID] { // Оставляем только те, которых НЕТ в списке на удаление
			filteredServerPresets = append(filteredServerPresets, p)
		}
	}

	// === 2. МЕРЖИМ ОСТАВШЕЕСЯ ===
	presetMap := make(map[string]domain.Preset)
	for _, p := range filteredServerPresets {
		presetMap[p.ID] = p
	}
	for _, p := range req.LocalPresets {
		presetMap[p.ID] = p
	}

	mergedPresets := make([]domain.Preset, 0, len(presetMap))
	for _, p := range presetMap {
		mergedPresets = append(mergedPresets, p)
	}

	// ... дальше идет код для LocalHistory (его не трогай, оставляй как было)

	// 2. Мержим историю по Date (дедубликация)
	historyMap := make(map[int64]domain.HistoryItem)
	for _, h := range serverHistory {
		historyMap[h.Date] = h
	}
	for _, h := range req.LocalHistory {
		historyMap[h.Date] = h
	}

	mergedHistory := make([]domain.HistoryItem, 0, len(historyMap))
	for _, h := range historyMap {
		mergedHistory = append(mergedHistory, h)
	}

	// Защита от возврата nil в JSON
	if mergedPresets == nil { mergedPresets = []domain.Preset{} }
	if mergedHistory == nil { mergedHistory = []domain.HistoryItem{} }

	// 3. Сохраняем объединенный стейт на сервере
	err = s.syncRepo.SaveSyncState(ctx, userUUID, mergedPresets, mergedHistory)
	if err != nil {
		return SyncResponse{}, err
	}

	return SyncResponse{
		MergedPresets: mergedPresets,
		MergedHistory: mergedHistory,
	}, nil
}
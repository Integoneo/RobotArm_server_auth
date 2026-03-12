package repository

import (
	"context"
	"sync"

	"robot-hand-server/internal/domain"
)

type InMemorySupportTicketRepository struct {
	mu      sync.Mutex
	tickets []*domain.SupportTicket
}

func NewInMemorySupportTicketRepository() *InMemorySupportTicketRepository {
	return &InMemorySupportTicketRepository{
		tickets: make([]*domain.SupportTicket, 0),
	}
}

func (r *InMemorySupportTicketRepository) Create(_ context.Context, ticket *domain.SupportTicket) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copyTicket := *ticket
	r.tickets = append(r.tickets, &copyTicket)
	return nil
}

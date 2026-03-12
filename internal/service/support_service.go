package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"robot-hand-server/internal/domain"
	"robot-hand-server/internal/mailer"
	"robot-hand-server/internal/repository"
	"robot-hand-server/internal/storage"
)

const maxImageSizeBytes = 5 << 20 // 5MB

type SupportService struct {
	users      repository.UserRepository
	tickets    repository.SupportTicketRepository
	imageStore storage.ImageStore
	mailer     mailer.SupportMailer
}

func NewSupportService(
	users repository.UserRepository,
	tickets repository.SupportTicketRepository,
	imageStore storage.ImageStore,
	supportMailer mailer.SupportMailer,
) *SupportService {
	if supportMailer == nil {
		supportMailer = noopSupportMailer{}
	}

	return &SupportService{
		users:      users,
		tickets:    tickets,
		imageStore: imageStore,
		mailer:     supportMailer,
	}
}

type noopSupportMailer struct{}

func (noopSupportMailer) SendSupportTicket(context.Context, domain.SupportTicket) error {
	return nil
}

func (s *SupportService) CreateTicket(ctx context.Context, req CreateSupportTicketRequest) error {
	header := strings.TrimSpace(req.Header)
	text := strings.TrimSpace(req.Text)

	if header == "" || text == "" {
		return fmt.Errorf("%w: header and text are required", ErrInvalidInput)
	}

	if req.AuthUserUUID == "" || req.UUID == "" {
		return fmt.Errorf("%w: uuid is required", ErrInvalidInput)
	}

	if _, err := uuid.Parse(req.UUID); err != nil {
		return fmt.Errorf("%w: invalid uuid", ErrInvalidInput)
	}

	if req.AuthUserUUID != req.UUID {
		return ErrUserMismatch
	}

	email, err := validateEmail(req.Email)
	if err != nil {
		return err
	}

	user, err := s.users.FindByUUID(ctx, req.UUID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrInvalidInput
		}
		return err
	}

	if !strings.EqualFold(user.Email, email) {
		return ErrUserMismatch
	}

	imagePath := ""
	if req.Image != nil {
		if len(req.Image.Content) == 0 {
			return fmt.Errorf("%w: empty image", ErrInvalidInput)
		}
		if len(req.Image.Content) > maxImageSizeBytes {
			return ErrImageTooLarge
		}

		contentType := http.DetectContentType(req.Image.Content)
		if contentType != "image/jpeg" && contentType != "image/png" {
			return ErrUnsupportedImage
		}

		imagePath, err = s.imageStore.Save(ctx, req.Image.Filename, req.Image.Content)
		if err != nil {
			return err
		}
	}

	ticket := &domain.SupportTicket{
		ID:        uuid.NewString(),
		UserUUID:  req.UUID,
		Email:     email,
		Header:    header,
		Text:      text,
		ImagePath: imagePath,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.mailer.SendSupportTicket(ctx, *ticket); err != nil {
		return err
	}

	if err := s.tickets.Create(ctx, ticket); err != nil {
		return err
	}

	return nil
}

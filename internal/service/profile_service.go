package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"robot-hand-server/internal/repository"
)

type ProfileService struct {
	users    repository.UserRepository
	syncRepo repository.SyncRepository // <-- Добавили
}

func NewProfileService(users repository.UserRepository, syncRepo repository.SyncRepository) *ProfileService {
	return &ProfileService{users: users, syncRepo: syncRepo}
}

func (s *ProfileService) GetProfile(ctx context.Context, userUUID string) (ProfileResponse, error) {
	user, err := s.users.FindByUUID(ctx, strings.TrimSpace(userUUID))
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ProfileResponse{}, ErrNotFound
		}
		return ProfileResponse{}, err
	}

	return ProfileResponse{
		UUID:    user.UUID,
		Email:   user.Email,
		Name:    user.Name,
		Surname: user.Surname,
	}, nil
}

func (s *ProfileService) UpdateProfile(ctx context.Context, userUUID string, req UpdateProfileRequest) (ProfileResponse, error) {
	user, err := s.users.FindByUUID(ctx, strings.TrimSpace(userUUID))
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ProfileResponse{}, ErrNotFound
		}
		return ProfileResponse{}, err
	}

	email, err := validateEmail(req.Email)
	if err != nil {
		return ProfileResponse{}, err
	}

	name := strings.TrimSpace(req.Name)
	surname := strings.TrimSpace(req.Surname)
	if name == "" || surname == "" {
		return ProfileResponse{}, fmt.Errorf("%w: name and surname are required", ErrInvalidInput)
	}

	user.Email = email
	user.Name = name
	user.Surname = surname

	currentPassword := req.CurrentPassword
	newPassword := req.NewPassword
	wantsPasswordChange := strings.TrimSpace(currentPassword) != "" || strings.TrimSpace(newPassword) != ""
	if wantsPasswordChange {
		if strings.TrimSpace(currentPassword) == "" || strings.TrimSpace(newPassword) == "" {
			return ProfileResponse{}, fmt.Errorf("%w: current and new password are required", ErrInvalidInput)
		}
		if len(newPassword) < 8 {
			return ProfileResponse{}, fmt.Errorf("%w: password must be at least 8 characters", ErrInvalidInput)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
			return ProfileResponse{}, ErrUnauthorized
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return ProfileResponse{}, err
		}
		user.PasswordHash = string(hash)
	}

	if err := s.users.Update(ctx, user); err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			return ProfileResponse{}, fmt.Errorf("%w: email already exists", ErrConflict)
		}
		if errors.Is(err, repository.ErrUserNotFound) {
			return ProfileResponse{}, ErrNotFound
		}
		return ProfileResponse{}, err
	}

	return ProfileResponse{
		UUID:    user.UUID,
		Email:   user.Email,
		Name:    user.Name,
		Surname: user.Surname,
	}, nil
}



func (s *ProfileService) DeleteProfile(ctx context.Context, userUUID string) error {
	uuid := strings.TrimSpace(userUUID)
	
	// 1. Удаляем самого юзера
	err := s.users.Delete(ctx, uuid)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrNotFound
		}
		return err
	}
	
	// 2. Каскадное удаление: сносим пресеты и историю
	_ = s.syncRepo.DeleteSyncState(ctx, uuid)
	
	return nil
}

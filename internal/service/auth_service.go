package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"robot-hand-server/internal/auth"
	"robot-hand-server/internal/domain"
	"robot-hand-server/internal/repository"
)

type tokenManager interface {
	Generate(userUUID string) (string, error)
	Parse(tokenString string) (string, error)
	ParseDetailed(tokenString string) (auth.ParsedToken, error)
}

type passwordResetMailer interface {
	SendPasswordResetCode(ctx context.Context, email, code string) error
}

type AuthService struct {
	users              repository.UserRepository
	revokedToken       repository.RevokedTokenRepository
	passwordResetCodes repository.PasswordResetCodeRepository
	tokens             tokenManager
	resetMailer        passwordResetMailer
	resetCodeTTL       time.Duration
	resetCodeSecret    []byte
}

func NewAuthService(
	users repository.UserRepository,
	revokedToken repository.RevokedTokenRepository,
	passwordResetCodes repository.PasswordResetCodeRepository,
	tokens tokenManager,
	resetMailer passwordResetMailer,
	resetCodeTTL time.Duration,
	resetCodeSecret string,
) *AuthService {
	if resetMailer == nil {
		resetMailer = noopPasswordResetMailer{}
	}

	if resetCodeTTL <= 0 {
		resetCodeTTL = 15 * time.Minute
	}

	if strings.TrimSpace(resetCodeSecret) == "" {
		resetCodeSecret = "fallback-reset-secret"
	}

	return &AuthService{
		users:              users,
		revokedToken:       revokedToken,
		passwordResetCodes: passwordResetCodes,
		tokens:             tokens,
		resetMailer:        resetMailer,
		resetCodeTTL:       resetCodeTTL,
		resetCodeSecret:    []byte(resetCodeSecret),
	}
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (AuthResponse, error) {
	email, err := validateEmail(req.Email)
	if err != nil {
		return AuthResponse{}, err
	}

	if len(req.Password) < 8 {
		return AuthResponse{}, fmt.Errorf("%w: password must be at least 8 characters", ErrInvalidInput)
	}

	name := strings.TrimSpace(req.Name)
	surname := strings.TrimSpace(req.Surname)
	if name == "" || surname == "" {
		return AuthResponse{}, fmt.Errorf("%w: name and surname are required", ErrInvalidInput)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResponse{}, err
	}

	user := &domain.User{
		UUID:         uuid.NewString(),
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		Surname:      surname,
	}

	err = s.users.Create(ctx, user)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			return AuthResponse{}, fmt.Errorf("%w: email already exists", ErrConflict)
		}
		return AuthResponse{}, err
	}

	token, err := s.tokens.Generate(user.UUID)
	if err != nil {
		return AuthResponse{}, err
	}

	 return AuthResponse{
    Token:   token,
    UUID:    user.UUID,
    Name:    user.Name,
    Surname: user.Surname,
  }, nil
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (AuthResponse, error) {
	email, err := validateEmail(req.Email)
	if err != nil {
		return AuthResponse{}, ErrUnauthorized
	}

	if strings.TrimSpace(req.Password) == "" {
		return AuthResponse{}, ErrUnauthorized
	}

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return AuthResponse{}, ErrUnauthorized
		}
		return AuthResponse{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return AuthResponse{}, ErrUnauthorized
	}

	token, err := s.tokens.Generate(user.UUID)
	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{
    Token:   token,
    UUID:    user.UUID,
    Name:    user.Name,
    Surname: user.Surname,
  }, nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) error {
	email, err := validateEmail(req.Email)
	if err != nil {
		return ErrNotFound
	}

	_, err = s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrNotFound
		}
		return err
	}

	code, err := generateResetCode()
	if err != nil {
		return err
	}

	codeHash := s.hashResetCode(email, code)
	expiresAt := time.Now().UTC().Add(s.resetCodeTTL)
	if err := s.passwordResetCodes.Save(ctx, email, codeHash, expiresAt); err != nil {
		return err
	}

	if err := s.resetMailer.SendPasswordResetCode(ctx, email, code); err != nil {
		return err
	}

	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	email, err := validateEmail(req.Email)
	if err != nil {
		return ErrInvalidInput
	}

	code := strings.TrimSpace(req.Code)
	if len(code) != 6 {
		return ErrInvalidCode
	}

	if len(req.NewPassword) < 8 {
		return ErrInvalidInput
	}

	if _, err := s.users.FindByEmail(ctx, email); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrNotFound
		}
		return err
	}

	consumed, err := s.passwordResetCodes.Consume(ctx, email, s.hashResetCode(email, code), time.Now().UTC())
	if err != nil {
		return err
	}
	if !consumed {
		return ErrInvalidCode
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.users.UpdatePasswordHashByEmail(ctx, email, string(hash)); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return ErrNotFound
		}
		return err
	}

	return nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	if strings.TrimSpace(token) == "" {
		return ErrInvalidToken
	}

	parsed, err := s.tokens.ParseDetailed(token)
	if err != nil || parsed.Subject == "" {
		return ErrInvalidToken
	}

	if err := s.revokedToken.Revoke(ctx, token, parsed.ExpiresAt); err != nil {
		return err
	}
	return nil
}

func (s *AuthService) AuthenticateToken(ctx context.Context, token string) (string, error) {
	if strings.TrimSpace(token) == "" {
		return "", ErrInvalidToken
	}

	revoked, err := s.revokedToken.IsRevoked(ctx, token)
	if err != nil {
		return "", err
	}
	if revoked {
		return "", ErrTokenRevoked
	}

	userUUID, err := s.tokens.Parse(token)
	if err != nil {
		return "", ErrInvalidToken
	}

	_, err = s.users.FindByUUID(ctx, userUUID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", ErrInvalidToken
		}
		return "", err
	}

	return userUUID, nil
}

func validateEmail(rawEmail string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(rawEmail))
	if email == "" {
		return "", fmt.Errorf("%w: email is required", ErrInvalidInput)
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return "", fmt.Errorf("%w: invalid email", ErrInvalidInput)
	}

	return email, nil
}

func generateResetCode() (string, error) {
	const digits = "0123456789"
	raw := make([]byte, 6)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}

	for i := range raw {
		raw[i] = digits[int(raw[i])%len(digits)]
	}

	return string(raw), nil
}

func (s *AuthService) hashResetCode(email, code string) string {
	mac := hmac.New(sha256.New, s.resetCodeSecret)
	mac.Write([]byte(strings.ToLower(strings.TrimSpace(email))))
	mac.Write([]byte(":"))
	mac.Write([]byte(code))
	return hex.EncodeToString(mac.Sum(nil))
}

type noopPasswordResetMailer struct{}

func (noopPasswordResetMailer) SendPasswordResetCode(context.Context, string, string) error {
	return nil
}

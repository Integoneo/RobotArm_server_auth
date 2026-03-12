package service

import "errors"

var (
	ErrInvalidInput     = errors.New("invalid input")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrNotFound         = errors.New("not found")
	ErrConflict         = errors.New("conflict")
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenRevoked     = errors.New("token revoked")
	ErrUserMismatch     = errors.New("user mismatch")
	ErrUnsupportedImage = errors.New("unsupported image")
	ErrImageTooLarge    = errors.New("image too large")
	ErrInvalidCode      = errors.New("invalid code")
)

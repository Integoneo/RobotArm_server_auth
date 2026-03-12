package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

type LocalImageStore struct {
	dir string
}

func NewLocalImageStore(dir string) *LocalImageStore {
	return &LocalImageStore{dir: dir}
}

func (s *LocalImageStore) Save(_ context.Context, originalFilename string, content []byte) (string, error) {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return "", err
	}

	name, err := randomName()
	if err != nil {
		return "", err
	}

	ext := filepath.Ext(originalFilename)
	if ext == "" {
		ext = ".bin"
	}

	fullPath := filepath.Join(s.dir, fmt.Sprintf("%s%s", name, ext))
	if err := os.WriteFile(fullPath, content, 0o644); err != nil {
		return "", err
	}

	return fullPath, nil
}

func randomName() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

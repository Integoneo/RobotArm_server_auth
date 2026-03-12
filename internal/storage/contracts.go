package storage

import "context"

type ImageStore interface {
	Save(ctx context.Context, originalFilename string, content []byte) (string, error)
}

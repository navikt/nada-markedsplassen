package service

import "context"

type CloudStorageAPI interface {
	GetObject(ctx context.Context, path string) (*ObjectWithData, error)
}

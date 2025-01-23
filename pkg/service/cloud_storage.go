package service

import "context"

type CloudStorageAPI interface {
	GetObjectAndUnmarshalYAML(ctx context.Context, path string, into any) error
	GetObject(ctx context.Context, path string) (*ObjectWithData, error)
}

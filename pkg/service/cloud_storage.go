package service

import "context"

type CloudStorageAPI interface {
	GetNumberOfObjectsWithPrefix(ctx context.Context, prefix string) (int, error)
	GetObjectsWithPrefix(ctx context.Context, prefix string) ([]*Object, error)
	GetObjectAndUnmarshalYAML(ctx context.Context, path string, into any) error
	GetObject(ctx context.Context, path string) (*ObjectWithData, error)
	WriteFileToBucket(ctx context.Context, pathPrefix string, file *UploadFile) error
	DeleteObjectsWithPrefix(ctx context.Context, prefix string) error
}

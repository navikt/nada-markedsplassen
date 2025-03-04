package service

import "context"

type CloudStorageAPI interface {
	GetNumberOfObjectsWithPrefix(ctx context.Context, bucket, prefix string) (int, error)
	GetObjectsWithPrefix(ctx context.Context, bucket, prefix string) ([]*Object, error)
	GetObjectAndUnmarshalYAML(ctx context.Context, bucket, path string, into any) error
	GetObject(ctx context.Context, bucket, path string) (*ObjectWithData, error)
	WriteFileToBucket(ctx context.Context, bucket, pathPrefix string, file *UploadFile) error
	DeleteObjectsWithPrefix(ctx context.Context, bucket, prefix string) error
}

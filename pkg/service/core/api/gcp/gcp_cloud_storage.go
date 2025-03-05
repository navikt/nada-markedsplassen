package gcp

import (
	"context"
	"errors"
	"fmt"
	"path"

	"github.com/navikt/nada-backend/pkg/cloudstorage"
	"gopkg.in/yaml.v3"

	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/rs/zerolog"
)

var _ service.CloudStorageAPI = &cloudStorageAPI{}

type cloudStorageAPI struct {
	ops cloudstorage.Operations
}

func (s *cloudStorageAPI) GetNumberOfObjectsWithPrefix(ctx context.Context, bucket, prefix string) (int, error) {
	const op errs.Op = "storyAPI.GetNumberOfObjectsWithPrefix"

	objects, err := s.ops.GetObjects(ctx, bucket, &cloudstorage.Query{Prefix: prefix})
	if err != nil {
		return 0, errs.E(errs.IO, service.CodeGCPStorage, op, err)
	}

	return len(objects), nil
}

func (s *cloudStorageAPI) GetObjectsWithPrefix(ctx context.Context, bucket, prefix string) ([]*service.Object, error) {
	const op errs.Op = "cloudStorageAPI.GetObjectsWithPrefix"

	raw, err := s.ops.GetObjects(ctx, bucket, &cloudstorage.Query{Prefix: prefix + "/"})
	if err != nil {
		return nil, errs.E(errs.IO, service.CodeGCPStorage, op, err)
	}

	objs := make([]*service.Object, len(raw))
	for idx, obj := range raw {
		objs[idx] = &service.Object{
			Name:   obj.Name,
			Bucket: obj.Bucket,
			Attrs: service.Attributes{
				ContentType:     obj.Attrs.ContentType,
				ContentEncoding: obj.Attrs.ContentEncoding,
				Size:            obj.Attrs.Size,
				SizeStr:         obj.Attrs.SizeStr,
			},
		}
	}

	return objs, nil
}

func (s *cloudStorageAPI) GetObjectAndUnmarshalYAML(ctx context.Context, bucket, path string, into any) error {
	const op errs.Op = "cloudStorageAPI.GetObjectAndUnmarshalYAML"

	obj, err := s.GetObject(ctx, bucket, path)
	if err != nil {
		return errs.E(errs.IO, service.CodeGCPStorage, op, err)
	}

	err = yaml.Unmarshal(obj.Data, into)
	if err != nil {
		return errs.E(errs.Invalid, service.CodeGCPStorage, op, err)
	}

	return nil
}

func (s *cloudStorageAPI) GetObject(ctx context.Context, bucket, path string) (*service.ObjectWithData, error) {
	const op errs.Op = "cloudStorageAPI.GetObject"

	obj, err := s.ops.GetObjectWithData(ctx, bucket, path)
	if err != nil {
		if errors.Is(err, cloudstorage.ErrObjectNotExist) {
			return nil, errs.E(errs.NotExist, service.CodeGCPStorage, op, fmt.Errorf("object %v does not exist", path), service.ParamObject)
		}

		return nil, errs.E(errs.IO, service.CodeGCPStorage, op, err)
	}

	return &service.ObjectWithData{
		Object: &service.Object{
			Name:   obj.Name,
			Bucket: obj.Bucket,
			Attrs: service.Attributes{
				ContentType:     obj.Attrs.ContentType,
				ContentEncoding: obj.Attrs.ContentEncoding,
				Size:            obj.Attrs.Size,
				SizeStr:         obj.Attrs.SizeStr,
			},
		},
		Data: obj.Data,
	}, nil
}

func (s *cloudStorageAPI) WriteFileToBucket(ctx context.Context, bucket, pathPrefix string, file *service.UploadFile) error {
	const op errs.Op = "storyAPI.WriteFileToBucket"

	err := s.ops.WriteObject(ctx, bucket, path.Join(pathPrefix, file.Path), file.ReadCloser, nil)
	if err != nil {
		return errs.E(errs.IO, service.CodeGCPStorage, op, err)
	}

	return nil
}

func (s *cloudStorageAPI) DeleteObjectsWithPrefix(ctx context.Context, bucket, prefix string) error {
	const op errs.Op = "cloudStorageAPI.DeleteObjectsWithPrefix"

	_, err := s.ops.DeleteObjects(ctx, bucket, &cloudstorage.Query{Prefix: prefix})
	if err != nil {
		return errs.E(errs.IO, service.CodeGCPStorage, op, err)
	}

	return nil
}

func NewCloudStorageAPI(ops cloudstorage.Operations, _ zerolog.Logger) *cloudStorageAPI {
	return &cloudStorageAPI{
		ops: ops,
	}
}

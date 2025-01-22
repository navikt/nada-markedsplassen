package gcp

import (
	"context"
	"errors"
	"fmt"

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

func (s *cloudStorageAPI) GetObjectAndUnmarshalYAML(ctx context.Context, path string, into any) error {
	const op errs.Op = "cloudStorageAPI.GetObjectAndUnmarshalYAML"

	obj, err := s.GetObject(ctx, path)
	if err != nil {
		return errs.E(errs.IO, service.CodeGCPStorage, op, err)
	}

	err = yaml.Unmarshal(obj.Data, into)
	if err != nil {
		return errs.E(errs.Invalid, service.CodeGCPStorage, op, err)
	}

	return nil
}

func (s *cloudStorageAPI) GetObject(ctx context.Context, path string) (*service.ObjectWithData, error) {
	const op errs.Op = "cloudStorageAPI.GetObject"

	obj, err := s.ops.GetObjectWithData(ctx, path)
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

func NewCloudStorageAPI(ops cloudstorage.Operations, _ zerolog.Logger) *cloudStorageAPI {
	return &cloudStorageAPI{
		ops: ops,
	}
}

package gcp

import (
	"context"
	"errors"
	"fmt"

	"github.com/navikt/nada-backend/pkg/cs"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/rs/zerolog"
)

var _ service.CloudStorageAPI = &cloudStorageAPI{}

type cloudStorageAPI struct {
	log zerolog.Logger
	ops cs.Operations
}

func (s *cloudStorageAPI) GetObject(ctx context.Context, path string) (*service.ObjectWithData, error) {
	const op errs.Op = "cloudStorageAPI.GetObject"

	obj, err := s.ops.GetObjectWithData(ctx, path)
	if err != nil {
		if errors.Is(err, cs.ErrObjectNotExist) {
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

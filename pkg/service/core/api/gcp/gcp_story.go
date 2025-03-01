package gcp

import (
	"context"
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/navikt/nada-backend/pkg/cloudstorage"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/rs/zerolog"
)

var _ service.StoryAPI = &storyAPI{}

type storyAPI struct {
	log zerolog.Logger
	ops cloudstorage.Operations
}

func (s *storyAPI) GetNumberOfObjectsWithPrefix(ctx context.Context, prefix string) (int, error) {
	const op errs.Op = "storyAPI.GetNumberOfObjectsWithPrefix"

	objects, err := s.ops.GetObjects(ctx, &cloudstorage.Query{Prefix: prefix})
	if err != nil {
		return 0, errs.E(errs.IO, service.CodeGCPStorage, op, err)
	}

	return len(objects), nil
}

func (s *storyAPI) GetIndexHtmlPath(ctx context.Context, prefix string) (string, error) {
	const op errs.Op = "storyAPI.GetIndexHtmlPath"

	prefix, _ = strings.CutSuffix(prefix, "/")

	objs, err := s.ops.GetObjects(ctx, &cloudstorage.Query{Prefix: prefix + "/"})
	if err != nil {
		return "", errs.E(errs.IO, service.CodeGCPStorage, op, err)
	}

	sort.Slice(objs, func(i, j int) bool {
		return len(objs[i].Name) < len(objs[j].Name)
	})

	var candidates []string
	for _, obj := range objs {
		if strings.HasSuffix(strings.ToLower(obj.Name), "/index.html") {
			return obj.Name, nil
		} else if strings.HasSuffix(strings.ToLower(obj.Name), ".html") {
			candidates = append(candidates, obj.Name)
		}
	}

	if len(candidates) == 0 {
		return "", errs.E(errs.NotExist, service.CodeGCPStorage, op, fmt.Errorf("no index.html found in %v", prefix), service.ParamObject)
	}

	return candidates[0], nil
}

func (s *storyAPI) GetObject(ctx context.Context, path string) (*service.ObjectWithData, error) {
	const op errs.Op = "storyAPI.GetObject"

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

func (s *storyAPI) DeleteObjectsWithPrefix(ctx context.Context, prefix string) (int, error) {
	const op errs.Op = "storyAPI.DeleteObjectsWithPrefix"

	n, err := s.ops.DeleteObjects(ctx, &cloudstorage.Query{Prefix: prefix})
	if err != nil {
		return 0, errs.E(errs.IO, service.CodeGCPStorage, op, err)
	}

	return n, nil
}

func (s *storyAPI) WriteFilesToBucket(ctx context.Context, storyID string, files []*service.UploadFile, cleanupOnFailure bool) error {
	const op errs.Op = "storyAPI.WriteFilesToBucket"

	var err error

	for _, file := range files {
		err = s.WriteFileToBucket(ctx, storyID, file)
		if err != nil {
			s.log.Error().Err(err).Msg("writing story file: " + path.Join(storyID, file.Path))
			break
		}
	}
	if err != nil && cleanupOnFailure {
		ed := s.DeleteStoryFolder(ctx, storyID)
		if ed != nil {
			s.log.Error().Err(ed).Msg("deleting story folder on cleanup: " + storyID)
		}
	}

	if err != nil {
		return errs.E(errs.IO, service.CodeGCPStorage, op, err)
	}

	return nil
}

func (s *storyAPI) WriteFileToBucket(ctx context.Context, pathPrefix string, file *service.UploadFile) error {
	const op errs.Op = "storyAPI.WriteFileToBucket"

	err := s.ops.WriteObject(ctx, path.Join(pathPrefix, file.Path), file.ReadCloser, nil)
	if err != nil {
		return errs.E(errs.IO, service.CodeGCPStorage, op, err)
	}

	return nil
}

func (s *storyAPI) DeleteStoryFolder(ctx context.Context, storyID string) error {
	const op errs.Op = "storyAPI.DeleteStoryFolder"

	if len(storyID) == 0 {
		return errs.E(errs.InvalidRequest, service.CodeGCPStorage, op, fmt.Errorf("story id %s is empty", storyID), service.ParamStory)
	}

	_, err := s.DeleteObjectsWithPrefix(ctx, storyID+"/")
	if err != nil {
		return errs.E(op, service.CodeGCPStorage, err)
	}

	return nil
}

func NewStoryAPI(ops cloudstorage.Operations, log zerolog.Logger) *storyAPI {
	return &storyAPI{
		log: log,
		ops: ops,
	}
}

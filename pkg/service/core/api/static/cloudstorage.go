package static

import (
	"context"

	"github.com/navikt/nada-backend/pkg/service"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

type cloudStorageAPI struct {
	log  zerolog.Logger
	data []byte
}

func (s *cloudStorageAPI) GetObjectAndUnmarshalYAML(ctx context.Context, path string, into any) error {
	return yaml.Unmarshal(s.data, into)
}

func (s *cloudStorageAPI) GetObject(ctx context.Context, path string) (*service.ObjectWithData, error) {
	return &service.ObjectWithData{
		Data: s.data,
	}, nil
}

func NewCloudStorageAPI(log zerolog.Logger, data []byte) *cloudStorageAPI {
	return &cloudStorageAPI{
		log:  log,
		data: data,
	}
}

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/database/gensql"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

type OpenMetabaseMetadata gensql.OpenMetabaseMetadatum

// Ensure that we always implement the service.MetabaseStorage interface
var _ service.OpenMetabaseStorage = &openMetabaseStorage{}

// FIXME: should define an interface that uses the subset of Querier methods that we need
// and reference that here
type openMetabaseStorage struct {
	db *database.Repo
}

func (s *openMetabaseStorage) SetDatabaseMetabaseMetadata(ctx context.Context, datasetID uuid.UUID, databaseID int) (*service.OpenMetabaseMetadata, error) {
	const op errs.Op = "openMetabaseStorage.SetDatabaseMetabaseMetadata"

	meta, err := s.db.Querier.SetDatabaseOpenMetabaseMetadata(ctx, gensql.SetDatabaseOpenMetabaseMetadataParams{
		DatabaseID: sql.NullInt32{Valid: true, Int32: int32(databaseID)},
		DatasetID:  datasetID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.E(errs.NotExist, service.CodeDatabase, op, fmt.Errorf("setting database %v: %w", databaseID, err), service.ParamDataset)
		}

		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return ToLocalOpen(meta).Convert(), nil
}

func (s *openMetabaseStorage) SetSyncCompletedMetabaseMetadata(ctx context.Context, datasetID uuid.UUID) error {
	const op errs.Op = "openMetabaseStorage.SetSyncCompletedMetabaseMetadata"

	err := s.db.Querier.SetSyncCompletedOpenMetabaseMetadata(ctx, datasetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errs.E(errs.NotExist, service.CodeDatabase, op, fmt.Errorf("setting sync completed: %w", err), service.ParamDataset)
		}

		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return nil
}

func (s *openMetabaseStorage) CreateMetadata(ctx context.Context, datasetID uuid.UUID) error {
	const op errs.Op = "openMetabaseStorage.CreateMetadata"

	err := s.db.Querier.CreateOpenMetabaseMetadata(ctx, datasetID)
	if err != nil {
		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return nil
}

func (s *openMetabaseStorage) GetMetadata(ctx context.Context, datasetID uuid.UUID, includeDeleted bool) (*service.OpenMetabaseMetadata, error) {
	const op errs.Op = "openMetabaseStorage.GetMetadata"

	if includeDeleted {
		meta, err := s.db.Querier.GetOpenMetabaseMetadataWithDeleted(ctx, datasetID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, errs.E(errs.NotExist, service.CodeDatabase, op, fmt.Errorf("getting dataset %v: %w", datasetID, err), service.ParamDataset)
			}

			return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
		}

		return ToLocalOpen(meta).Convert(), nil
	}

	meta, err := s.db.Querier.GetOpenMetabaseMetadata(ctx, datasetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.E(errs.NotExist, service.CodeDatabase, op, fmt.Errorf("getting dataset %v: %w", datasetID, err), service.ParamDataset)
		}

		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return ToLocalOpen(meta).Convert(), nil
}

func (s *openMetabaseStorage) GetAllMetadata(ctx context.Context) ([]*service.OpenMetabaseMetadata, error) {
	const op errs.Op = "openMetabaseStorage.GetAllMetadata"

	mbs, err := s.db.Querier.GetAllOpenMetabaseMetadata(ctx)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	mbMetas := make([]*service.OpenMetabaseMetadata, len(mbs))
	for idx, meta := range mbs {
		mbMetas[idx] = ToLocalOpen(meta).Convert()
	}

	return mbMetas, nil
}

func (s *openMetabaseStorage) GetOpenTablesInSameBigQueryDataset(ctx context.Context, projectID, dataset string) ([]string, error) {
	const op errs.Op = "openMetabaseStorage.GetOpenTablesInSameBigQueryDataset"

	tables, err := s.db.Querier.GetOpenMetabaseTablesInSameBigQueryDataset(ctx, gensql.GetOpenMetabaseTablesInSameBigQueryDatasetParams{
		ProjectID: projectID,
		Dataset:   dataset,
	})
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return tables, nil
}

func (s *openMetabaseStorage) DeleteMetadata(ctx context.Context, datasetID uuid.UUID) error {
	const op errs.Op = "openMetabaseStorage.DeleteMetadata"

	err := s.db.Querier.DeleteOpenMetabaseMetadata(ctx, datasetID)
	if err != nil {
		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return nil
}

func ToLocalOpen(m gensql.OpenMetabaseMetadatum) OpenMetabaseMetadata {
	return OpenMetabaseMetadata(m)
}

func (m OpenMetabaseMetadata) Convert() *service.OpenMetabaseMetadata {
	return &service.OpenMetabaseMetadata{
		DatasetID:     m.DatasetID,
		DatabaseID:    nullInt32ToIntPtr(m.DatabaseID),
		DeletedAt:     nullTimeToPtr(m.DeletedAt),
		SyncCompleted: nullTimeToPtr(m.SyncCompleted),
	}
}

func NewOpenMetabaseStorage(db *database.Repo) *openMetabaseStorage {
	return &openMetabaseStorage{
		db: db,
	}
}

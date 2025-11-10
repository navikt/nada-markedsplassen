package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/database/gensql"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

type RestrictedMetabaseMetadata gensql.RestrictedMetabaseMetadatum

var _ service.RestrictedMetabaseStorage = &restrictedMetabaseStorage{}

// FIXME: should define an interface that uses the subset of Querier methods that we need
// and reference that here
type restrictedMetabaseStorage struct {
	db *database.Repo
}

func (s *restrictedMetabaseStorage) SetServiceAccountPrivateKeyMetabaseMetadata(ctx context.Context, datasetID uuid.UUID, saPrivateKey []byte) (*service.RestrictedMetabaseMetadata, error) {
	const op errs.Op = "restrictedMetabaseStorage.SetServiceAccountPrivateKeyMetabaseMetadata"

	meta, err := s.db.Querier.SetServiceAccountPrivateKeyRestrictedMetabaseMetadata(ctx, gensql.SetServiceAccountPrivateKeyRestrictedMetabaseMetadataParams{
		SaPrivateKey: saPrivateKey,
		DatasetID:    datasetID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.E(errs.NotExist, service.CodeDatabase, op, fmt.Errorf("setting sa_private_key for %v: %w", datasetID, err), service.ParamDataset)
		}

		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return ToLocal(meta).Convert(), nil
}

func (s *restrictedMetabaseStorage) SetCollectionMetabaseMetadata(ctx context.Context, datasetID uuid.UUID, collectionID int) (*service.RestrictedMetabaseMetadata, error) {
	const op errs.Op = "restrictedMetabaseStorage.SetCollectionMetabaseMetadata"

	meta, err := s.db.Querier.SetCollectionRestrictedMetabaseMetadata(ctx, gensql.SetCollectionRestrictedMetabaseMetadataParams{
		CollectionID: sql.NullInt32{Valid: true, Int32: int32(collectionID)},
		DatasetID:    datasetID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.E(errs.NotExist, service.CodeDatabase, op, fmt.Errorf("setting collection %v: %w", collectionID, err), service.ParamDataset)
		}

		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return ToLocal(meta).Convert(), nil
}

func (s *restrictedMetabaseStorage) SetDatabaseMetabaseMetadata(ctx context.Context, datasetID uuid.UUID, databaseID int) (*service.RestrictedMetabaseMetadata, error) {
	const op errs.Op = "restrictedMetabaseStorage.SetDatabaseMetabaseMetadata"

	meta, err := s.db.Querier.SetDatabaseRestrictedMetabaseMetadata(ctx, gensql.SetDatabaseRestrictedMetabaseMetadataParams{
		DatabaseID: sql.NullInt32{Valid: true, Int32: int32(databaseID)},
		DatasetID:  datasetID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.E(errs.NotExist, service.CodeDatabase, op, fmt.Errorf("setting database %v: %w", databaseID, err), service.ParamDataset)
		}

		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return ToLocal(meta).Convert(), nil
}

func (s *restrictedMetabaseStorage) SetServiceAccountMetabaseMetadata(ctx context.Context, datasetID uuid.UUID, saEmail string) (*service.RestrictedMetabaseMetadata, error) {
	const op errs.Op = "restrictedMetabaseStorage.SetServiceAccountMetabaseMetadata"

	meta, err := s.db.Querier.SetServiceAccountRestrictedMetabaseMetadata(ctx, gensql.SetServiceAccountRestrictedMetabaseMetadataParams{
		SaEmail:   saEmail,
		DatasetID: datasetID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.E(errs.NotExist, service.CodeDatabase, op, fmt.Errorf("setting sa_email %v: %w", saEmail, err), service.ParamDataset)
		}

		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return ToLocal(meta).Convert(), nil
}

func (s *restrictedMetabaseStorage) SetSyncCompletedMetabaseMetadata(ctx context.Context, datasetID uuid.UUID) error {
	const op errs.Op = "restrictedMetabaseStorage.SetSyncCompletedMetabaseMetadata"

	err := s.db.Querier.SetSyncCompletedRestrictedMetabaseMetadata(ctx, datasetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errs.E(errs.NotExist, service.CodeDatabase, op, fmt.Errorf("setting sync completed: %w", err), service.ParamDataset)
		}

		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return nil
}

func (s *restrictedMetabaseStorage) SetPermissionGroupMetabaseMetadata(ctx context.Context, datasetID uuid.UUID, groupID int) (*service.RestrictedMetabaseMetadata, error) {
	const op errs.Op = "restrictedMetabaseStorage.SetPermissionGroupMetabaseMetadata"

	meta, err := s.db.Querier.SetPermissionGroupRestrictedMetabaseMetadata(ctx, gensql.SetPermissionGroupRestrictedMetabaseMetadataParams{
		PermissionGroupID: sql.NullInt32{Valid: true, Int32: int32(groupID)},
		DatasetID:         datasetID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.E(errs.NotExist, service.CodeDatabase, op, fmt.Errorf("setting permissions group %v: %w", groupID, err), service.ParamDataset)
		}

		return nil, errs.E(errs.Database, op, err)
	}

	return ToLocal(meta).Convert(), nil
}

func (s *restrictedMetabaseStorage) CreateMetadata(ctx context.Context, datasetID uuid.UUID) error {
	const op errs.Op = "restrictedMetabaseStorage.CreateMetadata"

	err := s.db.Querier.CreateRestrictedMetabaseMetadata(ctx, datasetID)
	if err != nil {
		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return nil
}

func (s *restrictedMetabaseStorage) GetMetadata(ctx context.Context, datasetID uuid.UUID) (*service.RestrictedMetabaseMetadata, error) {
	const op errs.Op = "restrictedMetabaseStorage.GetMetadata"

	meta, err := s.db.Querier.GetRestrictedMetabaseMetadata(ctx, datasetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.E(errs.NotExist, service.CodeDatabase, op, fmt.Errorf("%v: %w", datasetID, err), service.ParamDataset)
		}

		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return ToLocal(meta).Convert(), nil
}

func (s *restrictedMetabaseStorage) GetAllMetadata(ctx context.Context) ([]*service.RestrictedMetabaseMetadata, error) {
	const op errs.Op = "restrictedMetabaseStorage.GetAllMetadata"

	mbs, err := s.db.Querier.GetAllRestrictedMetabaseMetadata(ctx)
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	mbMetas := make([]*service.RestrictedMetabaseMetadata, len(mbs))
	for idx, meta := range mbs {
		mbMetas[idx] = ToLocal(meta).Convert()
	}

	return mbMetas, nil
}

func (s *restrictedMetabaseStorage) GetOpenTablesInSameBigQueryDataset(ctx context.Context, projectID, dataset string) ([]string, error) {
	const op errs.Op = "restrictedMetabaseStorage.GetOpenTablesInSameBigQueryDataset"

	tables, err := s.db.Querier.GetOpenMetabaseTablesInSameBigQueryDataset(ctx, gensql.GetOpenMetabaseTablesInSameBigQueryDatasetParams{
		ProjectID: projectID,
		Dataset:   dataset,
	})
	if err != nil {
		return nil, errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return tables, nil
}

func (s *restrictedMetabaseStorage) DeleteMetadata(ctx context.Context, datasetID uuid.UUID) error {
	const op errs.Op = "restrictedMetabaseStorage.DeleteMetadata"

	err := s.db.Querier.DeleteRestrictedMetabaseMetadata(ctx, datasetID)
	if err != nil {
		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return nil
}

func (s *restrictedMetabaseStorage) OpenPreviouslyRestrictedMetabaseBigqueryDatabase(ctx context.Context, datasetID uuid.UUID) error {
	const op errs.Op = "restrictedMetabaseStorage.OpenPreviouslyRestrictedMetabaseBigqueryDatabase"

	err := s.db.Querier.OpenPreviouslyRestrictedMetabaseMetadata(ctx, datasetID)
	if err != nil {
		return errs.E(errs.Database, service.CodeDatabase, op, err)
	}

	return nil
}

func ToLocal(m gensql.RestrictedMetabaseMetadatum) RestrictedMetabaseMetadata {
	return RestrictedMetabaseMetadata(m)
}

func (m RestrictedMetabaseMetadata) Convert() *service.RestrictedMetabaseMetadata {
	return &service.RestrictedMetabaseMetadata{
		DatasetID:         m.DatasetID,
		DatabaseID:        nullInt32ToIntPtr(m.DatabaseID),
		PermissionGroupID: nullInt32ToIntPtr(m.PermissionGroupID),
		CollectionID:      nullInt32ToIntPtr(m.CollectionID),
		SAEmail:           m.SaEmail,
		SyncCompleted:     nullTimeToPtr(m.SyncCompleted),
		SAPrivateKey:      m.SaPrivateKey,
	}
}

func nullTimeToPtr(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}

	return &nt.Time
}

func NewRestrictedMetabaseStorage(db *database.Repo) *restrictedMetabaseStorage {
	return &restrictedMetabaseStorage{
		db: db,
	}
}

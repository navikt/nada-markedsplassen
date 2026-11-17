package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/service/core/storage/postgres"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetabaseDocumentStorage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	log := zerolog.New(os.Stdout)
	containers := NewContainers(t, log)
	defer containers.Cleanup()

	pgCfg := containers.RunPostgres(NewPostgresConfig())
	repo, err := database.New(pgCfg.ConnectionURL(), 10, 10)
	require.NoError(t, err)
	storage := postgres.NewMetabaseDocumentStorage(repo)

	ownerGroup := "storage-test-" + uuid.NewString() + "@example.test"
	otherGroup := "storage-other-" + uuid.NewString() + "@example.test"
	teamID := uuid.New()
	publicDocumentID := uuid.New()
	created, err := storage.CreateMetabaseDocument(ctx, &service.NewPublicMetabaseDocument{
		Input: &service.PublicMetabaseDocumentInput{Keywords: []string{},
			Group:  ownerGroup,
			Link:   "https://metabase.example.test/public/document/storage-test",
			TeamID: nil,
		},
		CreatorEmail:     "storage-test@example.test",
		Name:             "Storage test document",
		PublicDocumentID: publicDocumentID,
		MetabaseID:       73190101,
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, publicDocumentID, created.PublicDocumentID)
	assert.Equal(t, 73190101, created.MetabaseID)
	assert.Nil(t, created.Description)
	assert.Nil(t, created.TeamkatalogenURL)
	assert.Nil(t, created.TeamID)
	assert.Empty(t, created.Keywords)
	assert.NotNil(t, created.Keywords)
	assert.Equal(t, ownerGroup, created.Group)

	byID, err := storage.GetMetabaseDocument(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created, byID)

	other, err := storage.CreateMetabaseDocument(ctx, &service.NewPublicMetabaseDocument{
		Input: &service.PublicMetabaseDocumentInput{Keywords: []string{},
			Group: otherGroup,
			Link:  "https://metabase.example.test/public/document/other",
		},
		CreatorEmail:     "storage-test@example.test",
		Name:             "Other group document",
		PublicDocumentID: uuid.New(),
		MetabaseID:       73190102,
	})
	require.NoError(t, err)
	documents, err := storage.GetMetabaseDocumentForGroups(ctx, []string{ownerGroup})
	require.NoError(t, err)
	require.Len(t, documents, 1)
	assert.Equal(t, created.ID, documents[0].ID)
	assert.NotEqual(t, other.ID, documents[0].ID)

	_, err = storage.CreateMetabaseDocument(ctx, &service.NewPublicMetabaseDocument{
		Input:            &service.PublicMetabaseDocumentInput{Keywords: []string{}, Group: ownerGroup, Link: "https://metabase.example.test/duplicate"},
		CreatorEmail:     "storage-test@example.test",
		Name:             "Duplicate Metabase ID",
		PublicDocumentID: uuid.New(),
		MetabaseID:       73190101,
	})
	assertMetabaseDocumentErrorKind(t, err, errs.Database)

	for _, invalidMetabaseID := range []int32{0, -1} {
		_, err = storage.CreateMetabaseDocument(ctx, &service.NewPublicMetabaseDocument{
			Input:            &service.PublicMetabaseDocumentInput{Keywords: []string{}, Group: ownerGroup, Link: "https://metabase.example.test/invalid"},
			CreatorEmail:     "storage-test@example.test",
			Name:             "Invalid Metabase ID",
			PublicDocumentID: uuid.New(),
			MetabaseID:       invalidMetabaseID,
		})
		assertMetabaseDocumentErrorKind(t, err, errs.Database)
	}

	description := "Updated description"
	teamURL := "https://teamkatalogen.example.test/team/storage-test"
	updated, err := storage.UpdateMetabaseDocument(ctx, &service.EditPublicMetabaseDocument{
		ID:   created.ID,
		Name: "Updated document",
		Input: &service.PublicMetabaseDocumentEditInput{
			Description:      &description,
			Keywords:         []string{"updated", "storage"},
			TeamkatalogenURL: &teamURL,
			TeamID:           &teamID,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated document", updated.Name)
	assert.Equal(t, &description, updated.Description)
	assert.Equal(t, []string{"updated", "storage"}, updated.Keywords)
	assert.Equal(t, &teamURL, updated.TeamkatalogenURL)
	assert.Equal(t, &teamID, updated.TeamID)
	assert.Equal(t, ownerGroup, updated.Group)

	cleared, err := storage.UpdateMetabaseDocument(ctx, &service.EditPublicMetabaseDocument{
		ID:   created.ID,
		Name: "Updated document",
		Input: &service.PublicMetabaseDocumentEditInput{
			Keywords: []string{},
		},
	})
	require.NoError(t, err)
	assert.Nil(t, cleared.Description)
	assert.Empty(t, cleared.Keywords)
	assert.NotNil(t, cleared.Keywords)
	assert.Nil(t, cleared.TeamkatalogenURL)
	assert.Nil(t, cleared.TeamID)
	assert.Equal(t, ownerGroup, cleared.Group)

	_, err = storage.GetMetabaseDocument(ctx, uuid.New())
	assertMetabaseDocumentErrorKind(t, err, errs.NotExist)

	historicalLastModified := time.Date(2001, time.January, 2, 3, 4, 5, 0, time.UTC)
	historicalID := uuid.New()
	historicalPublicDocumentID := uuid.New()
	_, err = repo.GetDB().ExecContext(ctx, `
		INSERT INTO metabase_document (
			id, name, description, "group", public_document_id, metabase_id,
			created_by, created, last_modified, keywords, teamkatalogen_url, team_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`,
		historicalID,
		"Historical timestamp fixture",
		nil,
		ownerGroup,
		historicalPublicDocumentID,
		73190103,
		"storage-test@example.test",
		historicalLastModified,
		historicalLastModified,
		[]string{},
		nil,
		nil,
	)
	require.NoError(t, err)
	historicalUpdate, err := storage.UpdateMetabaseDocument(ctx, &service.EditPublicMetabaseDocument{
		ID:   historicalID,
		Name: "Historical timestamp updated",
		Input: &service.PublicMetabaseDocumentEditInput{
			Keywords: []string{},
		},
	})
	require.NoError(t, err)
	assert.True(t, historicalUpdate.LastModified.After(historicalLastModified), "last_modified should be advanced by the update trigger")

	require.NoError(t, storage.DeleteMetabaseDocument(ctx, created.ID))
	_, err = storage.GetMetabaseDocument(ctx, created.ID)
	assertMetabaseDocumentErrorKind(t, err, errs.NotExist)
}

func assertMetabaseDocumentErrorKind(t *testing.T, err error, kind errs.Kind) {
	t.Helper()
	require.Error(t, err)
	require.True(t, errs.KindIs(kind, err), "expected error kind %v, got %v", kind, err)
}

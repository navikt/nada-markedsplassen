package integration

import (
	"context"
	"fmt"
	gohttp "net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/bq"
	bigQueryEmulator "github.com/navikt/nada-backend/pkg/bq/emulator"
	"github.com/navikt/nada-backend/pkg/config/v2"
	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/service/core"
	"github.com/navikt/nada-backend/pkg/service/core/api/gcp"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/routes"
	"github.com/navikt/nada-backend/pkg/service/core/storage"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// nolint: tparallel,maintidx,goconst
func TestDataproduct(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(20*time.Minute))
	defer cancel()

	log := zerolog.New(zerolog.NewConsoleWriter())
	log.Level(zerolog.DebugLevel)

	c := NewContainers(t, log)
	defer c.Cleanup()

	pgCfg := c.RunPostgres(NewPostgresConfig())

	repo, err := database.New(
		pgCfg.ConnectionURL(),
		10,
		10,
	)
	assert.NoError(t, err)

	fuelBqSchema := NewDatasetBiofuelConsumptionRatesSchema()

	bqe := bigQueryEmulator.New(log)
	bqe.WithProject(Project, fuelBqSchema...)
	bqe.EnableMock(false, log, bigQueryEmulator.NewPolicyMock(log).Mocks()...)

	bqHTTPPort := strconv.Itoa(GetFreePort(t))
	bqHTTPAddr := fmt.Sprintf("127.0.0.1:%s", bqHTTPPort)
	if len(os.Getenv("CI")) > 0 {
		bqHTTPAddr = fmt.Sprintf("0.0.0.0:%s", bqHTTPPort)
	}
	bqGRPCAddr := fmt.Sprintf("127.0.0.1:%s", strconv.Itoa(GetFreePort(t)))

	go func() {
		_ = bqe.Serve(ctx, bqHTTPAddr, bqGRPCAddr)
	}()
	bqClient := bq.NewClient("http://"+bqHTTPAddr, false, log)

	stores := storage.NewStores(nil, repo, config.Config{}, log)

	err = stores.NaisConsoleStorage.UpdateAllTeamProjects(ctx, []*service.NaisTeamMapping{
		{
			Slug:       NaisTeamNada,
			GroupEmail: GroupEmailNada,
			ProjectID:  Project,
		},
	})
	assert.NoError(t, err)

	zlog := zerolog.New(os.Stdout)
	datasetOwnerRouter := TestRouter(zlog)

	bqapi := gcp.NewBigQueryAPI(Project, Location, PseudoDataSet, bqClient)

	dataproductService := core.NewDataProductsService(
		stores.DataProductsStorage,
		stores.BigQueryStorage,
		bqapi,
		stores.NaisConsoleStorage,
		GroupEmailAllUsers,
	)

	{
		h := handlers.NewDataProductsHandler(dataproductService)
		e := routes.NewDataProductsEndpoints(zlog, h)
		fDatasetOwnerRoutes := routes.NewDataProductsRoutes(e, injectUser(UserOne))

		fDatasetOwnerRoutes(datasetOwnerRouter)
	}

	datasetOwnerServer := httptest.NewServer(datasetOwnerRouter)
	defer datasetOwnerServer.Close()

	dpName := "my dataproduct"
	dpDesc := "my dataproduct description"
	dpSlug := "my-dataproduct"

	var dpID uuid.UUID
	t.Run("Create dataproduct", func(t *testing.T) {
		got := &service.DataproductMinimal{}
		NewTester(t, datasetOwnerServer).
			Post(ctx, service.NewDataproduct{
				Name:        dpName,
				Description: strToStrPtr(dpDesc),
				Group:       GroupEmailNada,
				Slug:        strToStrPtr(dpSlug),
			}, "/api/dataproducts/new").
			HasStatusCode(gohttp.StatusOK).Value(got)

		expect := &service.DataproductMinimal{
			Name:        dpName,
			Description: strToStrPtr(dpDesc),
			Owner: &service.DataproductOwner{
				Group:            GroupEmailNada,
				TeamkatalogenURL: strToStrPtr(""),
				TeamContact:      strToStrPtr(""),
			},
			Slug: dpSlug,
		}

		dpID = got.ID

		diff := cmp.Diff(expect, got, cmpopts.IgnoreFields(service.DataproductMinimal{}, "ID", "Created", "LastModified"))
		assert.Empty(t, diff)
	})

	t.Run("Get dataproduct", func(t *testing.T) {
		got := &service.DataproductWithDataset{}
		NewTester(t, datasetOwnerServer).
			Get(ctx, fmt.Sprintf("/api/dataproducts/%s", dpID)).
			HasStatusCode(gohttp.StatusOK).Value(got)

		expect := &service.DataproductWithDataset{
			Dataproduct: service.Dataproduct{
				ID:          dpID,
				Name:        dpName,
				Description: strToStrPtr(dpDesc),
				Slug:        dpSlug,
				Owner: &service.DataproductOwner{
					Group: GroupEmailNada,
				},
				Keywords: []string{},
			},
			Datasets: []*service.DatasetInDataproduct{},
		}

		diff := cmp.Diff(expect, got, cmpopts.IgnoreFields(service.DataproductWithDataset{}, "Created", "LastModified"))
		assert.Empty(t, diff)
	})

	dataset := NewDatasetBiofuelConsumptionRatesSchema()[0]
	bqSource := service.NewBigQuery{
		ProjectID: Project,
		Dataset:   dataset.DatasetID,
		Table:     dataset.TableID,
	}

	dsName := "my dataset"
	dsDesc := "my dataset description"
	dsSlug := "my-dataset"

	t.Run("Create dataset", func(t *testing.T) {
		got := &service.Dataset{}
		NewTester(t, datasetOwnerServer).
			Post(ctx, service.NewDataset{
				Name:          dsName,
				Description:   strToStrPtr(dsDesc),
				Slug:          strToStrPtr(dsSlug),
				DataproductID: dpID,
				BigQuery:      bqSource,
				Pii:           service.PiiLevelNone,
			}, "/api/datasets/new").
			HasStatusCode(gohttp.StatusOK).Value(&got)

		expect := &service.Dataset{
			DataproductID: dpID,
			Name:          dsName,
			Description:   strToStrPtr(dsDesc),
			Slug:          dsSlug,
			Pii:           service.PiiLevelNone,
			Keywords:      []string{},
			Mappings:      []string{},
			Datasource: &service.BigQuery{
				ProjectID: Project,
				Dataset:   dataset.DatasetID,
				Table:     dataset.TableID,
				TableType: "TABLE",
			},
		}

		diff := cmp.Diff(expect, got, cmpopts.IgnoreFields(service.Dataset{},
			"ID", "Created", "LastModified",
			"Datasource.ID", "Datasource.DatasetID", "Datasource.Created",
			"Datasource.LastModified", "Datasource.Schema", "Datasource.Schema",
			"Datasource.PiiTags",
		))
		assert.Empty(t, diff)
	})

	dsAnonymizedName := "my anonymized dataset"
	dsAnonymizedDesc := "my dataset anonymized description"
	dsAnonymizationDesc := "description of applied anonymization"
	dsAnonymizedSlug := "my-anonymized-dataset"

	var dsAnonymizedID uuid.UUID
	t.Run("Create anonymized dataset", func(t *testing.T) {
		got := &service.Dataset{}
		NewTester(t, datasetOwnerServer).
			Post(ctx, service.NewDataset{
				Name:                     dsAnonymizedName,
				Description:              strToStrPtr(dsAnonymizedDesc),
				Slug:                     strToStrPtr(dsAnonymizedSlug),
				DataproductID:            dpID,
				BigQuery:                 bqSource,
				Pii:                      service.PiiLevelAnonymised,
				AnonymisationDescription: strToStrPtr(dsAnonymizationDesc),
				PseudoColumns:            []string{},
			}, "/api/datasets/new").
			HasStatusCode(gohttp.StatusOK).Value(&got)

		expect := &service.Dataset{
			DataproductID:            dpID,
			Name:                     dsAnonymizedName,
			Description:              strToStrPtr(dsAnonymizedDesc),
			Slug:                     dsAnonymizedSlug,
			Pii:                      service.PiiLevelAnonymised,
			AnonymisationDescription: strToStrPtr(dsAnonymizationDesc),
			Keywords:                 []string{},
			Mappings:                 []string{},
			Datasource: &service.BigQuery{
				ProjectID:     Project,
				Dataset:       dataset.DatasetID,
				Table:         dataset.TableID,
				TableType:     "TABLE",
				PseudoColumns: []string{},
			},
		}

		dsAnonymizedID = got.ID

		diff := cmp.Diff(expect, got, cmpopts.IgnoreFields(service.Dataset{},
			"ID", "Created", "LastModified",
			"Datasource.ID", "Datasource.DatasetID", "Datasource.Created",
			"Datasource.LastModified", "Datasource.Schema", "Datasource.PiiTags",
		))
		assert.Empty(t, diff)
	})

	t.Run("Get anonymized dataset", func(t *testing.T) {
		got := &service.DatasetWithAccess{}
		NewTester(t, datasetOwnerServer).
			Get(ctx, fmt.Sprintf("/api/datasets/%s", dsAnonymizedID)).
			HasStatusCode(gohttp.StatusOK).Value(got)

		expect := &service.DatasetWithAccess{
			ID:                       dsAnonymizedID,
			DataproductID:            dpID,
			Name:                     dsAnonymizedName,
			Description:              strToStrPtr(dsAnonymizedDesc),
			Slug:                     dsAnonymizedSlug,
			Pii:                      service.PiiLevelAnonymised,
			AnonymisationDescription: strToStrPtr(dsAnonymizationDesc),
			Keywords:                 []string{},
			Mappings:                 []string{},
			Datasource: &service.BigQuery{
				DatasetID:     dsAnonymizedID,
				ProjectID:     Project,
				Dataset:       dataset.DatasetID,
				Table:         dataset.TableID,
				TableType:     "TABLE",
				PseudoColumns: []string{},
			},
			Access: []*service.Access{},
		}

		diff := cmp.Diff(expect, got, cmpopts.IgnoreFields(service.DatasetWithAccess{}, "Created", "LastModified",
			"Datasource.ID", "Datasource.DatasetID", "Datasource.Created",
			"Datasource.LastModified", "Datasource.Schema", "Datasource.PiiTags",
		))
		assert.Empty(t, diff)
	})

	updatedDSAnonymizationDesc := "description of updated applied anonymization"
	t.Run("Update anonymized dataset", func(t *testing.T) {
		NewTester(t, datasetOwnerServer).
			Put(ctx, service.UpdateDatasetDto{
				Name:                     dsAnonymizedName,
				Description:              strToStrPtr(dsAnonymizedDesc),
				Slug:                     strToStrPtr(dsAnonymizedSlug),
				Pii:                      service.PiiLevelAnonymised,
				Keywords:                 []string{},
				PseudoColumns:            []string{},
				DataproductID:            &dpID,
				PiiTags:                  strToStrPtr(`{"id":"PII_KanVæreIndirekteIdentifiserende"}`),
				AnonymisationDescription: strToStrPtr(updatedDSAnonymizationDesc),
			}, fmt.Sprintf("/api/datasets/%s", dsAnonymizedID)).
			HasStatusCode(gohttp.StatusOK)
	})

	t.Run("Get updated anonymized dataset", func(t *testing.T) {
		got := &service.DatasetWithAccess{}
		NewTester(t, datasetOwnerServer).
			Get(ctx, fmt.Sprintf("/api/datasets/%s", dsAnonymizedID)).
			HasStatusCode(gohttp.StatusOK).Value(got)

		expect := &service.DatasetWithAccess{
			ID:                       dsAnonymizedID,
			DataproductID:            dpID,
			Name:                     dsAnonymizedName,
			Description:              strToStrPtr(dsAnonymizedDesc),
			Slug:                     dsAnonymizedSlug,
			Pii:                      service.PiiLevelAnonymised,
			AnonymisationDescription: strToStrPtr(updatedDSAnonymizationDesc),
			Keywords:                 []string{},
			Mappings:                 []string{},
			Datasource: &service.BigQuery{
				DatasetID:     dsAnonymizedID,
				ProjectID:     Project,
				Dataset:       dataset.DatasetID,
				Table:         dataset.TableID,
				TableType:     "TABLE",
				PseudoColumns: []string{},
			},
			Access: []*service.Access{},
		}

		diff := cmp.Diff(expect, got, cmpopts.IgnoreFields(service.DatasetWithAccess{}, "Created", "LastModified",
			"Datasource.ID", "Datasource.DatasetID", "Datasource.Created",
			"Datasource.LastModified", "Datasource.Schema", "Datasource.PiiTags",
		))
		assert.Empty(t, diff)
	})

	t.Run("Get dataproduct after adding datasets", func(t *testing.T) {
		got := &service.DataproductWithDataset{}
		NewTester(t, datasetOwnerServer).
			Get(ctx, fmt.Sprintf("/api/dataproducts/%s", dpID)).
			HasStatusCode(gohttp.StatusOK).Value(got)

		expect := &service.DataproductWithDataset{
			Dataproduct: service.Dataproduct{
				ID:          dpID,
				Name:        dpName,
				Description: strToStrPtr(dpDesc),
				Slug:        dpSlug,
				Owner: &service.DataproductOwner{
					Group: GroupEmailNada,
				},
				Keywords: []string{},
			},
		}

		assert.Equal(t, 2, len(got.Datasets))
		diff := cmp.Diff(expect, got, cmpopts.IgnoreFields(service.DataproductWithDataset{}, "Created", "LastModified", "Datasets"))
		assert.Empty(t, diff)
	})
}

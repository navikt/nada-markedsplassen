package integration_metabase

import (
	"context"
	"fmt"
	"golang.org/x/exp/rand"
	"testing"
	"time"

	"github.com/navikt/nada-backend/pkg/bq"
	crm "github.com/navikt/nada-backend/pkg/cloudresourcemanager"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/rs/zerolog"
)

var (
	MetabaseProject                = "nada-metabase-tests"
	MetabaseDatasetPrefix          = "integration_tests"
	MetabaseAllUsersServiceAccount = "all-metabase-users@nada-metabase-tests.iam.gserviceaccount.com"
	BigQueryDataViewerRole         = "roles/bigquery.dataViewer"
	BigQueryMetadataViewerRole     = "roles/bigquery.metadataViewer"
	NadaMetabaseRole               = fmt.Sprintf("projects/%s/roles/nada.metabase", MetabaseProject)
)

type CleanupFn func(ctx context.Context)

type GCPHelper struct {
	BigQueryTable service.NewBigQuery

	t               *testing.T
	log             zerolog.Logger
	bigqueryDataset string
	client          *bq.Client
}

func NewGCPHelper(t *testing.T, log zerolog.Logger) *GCPHelper {
	return &GCPHelper{
		log:    log.With().Str("component", "GCPHelper").Logger(),
		t:      t,
		client: bq.NewClient("", true, log.With().Str("component", "BigQueryClient").Logger()),
	}
}

func (g *GCPHelper) Start(ctx context.Context) CleanupFn {
	g.log.Info().Msg("Starting GCP helper for Metabase integration tests")

	var err error
	g.bigqueryDataset, err = g.prepareTestProject(ctx)
	if err != nil {
		g.log.Fatal().Err(err).Msg("preparing test project")
	}

	bqTable, err := g.createBigQueryTable(ctx, g.bigqueryDataset)
	if err != nil {
		g.log.Fatal().Err(err).Msg("creating BigQuery table")
	}

	g.BigQueryTable = bqTable

	return g.Cleanup
}

func (g *GCPHelper) Cleanup(ctx context.Context) {
	g.log.Info().Msg("Cleaning up GCP resources after Metabase integration tests")

	err := g.cleanupAfterTestRun(ctx, g.bigqueryDataset)
	if err != nil {
		g.log.Fatal().Err(err).Msg("cleaning up after test run")
	}

	g.log.Info().Msg("GCP resources cleaned up successfully")
}

// For each test run we create a new dataset
func (g *GCPHelper) prepareTestProject(ctx context.Context) (string, error) {
	datasetName := GenerateRandomName(MetabaseDatasetPrefix, 30)

	g.log.Info().Str("datasetName", datasetName).Msg("Creating new BigQuery dataset for test run")

	err := g.client.CreateDataset(ctx, MetabaseProject, datasetName, "europe-north1")
	if err != nil {
		return "", fmt.Errorf("error creating dataset: %v", err)
	}

	return datasetName, nil
}

// For each test case we create a new table in the dataset
func (g *GCPHelper) createBigQueryTable(ctx context.Context, dataset string) (service.NewBigQuery, error) {
	tableName := GenerateRandomName("test_table_", 30)

	g.log.Info().Str("tableName", tableName).Msg("Creating new BigQuery table for test run")

	err := g.client.CreateTable(ctx, &bq.Table{
		ProjectID: MetabaseProject,
		DatasetID: dataset,
		TableID:   tableName,
		Location:  "europe-north1",
		Schema: []*bq.Column{
			{
				Name: "id",
				Type: "INT64",
				Mode: bq.NullableMode,
			},
			{
				Name: "name",
				Type: "STRING",
				Mode: bq.NullableMode,
			},
		},
	})
	if err != nil {
		return service.NewBigQuery{}, fmt.Errorf("error creating table: %v", err)
	}

	return service.NewBigQuery{
		ProjectID: MetabaseProject,
		Dataset:   dataset,
		Table:     tableName,
	}, nil
}

// Remove all resources created for the test run
func (g *GCPHelper) cleanupAfterTestRun(ctx context.Context, bqDataset string) error {
	crmClient := crm.NewClient("", false, nil)

	// Deleting the dataset and its tables
	err := g.client.DeleteDataset(ctx, MetabaseProject, bqDataset, true)
	if err != nil {
		return err
	}

	// Cleaning up NADA metabase project iam role grants for deleted service accounts
	err = crmClient.UpdateProjectIAMPolicyBindingsMembers(
		ctx,
		MetabaseProject,
		crm.RemoveDeletedMembersWithRole([]string{NadaMetabaseRole}, g.log),
	)
	if err != nil {
		return fmt.Errorf("error updating project iam policy bindings: %v", err)
	}

	return nil
}

func strToStrPtr(s string) *string {
	return &s
}

func numberOfDatabasesWithAccessForPermissionGroup(permissionGraphForGroup map[string]service.PermissionGroup) int {
	accessCount := 0
	for _, permission := range permissionGraphForGroup {
		if permission.CreateQueries == "query-builder-and-native" {
			accessCount += 1
		}
	}

	return accessCount
}

func GenerateRandomName(prefix string, maxLength int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := prefix

	if len(result) >= maxLength {
		return result[:maxLength]
	}

	r := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))

	remaining := maxLength - len(prefix)
	b := make([]byte, remaining)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}

	return result + string(b)
}

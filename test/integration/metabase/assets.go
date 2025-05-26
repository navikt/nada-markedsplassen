package integration_metabase

import (
	"context"
	"fmt"
	"os"
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

// For each test run we create a new dataset
func prepareTestProject(ctx context.Context) (string, error) {
	bqClient := bq.NewClient("", true, zerolog.New(os.Stdout))

	dsName := fmt.Sprintf("%s_%d", MetabaseDatasetPrefix, time.Now().UnixNano())
	err := bqClient.CreateDataset(ctx, MetabaseProject, dsName, "europe-north1")
	if err != nil {
		return "", fmt.Errorf("error creating dataset: %v", err)
	}

	return dsName, nil
}

// For each test case we create a new table in the dataset
func createBigQueryTable(ctx context.Context, dataset, table string) (service.NewBigQuery, error) {
	bqClient := bq.NewClient("", true, zerolog.New(os.Stdout))

	err := bqClient.CreateTable(ctx, &bq.Table{
		ProjectID: MetabaseProject,
		DatasetID: dataset,
		TableID:   table,
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

	return service.NewBigQuery{ProjectID: MetabaseProject, Dataset: dataset, Table: table}, nil
}

// Remove all resources created for the test run
func cleanupAfterTestRun(ctx context.Context, bqDataset string) error {
	bqClient := bq.NewClient("", true, zerolog.New(os.Stdout))
	crmClient := crm.NewClient("", false, nil)

	// Deleting the dataset and its tables
	err := bqClient.DeleteDataset(ctx, MetabaseProject, bqDataset, true)
	if err != nil {
		return err
	}

	// Cleaning up NADA metabase project iam role grants for deleted service accounts
	err = crmClient.UpdateProjectIAMPolicyBindingsMembers(ctx, MetabaseProject, crm.RemoveDeletedMembersWithRole([]string{NadaMetabaseRole}, zerolog.Nop()))
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

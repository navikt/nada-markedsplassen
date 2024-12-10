package integration_metabase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/navikt/nada-backend/pkg/bq"
	crm "github.com/navikt/nada-backend/pkg/cloudresourcemanager"
	dmpSA "github.com/navikt/nada-backend/pkg/sa"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/rs/zerolog"
)

var (
	MetabaseProject                = "metabase-integration-tests"
	MetabaseDatasetPrefix          = "integration_tests"
	MetabaseAllUsersServiceAccount = "all-metabase-users@metabase-integration-tests.iam.gserviceaccount.com"
	BigQueryDataViewerRole         = "roles/bigquery.dataViewer"
	BigQueryMetadataViewerRole     = "roles/bigquery.metadataViewer"
	NadaMetabaseRole               = fmt.Sprintf("projects/%s/roles/nada.metabase", MetabaseProject)
)

// This generic cleanup function is called prior to the metabase integration tests
// to ensure that resources from previous failed test runs are cleaned up.
// To not interfere with any potential ongoing tests we only delete resources older than 10 minutes
func cleanupTestProject(ctx context.Context) error {
	saClient := dmpSA.NewClient("", false)
	bqClient := bq.NewClient("", true, zerolog.New(os.Stdout))
	crmClient := crm.NewClient("", false, nil)

	// Deleting all BigQuery datasets older than 10 minutes
	existingDatasets, err := bqClient.GetDatasets(ctx, MetabaseProject)
	if err != nil {
		return err
	}

	for _, ds := range existingDatasets {
		datasetWithMetadata, err := bqClient.GetDataset(ctx, MetabaseProject, ds.DatasetID)
		if err != nil {
			return err
		}

		if datasetWithMetadata.CreationTime.Before(time.Now().Add(-10 * time.Minute)) {
			err := bqClient.DeleteDataset(ctx, MetabaseProject, ds.DatasetID, true)
			if err != nil {
				return err
			}
		}
	}

	// Delete restricted database service accounts older than 10 minutes
	// Cannot fetch service account creation timestamp directly so instead fetching a service account key
	// and checking the valid after timestamp (which corresponds to creation time of the key)
	// NB: In cases where the key creation fails in the metabase integration tests
	// the service account will never be automatically cleaned up by this function
	serviceAccounts, err := saClient.ListServiceAccounts(ctx, MetabaseProject)
	if err != nil {
		return fmt.Errorf("error getting service accounts: %v", err)
	}
	for _, sa := range serviceAccounts {
		if !strings.HasPrefix(sa.Email, "nada-") {
			continue
		}

		keys, err := saClient.ListServiceAccountKeys(ctx, sa.Name)
		if err != nil {
			return fmt.Errorf("error listing service account keys: %v", err)
		}

		for _, key := range keys {
			createTime, err := time.Parse("2006-01-02T15:04:05Z", key.ValidAfterTime)
			if err != nil {
				return fmt.Errorf("error parsing key creation time: %v", err)
			}

			if createTime.Before(time.Now().Add(-10 * time.Minute)) {
				fmt.Printf("Deleting restricted database service account %s\n", sa.Email)
				err := saClient.DeleteServiceAccount(ctx, fmt.Sprintf("projects/%s/serviceAccounts/%s", MetabaseProject, sa.Email))
				if err != nil {
					if errors.Is(err, dmpSA.ErrNotFound) {
						continue
					}
					return fmt.Errorf("error deleting service account: %v", err)
				}
				break
			}
		}
	}

	// Cleaning up NADA metabase project iam role grants for deleted service accounts
	err = crmClient.UpdateProjectIAMPolicyBindingsMembers(ctx, MetabaseProject, crm.RemoveDeletedMembersWithRole([]string{NadaMetabaseRole}, zerolog.Nop()))
	if err != nil {
		return fmt.Errorf("error updating project iam policy bindings: %v", err)
	}

	return nil
}

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

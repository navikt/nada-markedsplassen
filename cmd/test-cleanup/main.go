package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/navikt/nada-backend/pkg/bq"
	crm "github.com/navikt/nada-backend/pkg/cloudresourcemanager"
	dmpSA "github.com/navikt/nada-backend/pkg/sa"
	"github.com/navikt/nada-backend/pkg/service"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog"
)

var project = flag.String("project", "nada-metabase-tests", "The GCP project used for integration tests")

func cleanupTestProject(ctx context.Context, project string, log zerolog.Logger) error {
	saClient := dmpSA.NewClient("", false)
	bqClient := bq.NewClient("", true, log)
	crmClient := crm.NewClient("", false, nil)

	datasets, err := bqClient.GetDatasets(ctx, project)
	if err != nil {
		return fmt.Errorf("getting datasets: %v", err)
	}

	for _, ds := range datasets {
		err := bqClient.DeleteDataset(ctx, project, ds.DatasetID, true)
		if err != nil && !errors.Is(err, bq.ErrNotExist) {
			return fmt.Errorf("deleting dataset %s: %v", ds.DatasetID, err)
		}
	}

	serviceAccounts, err := saClient.ListServiceAccounts(ctx, project)
	if err != nil {
		return fmt.Errorf("getting service accounts: %v", err)
	}

	for _, sa := range serviceAccounts {
		if !strings.HasPrefix(sa.Email, "nada-") {
			continue
		}

		err := saClient.DeleteServiceAccount(ctx, service.ServiceAccountNameFromEmail(project, sa.Email))
		if err != nil && !errors.Is(err, dmpSA.ErrNotFound) {
			return fmt.Errorf("error deleting service account: %v", err)
		}
	}

	// Cleaning up NADA metabase project iam role grants for deleted service accounts
	err = crmClient.UpdateProjectIAMPolicyBindingsMembers(
		ctx,
		project,
		crm.RemoveDeletedMembersWithRole(
			[]string{
				service.NadaMetabaseRole(project),
			},
			log,
		),
	)
	if err != nil {
		return fmt.Errorf("updating project IAM policy bindings: %v", err)
	}

	return nil
}

func main() {
	flag.Parse()

	if *project == "" {
		fmt.Println("error: -project flag must be set")
		os.Exit(1)
	}

	log := zerolog.New(os.Stdout).With().Str("system", "integration_test_cleaner").Timestamp().Logger()
	log.Info().Msg("starting integration test project cleanup")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		errCh <- cleanupTestProject(ctx, *project, log)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			log.Fatal().Err(err).Msg("cleanup")
		}
		log.Info().Msg("cleanup completed")
	case <-ctx.Done():
		log.Fatal().Err(ctx.Err()).Msg("integration test cleanup context timed out")
	case sig := <-sigCh:
		log.Info().Str("signal", sig.String()).Msg("received signal, shutting down")

		os.Exit(0)
	}
}

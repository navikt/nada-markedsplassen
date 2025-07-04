package integration

import (
	"context"
	"fmt"
	"github.com/navikt/nada-backend/pkg/kms"
	"github.com/navikt/nada-backend/pkg/kms/emulator"
	river2 "github.com/navikt/nada-backend/pkg/service/core/queue/river"
	"github.com/navikt/nada-backend/pkg/worker"
	"github.com/riverqueue/river"
	gohttp "net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	crm "github.com/navikt/nada-backend/pkg/cloudresourcemanager"
	crmEmulator "github.com/navikt/nada-backend/pkg/cloudresourcemanager/emulator"
	"google.golang.org/api/cloudresourcemanager/v1"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/navikt/nada-backend/pkg/config/v2"
	"github.com/navikt/nada-backend/pkg/sa"
	serviceAccountEmulator "github.com/navikt/nada-backend/pkg/sa/emulator"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/service/core/api/static"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/navikt/nada-backend/pkg/bq"
	bigQueryEmulator "github.com/navikt/nada-backend/pkg/bq/emulator"
	"github.com/navikt/nada-backend/pkg/database"
	"github.com/navikt/nada-backend/pkg/service/core"
	"github.com/navikt/nada-backend/pkg/service/core/api/gcp"
	"github.com/navikt/nada-backend/pkg/service/core/api/http"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/routes"
	"github.com/navikt/nada-backend/pkg/service/core/storage"
	"github.com/rs/zerolog"
)

// nolint: tparallel,maintidx,goconst
func TestAccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	const serviceaccountName = "my-sa@project-id.iam.gserviceaccount.com"

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

	mbCfg := c.RunMetabase(NewMetabaseConfig(), "../../.metabase_version")

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

	saEmulator := serviceAccountEmulator.New(log)
	saURL := saEmulator.Run()
	saClient := sa.NewClient(saURL, true)

	crmEmulator := crmEmulator.New(log)
	crmEmulator.SetPolicy(Project, &cloudresourcemanager.Policy{
		Bindings: []*cloudresourcemanager.Binding{
			{
				Role:    "roles/owner",
				Members: []string{fmt.Sprintf("user:%s", UserOne.Email)},
			},
		},
	})
	crmURL := crmEmulator.Run()
	crmClient := crm.NewClient(crmURL, true, nil)

	stores := storage.NewStores(nil, repo, config.Config{}, log)

	zlog := zerolog.New(os.Stdout)
	datasetOwnerRouter := TestRouter(zlog)
	accessRequesterRouter := TestRouter(zlog)

	bigQueryContainerHostPort := "http://host.docker.internal:" + bqHTTPPort
	if len(os.Getenv("CI")) > 0 {
		bigQueryContainerHostPort = "http://172.17.0.1:" + bqHTTPPort
	}

	kmsEmulator := emulator.New(log)
	kmsEmulator.AddSymmetricKey(MetabaseProject, Location, Keyring, MetabaseKeyName, []byte("7b483b28d6e67cfd3b9b5813a286c763"))
	kmsURL := kmsEmulator.Run()

	kmsClient := kms.NewClient(kmsURL, true)

	crmapi := gcp.NewCloudResourceManagerAPI(crmClient)
	saapi := gcp.NewServiceAccountAPI(saClient)
	bqapi := gcp.NewBigQueryAPI(Project, Location, PseudoDataSet, bqClient)
	kmsapi := gcp.NewKMSAPI(kmsClient)

	// FIXME: should we just add /api to the connectionurl returned
	mbapi := http.NewMetabaseHTTP(
		mbCfg.ConnectionURL()+"/api",
		mbCfg.Email,
		mbCfg.Password,
		// We want metabase to connect with the big query emulator
		// running on the host
		bigQueryContainerHostPort,
		true,
		false,
		log,
	)

	workers := river.NewWorkers()
	riverConfig := worker.RiverConfig(&zlog, workers)
	mbqueue := river2.NewMetabaseQueue(repo, riverConfig)

	mbService := core.NewMetabaseService(
		Project,
		Location,
		Keyring,
		MetabaseKeyName,
		fakeMetabaseSA,
		"nada-metabase@test.iam.gserviceaccount.com",
		"group:"+GroupEmailAllUsers,
		GroupEmailAllUsers,
		mbqueue,
		kmsapi,
		mbapi,
		bqapi,
		saapi,
		crmapi,
		stores.MetaBaseStorage,
		stores.BigQueryStorage,
		stores.DataProductsStorage,
		stores.AccessStorage,
		zlog,
	)

	err = worker.MetabaseAddWorkers(riverConfig, mbService, repo)
	assert.NoError(t, err)

	err = stores.NaisConsoleStorage.UpdateAllTeamProjects(ctx, []*service.NaisTeamMapping{
		{
			Slug:       NaisTeamNada,
			GroupEmail: GroupEmailNada,
			ProjectID:  Project,
		},
	})
	assert.NoError(t, err)

	dataproductService := core.NewDataProductsService(
		stores.DataProductsStorage,
		stores.BigQueryStorage,
		bqapi,
		stores.NaisConsoleStorage,
		GroupEmailAllUsers,
	)

	userService := core.NewUserService(
		stores.AccessStorage,
		stores.PollyStorage,
		stores.TokenStorage,
		stores.StoryStorage,
		stores.DataProductsStorage,
		stores.InsightProductStorage,
		stores.NaisConsoleStorage,
		zlog,
	)

	StorageCreateProductAreasAndTeams(t, stores.ProductAreaStorage)
	fuel, err := dataproductService.CreateDataproduct(ctx, UserOne, NewDataProductBiofuelProduction(GroupEmailNada, TeamSeagrassID))
	assert.NoError(t, err)

	fuelData, err := dataproductService.CreateDataset(ctx, UserOne, NewDatasetBiofuelConsumptionRates(fuel.ID))
	assert.NoError(t, err)

	{
		h := handlers.NewUserHandler(userService)
		e := routes.NewUserEndpoints(zlog, h)
		fAccessRequesterRoutes := routes.NewUserRoutes(e, InjectUser(UserTwo))
		fAccessRequesterRoutes(accessRequesterRouter)
	}

	{
		h := handlers.NewDataProductsHandler(dataproductService)
		e := routes.NewDataProductsEndpoints(zlog, h)
		fDatasetOwnerRoutes := routes.NewDataProductsRoutes(e, InjectUser(UserOne))
		fAccessRequesterRoutes := routes.NewDataProductsRoutes(e, InjectUser(UserTwo))

		fDatasetOwnerRoutes(datasetOwnerRouter)
		fAccessRequesterRoutes(accessRequesterRouter)
	}

	{
		h := handlers.NewMetabaseHandler(mbService)
		e := routes.NewMetabaseEndpoints(zlog, h)
		fDatasetOwnerRoutes := routes.NewMetabaseRoutes(e, InjectUser(UserOne))

		fDatasetOwnerRoutes(datasetOwnerRouter)
	}

	{
		slack := static.NewSlackAPI(log)
		s := core.NewAccessService(
			"https://data.nav.no",
			slack,
			stores.PollyStorage,
			stores.AccessStorage,
			stores.DataProductsStorage,
			stores.BigQueryStorage,
			stores.JoinableViewsStorage,
			bqapi,
		)
		h := handlers.NewAccessHandler(s, mbService, Project)
		e := routes.NewAccessEndpoints(zlog, h)
		fDatasetOwnerRoutes := routes.NewAccessRoutes(e, InjectUser(UserOne))
		fAccessRequesterRoutes := routes.NewAccessRoutes(e, InjectUser(UserTwo))

		fDatasetOwnerRoutes(datasetOwnerRouter)
		fAccessRequesterRoutes(accessRequesterRouter)
	}

	datasetOwnerServer := httptest.NewServer(datasetOwnerRouter)
	defer datasetOwnerServer.Close()

	accessRequesterServer := httptest.NewServer(accessRequesterRouter)
	defer accessRequesterServer.Close()

	t.Run("Create dataset access request", func(t *testing.T) {
		NewTester(t, accessRequesterServer).
			Post(ctx, service.NewAccessRequestDTO{
				DatasetID:   fuelData.ID,
				Expires:     nil,
				Subject:     strToStrPtr(UserTwo.Email),
				SubjectType: strToStrPtr(service.SubjectTypeUser),
			}, "/api/accessRequests/new").
			HasStatusCode(gohttp.StatusNoContent)

		expect := &service.AccessRequestsWrapper{
			AccessRequests: []service.AccessRequest{
				{
					DatasetID:   fuelData.ID,
					Subject:     UserTwoEmail,
					SubjectType: service.SubjectTypeUser,
					Owner:       UserTwoEmail,
					Status:      service.AccessRequestStatusPending,
					Polly:       &service.Polly{},
				},
			},
		}
		got := &service.AccessRequestsWrapper{}

		NewTester(t, datasetOwnerServer).Get(ctx, "/api/accessRequests", "datasetId", fuelData.ID.String()).
			HasStatusCode(gohttp.StatusOK).
			Value(got)

		require.Len(t, got.AccessRequests, 1)
		diff := cmp.Diff(expect.AccessRequests[0], got.AccessRequests[0], cmpopts.IgnoreFields(service.AccessRequest{}, "ID", "Created"))
		assert.Empty(t, diff)
	})

	existingAR := &service.AccessRequest{}

	t.Run("Approve dataset access request", func(t *testing.T) {
		existingARs := &service.AccessRequestsWrapper{}
		NewTester(t, datasetOwnerServer).Get(ctx, "/api/accessRequests", "datasetId", fuelData.ID.String()).
			HasStatusCode(gohttp.StatusOK).
			Value(existingARs)

		existingAR = &existingARs.AccessRequests[0]

		NewTester(t, datasetOwnerServer).Post(ctx, nil, fmt.Sprintf("/api/accessRequests/process/%v", existingAR.ID), "action", "approve").
			HasStatusCode(gohttp.StatusNoContent)

		expect := &service.DatasetWithAccess{
			Access: []*service.Access{
				{
					DatasetID:     existingAR.DatasetID,
					Subject:       "user:" + UserTwo.Email,
					Owner:         UserTwo.Email,
					Granter:       UserOne.Email,
					AccessRequest: existingAR,
				},
			},
		}

		got := &service.DatasetWithAccess{}
		NewTester(t, datasetOwnerServer).Get(ctx, fmt.Sprintf("/api/datasets/%v", existingAR.DatasetID)).
			HasStatusCode(gohttp.StatusOK).
			Value(got)

		require.Len(t, got.Access, 1)
		diff := cmp.Diff(expect.Access[0], got.Access[0], cmpopts.IgnoreFields(service.Access{}, "ID", "Created", "AccessRequest"))
		assert.Empty(t, diff)
	})

	t.Run("Delete dataset access request", func(t *testing.T) {
		got := &service.UserInfo{}
		NewTester(t, accessRequesterServer).Get(ctx, "/api/userData").
			HasStatusCode(gohttp.StatusOK).
			Value(got)

		require.Len(t, got.AccessRequests, 1)

		NewTester(t, accessRequesterServer).Delete(ctx, fmt.Sprintf("/api/accessRequests/%v", existingAR.ID)).
			HasStatusCode(gohttp.StatusNoContent)

		NewTester(t, accessRequesterServer).Get(ctx, "/api/userData").
			HasStatusCode(gohttp.StatusOK).
			Value(got)

		require.Empty(t, got.AccessRequests)
	})

	t.Run("Deny dataset access request", func(t *testing.T) {
		denyReason := "you must provide a purpose for access to this dataset"
		NewTester(t, accessRequesterServer).
			Post(ctx, service.NewAccessRequestDTO{
				DatasetID:   fuelData.ID,
				Expires:     nil,
				Subject:     strToStrPtr(UserTwo.Email),
				SubjectType: strToStrPtr(service.SubjectTypeUser),
			}, "/api/accessRequests/new").
			HasStatusCode(gohttp.StatusNoContent)

		existingARs := &service.AccessRequestsWrapper{}
		NewTester(t, datasetOwnerServer).Get(ctx, "/api/accessRequests", "datasetId", fuelData.ID.String()).
			HasStatusCode(gohttp.StatusOK).
			Value(existingARs)

		ar := existingARs.AccessRequests[0]

		NewTester(t, datasetOwnerServer).Post(ctx, nil, fmt.Sprintf("/api/accessRequests/process/%v", ar.ID), "action", "deny", "reason", url.QueryEscape(denyReason)).
			HasStatusCode(gohttp.StatusNoContent)

		expect := &service.UserInfo{
			AccessRequests: []service.AccessRequest{
				{
					ID:          ar.ID,
					DatasetID:   ar.DatasetID,
					Granter:     strToStrPtr(UserOne.Email),
					Owner:       UserTwo.Email,
					Subject:     UserTwo.Email,
					SubjectType: service.SubjectTypeUser,
					Status:      service.AccessRequestStatusDenied,
					Reason:      &denyReason,
					Polly:       nil,
				},
			},
		}

		got := &service.UserInfo{}
		NewTester(t, accessRequesterServer).Get(ctx, "/api/userData").
			HasStatusCode(gohttp.StatusOK).
			Value(got)

		require.Len(t, got.AccessRequests, 1)
		diff := cmp.Diff(expect.AccessRequests[0], got.AccessRequests[0], cmpopts.IgnoreFields(service.AccessRequest{}, "Created", "Closed"))
		assert.Empty(t, diff)

		NewTester(t, accessRequesterServer).Delete(ctx, fmt.Sprintf("/api/accessRequests/%v", ar.ID)).
			HasStatusCode(gohttp.StatusNoContent)

		NewTester(t, accessRequesterServer).Get(ctx, "/api/userData").
			HasStatusCode(gohttp.StatusOK).
			Value(got)

		require.Empty(t, got.AccessRequests)
	})

	t.Run("Grant dataset access request for service account", func(t *testing.T) {
		NewTester(t, accessRequesterServer).
			Post(ctx, service.NewAccessRequestDTO{
				DatasetID:   fuelData.ID,
				Expires:     nil,
				Subject:     strToStrPtr(serviceaccountName),
				SubjectType: strToStrPtr(service.SubjectTypeServiceAccount),
				Owner:       strToStrPtr(GroupEmailAllUsers),
			}, "/api/accessRequests/new").
			HasStatusCode(gohttp.StatusNoContent)

		existingARs := &service.AccessRequestsWrapper{}
		NewTester(t, datasetOwnerServer).Get(ctx, "/api/accessRequests", "datasetId", fuelData.ID.String()).
			HasStatusCode(gohttp.StatusOK).
			Value(existingARs)

		ar := existingARs.AccessRequests[0]

		NewTester(t, datasetOwnerServer).Post(ctx, nil, fmt.Sprintf("/api/accessRequests/process/%v", ar.ID), "action", "approve").
			HasStatusCode(gohttp.StatusNoContent)

		expect := &service.UserInfo{
			AccessRequests: []service.AccessRequest{
				{
					ID:          ar.ID,
					DatasetID:   ar.DatasetID,
					Granter:     strToStrPtr(UserOne.Email),
					Owner:       GroupEmailAllUsers,
					Subject:     serviceaccountName,
					SubjectType: service.SubjectTypeServiceAccount,
					Status:      service.AccessRequestStatusApproved,
					Polly:       nil,
				},
			},
			Accessable: service.AccessibleDatasets{ // nolint: misspell
				ServiceAccountGranted: []*service.AccessibleDataset{
					{
						Subject: strToStrPtr("serviceAccount:" + serviceaccountName),
						Dataset: service.Dataset{
							ID:            fuelData.ID,
							DataproductID: fuel.ID,
						},
					},
				},
			},
		}

		got := &service.UserInfo{}
		NewTester(t, accessRequesterServer).Get(ctx, "/api/userData").
			HasStatusCode(gohttp.StatusOK).
			Value(got)

		require.Len(t, got.AccessRequests, 1)
		diff := cmp.Diff(expect.AccessRequests[0], got.AccessRequests[0], cmpopts.IgnoreFields(service.AccessRequest{}, "Created", "Closed"))
		assert.Empty(t, diff)

		require.Len(t, got.Accessable.ServiceAccountGranted, 1)
		assert.Equal(t, *expect.Accessable.ServiceAccountGranted[0].Subject, *got.Accessable.ServiceAccountGranted[0].Subject)
		assert.Equal(t, expect.Accessable.ServiceAccountGranted[0].DataproductID, got.Accessable.ServiceAccountGranted[0].DataproductID)
		assert.Equal(t, expect.Accessable.ServiceAccountGranted[0].ID, got.Accessable.ServiceAccountGranted[0].ID)
	})

	t.Run("Verify only relevant accesses are returned from dataset api", func(t *testing.T) {
		expected := []*service.Access{
			{
				Subject:   "user:" + UserTwoEmail,
				Owner:     UserTwoEmail,
				Granter:   UserOneEmail,
				DatasetID: fuelData.ID,
			},
		}

		got := &service.DatasetWithAccess{}
		NewTester(t, accessRequesterServer).Get(ctx, fmt.Sprintf("/api/datasets/%v", fuelData.ID)).
			HasStatusCode(gohttp.StatusOK).
			Value(got)

		require.Len(t, got.Access, 1)
		diff := cmp.Diff(expected, got.Access, cmpopts.IgnoreFields(service.Access{}, "ID", "Created", "Expires", "Revoked", "AccessRequest"))
		assert.Empty(t, diff)

		expected = []*service.Access{
			{
				Subject:   "serviceAccount:" + serviceaccountName,
				Owner:     GroupEmailAllUsers,
				Granter:   UserOneEmail,
				DatasetID: fuelData.ID,
			},
			{
				Subject:   "user:" + UserTwoEmail,
				Owner:     UserTwoEmail,
				Granter:   UserOneEmail,
				DatasetID: fuelData.ID,
			},
		}

		NewTester(t, datasetOwnerServer).Get(ctx, fmt.Sprintf("/api/datasets/%v", fuelData.ID)).
			HasStatusCode(gohttp.StatusOK).
			Value(got)

		require.Len(t, got.Access, 2)
		diff = cmp.Diff(expected, got.Access, cmpopts.IgnoreFields(service.Access{}, "ID", "Created", "Expires", "Revoked", "AccessRequest"))
		assert.Empty(t, diff)
	})

	t.Run("Revoke access grants as dataproduct owner", func(t *testing.T) {
		got := &service.UserInfo{}
		NewTester(t, accessRequesterServer).Get(ctx, "/api/userData").
			HasStatusCode(gohttp.StatusOK).
			Value(got)

		require.Len(t, got.Accessable.ServiceAccountGranted, 1)
		require.NotNil(t, got.Accessable.ServiceAccountGranted[0].ID)
		accessID := *got.Accessable.ServiceAccountGranted[0].AccessID

		NewTester(t, datasetOwnerServer).
			Post(ctx, nil, fmt.Sprintf("/api/accesses/revoke?accessId=%s", accessID)).
			HasStatusCode(gohttp.StatusNoContent)

		require.Len(t, got.Accessable.Granted, 1)
		require.NotNil(t, got.Accessable.Granted[0].ID)
		accessID = *got.Accessable.Granted[0].AccessID

		NewTester(t, datasetOwnerServer).
			Post(ctx, nil, fmt.Sprintf("/api/accesses/revoke?accessId=%s", accessID)).
			HasStatusCode(gohttp.StatusNoContent)

		got = &service.UserInfo{}
		NewTester(t, accessRequesterServer).Get(ctx, "/api/userData").
			HasStatusCode(gohttp.StatusOK).
			Value(got)

		require.Len(t, got.Accessable.ServiceAccountGranted, 0)
		require.Len(t, got.Accessable.Granted, 0)
	})

	t.Run("Verify users can only revoke grants where they are access owners", func(t *testing.T) {
		got := &service.UserInfo{}
		NewTester(t, accessRequesterServer).Get(ctx, "/api/userData").
			HasStatusCode(gohttp.StatusOK).Value(got)

		require.Len(t, got.Accessable.Owned, 0)
		require.Len(t, got.Accessable.ServiceAccountGranted, 0)
		require.Len(t, got.Accessable.Granted, 0)

		NewTester(t, datasetOwnerServer).
			Post(ctx, service.GrantAccessData{
				DatasetID:   fuelData.ID,
				Subject:     strToStrPtr(GroupEmailAllUsers),
				SubjectType: strToStrPtr(service.SubjectTypeGroup),
				Owner:       strToStrPtr(GroupEmailAllUsers),
			}, "/api/accesses/grant").
			HasStatusCode(gohttp.StatusNoContent)

		NewTester(t, datasetOwnerServer).
			Post(ctx, service.GrantAccessData{
				DatasetID:   fuelData.ID,
				Subject:     strToStrPtr(serviceaccountName),
				SubjectType: strToStrPtr(service.SubjectTypeServiceAccount),
				Owner:       strToStrPtr(GroupEmailAllUsers),
			}, "/api/accesses/grant").
			HasStatusCode(gohttp.StatusNoContent)

		NewTester(t, accessRequesterServer).Get(ctx, "/api/userData").
			HasStatusCode(gohttp.StatusOK).Value(got)

		require.Len(t, got.Accessable.Owned, 0)

		require.Len(t, got.Accessable.Granted, 1)
		require.NotNil(t, got.Accessable.Granted[0].Subject)
		assert.Equal(t, fmt.Sprintf("%s:%s", service.SubjectTypeGroup, GroupEmailAllUsers), *got.Accessable.Granted[0].Subject)
		require.NotNil(t, got.Accessable.Granted[0].AccessID)
		groupGrantedAccessID := *got.Accessable.Granted[0].AccessID

		require.Len(t, got.Accessable.ServiceAccountGranted, 1)
		require.NotNil(t, got.Accessable.ServiceAccountGranted[0].Subject)
		assert.Equal(t, fmt.Sprintf("%s:%s", service.SubjectTypeServiceAccount, serviceaccountName), *got.Accessable.ServiceAccountGranted[0].Subject)
		require.NotNil(t, got.Accessable.ServiceAccountGranted[0].AccessID)
		serviceAccountGrantedAccessID := *got.Accessable.ServiceAccountGranted[0].AccessID

		NewTester(t, accessRequesterServer).
			Post(ctx, nil, fmt.Sprintf("/api/accesses/revoke?accessId=%s", groupGrantedAccessID)).
			HasStatusCode(gohttp.StatusNoContent)

		NewTester(t, accessRequesterServer).
			Post(ctx, nil, fmt.Sprintf("/api/accesses/revoke?accessId=%s", serviceAccountGrantedAccessID)).
			HasStatusCode(gohttp.StatusNoContent)

		got = &service.UserInfo{}
		NewTester(t, accessRequesterServer).Get(ctx, "/api/userData").
			HasStatusCode(gohttp.StatusOK).Value(got)

		require.Len(t, got.Accessable.Owned, 0)
		require.Len(t, got.Accessable.ServiceAccountGranted, 0)
		require.Len(t, got.Accessable.Granted, 0)
	})
}

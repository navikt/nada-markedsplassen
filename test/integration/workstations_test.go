package integration

import (
	"cloud.google.com/go/artifactregistry/apiv1/artifactregistrypb"
	"cloud.google.com/go/iam/apiv1/iampb"
	"context"
	"fmt"
	"github.com/navikt/nada-backend/pkg/artifactregistry"
	gohttp "net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/navikt/nada-backend/pkg/worker"

	"github.com/navikt/nada-backend/pkg/database"
	riverstore "github.com/navikt/nada-backend/pkg/service/core/storage/river"
	riverapi "github.com/riverqueue/river"
	"google.golang.org/api/iam/v1"

	logpb "cloud.google.com/go/logging/apiv2/loggingpb"
	"github.com/navikt/nada-backend/pkg/cloudlogging"
	"github.com/stretchr/testify/require"
	ltype "google.golang.org/genproto/googleapis/logging/type"
	"google.golang.org/protobuf/types/known/timestamppb"

	"cloud.google.com/go/compute/apiv1/computepb"

	artifactRegistryEmulator "github.com/navikt/nada-backend/pkg/artifactregistry/emulator"
	cloudLoggingEmulator "github.com/navikt/nada-backend/pkg/cloudlogging/emulator"
	crm "github.com/navikt/nada-backend/pkg/cloudresourcemanager"
	crmEmulator "github.com/navikt/nada-backend/pkg/cloudresourcemanager/emulator"
	"github.com/navikt/nada-backend/pkg/computeengine"
	computeEmulator "github.com/navikt/nada-backend/pkg/computeengine/emulator"

	"google.golang.org/api/cloudresourcemanager/v1"

	"cloud.google.com/go/workstations/apiv1/workstationspb"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/navikt/nada-backend/pkg/sa"
	serviceAccountEmulator "github.com/navikt/nada-backend/pkg/sa/emulator"
	"github.com/navikt/nada-backend/pkg/securewebproxy"
	secureWebProxyEmulator "github.com/navikt/nada-backend/pkg/securewebproxy/emulator"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/service/core"
	"github.com/navikt/nada-backend/pkg/service/core/api/gcp"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/routes"
	"github.com/navikt/nada-backend/pkg/workstations"
	"github.com/navikt/nada-backend/pkg/workstations/emulator"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestWorkstations(t *testing.T) {
	t.Parallel()

	log := zerolog.New(zerolog.NewConsoleWriter()).Level(zerolog.InfoLevel)
	ctx := context.Background()
	e := emulator.New(log)
	apiURL := e.Run()
	project := "test"
	location := "europe-north1"
	repository := "knast-images"
	clusterID := "clusterID"
	slug := UserOneIdent

	c := NewContainers(t, log)
	defer c.Cleanup()

	pgCfg := c.RunPostgres(NewPostgresConfig())

	repo, err := database.New(
		pgCfg.ConnectionURL(),
		10,
		10,
	)
	assert.NoError(t, err)

	client := workstations.New(project, location, clusterID, apiURL, true)

	saEmulator := serviceAccountEmulator.New(log)
	saURL := saEmulator.Run()
	saClient := sa.NewClient(saURL, true)
	saEmulator.SetIamPolicy(
		fmt.Sprintf("%s@%s.iam.gserviceaccount.com", service.WorkstationServiceAccountID(slug), project),
		&iam.Policy{},
	)

	crmEmulator := crmEmulator.New(log)
	crmEmulator.SetPolicy(project, &cloudresourcemanager.Policy{
		Bindings: []*cloudresourcemanager.Binding{
			{
				Role:    "roles/owner",
				Members: []string{fmt.Sprintf("user:%s", UserOne.Email)},
			},
		},
	})
	crmURL := crmEmulator.Run()
	tagBindingClient := crmEmulator.TagBindingPolicyClient(
		[]string{"europe-north1-a", "europe-north1-b"},
		gohttp.StatusOK,
		log,
	)
	crmClient := crm.NewClient(crmURL, true, tagBindingClient)

	swpEmulator := secureWebProxyEmulator.New(log)
	swpURL := swpEmulator.Run()
	swpClient := securewebproxy.New(swpURL, true)

	arEmulator := artifactRegistryEmulator.New(log)
	arEmulator.SetImages(map[string][]*artifactregistrypb.DockerImage{
		fmt.Sprintf("projects/%s/locations/%s/repositories/%s", project, location, repository): {
			{
				Name: "nginx",
				Uri:  "eu-north1-docker.pkg.dev/test/test/nginx@sha256:e9954c1fc875017be1c3e36eca16be2d9e9bccc4bf072163515467d6a823c7cf",
				Tags: []string{"latest"},
			},
			{
				Name: "something",
				Uri:  "eu-north1-docker.pkg.dev/test/test/something@sha256:e9954c1fc875017be1c3e36eca16be2d9e9bccc4bf072163515467d6a823c7cf",
				Tags: []string{"v1"},
			},
		},
	})
	arEmulator.SetIamPolicy(repository, &iampb.Policy{})
	arURL := arEmulator.Run()
	arClient := artifactregistry.New(arURL, true)

	ceEmulator := computeEmulator.New(log)
	ceEmulator.SetInstances(map[string][]*computepb.Instance{
		"europe-north1-a": {
			{
				Name: strToStrPtr("instance-1"),
				Labels: map[string]string{
					service.WorkstationConfigIDLabel: slug,
				},
				Id: uint64Ptr(12345),
			},
		},
	})
	ceEmulator.SetFirewallPolicies(map[string]*computepb.FirewallPolicy{
		"test-europe-north1-onprem-access": {
			Name: strToStrPtr("onprem-access"),
			Rules: []*computepb.FirewallPolicyRule{
				{
					RuleName: strToStrPtr("rule-1"),
					TargetSecureTags: []*computepb.FirewallPolicyRuleSecureTag{
						{
							Name: strToStrPtr("test-project/my-resource-tag/my-resource-tag"),
						},
					},
				},
			},
		},
	})

	clEmulator := cloudLoggingEmulator.NewEmulator(log)
	clEmulator.AppendLogEntries([]*logpb.LogEntry{
		{
			LogName: "ehh",
			HttpRequest: &ltype.HttpRequest{
				RequestMethod: gohttp.MethodGet,
				RequestUrl:    "https://github.com/navikt",
				UserAgent:     "Chrome 91",
			},
			Timestamp: timestamppb.Now(),
		},
	})
	clURL, err := clEmulator.Start()
	require.NoError(t, err)
	clClient := cloudlogging.NewClient(clURL, true)
	clAPI := gcp.NewCloudLoggingAPI(clClient)

	computeURL := ceEmulator.Run()
	computeClient := computeengine.NewClient(computeURL, true)
	computeAPI := gcp.NewComputeAPI(Project, computeClient)

	workers := riverapi.NewWorkers()

	config := worker.WorkstationConfig(&log, workers)
	config.TestOnly = true

	workstationsStorage := riverstore.NewWorkstationsStorage(config, repo)

	arAPI := gcp.NewArtifactRegistryAPI(arClient, log)

	router := TestRouter(log)
	crmAPI := gcp.NewCloudResourceManagerAPI(crmClient)
	gAPI := gcp.NewWorkstationsAPI(client)
	saAPI := gcp.NewServiceAccountAPI(saClient)
	swpAPI := gcp.NewSecureWebProxyAPI(swpClient)
	workstationService := core.NewWorkstationService(
		project,
		project,
		location,
		"my-policy",
		"onprem-access",
		"my-bucket",
		"my-view",
		fmt.Sprintf("admin-sa@%s.iam.gserviceaccount.com", project),
		repository,
		project,
		saAPI,
		crmAPI,
		swpAPI,
		gAPI,
		computeAPI,
		clAPI,
		workstationsStorage,
		arAPI,
	)

	{
		h := handlers.NewWorkstationsHandler(workstationService)
		e := routes.NewWorkstationsEndpoints(log, h)
		f := routes.NewWorkstationsRoutes(e, injectUser(UserOne))
		f(router)
	}

	workstationWorker, err := worker.NewWorkstationWorker(config, workstationService, repo)
	require.NoError(t, err)
	err = workstationWorker.Start(ctx)
	require.NoError(t, err)
	defer workstationWorker.Stop(ctx)

	server := httptest.NewServer(router)

	fullyQualifiedConfigName := fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s/workstationConfigs/%s", project, location, clusterID, slug)

	workstation := &service.WorkstationOutput{}

	t.Run("Get workstation options", func(t *testing.T) {
		expected := &service.WorkstationOptions{
			MachineTypes: service.WorkstationMachineTypes(),
			ContainerImages: []*service.WorkstationContainer{
				{
					Image:       "eu-north1-docker.pkg.dev/test/test/nginx:latest",
					Description: "nginx",
					Labels:      map[string]string{},
				},
			},
			FirewallTags: []*service.FirewallTag{
				{
					Name:      "rule-1",
					SecureTag: "test-project/my-resource-tag/my-resource-tag",
				},
			},
		}

		got := &service.WorkstationOptions{}

		NewTester(t, server).
			Get(ctx, "/api/workstations/options").
			HasStatusCode(gohttp.StatusOK).
			Expect(expected, got)
	})

	t.Run("Create workstation", func(t *testing.T) {
		expectedJob := &service.WorkstationJob{
			ID:              1,
			Name:            "User Userson",
			Email:           "user.userson@email.com",
			Ident:           "v101010",
			MachineType:     service.MachineTypeN2DStandard16,
			ContainerImage:  service.ContainerImageVSCode,
			URLAllowList:    []string{"github.com/navikt"},
			OnPremAllowList: nil,
			State:           service.WorkstationJobStateRunning,
			Duplicate:       false,
			Errors:          []string{},
		}

		subscribeChan, subscribeCancel := workstationWorker.Subscribe(riverapi.EventKindJobCompleted)
		go func() {
			time.Sleep(5 * time.Second)
			subscribeCancel()
		}()

		job := &service.WorkstationJob{}

		NewTester(t, server).
			Post(ctx, service.WorkstationInput{
				MachineType:    service.MachineTypeN2DStandard16,
				ContainerImage: service.ContainerImageVSCode,
				URLAllowList:   []string{"github.com/navikt"},
			}, "/api/workstations/job").
			HasStatusCode(gohttp.StatusAccepted).
			Expect(expectedJob, job, cmpopts.IgnoreFields(service.WorkstationJob{}, "StartTime"))

		event := <-subscribeChan
		assert.Equal(t, riverapi.EventKindJobCompleted, event.Kind)

		expectedJob.State = service.WorkstationJobStateCompleted

		NewTester(t, server).
			Get(ctx, fmt.Sprintf("/api/workstations/job/%d", job.ID)).
			HasStatusCode(gohttp.StatusOK).
			Expect(expectedJob, job, cmpopts.IgnoreFields(service.WorkstationJob{}, "StartTime"))

		expectedWorkstation := &service.WorkstationOutput{
			Slug:         slug,
			DisplayName:  "User Userson (user.userson@email.com)",
			State:        service.Workstation_STATE_STARTING,
			URLAllowList: []string{"github.com/navikt"},
			Config: &service.WorkstationConfigOutput{
				MachineType:            service.MachineTypeN2DStandard16,
				Image:                  service.ContainerImageVSCode,
				IdleTimeout:            2 * time.Hour,
				RunningTimeout:         12 * time.Hour,
				FirewallRulesAllowList: []string{},
				Env:                    map[string]string{"WORKSTATION_NAME": slug},
			},
		}

		workstation = &service.WorkstationOutput{}
		NewTester(t, server).
			Get(ctx, "/api/workstations/").
			HasStatusCode(gohttp.StatusOK).
			Expect(expectedWorkstation, workstation, cmpopts.IgnoreFields(service.WorkstationOutput{}, "CreateTime", "StartTime", "Config.CreateTime"))
	})

	t.Run("Change running state", func(t *testing.T) {
		e.SetWorkstationState(fullyQualifiedConfigName, slug, workstationspb.Workstation_STATE_RUNNING)

		expected := &service.WorkstationOutput{
			Slug:        slug,
			DisplayName: "User Userson (user.userson@email.com)",
			Reconciling: false,
			UpdateTime:  nil,
			StartTime:   nil,
			State:       service.Workstation_STATE_RUNNING,
			Config: &service.WorkstationConfigOutput{
				UpdateTime:             nil,
				IdleTimeout:            2 * time.Hour,
				RunningTimeout:         12 * time.Hour,
				MachineType:            service.MachineTypeN2DStandard16,
				FirewallRulesAllowList: []string{},
				Image:                  service.ContainerImageVSCode,
				Env:                    map[string]string{"WORKSTATION_NAME": slug},
			},
			URLAllowList: []string{"github.com/navikt"},
		}

		NewTester(t, server).
			Get(ctx, "/api/workstations/").Debug(os.Stdout).
			HasStatusCode(gohttp.StatusOK).
			Expect(expected, workstation, cmpopts.IgnoreFields(service.WorkstationOutput{}, "CreateTime", "StartTime", "Config.CreateTime"))
		assert.NotNil(t, workstation.StartTime)
	})

	t.Run("Get workstation logs", func(t *testing.T) {
		expected := &service.WorkstationLogs{
			ProxyDeniedHostPaths: []*service.LogEntry{
				{
					Timestamp: time.Time{},
					HTTPRequest: &service.HTTPRequest{
						URL: &url.URL{
							Host: "github.com",
							Path: "/navikt",
						},
						Method:    gohttp.MethodGet,
						UserAgent: "Chrome 91",
					},
				},
			},
		}

		got := &service.WorkstationLogs{}

		NewTester(t, server).
			Get(ctx, "/api/workstations/logs").
			HasStatusCode(gohttp.StatusOK).
			Expect(expected, got, cmpopts.IgnoreFields(service.LogEntry{}, "Timestamp", "HTTPRequest.URL.Scheme"))
	})

	t.Run("Update workstation", func(t *testing.T) {
		expectedJob := &service.WorkstationJob{
			ID:              2,
			Name:            "User Userson",
			Email:           "user.userson@email.com",
			Ident:           "v101010",
			MachineType:     service.MachineTypeN2DStandard32,
			ContainerImage:  service.ContainerImageIntellijUltimate,
			URLAllowList:    []string{"github.com/navikt"},
			OnPremAllowList: []string{"rule-1"},
			State:           service.WorkstationJobStateRunning,
			Duplicate:       false,
			Errors:          []string{},
		}

		subscribeChan, subscribeCancel := workstationWorker.Subscribe(riverapi.EventKindJobCompleted)
		go func() {
			time.Sleep(5 * time.Second)
			subscribeCancel()
		}()

		job := &service.WorkstationJob{}

		NewTester(t, server).
			Post(ctx, service.WorkstationInput{
				MachineType:     service.MachineTypeN2DStandard32,
				ContainerImage:  service.ContainerImageIntellijUltimate,
				URLAllowList:    []string{"github.com/navikt"},
				OnPremAllowList: []string{"rule-1"},
			}, "/api/workstations/job").
			HasStatusCode(gohttp.StatusAccepted).
			Expect(expectedJob, job, cmpopts.IgnoreFields(service.WorkstationJob{}, "StartTime"))

		event := <-subscribeChan
		assert.Equal(t, riverapi.EventKindJobCompleted, event.Kind)

		expectedJob.State = service.WorkstationJobStateCompleted

		NewTester(t, server).
			Get(ctx, fmt.Sprintf("/api/workstations/job/%d", job.ID)).
			HasStatusCode(gohttp.StatusOK).
			Expect(expectedJob, job, cmpopts.IgnoreFields(service.WorkstationJob{}, "StartTime"))
	})

	t.Run("Get updated workstation", func(t *testing.T) {
		expected := &service.WorkstationOutput{
			Slug:        slug,
			DisplayName: "User Userson (user.userson@email.com)",
			Reconciling: false,
			StartTime:   nil,
			State:       service.Workstation_STATE_RUNNING,
			Config: &service.WorkstationConfigOutput{
				UpdateTime:     nil,
				IdleTimeout:    2 * time.Hour,
				RunningTimeout: 12 * time.Hour,
				MachineType:    service.MachineTypeN2DStandard32,
				FirewallRulesAllowList: []string{
					"rule-1",
				},
				Image: service.ContainerImageIntellijUltimate,
				Env:   map[string]string{"WORKSTATION_NAME": slug},
			},
			URLAllowList: []string{"github.com/navikt"},
		}

		NewTester(t, server).
			Get(ctx, "/api/workstations/").Debug(os.Stdout).
			HasStatusCode(gohttp.StatusOK).
			Expect(expected, workstation, cmpopts.IgnoreFields(service.WorkstationOutput{}, "CreateTime", "StartTime", "UpdateTime", "Config.CreateTime", "Config.UpdateTime"))
		assert.NotNil(t, workstation.StartTime)
	})

	t.Run("Update URLList", func(t *testing.T) {
		expected := &service.WorkstationOutput{
			Slug:        slug,
			DisplayName: "User Userson (user.userson@email.com)",
			Reconciling: false,
			StartTime:   nil,
			State:       service.Workstation_STATE_RUNNING,
			Config: &service.WorkstationConfigOutput{
				UpdateTime:     nil,
				IdleTimeout:    2 * time.Hour,
				RunningTimeout: 12 * time.Hour,
				MachineType:    service.MachineTypeN2DStandard32,
				Image:          service.ContainerImageIntellijUltimate,
				FirewallRulesAllowList: []string{
					"rule-1",
				},
				Env: map[string]string{"WORKSTATION_NAME": slug},
			},
			URLAllowList: []string{"github.com/navikt", "github.com/navikt2"},
		}

		NewTester(t, server).
			Put(ctx, &service.WorkstationURLList{URLAllowList: []string{"github.com/navikt", "github.com/navikt2"}}, "/api/workstations/urllist").Debug(os.Stdout).
			HasStatusCode(gohttp.StatusNoContent)

		NewTester(t, server).
			Get(ctx, "/api/workstations/").Debug(os.Stdout).
			HasStatusCode(gohttp.StatusOK).
			Expect(expected, workstation, cmpopts.IgnoreFields(service.WorkstationOutput{}, "CreateTime", "StartTime", "UpdateTime", "Config.CreateTime", "Config.UpdateTime"))
		assert.NotNil(t, workstation.StartTime)
	})

	t.Run("Start workstation", func(t *testing.T) {
		NewTester(t, server).
			Post(ctx, nil, "/api/workstations/start").
			HasStatusCode(gohttp.StatusNoContent)
	})

	t.Run("Stop workstation", func(t *testing.T) {
		NewTester(t, server).
			Post(ctx, nil, "/api/workstations/stop").
			HasStatusCode(gohttp.StatusNoContent)
	})

	t.Run("Delete workstation", func(t *testing.T) {
		NewTester(t, server).
			Delete(ctx, "/api/workstations/").
			HasStatusCode(gohttp.StatusNoContent)
	})

	t.Run("Start workstation that does not exist", func(t *testing.T) {
		NewTester(t, server).
			Post(ctx, nil, "/api/workstations/start").
			HasStatusCode(gohttp.StatusNotFound)
	})

	t.Run("Stop workstation that does not exist", func(t *testing.T) {
		NewTester(t, server).
			Post(ctx, nil, "/api/workstations/stop").
			HasStatusCode(gohttp.StatusNotFound)
	})

	t.Run("Get workstation that does not exist", func(t *testing.T) {
		NewTester(t, server).
			Get(ctx, "/api/workstations/").
			HasStatusCode(gohttp.StatusNotFound)
		assert.Empty(t, e.GetWorkstationConfigs())
		assert.Empty(t, e.GetWorkstations())
	})

	t.Run("Get workstation jobs for user", func(t *testing.T) {
		expected := &service.WorkstationJobs{
			Jobs: []*service.WorkstationJob{
				{
					ID:              2,
					Name:            "User Userson",
					Email:           "user.userson@email.com",
					Ident:           "v101010",
					MachineType:     service.MachineTypeN2DStandard32,
					ContainerImage:  service.ContainerImageIntellijUltimate,
					URLAllowList:    []string{"github.com/navikt"},
					OnPremAllowList: []string{"rule-1"},
					State:           service.WorkstationJobStateCompleted,
					Errors:          []string{},
					Diff: map[string]*service.Diff{
						service.WorkstationDiffMachineType: {
							Value: "n2d-standard-32",
						},
						service.WorkstationDiffContainerImage: {
							Value: "europe-north1-docker.pkg.dev/cloud-workstations-images/predefined/intellij-ultimate:latest",
						},
						service.WorkstationDiffOnPremAllowList: {
							Added: []string{"rule-1"},
						},
					},
				},
				{
					ID:             1,
					Name:           "User Userson",
					Email:          "user.userson@email.com",
					Ident:          "v101010",
					MachineType:    service.MachineTypeN2DStandard16,
					ContainerImage: service.ContainerImageVSCode,
					URLAllowList:   []string{"github.com/navikt"},
					State:          service.WorkstationJobStateCompleted,
					Errors:         []string{},
				},
			},
		}

		got := &service.WorkstationJobs{}

		NewTester(t, server).
			Get(ctx, "/api/workstations/job").
			HasStatusCode(gohttp.StatusOK).
			Expect(expected, got, cmpopts.IgnoreFields(service.WorkstationJob{}, "StartTime"))
	})
}

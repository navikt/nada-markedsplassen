package integration

import (
	"context"
	"fmt"
	gohttp "net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	logpb "cloud.google.com/go/logging/apiv2/loggingpb"
	"github.com/navikt/nada-backend/pkg/cloudlogging"
	"github.com/stretchr/testify/require"
	ltype "google.golang.org/genproto/googleapis/logging/type"
	"google.golang.org/protobuf/types/known/timestamppb"

	"cloud.google.com/go/compute/apiv1/computepb"

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

	log := zerolog.New(zerolog.NewConsoleWriter())
	ctx := context.Background()
	e := emulator.New(log)
	apiURL := e.Run()
	project := "test"
	location := "europe-north1"
	clusterID := "clusterID"
	slug := UserOneIdent

	client := workstations.New(project, location, clusterID, apiURL, true)

	saEmulator := serviceAccountEmulator.New(log)
	saURL := saEmulator.Run()
	saClient := sa.NewClient(saURL, true)

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

	router := TestRouter(log)
	{
		crmAPI := gcp.NewCloudResourceManagerAPI(crmClient)
		gAPI := gcp.NewWorkstationsAPI(client)
		saAPI := gcp.NewServiceAccountAPI(saClient)
		swpAPI := gcp.NewSecureWebProxyAPI(swpClient)
		s := core.NewWorkstationService(project, project, location, "my-policy", "onprem-access", "my-bucket", "my-view", saAPI, crmAPI, swpAPI, gAPI, computeAPI, clAPI)
		h := handlers.NewWorkstationsHandler(s)
		e := routes.NewWorkstationsEndpoints(log, h)
		f := routes.NewWorkstationsRoutes(e, injectUser(UserOne))
		f(router)
	}
	server := httptest.NewServer(router)

	fullyQualifiedConfigName := fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s/workstationConfigs/%s", project, location, clusterID, slug)

	workstation := &service.WorkstationOutput{}

	t.Run("Get workstation options", func(t *testing.T) {
		expected := &service.WorkstationOptions{
			MachineTypes:    service.WorkstationMachineTypes(),
			ContainerImages: service.WorkstationContainers(),
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
		expected := &service.WorkstationOutput{
			Slug:        slug,
			DisplayName: "User Userson (user.userson@email.com)",
			Reconciling: false,
			UpdateTime:  nil,
			StartTime:   nil,
			State:       service.Workstation_STATE_STARTING,
			Config: &service.WorkstationConfigOutput{
				UpdateTime:     nil,
				IdleTimeout:    2 * time.Hour,
				RunningTimeout: 12 * time.Hour,
				MachineType:    service.MachineTypeN2DStandard16,
				Image:          service.ContainerImageVSCode,
				Env:            map[string]string{"WORKSTATION_NAME": slug},
			},
			URLAllowList: []string{"github.com/navikt"},
		}

		NewTester(t, server).
			Post(ctx, service.WorkstationInput{
				MachineType:    service.MachineTypeN2DStandard16,
				ContainerImage: service.ContainerImageVSCode,
				URLAllowList:   []string{"github.com/navikt"},
			}, "/api/workstations/").
			HasStatusCode(gohttp.StatusOK).
			Expect(expected, workstation, cmpopts.IgnoreFields(service.WorkstationOutput{}, "CreateTime", "Config.CreateTime"))
		assert.NotNil(t, workstation.CreateTime)
		assert.NotNil(t, workstation.Config.CreateTime)
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
		expected := &service.WorkstationOutput{
			Slug:        slug,
			DisplayName: "User Userson (user.userson@email.com)",
			Reconciling: false,
			State:       service.Workstation_STATE_RUNNING,
			Config: &service.WorkstationConfigOutput{
				UpdateTime:     nil,
				IdleTimeout:    2 * time.Hour,
				RunningTimeout: 12 * time.Hour,
				MachineType:    service.MachineTypeN2DStandard32,
				Image:          service.ContainerImageIntellijUltimate,
				Env:            map[string]string{"WORKSTATION_NAME": slug},
			},
			URLAllowList: []string{"github.com/navikt"},
		}

		NewTester(t, server).
			Post(ctx, service.WorkstationInput{
				MachineType:    service.MachineTypeN2DStandard32,
				ContainerImage: service.ContainerImageIntellijUltimate,
				URLAllowList:   []string{"github.com/navikt"},
			}, "/api/workstations/").
			HasStatusCode(gohttp.StatusOK).
			Expect(expected, workstation, cmpopts.IgnoreFields(service.WorkstationOutput{}, "CreateTime", "UpdateTime", "StartTime", "Config.CreateTime", "Config.UpdateTime"))
		assert.NotNil(t, workstation.StartTime)
		assert.NotNil(t, workstation.UpdateTime)
		assert.NotNil(t, workstation.Config.UpdateTime)
	})

	t.Run("Get updated workstation", func(t *testing.T) {
		expected := &service.WorkstationOutput{
			Slug:        slug,
			DisplayName: "User Userson (user.userson@email.com)",
			Reconciling: false,
			StartTime:   nil,
			State:       service.Workstation_STATE_RUNNING,
			Config: &service.WorkstationConfigOutput{
				UpdateTime:             nil,
				IdleTimeout:            2 * time.Hour,
				RunningTimeout:         12 * time.Hour,
				MachineType:            service.MachineTypeN2DStandard32,
				FirewallRulesAllowList: []string{},
				Image:                  service.ContainerImageIntellijUltimate,
				Env:                    map[string]string{"WORKSTATION_NAME": slug},
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
				UpdateTime:             nil,
				IdleTimeout:            2 * time.Hour,
				RunningTimeout:         12 * time.Hour,
				MachineType:            service.MachineTypeN2DStandard32,
				Image:                  service.ContainerImageIntellijUltimate,
				FirewallRulesAllowList: []string{},
				Env:                    map[string]string{"WORKSTATION_NAME": slug},
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
}

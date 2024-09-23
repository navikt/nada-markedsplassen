package integration

import (
	"context"
	"fmt"
	gohttp "net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/workstations/apiv1/workstationspb"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/navikt/nada-backend/pkg/sa"
	serviceAccountEmulator "github.com/navikt/nada-backend/pkg/sa/emulator"
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
	slug := "user-userson-at-email-com"

	client := workstations.New(project, location, clusterID, apiURL, true)

	saEmulator := serviceAccountEmulator.New(log)
	saURL := saEmulator.Run()
	saClient := sa.NewClient(saURL, true)
	router := TestRouter(log)
	{
		gAPI := gcp.NewWorkstationsAPI(client)
		saAPI := gcp.NewServiceAccountAPI(saClient)
		s := core.NewWorkstationService(project, saAPI, gAPI)
		h := handlers.NewWorkstationsHandler(s)
		e := routes.NewWorkstationsEndpoints(log, h)
		f := routes.NewWorkstationsRoutes(e, injectUser(UserOne))
		f(router)
	}
	server := httptest.NewServer(router)

	fullyQualifiedConfigName := fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s/workstationConfigs/%s", project, location, clusterID, slug)

	workstation := &service.WorkstationOutput{}

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
		}

		NewTester(t, server).
			Post(ctx, service.WorkstationInput{
				MachineType:    service.MachineTypeN2DStandard16,
				ContainerImage: service.ContainerImageVSCode,
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
				UpdateTime:     nil,
				IdleTimeout:    2 * time.Hour,
				RunningTimeout: 12 * time.Hour,
				MachineType:    service.MachineTypeN2DStandard16,
				Image:          service.ContainerImageVSCode,
				Env:            map[string]string{"WORKSTATION_NAME": slug},
			},
		}

		NewTester(t, server).
			Get(ctx, "/api/workstations/").Debug(os.Stdout).
			HasStatusCode(gohttp.StatusOK).
			Expect(expected, workstation, cmpopts.IgnoreFields(service.WorkstationOutput{}, "CreateTime", "StartTime", "Config.CreateTime"))
		assert.NotNil(t, workstation.StartTime)
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
		}

		NewTester(t, server).
			Post(ctx, service.WorkstationInput{
				MachineType:    service.MachineTypeN2DStandard32,
				ContainerImage: service.ContainerImageIntellijUltimate,
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
				UpdateTime:     nil,
				IdleTimeout:    2 * time.Hour,
				RunningTimeout: 12 * time.Hour,
				MachineType:    service.MachineTypeN2DStandard32,
				Image:          service.ContainerImageIntellijUltimate,
				Env:            map[string]string{"WORKSTATION_NAME": slug},
			},
		}

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
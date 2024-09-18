package integration

import (
	"context"
	gohttp "net/http"
	"net/http/httptest"
	"testing"

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

	t.Run("Create workstation", func(t *testing.T) {
		got := &service.WorkstationOutput{}
		expected := &service.WorkstationOutput{
			Workstation:       &service.Workstation{Name: ""},
			WorkstationConfig: &service.WorkstationConfig{Name: ""},
		}

		NewTester(t, server).
			Post(ctx, service.WorkstationInput{
				MachineType:    service.MachineTypeN2DStandard16,
				ContainerImage: service.ContainerImageVSCode,
			}, "/api/workstations/").
			HasStatusCode(gohttp.StatusOK).Expect(expected, got)
	})
}

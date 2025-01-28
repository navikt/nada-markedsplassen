package workstations_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"cloud.google.com/go/iam/apiv1/iampb"
	"cloud.google.com/go/workstations/apiv1/workstationspb"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/navikt/nada-backend/pkg/workstations"
	"github.com/navikt/nada-backend/pkg/workstations/emulator"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkstationOperations(t *testing.T) {
	t.Parallel()

	log := zerolog.New(zerolog.NewConsoleWriter())
	ctx := context.Background()
	e := emulator.New(log)
	apiURL := e.Run()

	project := "test"
	location := "europe-north1"
	clusterID := "clusterID"
	workstationHost := "ident.workstations.domain"
	configSlug := "nada"
	configDisplayName := "Team nada"
	fullyQualifiedWorkstationCfgName := fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s/workstationConfigs/%s", project, location, clusterID, configSlug)
	saEmail := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", configSlug, project)
	client := workstations.New(project, location, clusterID, apiURL, true, http.DefaultClient)

	workstationConfig := &workstations.WorkstationConfig{
		Slug:               configSlug,
		FullyQualifiedName: fullyQualifiedWorkstationCfgName,
		DisplayName:        configDisplayName,
		Annotations:        map[string]string{"onprem-allow-list": "host1,host2"},
		MachineType:        workstations.MachineTypeN2DStandard2,
		ServiceAccount:     saEmail,
		Image:              workstations.ContainerImageVSCode,
		IdleTimeout:        workstations.DefaultIdleTimeoutInSec * time.Second,
		RunningTimeout:     workstations.DefaultRunningTimeoutInSec * time.Second,
	}

	workstationSlug := "nada-workstation"
	workstationDisplayName := "Team nada workstation"

	workstation := &workstations.Workstation{
		Slug:               workstationSlug,
		FullyQualifiedName: fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s/workstationConfigs/%s/workstations/%s", project, location, clusterID, configSlug, workstationSlug),
		DisplayName:        workstationDisplayName,
		State:              workstations.Workstation_STATE_STARTING,
		Reconciling:        false,
		UpdateTime:         nil,
		StartTime:          nil,
		Host:               workstationHost,
	}

	t.Run("Get workstation config that does not exist", func(t *testing.T) {
		_, err := client.GetWorkstationConfig(ctx, &workstations.WorkstationConfigGetOpts{
			Slug: configSlug,
		})

		require.ErrorIs(t, err, workstations.ErrNotExist)
	})

	t.Run("Create workstation config", func(t *testing.T) {
		got, err := client.CreateWorkstationConfig(ctx, &workstations.WorkstationConfigOpts{
			Slug:                configSlug,
			DisplayName:         configDisplayName,
			Annotations:         map[string]string{"onprem-allow-list": "host1,host2"},
			MachineType:         workstations.MachineTypeN2DStandard2,
			ServiceAccountEmail: saEmail,
			SubjectEmail:        "nada@nav.no",
			ContainerImage:      workstations.ContainerImageVSCode,
		})

		require.NoError(t, err)
		diff := cmp.Diff(workstationConfig, got, cmpopts.IgnoreFields(workstations.WorkstationConfig{}, "CreateTime", "CompleteConfigAsJSON"))
		assert.Empty(t, diff)
		assert.NotNil(t, got.CreateTime)

		gotGoogleWorkstationsConfig := &workstationspb.WorkstationConfig{}
		err = protojson.UnmarshalOptions{AllowPartial: true}.Unmarshal(got.CompleteConfigAsJSON, gotGoogleWorkstationsConfig)
		require.NoError(t, err)

		assert.Equal(t, workstations.ContainerImageVSCode, gotGoogleWorkstationsConfig.Container.Image)
		assert.Equal(t, workstations.MachineTypeN2DStandard2, gotGoogleWorkstationsConfig.Host.Config.(*workstationspb.WorkstationConfig_Host_GceInstance_).GceInstance.MachineType)
	})

	t.Run("Get workstation config", func(t *testing.T) {
		got, err := client.GetWorkstationConfig(ctx, &workstations.WorkstationConfigGetOpts{
			Slug: configSlug,
		})

		require.NoError(t, err)
		diff := cmp.Diff(workstationConfig, got, cmpopts.IgnoreFields(workstations.WorkstationConfig{}, "CreateTime", "UpdateTime", "CompleteConfigAsJSON"))
		assert.Empty(t, diff)

		gotGoogleWorkstationsConfig := &workstationspb.WorkstationConfig{}
		err = protojson.UnmarshalOptions{AllowPartial: true}.Unmarshal(got.CompleteConfigAsJSON, gotGoogleWorkstationsConfig)
		require.NoError(t, err)

		assert.Equal(t, workstations.ContainerImageVSCode, gotGoogleWorkstationsConfig.Container.Image)
		assert.Equal(t, workstations.MachineTypeN2DStandard2, gotGoogleWorkstationsConfig.Host.Config.(*workstationspb.WorkstationConfig_Host_GceInstance_).GceInstance.MachineType)
	})

	t.Run("Update workstation config", func(t *testing.T) {
		workstationConfig.MachineType = workstations.MachineTypeN2DStandard32
		workstationConfig.Image = workstations.ContainerImagePosit
		workstationConfig.Annotations = map[string]string{"onprem-allow-list": "host1,host2,host3"}

		got, err := client.UpdateWorkstationConfig(ctx, &workstations.WorkstationConfigUpdateOpts{
			Slug:           configSlug,
			Annotations:    map[string]string{"onprem-allow-list": "host1,host2,host3"},
			MachineType:    workstations.MachineTypeN2DStandard32,
			ContainerImage: workstations.ContainerImagePosit,
		})

		require.NoError(t, err)
		diff := cmp.Diff(workstationConfig, got, cmpopts.IgnoreFields(workstations.WorkstationConfig{}, "CreateTime", "UpdateTime", "CompleteConfigAsJSON"))
		assert.Empty(t, diff)
		assert.NotNil(t, got.UpdateTime)

		gotGoogleWorkstationsConfig := &workstationspb.WorkstationConfig{}
		err = protojson.UnmarshalOptions{AllowPartial: true}.Unmarshal(got.CompleteConfigAsJSON, gotGoogleWorkstationsConfig)
		require.NoError(t, err)

		assert.Equal(t, workstations.ContainerImagePosit, gotGoogleWorkstationsConfig.Container.Image)
		assert.Equal(t, workstations.MachineTypeN2DStandard32, gotGoogleWorkstationsConfig.Host.Config.(*workstationspb.WorkstationConfig_Host_GceInstance_).GceInstance.MachineType)
	})

	t.Run("Get updated workstation config", func(t *testing.T) {
		got, err := client.GetWorkstationConfig(ctx, &workstations.WorkstationConfigGetOpts{
			Slug: configSlug,
		})
		workstationConfig.Annotations = map[string]string{"onprem-allow-list": "host1,host2,host3"}
		require.NoError(t, err)
		diff := cmp.Diff(workstationConfig, got, cmpopts.IgnoreFields(workstations.WorkstationConfig{}, "CreateTime", "UpdateTime", "CompleteConfigAsJSON"))
		assert.Empty(t, diff)
		assert.NotNil(t, got.UpdateTime)

		gotGoogleWorkstationsConfig := &workstationspb.WorkstationConfig{}
		err = protojson.UnmarshalOptions{AllowPartial: true}.Unmarshal(got.CompleteConfigAsJSON, gotGoogleWorkstationsConfig)
		require.NoError(t, err)

		assert.Equal(t, workstations.ContainerImagePosit, gotGoogleWorkstationsConfig.Container.Image)
		assert.Equal(t, workstations.MachineTypeN2DStandard32, gotGoogleWorkstationsConfig.Host.Config.(*workstationspb.WorkstationConfig_Host_GceInstance_).GceInstance.MachineType)
	})

	t.Run("Update workstation config that does not exist", func(t *testing.T) {
		_, err := client.UpdateWorkstationConfig(ctx, &workstations.WorkstationConfigUpdateOpts{
			Slug:           "non-existent",
			MachineType:    workstations.MachineTypeN2DStandard32,
			ContainerImage: workstations.ContainerImagePosit,
		})

		require.ErrorIs(t, err, workstations.ErrNotExist)
	})

	t.Run("Get workstation that does not exist", func(t *testing.T) {
		_, err := client.GetWorkstation(ctx, &workstations.WorkstationIdentifier{
			Slug:                  workstationSlug,
			WorkstationConfigSlug: configSlug,
		})

		require.ErrorIs(t, err, workstations.ErrNotExist)
	})

	t.Run("Start workstation that does not exist", func(t *testing.T) {
		err := client.StartWorkstation(ctx, &workstations.WorkstationIdentifier{
			Slug:                  workstationSlug,
			WorkstationConfigSlug: configSlug,
		})

		require.ErrorIs(t, err, workstations.ErrNotExist)
	})

	t.Run("Stop workstation that does not exist", func(t *testing.T) {
		err := client.StopWorkstation(ctx, &workstations.WorkstationIdentifier{
			Slug:                  workstationSlug,
			WorkstationConfigSlug: configSlug,
		})

		require.ErrorIs(t, err, workstations.ErrNotExist)
	})

	t.Run("Create workstation", func(t *testing.T) {
		got, err := client.CreateWorkstation(ctx, &workstations.WorkstationOpts{
			Slug:                  workstationSlug,
			DisplayName:           workstationDisplayName,
			Labels:                map[string]string{},
			WorkstationConfigSlug: configSlug,
		})

		require.NoError(t, err)
		diff := cmp.Diff(workstation, got, cmpopts.IgnoreFields(workstations.Workstation{}, "CreateTime"))
		assert.Empty(t, diff)
		assert.NotNil(t, got.CreateTime)
	})

	t.Run("Get, start, and stop workstation", func(t *testing.T) {
		got, err := client.GetWorkstation(ctx, &workstations.WorkstationIdentifier{
			Slug:                  workstationSlug,
			WorkstationConfigSlug: configSlug,
		})

		require.NoError(t, err)
		diff := cmp.Diff(workstation, got, cmpopts.IgnoreFields(workstations.Workstation{}, "CreateTime"))
		assert.Empty(t, diff)

		err = client.StartWorkstation(ctx, &workstations.WorkstationIdentifier{
			Slug:                  workstationSlug,
			WorkstationConfigSlug: configSlug,
		})
		require.NoError(t, err)

		err = client.StopWorkstation(ctx, &workstations.WorkstationIdentifier{
			Slug:                  workstationSlug,
			WorkstationConfigSlug: configSlug,
		})
		require.NoError(t, err)
	})

	t.Run("Update workstation policy", func(t *testing.T) {
		expect := &iampb.Policy{
			Bindings: []*iampb.Binding{
				{
					Role:    "roles/viewer",
					Members: []string{"user:nada@nav.no"},
				},
			},
		}

		err := client.UpdateWorkstationIAMPolicyBindings(ctx, &workstations.WorkstationIdentifier{
			Slug:                  workstationSlug,
			WorkstationConfigSlug: configSlug,
		}, func(bindings []*workstations.Binding) []*workstations.Binding {
			return append(bindings, &workstations.Binding{
				Role:    "roles/viewer",
				Members: []string{"user:nada@nav.no"},
			})
		})
		require.NoError(t, err)

		policies := e.GetIamPolicies()
		assert.Len(t, policies, 1)
		for _, policy := range policies {
			diff := cmp.Diff(expect.Bindings, policy.Bindings, cmpopts.IgnoreUnexported(iampb.Binding{}))
			assert.Empty(t, diff)
		}
	})

	t.Run("Delete workstation config", func(t *testing.T) {
		err := client.DeleteWorkstationConfig(ctx, &workstations.WorkstationConfigDeleteOpts{
			Slug: configSlug,
		})

		require.NoError(t, err)

		assert.Empty(t, e.GetWorkstationConfigs())
		assert.Empty(t, e.GetWorkstations())
	})
}

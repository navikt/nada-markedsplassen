package workstations

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/iam/apiv1/iampb"

	workstations "cloud.google.com/go/workstations/apiv1"
	"cloud.google.com/go/workstations/apiv1/workstationspb"

	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const (
	DefaultIdleTimeoutInSec    = 7200  // 2 hours
	DefaultRunningTimeoutInSec = 43200 // 12 hours
	DefaultBootDiskSizeInGB    = 120
	DefaultHomeDiskSizeInGB    = 100
	DefaultHomeDiskType        = "pd-ssd"
	DefaultHomeDiskFsType      = "ext4"

	MachineTypeN2DStandard2  = "n2d-standard-2"
	MachineTypeN2DStandard4  = "n2d-standard-4"
	MachineTypeN2DStandard8  = "n2d-standard-8"
	MachineTypeN2DStandard16 = "n2d-standard-16"
	MachineTypeN2DStandard32 = "n2d-standard-32"

	ContainerImageVSCode           = "us-central1-docker.pkg.dev/cloud-workstations-images/predefined/code-oss:latest"
	ContainerImageIntellijUltimate = "us-central1-docker.pkg.dev/cloud-workstations-images/predefined/intellij-ultimate:latest"
	ContainerImagePosit            = "us-central1-docker.pkg.dev/posit-images/cloud-workstations/workbench:latest"

	MountPathHome = "/home"
)

var (
	ErrNotExist = errors.New("not exist")
	ErrExist    = errors.New("already exists")
)

var _ Operations = &Client{}

type Operations interface {
	GetWorkstationConfig(ctx context.Context, opts *WorkstationConfigGetOpts) (*WorkstationConfig, error)
	CreateWorkstationConfig(ctx context.Context, opts *WorkstationConfigOpts) (*WorkstationConfig, error)
	UpdateWorkstationConfig(ctx context.Context, opts *WorkstationConfigUpdateOpts) (*WorkstationConfig, error)
	DeleteWorkstationConfig(ctx context.Context, opts *WorkstationConfigDeleteOpts) error
	GetWorkstation(ctx context.Context, opts *WorkstationIdentifier) (*Workstation, error)
	CreateWorkstation(ctx context.Context, opts *WorkstationOpts) (*Workstation, error)
	StartWorkstation(ctx context.Context, opts *WorkstationIdentifier) error
	StopWorkstation(ctx context.Context, opts *WorkstationIdentifier) error
	UpdateWorkstationIAMPolicyBindings(ctx context.Context, opts *WorkstationIdentifier, fn UpdateWorkstationIAMPolicyBindingsFn) error
}

type UpdateWorkstationIAMPolicyBindingsFn func(bindings []*Binding) []*Binding

type Binding struct {
	Role    string
	Members []string
}

type WorkstationConfigGetOpts struct {
	// Slug is the unique identifier of the workstation
	Slug string
}

type WorkstationConfigOpts struct {
	// Slug is the unique identifier of the workstation
	Slug string

	// DisplayName is the human-readable name of the workstation
	DisplayName string

	// MachineType is the type of machine that will be used for the workstation, e.g.:
	// - n2d-standard-2
	// - n2d-standard-4
	// - n2d-standard-8
	// - n2d-standard-16
	// - n2d-standard-32
	MachineType string

	// ServiceAccountEmail is the email address of the service account that will be associated with the workstation,
	// which we can use to grant permissions to the workstation, e.g.:
	// - Secure Web Proxy rules
	// - VPC Service controls
	// - Login
	ServiceAccountEmail string

	// SubjectEmail is the email address of the subject that will be using the workstation
	SubjectEmail string

	// Map of labels applied to Workstation resources
	Labels map[string]string

	// ContainerImage is the image that will be used to run the workstation
	ContainerImage string
}

type WorkstationConfigUpdateOpts struct {
	// Slug is the unique identifier of the workstation
	Slug string

	// MachineType is the type of machine that will be used for the workstation, e.g.:
	// - n2d-standard-2
	// - n2d-standard-4
	// - n2d-standard-8
	// - n2d-standard-16
	// - n2d-standard-32
	MachineType string

	// ContainerImage is the image that will be used to run the workstation
	ContainerImage string
}

type WorkstationConfigDeleteOpts struct {
	// Slug is the unique identifier of the workstation
	Slug string
}

type WorkstationIdentifier struct {
	// Slug is the unique identifier of the workstation
	Slug string

	// Slug identifying the workstation config
	WorkstationConfigSlug string
}

type WorkstationOpts struct {
	// Slug is the unique identifier of the workstation
	Slug string

	// DisplayName is the human-readable name of the workstation
	DisplayName string

	// Labels applied to the resource and propagated to the underlying Compute Engine resources.
	Labels map[string]string

	// Workstation configuration
	WorkstationConfigSlug string
}

type WorkstationConfig struct {
	// Name of this workstation configuration.
	Slug string

	// The fully qualified name of this workstation configuration
	FullyQualifiedName string

	// Human-readable name for this workstation configuration.
	DisplayName string

	// [Labels](https://cloud.google.com/workstations/docs/label-resources) that
	// are applied to the workstation configuration and that are also propagated
	// to the underlying Compute Engine resources.
	Labels map[string]string

	// Time when this workstation configuration was created.
	CreateTime time.Time

	// Time when this workstation configuration was most recently
	// updated.
	UpdateTime *time.Time

	// Number of seconds to wait before automatically stopping a
	// workstation after it last received user traffic.
	IdleTimeout time.Duration

	// Number of seconds that a workstation can run until it is
	// automatically shut down. We recommend that workstations be shut down daily
	RunningTimeout time.Duration

	// The type of machine to use for VM instances—for example,
	// `"e2-standard-4"`. For more information about machine types that
	// Cloud Workstations supports, see the list of
	// [available machine
	// types](https://cloud.google.com/workstations/docs/available-machine-types).
	MachineType string

	// The email address of the service account for Cloud
	// Workstations VMs created with this configuration.
	ServiceAccount string

	// The container image to use for the workstation.
	Image string

	// Environment variables passed to the container's entrypoint.
	Env map[string]string
}

type WorkstationState int32

const (
	Workstation_STATE_STARTING WorkstationState = 1
	Workstation_STATE_RUNNING  WorkstationState = 2
	Workstation_STATE_STOPPING WorkstationState = 3
	Workstation_STATE_STOPPED  WorkstationState = 4
)

type Workstation struct {
	// Name of this workstation.
	Slug string

	// The fully qualified name of this workstation.
	FullyQualifiedName string

	// Human-readable name for this workstation.
	DisplayName string

	// Indicates whether this workstation is currently being updated
	// to match its intended state.
	Reconciling bool

	// Time when this workstation was created.
	CreateTime time.Time

	// Time when this workstation was most recently updated.
	UpdateTime *time.Time

	// Time when this workstation was most recently successfully
	// started, regardless of the workstation's initial state.
	StartTime *time.Time

	State WorkstationState

	// Host to which clients can send HTTPS traffic that will be
	// received by the workstation. Authorized traffic will be received to the
	// workstation as HTTP on port 80. To send traffic to a different port,
	// clients may prefix the host with the destination port in the format
	// `{port}-{host}`.
	Host string
}

type Client struct {
	project              string
	location             string
	workstationClusterID string

	apiEndpoint string
	disableAuth bool
}

func (c *Client) UpdateWorkstationIAMPolicyBindings(ctx context.Context, opts *WorkstationIdentifier, fn UpdateWorkstationIAMPolicyBindingsFn) error {
	client, err := c.newClient(ctx)
	if err != nil {
		return err
	}

	policy, err := client.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
		Resource: c.FullyQualifiedWorkstationName(opts.WorkstationConfigSlug, opts.Slug),
	})
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusNotFound {
			return ErrNotExist
		}

		return fmt.Errorf("getting IAM policy: %w", err)
	}

	bindings := []*Binding{}
	for _, binding := range policy.Bindings {
		bindings = append(bindings, &Binding{
			Role:    binding.Role,
			Members: binding.Members,
		})
	}

	bindings = fn(bindings)

	var pbBindings []*iampb.Binding
	for _, binding := range bindings {
		pbBindings = append(pbBindings, &iampb.Binding{
			Role:    binding.Role,
			Members: binding.Members,
		})
	}

	policy.Bindings = pbBindings

	_, err = client.SetIamPolicy(ctx, &iampb.SetIamPolicyRequest{
		Resource: c.FullyQualifiedWorkstationName(opts.WorkstationConfigSlug, opts.Slug),
		Policy:   policy,
	})
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusNotFound {
			return ErrNotExist
		}

		return fmt.Errorf("setting IAM policy: %w", err)
	}

	return nil
}

func (c *Client) StartWorkstation(ctx context.Context, opts *WorkstationIdentifier) error {
	client, err := c.newClient(ctx)
	if err != nil {
		return err
	}

	raw, err := client.StartWorkstation(ctx, &workstationspb.StartWorkstationRequest{
		Name: c.FullyQualifiedWorkstationName(opts.WorkstationConfigSlug, opts.Slug),
	})
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusNotFound {
			return ErrNotExist
		}

		return err
	}

	_, err = raw.Wait(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) StopWorkstation(ctx context.Context, opts *WorkstationIdentifier) error {
	client, err := c.newClient(ctx)
	if err != nil {
		return err
	}

	raw, err := client.StopWorkstation(ctx, &workstationspb.StopWorkstationRequest{
		Name: c.FullyQualifiedWorkstationName(opts.WorkstationConfigSlug, opts.Slug),
	})
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusNotFound {
			return ErrNotExist
		}

		return err
	}

	_, err = raw.Wait(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetWorkstationConfig(ctx context.Context, opts *WorkstationConfigGetOpts) (*WorkstationConfig, error) {
	client, err := c.newClient(ctx)
	if err != nil {
		return nil, err
	}

	raw, err := client.GetWorkstationConfig(ctx, &workstationspb.GetWorkstationConfigRequest{
		Name: c.FullyQualifiedWorkstationConfigName(opts.Slug),
	})
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusNotFound {
			return nil, ErrNotExist
		}

		return nil, err
	}

	var updateTime *time.Time
	if raw.UpdateTime != nil {
		t := raw.UpdateTime.AsTime()
		updateTime = &t
	}

	var idleTimeout time.Duration
	if raw.IdleTimeout != nil {
		idleTimeout = raw.IdleTimeout.AsDuration()
	}

	var runningTimeout time.Duration
	if raw.RunningTimeout != nil {
		runningTimeout = raw.RunningTimeout.AsDuration()
	}

	return &WorkstationConfig{
		Slug:               opts.Slug,
		FullyQualifiedName: raw.Name,
		DisplayName:        raw.DisplayName,
		Labels:             raw.Labels,
		CreateTime:         raw.CreateTime.AsTime(),
		UpdateTime:         updateTime,
		IdleTimeout:        idleTimeout,
		RunningTimeout:     runningTimeout,
		MachineType:        raw.Host.GetGceInstance().MachineType,
		ServiceAccount:     raw.Host.GetGceInstance().ServiceAccount,
		Image:              raw.Container.Image,
		Env:                raw.Container.Env,
	}, nil
}

func (c *Client) CreateWorkstationConfig(ctx context.Context, opts *WorkstationConfigOpts) (*WorkstationConfig, error) {
	client, err := c.newClient(ctx)
	if err != nil {
		return nil, err
	}

	op, err := client.CreateWorkstationConfig(ctx, &workstationspb.CreateWorkstationConfigRequest{
		Parent:              c.FullyQualifiedWorkstationClusterName(),
		WorkstationConfigId: opts.Slug,
		WorkstationConfig: &workstationspb.WorkstationConfig{
			Name:        c.FullyQualifiedWorkstationConfigName(opts.Slug),
			DisplayName: opts.DisplayName,
			Labels:      opts.Labels,
			IdleTimeout: &durationpb.Duration{
				Seconds: DefaultIdleTimeoutInSec,
			},
			RunningTimeout: &durationpb.Duration{
				Seconds: DefaultRunningTimeoutInSec,
			},
			Host: &workstationspb.WorkstationConfig_Host{
				Config: &workstationspb.WorkstationConfig_Host_GceInstance_{
					GceInstance: &workstationspb.WorkstationConfig_Host_GceInstance{
						MachineType:                opts.MachineType,
						ServiceAccount:             opts.ServiceAccountEmail,
						Tags:                       nil, // FIXME:  lets try to avoid using this, but we might need it for some default rules
						DisablePublicIpAddresses:   true,
						EnableNestedVirtualization: false,
						ShieldedInstanceConfig: &workstationspb.WorkstationConfig_Host_GceInstance_GceShieldedInstanceConfig{
							EnableSecureBoot:          true,
							EnableVtpm:                true,
							EnableIntegrityMonitoring: true,
						},
						ConfidentialInstanceConfig: &workstationspb.WorkstationConfig_Host_GceInstance_GceConfidentialInstanceConfig{
							EnableConfidentialCompute: true,
						},
						BootDiskSizeGb: DefaultBootDiskSizeInGB,
					},
				},
			},
			PersistentDirectories: []*workstationspb.WorkstationConfig_PersistentDirectory{
				{
					DirectoryType: &workstationspb.WorkstationConfig_PersistentDirectory_GcePd{
						GcePd: &workstationspb.WorkstationConfig_PersistentDirectory_GceRegionalPersistentDisk{
							SizeGb:        DefaultHomeDiskSizeInGB,
							FsType:        DefaultHomeDiskFsType,
							DiskType:      DefaultHomeDiskType,
							ReclaimPolicy: workstationspb.WorkstationConfig_PersistentDirectory_GceRegionalPersistentDisk_DELETE,
						},
					},
					MountPath: MountPathHome,
				},
			},
			Container: &workstationspb.WorkstationConfig_Container{
				Image:   opts.ContainerImage,
				Command: nil,
				Args:    nil,
				// FIXME: we need to set PIP_INDEX_URL=..., HTTP_PROXY=..., NO_PROXY=.adeo.no, ...
				Env: map[string]string{
					"WORKSTATION_NAME": opts.Slug,
				},
			},
		},
	})
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusConflict {
			return nil, ErrExist
		}

		return nil, err
	}

	raw, err := op.Wait(ctx)
	if err != nil {
		return nil, err
	}

	var updateTime *time.Time
	if raw.UpdateTime != nil {
		t := raw.UpdateTime.AsTime()
		updateTime = &t
	}

	var idleTimeout time.Duration
	if raw.IdleTimeout != nil {
		idleTimeout = raw.IdleTimeout.AsDuration()
	}

	var runningTimeout time.Duration
	if raw.RunningTimeout != nil {
		runningTimeout = raw.RunningTimeout.AsDuration()
	}

	return &WorkstationConfig{
		Slug:               opts.Slug,
		FullyQualifiedName: raw.Name,
		DisplayName:        raw.DisplayName,
		Labels:             raw.Labels,
		CreateTime:         raw.CreateTime.AsTime(),
		UpdateTime:         updateTime,
		IdleTimeout:        idleTimeout,
		RunningTimeout:     runningTimeout,
		MachineType:        raw.Host.GetGceInstance().MachineType,
		ServiceAccount:     raw.Host.GetGceInstance().ServiceAccount,
		Image:              raw.Container.Image,
		Env:                raw.Container.Env,
	}, nil
}

func (c *Client) GetWorkstation(ctx context.Context, opts *WorkstationIdentifier) (*Workstation, error) {
	client, err := c.newClient(ctx)
	if err != nil {
		return nil, err
	}

	raw, err := client.GetWorkstation(ctx, &workstationspb.GetWorkstationRequest{
		Name: c.FullyQualifiedWorkstationName(opts.WorkstationConfigSlug, opts.Slug),
	})
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusNotFound {
			return nil, ErrNotExist
		}

		return nil, err
	}

	var updateTime *time.Time
	if raw.UpdateTime != nil {
		t := raw.UpdateTime.AsTime()
		updateTime = &t
	}

	var startTime *time.Time
	if raw.StartTime != nil {
		t := raw.StartTime.AsTime()
		startTime = &t
	}

	return &Workstation{
		Slug:               opts.Slug,
		FullyQualifiedName: raw.Name,
		DisplayName:        raw.DisplayName,
		CreateTime:         raw.CreateTime.AsTime(),
		UpdateTime:         updateTime,
		StartTime:          startTime,
		State:              WorkstationState(raw.State),
		Host:               raw.Host,
	}, nil
}

func (c *Client) CreateWorkstation(ctx context.Context, opts *WorkstationOpts) (*Workstation, error) {
	client, err := c.newClient(ctx)
	if err != nil {
		return nil, err
	}

	op, err := client.CreateWorkstation(ctx, &workstationspb.CreateWorkstationRequest{
		Parent:        c.FullyQualifiedWorkstationConfigName(opts.WorkstationConfigSlug),
		WorkstationId: opts.Slug,
		Workstation: &workstationspb.Workstation{
			Name:        c.FullyQualifiedWorkstationName(opts.WorkstationConfigSlug, opts.Slug),
			DisplayName: opts.DisplayName,
			Labels:      opts.Labels,
		},
	})
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusConflict {
			return nil, ErrExist
		}

		return nil, err
	}

	raw, err := op.Wait(ctx)
	if err != nil {
		return nil, err
	}

	var updateTime *time.Time
	if raw.UpdateTime != nil {
		t := raw.UpdateTime.AsTime()
		updateTime = &t
	}

	var startTime *time.Time
	if raw.StartTime != nil {
		t := raw.StartTime.AsTime()
		startTime = &t
	}

	return &Workstation{
		Slug:               opts.Slug,
		FullyQualifiedName: raw.Name,
		DisplayName:        raw.DisplayName,
		CreateTime:         raw.CreateTime.AsTime(),
		UpdateTime:         updateTime,
		StartTime:          startTime,
		State:              WorkstationState(raw.State),
		Host:               raw.Host,
	}, nil
}

func (c *Client) UpdateWorkstationConfig(ctx context.Context, opts *WorkstationConfigUpdateOpts) (*WorkstationConfig, error) {
	client, err := c.newClient(ctx)
	if err != nil {
		return nil, err
	}

	op, err := client.UpdateWorkstationConfig(ctx, &workstationspb.UpdateWorkstationConfigRequest{
		WorkstationConfig: &workstationspb.WorkstationConfig{
			Name: c.FullyQualifiedWorkstationConfigName(opts.Slug),
			Host: &workstationspb.WorkstationConfig_Host{
				Config: &workstationspb.WorkstationConfig_Host_GceInstance_{
					GceInstance: &workstationspb.WorkstationConfig_Host_GceInstance{
						MachineType: opts.MachineType,
					},
				},
			},
			Container: &workstationspb.WorkstationConfig_Container{
				Image: opts.ContainerImage,
			},
		},
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{
				"host.gce_instance.machine_type",
				"container.image",
			},
		},
		ValidateOnly: false,
		AllowMissing: false,
	})
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusNotFound {
			return nil, ErrNotExist
		}

		return nil, err
	}

	raw, err := op.Wait(ctx)
	if err != nil {
		return nil, err
	}

	var updateTime *time.Time
	if raw.UpdateTime != nil {
		t := raw.UpdateTime.AsTime()
		updateTime = &t
	}

	var idleTimeout time.Duration
	if raw.IdleTimeout != nil {
		idleTimeout = raw.IdleTimeout.AsDuration()
	}

	var runningTimeout time.Duration
	if raw.RunningTimeout != nil {
		runningTimeout = raw.RunningTimeout.AsDuration()
	}

	return &WorkstationConfig{
		Slug:               opts.Slug,
		FullyQualifiedName: raw.Name,
		DisplayName:        raw.DisplayName,
		Labels:             raw.Labels,
		CreateTime:         raw.CreateTime.AsTime(),
		UpdateTime:         updateTime,
		IdleTimeout:        idleTimeout,
		RunningTimeout:     runningTimeout,
		MachineType:        raw.Host.GetGceInstance().MachineType,
		ServiceAccount:     raw.Host.GetGceInstance().ServiceAccount,
		Image:              raw.Container.Image,
		Env:                raw.Container.Env,
	}, nil
}

func (c *Client) DeleteWorkstationConfig(ctx context.Context, opts *WorkstationConfigDeleteOpts) error {
	client, err := c.newClient(ctx)
	if err != nil {
		return err
	}

	op, err := client.DeleteWorkstationConfig(ctx, &workstationspb.DeleteWorkstationConfigRequest{
		Name:  c.FullyQualifiedWorkstationConfigName(opts.Slug),
		Force: true, // If set, any workstations in the workstation configuration are also deleted.
	})
	if err != nil {
		return err
	}

	_, err = op.Wait(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) newClient(ctx context.Context) (*workstations.Client, error) {
	var options []option.ClientOption

	if c.apiEndpoint != "" {
		options = append(options, option.WithEndpoint(c.apiEndpoint))
	}

	if c.disableAuth {
		options = append(options,
			option.WithoutAuthentication(),
		)
	}

	client, err := workstations.NewRESTClient(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("creating workstations client: %w", err)
	}

	return client, nil
}

func (c *Client) FullyQualifiedWorkstationClusterName() string {
	return fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s", c.project, c.location, c.workstationClusterID)
}

func (c *Client) FullyQualifiedWorkstationConfigName(configName string) string {
	return fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s/workstationConfigs/%s", c.project, c.location, c.workstationClusterID, configName)
}

func (c *Client) FullyQualifiedWorkstationName(configName, workstationName string) string {
	return fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s/workstationConfigs/%s/workstations/%s", c.project, c.location, c.workstationClusterID, configName, workstationName)
}

func New(project, location, workstationClusterID, apiEndpoint string, disableAuth bool) *Client {
	return &Client{
		project:              project,
		location:             location,
		workstationClusterID: workstationClusterID,
		apiEndpoint:          apiEndpoint,
		disableAuth:          disableAuth,
	}
}

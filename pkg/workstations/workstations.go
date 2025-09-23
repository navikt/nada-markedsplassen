package workstations

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"cloud.google.com/go/iam/apiv1/iampb"
	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"golang.org/x/oauth2/google"

	workstations "cloud.google.com/go/workstations/apiv1"
	"cloud.google.com/go/workstations/apiv1/workstationspb"

	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/encoding/protojson"
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

	ContainerImageVSCode           = "us-central1-docker.pkg.dev/cloud-workstations-images/predefined/code-oss:latest"
	ContainerImageIntellijUltimate = "us-central1-docker.pkg.dev/cloud-workstations-images/predefined/intellij-ultimate:latest"
	ContainerImagePosit            = "us-central1-docker.pkg.dev/posit-images/cloud-workstations/workbench:latest"

	MountPathHome = "/home"

	workstationsAPIHost = "https://workstations.googleapis.com"
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
	ListWorkstationConfigs(ctx context.Context) ([]*WorkstationConfig, error)
	GetSSHConnectivity(ctx context.Context, slug string) (bool, error)
	UpdateSSHConnectivity(ctx context.Context, slug string, allow bool) error
}

type UpdateWorkstationIAMPolicyBindingsFn func(bindings []*Binding) []*Binding

type Binding struct {
	Role    string
	Members []string
}

type ReadinessCheck struct {
	Path string
	Port int
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

	// Annotations are free-form annotations used to persist information about firewall openings
	Annotations map[string]string

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

	// Environment variables passed to the container's entrypoint.
	Env map[string]string

	// Map of labels applied to Workstation resources
	Labels map[string]string

	// ContainerImage is the image that will be used to run the workstation
	ContainerImage string

	// Extra readiness checks to be added to the workstation configuration.
	ReadinessChecks []*ReadinessCheck
}

type WorkstationConfigUpdateOpts struct {
	// Slug is the unique identifier of the workstation
	Slug string

	// Annotations are free-form annotations used to persist information about firewall openings
	Annotations map[string]string

	// MachineType is the type of machine that will be used for the workstation, e.g.:
	// - n2d-standard-2
	// - n2d-standard-4
	// - n2d-standard-8
	// - n2d-standard-16
	// - n2d-standard-32
	MachineType string

	// ContainerImage is the image that will be used to run the workstation
	ContainerImage string

	// Environment variables passed to the container's entrypoint.
	Env map[string]string

	// Extra readiness checks to be added to the workstation configuration.
	ReadinessChecks []*ReadinessCheck
}

type WorkstationConfigDeleteOpts struct {
	// Slug is the unique identifier of the workstation
	Slug string
	// NoWait is a flag to indicate if the operation should wait for completion
	NoWait bool
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

	// Annotations are free-form annotations used to persist information
	Annotations map[string]string

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

	// ReplicaZones are the zones within a region for which vm instances are created
	ReplicaZones []string

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

	// Complete workstation configuration as returned by the Google API.
	CompleteConfigAsJSON []byte

	// ReadinessChecks are additional checks that are run to determine if the workstation is ready to use.
	ReadinessChecks []*ReadinessCheck

	// Indicates whether this workstation configuration is currently being updated to match its intended state.
	Reconciling bool
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

	apiEndpoint          string
	disableAuth          bool
	disableSSHHTTPClient *http.Client
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

	var machineType, serviceAccount string
	if raw.Host != nil && raw.Host.GetGceInstance() != nil {
		machineType = raw.Host.GetGceInstance().MachineType
		serviceAccount = raw.Host.GetGceInstance().ServiceAccount
	}

	configBytes, err := protojson.Marshal(raw)
	if err != nil {
		return nil, err
	}

	var readinessChecksOut []*ReadinessCheck
	for _, rc := range raw.ReadinessChecks {
		readinessChecksOut = append(readinessChecksOut, &ReadinessCheck{
			Path: rc.Path,
			Port: int(rc.Port),
		})
	}

	return &WorkstationConfig{
		Slug:                 opts.Slug,
		FullyQualifiedName:   raw.Name,
		DisplayName:          raw.DisplayName,
		Annotations:          raw.Annotations,
		Labels:               raw.Labels,
		CreateTime:           raw.CreateTime.AsTime(),
		UpdateTime:           updateTime,
		IdleTimeout:          idleTimeout,
		RunningTimeout:       runningTimeout,
		ReplicaZones:         raw.GetReplicaZones(),
		MachineType:          machineType,
		ServiceAccount:       serviceAccount,
		Image:                raw.Container.Image,
		Env:                  raw.Container.Env,
		CompleteConfigAsJSON: configBytes,
		ReadinessChecks:      readinessChecksOut,
		Reconciling:          raw.Reconciling,
	}, nil
}

func (c *Client) CreateWorkstationConfig(ctx context.Context, opts *WorkstationConfigOpts) (*WorkstationConfig, error) {
	client, err := c.newClient(ctx)
	if err != nil {
		return nil, err
	}

	var readinessChecks []*workstationspb.WorkstationConfig_ReadinessCheck
	for _, rc := range opts.ReadinessChecks {
		readinessChecks = append(readinessChecks, &workstationspb.WorkstationConfig_ReadinessCheck{
			Path: rc.Path,
			Port: int32(rc.Port),
		})
	}

	config := &workstationspb.CreateWorkstationConfigRequest{
		Parent:              c.FullyQualifiedWorkstationClusterName(),
		WorkstationConfigId: opts.Slug,
		WorkstationConfig: &workstationspb.WorkstationConfig{
			Name:            c.FullyQualifiedWorkstationConfigName(opts.Slug),
			Annotations:     opts.Annotations,
			DisplayName:     opts.DisplayName,
			Labels:          opts.Labels,
			ReadinessChecks: readinessChecks,
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
				Env:     opts.Env,
			},
		},
	}

	op, err := client.CreateWorkstationConfig(ctx, config)
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

	err = c.disableSSHUnderlyingVM(ctx, opts.Slug)
	if err != nil {
		return nil, err
	}

	configBytes, err := protojson.Marshal(raw)
	if err != nil {
		return nil, err
	}

	var readinessChecksOut []*ReadinessCheck
	for _, rc := range raw.ReadinessChecks {
		readinessChecksOut = append(readinessChecksOut, &ReadinessCheck{
			Path: rc.Path,
			Port: int(rc.Port),
		})
	}

	return &WorkstationConfig{
		Slug:                 opts.Slug,
		FullyQualifiedName:   raw.Name,
		DisplayName:          raw.DisplayName,
		Annotations:          raw.Annotations,
		Labels:               raw.Labels,
		CreateTime:           raw.CreateTime.AsTime(),
		UpdateTime:           updateTime,
		IdleTimeout:          idleTimeout,
		RunningTimeout:       runningTimeout,
		ReplicaZones:         raw.GetReplicaZones(),
		MachineType:          raw.Host.GetGceInstance().MachineType,
		ServiceAccount:       raw.Host.GetGceInstance().ServiceAccount,
		Image:                raw.Container.Image,
		Env:                  raw.Container.Env,
		CompleteConfigAsJSON: configBytes,
		ReadinessChecks:      readinessChecksOut,
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
		Reconciling:        raw.Reconciling,
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

	var readinessChecks []*workstationspb.WorkstationConfig_ReadinessCheck
	for _, rc := range opts.ReadinessChecks {
		readinessChecks = append(readinessChecks, &workstationspb.WorkstationConfig_ReadinessCheck{
			Path: rc.Path,
			Port: int32(rc.Port),
		})
	}

	config := &workstationspb.UpdateWorkstationConfigRequest{
		WorkstationConfig: &workstationspb.WorkstationConfig{
			Name:            c.FullyQualifiedWorkstationConfigName(opts.Slug),
			Annotations:     opts.Annotations,
			ReadinessChecks: readinessChecks,
			Host: &workstationspb.WorkstationConfig_Host{
				Config: &workstationspb.WorkstationConfig_Host_GceInstance_{
					GceInstance: &workstationspb.WorkstationConfig_Host_GceInstance{
						MachineType: opts.MachineType,
					},
				},
			},
			Container: &workstationspb.WorkstationConfig_Container{
				Image: opts.ContainerImage,
				Env:   opts.Env,
			},
		},
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{
				"host.gce_instance.machine_type",
				"container.image",
				"annotations",
				"readiness_checks",
				"container.env",
			},
		},
		ValidateOnly: false,
		AllowMissing: false,
	}

	op, err := client.UpdateWorkstationConfig(ctx, config)
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

	err = c.disableSSHUnderlyingVM(ctx, opts.Slug)
	if err != nil {
		return nil, err
	}

	configBytes, err := protojson.Marshal(raw)
	if err != nil {
		return nil, err
	}

	var readinessChecksOut []*ReadinessCheck
	for _, rc := range raw.ReadinessChecks {
		readinessChecksOut = append(readinessChecksOut, &ReadinessCheck{
			Path: rc.Path,
			Port: int(rc.Port),
		})
	}

	return &WorkstationConfig{
		Slug:                 opts.Slug,
		FullyQualifiedName:   raw.Name,
		DisplayName:          raw.DisplayName,
		Annotations:          raw.Annotations,
		Labels:               raw.Labels,
		CreateTime:           raw.CreateTime.AsTime(),
		UpdateTime:           updateTime,
		IdleTimeout:          idleTimeout,
		RunningTimeout:       runningTimeout,
		ReplicaZones:         raw.GetReplicaZones(),
		MachineType:          raw.Host.GetGceInstance().MachineType,
		ServiceAccount:       raw.Host.GetGceInstance().ServiceAccount,
		Image:                raw.Container.Image,
		Env:                  raw.Container.Env,
		CompleteConfigAsJSON: configBytes,
		ReadinessChecks:      readinessChecksOut,
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

	if !opts.NoWait {
		_, err = op.Wait(ctx)
		if err != nil {
			return err
		}
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

func (c *Client) disableSSHUnderlyingVM(ctx context.Context, workstationID string) error {
	var err error

	host := workstationsAPIHost
	if c.apiEndpoint != "" {
		host = c.apiEndpoint
	}

	token := ""
	if !c.disableAuth {
		token, err = c.getGoogleAPIToken(ctx)
		if err != nil {
			return fmt.Errorf("getting token: %w", err)
		}
	}

	type GCEInstance struct {
		DisableSSH bool `json:"disableSsh"`
	}
	type WorkstationHost struct {
		GCEInstance *GCEInstance `json:"gceInstance"`
	}
	type disableSSHBody struct {
		Container map[string]any   `json:"container"`
		Host      *WorkstationHost `json:"host"`
	}

	body := disableSSHBody{
		Container: map[string]any{},
		Host: &WorkstationHost{
			GCEInstance: &GCEInstance{
				DisableSSH: true,
			},
		},
	}

	dataBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshalling request body: %w", err)
	}

	url := fmt.Sprintf("%s/v1/projects/%s/locations/%s/workstationClusters/%s/workstationConfigs/%s", host, c.project, c.location, c.workstationClusterID, workstationID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(dataBytes))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	q := req.URL.Query()
	q.Add("alt", "json")
	q.Add("updateMask", "host.gce_instance.disable_ssh")
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	res, err := c.disableSSHHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	if res.StatusCode > 299 {
		resBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("workstation config disable ssh status code %d, error reading response body: %w", res.StatusCode, err)
		}
		return fmt.Errorf("workstation config disable ssh status code %d: %s", res.StatusCode, string(resBytes))
	}

	resBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	op := &longrunningpb.Operation{}
	err = protojson.UnmarshalOptions{AllowPartial: true}.Unmarshal(resBytes, op)
	if err != nil {
		return fmt.Errorf("unmarshalling workstation operations object %w", err)
	}

	err = c.waitGoogleAPIOperationDone(ctx, host, token, op.Name)
	if err != nil {
		return fmt.Errorf("waiting for operation to complete: %w", err)
	}

	return nil
}

func (c *Client) getGoogleAPIToken(ctx context.Context) (string, error) {
	tokenSource, err := google.DefaultTokenSource(ctx)
	if err != nil {
		return "", fmt.Errorf("getting default token source: %w", err)
	}

	token, err := tokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("getting token: %w", err)
	}

	return token.AccessToken, nil
}

func (c *Client) waitGoogleAPIOperationDone(ctx context.Context, host, token, opName string) error {
	url := fmt.Sprintf("%s/v1/%s", host, opName)

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("Content-Type", "application/json")

		res, err := c.disableSSHHTTPClient.Do(req)
		if err != nil {
			return fmt.Errorf("sending request: %w", err)
		}
		defer res.Body.Close()

		resBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("operation returned status code %d, error reading response body: %w", res.StatusCode, err)
		}

		op := &longrunningpb.Operation{}
		err = protojson.UnmarshalOptions{AllowPartial: true, DiscardUnknown: true}.Unmarshal(resBytes, op)
		if err != nil {
			return err
		}

		if op.Done {
			break
		}

		time.Sleep(time.Second)
	}

	return nil
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

func extractWorkstationConfigSlug(path string) (string, error) {
	re := regexp.MustCompile(`^projects/[^/]+/locations/[^/]+/workstationClusters/[^/]+/workstationConfigs/([^/]+)$`)
	matches := re.FindStringSubmatch(path)
	if len(matches) < 2 {
		return "", errors.New("unsupported format: " + path)
	}
	return matches[1], nil
}

func (c *Client) ListWorkstationConfigs(ctx context.Context) ([]*WorkstationConfig, error) {
	client, err := c.newClient(ctx)
	if err != nil {
		return nil, err
	}

	wcIter := client.ListWorkstationConfigs(ctx, &workstationspb.ListWorkstationConfigsRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s/workstationClusters/%s", c.project, c.location, c.workstationClusterID),
	})

	wcs := []*WorkstationConfig{}
	for wc, err := wcIter.Next(); err == nil; wc, err = wcIter.Next() {
		var updateTime *time.Time
		if wc.UpdateTime != nil {
			t := wc.UpdateTime.AsTime()
			updateTime = &t
		}

		slug, err := extractWorkstationConfigSlug(wc.Name)
		if err != nil {
			continue
		}

		wcs = append(wcs, &WorkstationConfig{
			Slug:               slug,
			FullyQualifiedName: wc.Name,
			DisplayName:        wc.DisplayName,
			Annotations:        wc.Annotations,
			Labels:             wc.Labels,
			CreateTime:         wc.CreateTime.AsTime(),
			UpdateTime:         updateTime,
			IdleTimeout:        wc.IdleTimeout.AsDuration(),
			RunningTimeout:     wc.RunningTimeout.AsDuration(),
			ReplicaZones:       wc.GetReplicaZones(),
			MachineType:        wc.Host.GetGceInstance().MachineType,
			ServiceAccount:     wc.Host.GetGceInstance().ServiceAccount,
			Image:              wc.Container.Image,
			Env:                wc.Container.Env,
			Reconciling:        wc.Reconciling,
		})
	}
	return wcs, nil
}

func New(project, location, workstationClusterID, apiEndpoint string, disableAuth bool, disableSSHHTTPClient *http.Client) *Client {
	return &Client{
		project:              project,
		location:             location,
		workstationClusterID: workstationClusterID,
		apiEndpoint:          apiEndpoint,
		disableAuth:          disableAuth,
		disableSSHHTTPClient: disableSSHHTTPClient,
	}
}

type workstationConfigPortRange struct {
	First int32 `json:"first"`
	Last  int32 `json:"last"`
}

type workstationConfigAllowedPorts struct {
	AllowedPorts []workstationConfigPortRange `json:"allowedPorts"`
}

func (c *Client) getWorkstationConfigAllowedPorts(ctx context.Context, slug string) (*workstationConfigAllowedPorts, error) {
	host := workstationsAPIHost
	if c.apiEndpoint != "" {
		host = c.apiEndpoint
	}

	token := ""
	var err error
	if !c.disableAuth {
		token, err = c.getGoogleAPIToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("getting token: %w", err)
		}
	}

	name := c.FullyQualifiedWorkstationConfigName(slug)

	url := fmt.Sprintf("%s/v1/%s", host, name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating get workstation allowed port request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := c.disableSSHHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making get workstation allowed port request: %w", err)
	}
	if res.StatusCode > 299 {
		resBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("get workstation allowed port, error reading response body: %w", err)
		}
		return nil, fmt.Errorf("get workstation allowed port, status code %d: %s", res.StatusCode, string(resBytes))
	}

	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)

	var config workstationConfigAllowedPorts
	if err := json.Unmarshal(b, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *Client) GetSSHConnectivity(ctx context.Context, slug string) (bool, error) {
	config, err := c.getWorkstationConfigAllowedPorts(ctx, slug)
	if err != nil {
		return false, fmt.Errorf("failed to get workstation config allowed ports: %w", err)
	}

	for _, p := range config.AllowedPorts {
		if p.First <= 22 && p.Last >= 22 {
			return true, nil
		}
	}
	return false, nil
}

func (c *Client) UpdateSSHConnectivity(ctx context.Context, slug string, allow bool) error {
	host := workstationsAPIHost
	if c.apiEndpoint != "" {
		host = c.apiEndpoint
	}

	token := ""
	var err error
	if !c.disableAuth {
		token, err = c.getGoogleAPIToken(ctx)
		if err != nil {
			return fmt.Errorf("getting token: %w", err)
		}
	}

	name := c.FullyQualifiedWorkstationConfigName(slug)

	url := fmt.Sprintf("%s/v1/%s", host, name)

	url = fmt.Sprintf(
		"%s?updateMask=disableTcpConnections,allowedPorts",
		url,
	)

	updatedAllowedPorts := []workstationConfigPortRange{
		{First: 80, Last: 80}, // HTTP
	}

	if allow {
		updatedAllowedPorts = append(updatedAllowedPorts, workstationConfigPortRange{First: 22, Last: 22})      // SSH
		updatedAllowedPorts = append(updatedAllowedPorts, workstationConfigPortRange{First: 1024, Last: 65535}) // high port
	}

	body, err := json.Marshal(workstationConfigAllowedPorts{AllowedPorts: updatedAllowedPorts})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.disableSSHHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("making update workstation ssh connectivity request: %w", err)
	} else if res != nil && res.StatusCode >= 300 {
		return fmt.Errorf("update workstation ssh connectivity response: %s", res.Status)
	}

	defer res.Body.Close()

	b, _ := io.ReadAll(res.Body)
	op := &longrunningpb.Operation{}
	err = protojson.UnmarshalOptions{AllowPartial: true}.Unmarshal(b, op)
	if err != nil {
		return fmt.Errorf("unmarshalling workstation operations object %w", err)
	}

	err = c.waitGoogleAPIOperationDone(ctx, host, token, op.Name)
	if err != nil {
		return fmt.Errorf("waiting for operation to complete: %w", err)
	}
	return nil
}

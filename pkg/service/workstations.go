package service

import (
	"context"
	"regexp"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

const (
	LabelCreatedBy    = "created-by"
	LabelSubjectEmail = "subject-email"

	DefaultCreatedBy = "datamarkedsplassen"

	MachineTypeN2DStandard2  = "n2d-standard-2"
	MachineTypeN2DStandard4  = "n2d-standard-4"
	MachineTypeN2DStandard8  = "n2d-standard-8"
	MachineTypeN2DStandard16 = "n2d-standard-16"
	MachineTypeN2DStandard32 = "n2d-standard-32"

	ContainerImageVSCode           = "us-central1-docker.pkg.dev/cloud-workstations-images/predefined/code-oss:latest"
	ContainerImageIntellijUltimate = "us-central1-docker.pkg.dev/cloud-workstations-images/predefined/intellij-ultimate:latest"
	ContainerImagePosit            = "us-central1-docker.pkg.dev/posit-images/cloud-workstations/workbench:latest"
)

type WorkstationsService interface {
	// GetWorkstation gets the workstation for the given user including the configuration
	GetWorkstation(ctx context.Context, user *User) (*WorkstationOutput, error)

	// EnsureWorkstation creates a new workstation including the necessary service account, permissions and configuration
	EnsureWorkstation(ctx context.Context, user *User, input *WorkstationInput) (*WorkstationOutput, error)

	// DeleteWorkstation deletes the workstation configuration which also will delete the running workstation
	DeleteWorkstation(ctx context.Context, user *User) error

	// StartWorkstation starts the workstation
	StartWorkstation(ctx context.Context, user *User) error

	// StopWorkstation stops the workstation
	StopWorkstation(ctx context.Context, user *User) error
}

type WorkstationsAPI interface {
	EnsureWorkstationWithConfig(ctx context.Context, opts *EnsureWorkstationOpts) (*WorkstationConfig, *Workstation, error)
	CreateWorkstationConfig(ctx context.Context, opts *WorkstationConfigOpts) (*WorkstationConfig, error)
	UpdateWorkstationConfig(ctx context.Context, opts *WorkstationConfigUpdateOpts) (*WorkstationConfig, error)
	DeleteWorkstationConfig(ctx context.Context, opts *WorkstationConfigDeleteOpts) error
	CreateWorkstation(ctx context.Context, opts *WorkstationOpts) (*Workstation, error)
	GetWorkstation(ctx context.Context, opts *WorkstationGetOpts) (*Workstation, error)
	GetWorkstationConfig(ctx context.Context, opts *WorkstationConfigGetOpts) (*WorkstationConfig, error)
	StartWorkstation(ctx context.Context, opts *WorkstationStartOpts) error
	StopWorkstation(ctx context.Context, opts *WorkstationStopOpts) error
}

type WorkstationInput struct {
	// MachineType is the type of machine that will be used for the workstation, e.g.:
	// - n2d-standard-2
	// - n2d-standard-4
	// - n2d-standard-8
	// - n2d-standard-16
	// - n2d-standard-32
	MachineType string `json:"machineType"`

	// ContainerImage is the image that will be used to run the workstation
	ContainerImage string `json:"containerImage"`

	// URLAllowList is a list of the URLs allowed to access from workstation
	URLAllowList []string `json:"urlAllowList"`
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

type EnsureWorkstationOpts struct {
	Workstation WorkstationOpts
	Config      WorkstationConfigOpts
}

func (o WorkstationConfigOpts) Validate() error {
	return validation.ValidateStruct(&o,
		validation.Field(&o.Slug, validation.Required, validation.Length(3, 63), validation.Match(regexp.MustCompile(`^[a-z][a-z0-9-]+[a-z0-9]$`))),
		validation.Field(&o.DisplayName, validation.Required),
		validation.Field(&o.MachineType, validation.Required, validation.In(
			MachineTypeN2DStandard2,
			MachineTypeN2DStandard4,
			MachineTypeN2DStandard8,
			MachineTypeN2DStandard16,
			MachineTypeN2DStandard32,
		)),
		validation.Field(&o.ServiceAccountEmail, validation.Required, is.EmailFormat),
		validation.Field(&o.SubjectEmail, validation.Required, is.EmailFormat),
		validation.Field(&o.ContainerImage, validation.Required, validation.In(
			ContainerImageVSCode,
			ContainerImageIntellijUltimate,
			ContainerImagePosit,
		)),
	)
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

type WorkstationOpts struct {
	// Slug is the unique identifier of the workstation
	Slug string

	// DisplayName is the human-readable name of the workstation
	DisplayName string

	// Labels applied to the resource and propagated to the underlying Compute Engine resources.
	Labels map[string]string

	// Workstation configuration
	ConfigName string
}

type WorkstationConfigGetOpts struct {
	Slug string
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

type WorkstationConfigOutput struct {
	// Time when this workstation configuration was created.
	CreateTime time.Time `json:"createTime"`

	// Time when this workstation configuration was most recently
	// updated.
	UpdateTime *time.Time `json:"updateTime"`

	// Number of seconds to wait before automatically stopping a
	// workstation after it last received user traffic.
	IdleTimeout time.Duration `json:"idleTimeout"`

	// Number of seconds that a workstation can run until it is
	// automatically shut down. We recommend that workstations be shut down daily
	RunningTimeout time.Duration `json:"runningTimeout"`

	// The type of machine to use for VM instances—for example,
	// `"e2-standard-4"`. For more information about machine types that
	// Cloud Workstations supports, see the list of
	// [available machine
	// types](https://cloud.google.com/workstations/docs/available-machine-types).
	MachineType string `json:"machineType"`

	// The container image to use for the workstation.
	Image string `json:"image"`

	// Environment variables passed to the container's entrypoint.
	Env map[string]string `json:"env"`
}

type WorkstationOutput struct {
	Slug        string `json:"slug"`
	DisplayName string `json:"displayName"`

	// Indicates whether this workstation is currently being updated
	// to match its intended state.
	Reconciling bool `json:"reconciling"`

	// Time when this workstation was created.
	CreateTime time.Time `json:"createTime"`

	// Time when this workstation was most recently updated.
	UpdateTime *time.Time `json:"updateTime"`

	// Time when this workstation was most recently successfully
	// started, regardless of the workstation's initial state.
	StartTime *time.Time `json:"startTime"`

	State WorkstationState `json:"state"`

	// List of allowed URLs for the workstation
	URLAllowList []string `json:"urlAllowList"`

	Config *WorkstationConfigOutput `json:"config"`
}

type WorkstationGetOpts struct {
	// Slug is the unique identifier of the workstation
	Slug       string
	ConfigName string
}

type WorkstationStartOpts struct {
	// Slug is the unique identifier of the workstation
	Slug       string
	ConfigName string
}

type WorkstationStopOpts struct {
	// Slug is the unique identifier of the workstation
	Slug       string
	ConfigName string
}

func DefaultWorkstationLabels(subjectEmail string) map[string]string {
	return map[string]string{
		LabelCreatedBy:    DefaultCreatedBy,
		LabelSubjectEmail: subjectEmail,
	}
}

package service

import (
	"context"
	"fmt"
	"regexp"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

const (
	LabelCreatedBy    = "created-by"
	LabelSubjectIdent = "subject-ident"

	DefaultCreatedBy = "datamarkedsplassen"

	MachineTypeN2DStandard2  = "n2d-standard-2"
	MachineTypeN2DStandard4  = "n2d-standard-4"
	MachineTypeN2DStandard8  = "n2d-standard-8"
	MachineTypeN2DStandard16 = "n2d-standard-16"
	MachineTypeN2DStandard32 = "n2d-standard-32"

	ContainerImageVSCode           = "europe-north1-docker.pkg.dev/cloud-workstations-images/predefined/code-oss:latest"
	ContainerImageIntellijUltimate = "europe-north1-docker.pkg.dev/cloud-workstations-images/predefined/intellij-ultimate:latest"
	ContainerImagePosit            = "europe-north1-docker.pkg.dev/posit-images/cloud-workstations/workbench:latest"

	WorkstationUserRole = "roles/workstations.user"

	WorkstationImagesTag = "latest"

	WorkstationOnpremAllowlistAnnotation = "onprem-allowlist"
	WorkstationConfigIDLabel             = "workstation_config_id"
)

type WorkstationsService interface {
	// CreateWorkstationJob creates an asynchronous job to ensure that a workstation with the provided configuration exists
	CreateWorkstationJob(ctx context.Context, user *User, input *WorkstationInput) (*WorkstationJob, error)

	// GetWorkstationJob gets the workstation job with the given id
	GetWorkstationJob(ctx context.Context, jobID int64) (*WorkstationJob, error)

	// GetWorkstationJobsForUser gets the running workstation jobs for the given user
	GetWorkstationJobsForUser(ctx context.Context, ident string) (*WorkstationJobs, error) // filter: running

	// GetWorkstation gets the workstation for the given user including the configuration
	GetWorkstation(ctx context.Context, user *User) (*WorkstationOutput, error)

	// GetWorkstationLogs gets the logs for the workstation
	GetWorkstationLogs(ctx context.Context, user *User) (*WorkstationLogs, error)

	// GetWorkstationOptions gets the options for creating a new workstation
	GetWorkstationOptions(ctx context.Context) (*WorkstationOptions, error)

	// EnsureWorkstation creates a new workstation including the necessary service account, permissions and configuration
	EnsureWorkstation(ctx context.Context, user *User, input *WorkstationInput) (*WorkstationOutput, error)

	// DeleteWorkstation deletes the workstation configuration which also will delete the running workstation
	DeleteWorkstation(ctx context.Context, user *User) error

	// UpdateWorkstationURLList updates the URL allow list for the workstation
	UpdateWorkstationURLList(ctx context.Context, user *User, input *WorkstationURLList) error

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
	GetWorkstationConfig(ctx context.Context, opts *WorkstationConfigGetOpts) (*WorkstationConfig, error)

	CreateWorkstation(ctx context.Context, opts *WorkstationOpts) (*Workstation, error)
	GetWorkstation(ctx context.Context, id *WorkstationIdentifier) (*Workstation, error)
	StartWorkstation(ctx context.Context, id *WorkstationIdentifier) error
	StopWorkstation(ctx context.Context, id *WorkstationIdentifier) error

	AddWorkstationUser(ctx context.Context, id *WorkstationIdentifier, email string) error
}

type WorkstationsStorage interface {
	GetWorkstationJob(ctx context.Context, jobID int64) (*WorkstationJob, error)
	CreateWorkstationJob(ctx context.Context, opts *WorkstationJobOpts) (*WorkstationJob, error)
	GetWorkstationJobsForUser(ctx context.Context, ident string) ([]*WorkstationJob, error) // filter: running
}

type WorkstationJobs struct {
	Jobs []*WorkstationJob `json:"jobs"`
}

type WorkstationJob struct {
	ID int64 `json:"id"`

	Name  string `json:"name"`
	Email string `json:"email"`
	Ident string `json:"ident"`

	MachineType     string   `json:"machineType"`
	ContainerImage  string   `json:"containerImage"`
	URLAllowList    []string `json:"urlAllowList"`
	OnPremAllowList []string `json:"onPremAllowList"`

	StartTime time.Time           `json:"startTime"`
	State     WorkstationJobState `json:"state"`
	Duplicate bool                `json:"duplicate"`
	Errors    []string            `json:"errors"`
}

type WorkstationJobOpts struct {
	User  *User
	Input *WorkstationInput
}

type WorkstationJobState string

const (
	WorkstationJobStateCompleted WorkstationJobState = "COMPLETED"
	WorkstationJobStateRunning   WorkstationJobState = "RUNNING"
	WorkstationJobStateFailed    WorkstationJobState = "FAILED"
)

type WorkstationMachineType struct {
	MachineType string `json:"machineType"`
	VCPU        int    `json:"vCPU"`
	MemoryGB    int    `json:"memoryGB"`
}

// WorkstationMachineTypes returns the available machine types for workstations
// - https://cloud.google.com/compute/docs/general-purpose-machines#n2d_machine_types
func WorkstationMachineTypes() []*WorkstationMachineType {
	return []*WorkstationMachineType{
		{
			MachineType: MachineTypeN2DStandard2,
			VCPU:        2,
			MemoryGB:    8,
		},
		{
			MachineType: MachineTypeN2DStandard4,
			VCPU:        4,
			MemoryGB:    16,
		},
		{
			MachineType: MachineTypeN2DStandard8,
			VCPU:        8,
			MemoryGB:    32,
		},
		{
			MachineType: MachineTypeN2DStandard16,
			VCPU:        16,
			MemoryGB:    64,
		},
		{
			MachineType: MachineTypeN2DStandard32,
			VCPU:        32,
			MemoryGB:    128,
		},
	}
}

type WorkstationContainer struct {
	Image       string            `json:"image"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
}

func WorkstationContainers() []*WorkstationContainer {
	return []*WorkstationContainer{
		{
			Image:       ContainerImageVSCode,
			Description: "Visual Studio Code",
		},
		{
			Image:       ContainerImageIntellijUltimate,
			Description: "IntelliJ Ultimate",
		},
		{
			Image:       ContainerImagePosit,
			Description: "Posit Workbench",
		},
	}
}

type WorkstationLogs struct {
	ProxyDeniedHostPaths []*LogEntry `json:"proxyDeniedHostPaths"`
}

type WorkstationOptions struct {
	// FirewallTags is a list of possible firewall tags that can be used
	FirewallTags []*FirewallTag `json:"firewallTags"`

	// Container images that are allowed to be used
	ContainerImages []*WorkstationContainer `json:"containerImages"`

	// Machine types that are allowed to be used
	MachineTypes []*WorkstationMachineType `json:"machineTypes"`
}

type FirewallTag struct {
	Name      string `json:"name"`
	SecureTag string `json:"secureTag"`
}

type WorkstationURLList struct {
	// URLAllowList is a list of the URLs allowed to access from workstation
	URLAllowList []string `json:"urlAllowList"`
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

	// OnPremAllowList is a list of the on-premises hosts allowed to access from workstation
	OnPremAllowList []string `json:"onPremAllowList"`
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

	// Annotations are free-form annotations used to persist information
	Annotations map[string]string

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
		validation.Field(&o.ContainerImage, validation.Required),
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

	// Annotations are free-form annotations used to persist information
	Annotations map[string]string

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

	// The firewall rules that the user has associated with their workstation
	FirewallRulesAllowList []string `json:"firewallRulesAllowList"`

	// Environment variables passed to the container's entrypoint.
	Env map[string]string `json:"env"`
}

type WorkstationOutput struct {
	Slug        string `json:"slug"`
	DisplayName string `json:"displayName"`

	// Creating indicates whether this workstation is currently being created.
	Creating bool `json:"creating"`

	// Indicates whether the request was considered a duplicate
	DuplicateRequest bool `json:"duplicateRequest"`

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

type WorkstationIdentifier struct {
	Slug                  string
	WorkstationConfigSlug string
}

func DefaultWorkstationLabels(slug string) map[string]string {
	return map[string]string{
		LabelCreatedBy:    DefaultCreatedBy,
		LabelSubjectIdent: slug,
	}
}

func WorkstationServiceAccountID(slug string) string {
	return fmt.Sprintf("workstation-%s", slug)
}

func WorkstationDeniedRequestsLoggingResourceName(project, location, bucket, view string) string {
	return fmt.Sprintf("projects/%s/locations/%s/buckets/%s/views/%s", project, location, bucket, view)
}

func WorkstationDeniedRequestsLoggingFilter(policyName, ruleName, timestamp string) string {
	return fmt.Sprintf(`jsonPayload.enforcedGatewaySecurityPolicy.matchedRules.name:"/gatewaySecurityPolicies/%s/rules/%s" AND timestamp>"%s"`, policyName, ruleName, timestamp)
}

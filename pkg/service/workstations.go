package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

const (
	LabelApp          = "app"
	LabelCreatedBy    = "created-by"
	LabelSubjectIdent = "subject-ident"

	DefaultCreatedBy = "datamarkedsplassen"
	DefaultAppKnast  = "knast"

	MachineTypeN2DStandard2  = "n2d-standard-2"
	MachineTypeN2DStandard4  = "n2d-standard-4"
	MachineTypeN2DStandard8  = "n2d-standard-8"
	MachineTypeN2DStandard16 = "n2d-standard-16"
	MachineTypeN2DStandard32 = "n2d-standard-32"

	ContainerImageVSCode           = "europe-north1-docker.pkg.dev/cloud-workstations-images/predefined/code-oss:latest"
	ContainerImageIntellijUltimate = "europe-north1-docker.pkg.dev/cloud-workstations-images/predefined/intellij-ultimate:latest"
	ContainerImagePosit            = "europe-north1-docker.pkg.dev/posit-images/cloud-workstations/workbench:latest"

	WorkstationDiffDisableGlobalURLAllowList = "disable_global_url_allow_list"
	WorkstationDiffContainerImage            = "container_image"
	WorkstationDiffMachineType               = "machine_type"
	WorkstationDiffURLAllowList              = "url_allow_list"
	WorkstationDiffOnPremAllowList           = "on_prem_allow_list"

	WorkstationUserRole = "roles/workstations.user"

	WorkstationImagesTag = "latest"

	WorkstationDisableGlobalURLAllowListAnnotation = "disable-global-url-allow-list"
	WorkstationOnpremAllowlistAnnotation           = "onprem-allowlist"

	// WorkstationConfigIDLabel is a label applied to the running workstation by GCP
	WorkstationConfigIDLabel = "workstation_config_id"

	DefaultWorkstationProxyURL    = "http://proxy.knada.local:443"
	DefaultWorkstationNoProxyList = ".adeo.no,.preprod.local,.test.local,.intern.nav.no,.intern.dev.nav.no,.nais.adeo.no,localhost,metadata.google.internal"

	SecureWebProxyCertFile = "/usr/local/share/ca-certificates/swp.crt"

	// WorkstationEffectiveTagGCPKeyParentName is the key for the parent name in the effective tag set by Google themselves
	// we use this to filter out the tags that are not created by us
	WorkstationEffectiveTagGCPKeyParentName = "organizations/433637338589"
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

	// DeleteWorkstationByUser deletes the workstation configuration which also will delete the running workstation
	DeleteWorkstationByUser(ctx context.Context, user *User) error

	// DeleteWorkstationBySlug deletes the workstation configuration which also will delete the running workstation
	DeleteWorkstationBySlug(ctx context.Context, slug string) error

	// GetWorkstationURLList gets the URL allow list for the workstation
	GetWorkstationURLList(ctx context.Context, user *User) (*WorkstationURLList, error)

	// UpdateWorkstationURLList updates the URL allow list for the workstation
	UpdateWorkstationURLList(ctx context.Context, user *User, input *WorkstationURLList) error

	// CreateWorkstationStartJob creates a job to start the workstation
	CreateWorkstationStartJob(ctx context.Context, user *User) (*WorkstationStartJob, error)

	// GetWorkstationStartJob gets the start jobs for the given id
	GetWorkstationStartJob(ctx context.Context, id int64) (*WorkstationStartJob, error)

	// GetWorkstationStartJobsForUser gets the start jobs for the given user
	GetWorkstationStartJobsForUser(ctx context.Context, ident string) (*WorkstationStartJobs, error)

	// StartWorkstation starts the workstation
	StartWorkstation(ctx context.Context, userStart *User) error

	// GetWorkstationOnpremMapping gets the on-prem allowlist for the workstation
	GetWorkstationOnpremMapping(ctx context.Context, user *User) (*WorkstationOnpremAllowList, error)

	// UpdateWorkstationOnpremMapping updates the on-prem allowlist for the workstation
	UpdateWorkstationOnpremMapping(ctx context.Context, user *User, onpremAllowList *WorkstationOnpremAllowList) error

	// UpdateWorkstationZonalTagBindingsForUser updates the zonal tag bindings for the workstation
	UpdateWorkstationZonalTagBindingsForUser(ctx context.Context, ident string) error

	// CreateWorkstationZonalTagBindingsJobForUser creates a job to add or remove a zonal tag binding to the workstation
	CreateWorkstationZonalTagBindingsJobForUser(ctx context.Context, ident string) (*WorkstationZonalTagBindingsJob, error)

	// GetWorkstationZonalTagBindingsJobsForUser gets the zonal tag binding job with the given id
	GetWorkstationZonalTagBindingsJobsForUser(ctx context.Context, ident string) ([]*WorkstationZonalTagBindingsJob, error)

	// GetWorkstationZonalTagBindings gets the zonal tag bindings for the workstation
	GetWorkstationZonalTagBindings(ctx context.Context, ident string) ([]*EffectiveTag, error)

	// AddWorkstationZonalTagBinding adds a zonal tag binding to the workstation
	AddWorkstationZonalTagBinding(ctx context.Context, zone, parent, tagNamespacedName string) error

	// RemoveWorkstationZonalTagBinding removes a zonal tag binding from the workstation
	RemoveWorkstationZonalTagBinding(ctx context.Context, zone, parent, tagValue string) error

	// StopWorkstation stops the workstation
	StopWorkstation(ctx context.Context, user *User) error

	// ListWorkstations lists all workstations
	ListWorkstations(ctx context.Context) ([]*WorkstationOutput, error)
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

	ListWorkstationConfigs(ctx context.Context) ([]*WorkstationConfig, error)
}

type WorkstationsQueue interface {
	GetWorkstationJob(ctx context.Context, jobID int64) (*WorkstationJob, error)
	CreateWorkstationJob(ctx context.Context, opts *WorkstationJobOpts) (*WorkstationJob, error)
	GetWorkstationJobsForUser(ctx context.Context, ident string) ([]*WorkstationJob, error) // filter: running

	GetWorkstationStartJob(ctx context.Context, id int64) (*WorkstationStartJob, error)
	CreateWorkstationStartJob(ctx context.Context, ident string) (*WorkstationStartJob, error)
	GetWorkstationStartJobsForUser(ctx context.Context, ident string) ([]*WorkstationStartJob, error)

	GetWorkstationZonalTagBindingsJob(ctx context.Context, jobID int64) (*WorkstationZonalTagBindingsJob, error)
	CreateWorkstationZonalTagBindingsJob(ctx context.Context, ident string) (*WorkstationZonalTagBindingsJob, error)
	GetWorkstationZonalTagBindingsJobsForUser(ctx context.Context, ident string) ([]*WorkstationZonalTagBindingsJob, error)
}

type WorkstationsStorage interface {
	CreateWorkstationsConfigChange(ctx context.Context, navIdent string, config json.RawMessage) error
	CreateWorkstationsOnpremAllowListChange(ctx context.Context, navIdent string, hosts []string) error
	GetLastWorkstationsOnpremAllowList(ctx context.Context, navIdent string) ([]string, error)

	CreateWorkstationsURLListChange(ctx context.Context, navIdent string, input *WorkstationURLList) error
	GetLastWorkstationsURLList(ctx context.Context, navIdent string) (*WorkstationURLList, error)
}

type WorkstationOnpremAllowList struct {
	Hosts []string `json:"hosts"`
}

type WorkstationZonalTagBindingsJobOpts struct {
	Ident string `json:"ident"`
}

type WorkstationZonalTagBindingsJobs struct {
	Jobs []*WorkstationZonalTagBindingsJob `json:"jobs"`
}

type WorkstationZonalTagBindingsJob struct {
	ID int64 `json:"id"`

	Ident string `json:"ident"`

	StartTime time.Time           `json:"startTime"`
	State     WorkstationJobState `json:"state"`
	Duplicate bool                `json:"duplicate"`
	Errors    []string            `json:"errors"`
}

type WorkstationStartJobs struct {
	Jobs []*WorkstationStartJob `json:"jobs"`
}

type WorkstationStartJob struct {
	ID int64 `json:"id"`

	Ident string `json:"ident"`

	StartTime time.Time           `json:"startTime"`
	State     WorkstationJobState `json:"state"`
	Duplicate bool                `json:"duplicate"`
	Errors    []string            `json:"errors"`
}

type WorkstationJobs struct {
	Jobs []*WorkstationJob `json:"jobs"`
}

type WorkstationJob struct {
	ID int64 `json:"id"`

	Name  string `json:"name"`
	Email string `json:"email"`
	Ident string `json:"ident"`

	MachineType    string `json:"machineType"`
	ContainerImage string `json:"containerImage"`

	StartTime time.Time           `json:"startTime"`
	State     WorkstationJobState `json:"state"`
	Duplicate bool                `json:"duplicate"`
	Errors    []string            `json:"errors"`

	Diff map[string]*Diff `json:"diff"`
}

type Diff struct {
	Added   []string `json:"added"`
	Removed []string `json:"removed"`
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
	Image         string            `json:"image"`
	Description   string            `json:"description"`
	Labels        map[string]string `json:"labels"`
	Documentation string            `json:"documentation"`
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

	// Global URL allow list
	GlobalURLAllowList []string `json:"globalURLAllowList"`
}

type FirewallTag struct {
	Name      string `json:"name"`
	SecureTag string `json:"secureTag"`
}

type WorkstationURLList struct {
	// URLAllowList is a list of the URLs allowed to access from workstation
	URLAllowList []string `json:"urlAllowList"`

	// DisableGlobalAllowList is a flag to disable the global URL allow list
	DisableGlobalAllowList bool `json:"disable"`
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

	Env map[string]string

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

	// Environment variables passed to the container's entrypoint.
	Env map[string]string
}

type WorkstationConfigDeleteOpts struct {
	// Slug is the unique identifier of the workstation
	Slug string
	// NoWait is a flag to indicate if the request should wait for the operation to complete
	NoWait bool
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

	// Complete workstation configuration as returned by the Google API.
	CompleteConfigAsJSON []byte
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

	Config *WorkstationConfigOutput `json:"config"`

	Host string `json:"host"`
}

type WorkstationIdentifier struct {
	Slug                  string
	WorkstationConfigSlug string
}

func DefaultWorkstationLabels(slug string) map[string]string {
	return map[string]string{
		LabelApp:          DefaultAppKnast,
		LabelCreatedBy:    DefaultCreatedBy,
		LabelSubjectIdent: slug,
	}
}

func DefaultWorkstationEnv(ident, email, fullName string) map[string]string {
	return map[string]string{
		"WORKSTATION_NAME":           ident,
		"WORKSTATION_USER_EMAIL":     email,
		"WORKSTATION_USER_FULL_NAME": fullName,

		"http_proxy": DefaultWorkstationProxyURL,
		"HTTP_PROXY": DefaultWorkstationProxyURL,

		"https_proxy": DefaultWorkstationProxyURL,
		"HTTPS_PROXY": DefaultWorkstationProxyURL,

		"no_proxy": DefaultWorkstationNoProxyList,
		"NO_PROXY": DefaultWorkstationNoProxyList,

		"NODE_EXTRA_CA_CERTS": SecureWebProxyCertFile,
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

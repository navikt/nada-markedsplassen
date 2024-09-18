package service

import (
	"context"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

const (
	LabelCreatedBy    = "created-by"
	LabelSubjectEmail = "subject-email"

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
}

type WorkstationsAPI interface {
	EnsureWorkstationWithConfig(ctx context.Context, opts *EnsureWorkstationOpts) error
	CreateWorkstationConfig(ctx context.Context, opts *WorkstationConfigOpts) (*WorkstationConfig, error)
	UpdateWorkstationConfig(ctx context.Context, opts *WorkstationConfigUpdateOpts) (*WorkstationConfig, error)
	DeleteWorkstationConfig(ctx context.Context, opts *WorkstationConfigDeleteOpts) error
	CreateWorkstation(ctx context.Context, opts *WorkstationOpts) (*Workstation, error)
	GetWorkstation(ctx context.Context, opts *WorkstationGetOpts) (*Workstation, error)
	GetWorkstationConfig(ctx context.Context, opts *WorkstationConfigGetOpts) (*WorkstationConfig, error)
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

type WorkstationConfig struct {
	Name        string
	DisplayName string
}

type WorkstationConfigGetOpts struct {
	Slug string
}

type Workstation struct {
	Name string
}

type WorkstationOutput struct {
	Workstation       *Workstation
	WorkstationConfig *WorkstationConfig
}

type WorkstationGetOpts struct {
	// Slug is the unique identifier of the workstation
	Slug       string
	ConfigName string
}

func DefaultWorkstationLabels(subjectEmail string) map[string]string {
	return map[string]string{
		LabelCreatedBy:    "datamarkedsplassen",
		LabelSubjectEmail: subjectEmail,
	}
}

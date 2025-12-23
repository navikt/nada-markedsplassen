package service

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AccessStorage interface {
	CreateAccessRequestForDataset(ctx context.Context, datasetID uuid.UUID, pollyDocumentationID uuid.NullUUID, subject, owner, platform string, expires *time.Time) (*AccessRequest, error)
	DeleteAccessRequest(ctx context.Context, accessRequestID uuid.UUID) error
	DenyAccessRequest(ctx context.Context, user *User, accessRequestID uuid.UUID, reason *string) error
	GetAccessRequest(ctx context.Context, accessRequestID uuid.UUID) (*AccessRequest, error)
	GetAccessToDataset(ctx context.Context, id uuid.UUID) (*Access, error)
	GetUnrevokedExpiredAccess(ctx context.Context) ([]*Access, error)
	GrantAccessToDatasetAndApproveRequest(ctx context.Context, user *User, subjectWithType string, accessRequest *AccessRequest) error
	GrantAccessToDatasetAndRenew(ctx context.Context, datasetID uuid.UUID, expires *time.Time, subject, owner, granter, platform string) error
	ListAccessRequestsForDataset(ctx context.Context, datasetID uuid.UUID) ([]AccessRequest, error)
	ListAccessRequestsForOwner(ctx context.Context, owner []string) ([]AccessRequest, error)
	ListActiveAccessToDataset(ctx context.Context, datasetID uuid.UUID) ([]*Access, error)
	RevokeAccessToDataset(ctx context.Context, id uuid.UUID) error
	UpdateAccessRequest(ctx context.Context, input UpdateAccessRequestDTO) error
	GetDatasetIDFromAccessRequest(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	GetUserAccesses(ctx context.Context, user *User) (*UserAccesses, error)
	GetAllUserAccesses(ctx context.Context) ([]UserAccessDataproduct, error)
}

type AccessService interface {
	GetAccessRequests(ctx context.Context, datasetID uuid.UUID) (*AccessRequestsWrapper, error)
	CreateAccessRequest(ctx context.Context, user *User, input NewAccessRequestDTO) error
	DeleteAccessRequest(ctx context.Context, user *User, accessRequestID uuid.UUID) error
	UpdateAccessRequest(ctx context.Context, input UpdateAccessRequestDTO) error
	ApproveAccessRequest(ctx context.Context, user *User, accessRequestID uuid.UUID) error
	DenyAccessRequest(ctx context.Context, user *User, accessRequestID uuid.UUID, reason *string) error
	GrantBigQueryAccessToDataset(ctx context.Context, user *User, input GrantAccessData, gcpProjectID string) error
	GrantMetabaseAccessAllUsersToDataset(ctx context.Context, user *User, input GrantAccessData) error
	GrantMetabaseAccessRestrictedToDataset(ctx context.Context, user *User, input GrantAccessData) error
	RevokeMetabaseRestrictedAccessToDataset(ctx context.Context, access *Access) error
	RevokeMetabaseAllUsersAccessToDataset(ctx context.Context, access *Access) error
	RevokeBigQueryAccessToDataset(ctx context.Context, access *Access, gcpProjectID string) error
	GetAccessToDataset(ctx context.Context, id uuid.UUID) (*Access, error)
	EnsureUserIsAuthorizedToRevokeAccess(ctx context.Context, user *User, access *Access) error
	EnsureUserIsDatasetOwner(ctx context.Context, user *User, datasetID uuid.UUID) error
	EnsureUserIsAuthorizedToApproveRequest(ctx context.Context, user *User, accessID uuid.UUID) error
	GetUserAccesses(ctx context.Context, user *User) (*UserAccesses, error)
	GetAllUserAccesses(ctx context.Context, user *User) ([]UserAccessDataproduct, error)
}

type DatasetAccess struct {
	Subject string    `json:"subject"`
	Active  []*Access `json:"active"`
	Revoked []*Access `json:"revoked"`
}

type Access struct {
	ID            uuid.UUID      `json:"id"`
	Subject       string         `json:"subject"`
	Owner         string         `json:"owner"`
	Granter       string         `json:"granter"`
	Expires       *time.Time     `json:"expires"`
	Created       time.Time      `json:"created"`
	Revoked       *time.Time     `json:"revoked"`
	DatasetID     uuid.UUID      `json:"datasetID"`
	AccessRequest *AccessRequest `json:"accessRequest"`
	Platform      string         `json:"platform"`
}

type NewAccessRequestDTO struct {
	DatasetID   uuid.UUID   `json:"datasetID"`
	Subject     string      `json:"subject"`
	SubjectType string      `json:"subjectType"`
	Owner       string      `json:"owner"`
	Expires     *time.Time  `json:"expires"`
	Polly       *PollyInput `json:"polly"`
	Platform    string      `json:"platform"`
}

type UpdateAccessRequestDTO struct {
	ID      uuid.UUID   `json:"id"`
	Owner   string      `json:"owner"`
	Expires *time.Time  `json:"expires"`
	Polly   *PollyInput `json:"polly"`
}

type AccessRequest struct {
	ID          uuid.UUID           `json:"id"`
	DatasetID   uuid.UUID           `json:"datasetID"`
	Subject     string              `json:"subject"`
	SubjectType string              `json:"subjectType"`
	Created     time.Time           `json:"created"`
	Status      AccessRequestStatus `json:"status"`
	Closed      *time.Time          `json:"closed"`
	Expires     *time.Time          `json:"expires"`
	Granter     *string             `json:"granter"`
	Owner       string              `json:"owner"`
	Polly       *Polly              `json:"polly"`
	Reason      *string             `json:"reason"`
	Platform    string              `json:"platform"`
}

type AccessRequestForGranter struct {
	AccessRequest   `tstype:",extends"`
	DataproductID   uuid.UUID `json:"dataproductID"`
	DataproductSlug string    `json:"dataproductSlug"`
	DatasetName     string    `json:"datasetName"`
	DataproductName string    `json:"dataproductName"`
}

type AccessRequestsWrapper struct {
	AccessRequests []AccessRequest `json:"accessRequests"`
}

const (
	SubjectTypeUser           string = "user"
	SubjectTypeGroup          string = "group"
	SubjectTypeServiceAccount string = "serviceAccount"
)

type GrantAccessData struct {
	DatasetID   uuid.UUID  `json:"datasetID"`
	Expires     *time.Time `json:"expires"`
	Subject     string     `json:"subject"`
	Owner       string     `json:"owner"`
	SubjectType string     `json:"subjectType"`
}

type UserAccessDatasets struct {
	DatasetID          uuid.UUID `json:"datasetID"`
	DatasetName        string    `json:"datasetName"`
	DatasetDescription string    `json:"datasetDescription"`
	DatasetSlug        string    `json:"datasetSlug"`
	Accesses           []Access  `json:"accesses"`
}

type UserAccessDataproduct struct {
	DataproductID          uuid.UUID            `json:"dataproductID"`
	DataproductName        string               `json:"dataproductName"`
	DataproductDescription string               `json:"dataproductDescription"`
	DataproductSlug        string               `json:"dataproductSlug"`
	DataproductGroup       string               `json:"dataproductGroup"`
	Datasets               []UserAccessDatasets `json:"datasets"`
}

type UserAccesses struct {
	Personal        []UserAccessDataproduct            `json:"personal"`
	ServiceAccounts map[string][]UserAccessDataproduct `json:"serviceAccountGranted"`
}

type AccessRequestStatus string

const (
	AccessRequestStatusPending  AccessRequestStatus = "pending"
	AccessRequestStatusApproved AccessRequestStatus = "approved"
	AccessRequestStatusDenied   AccessRequestStatus = "denied"
)

const (
	AccessPlatformBigQuery = "bigquery"
	AccessPlatformMetabase = "metabase"
)

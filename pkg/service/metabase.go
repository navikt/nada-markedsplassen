package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	MetabaseRestrictedCollectionTag = "🔐"
	MetabaseAllUsersGroupID         = 1
)

type MetabaseStorage interface {
	CreateMetadata(ctx context.Context, datasetID uuid.UUID) error
	DeleteMetadata(ctx context.Context, datasetID uuid.UUID) error
	DeleteRestrictedMetadata(ctx context.Context, datasetID uuid.UUID) error
	GetAllMetadata(ctx context.Context) ([]*MetabaseMetadata, error)
	GetMetadata(ctx context.Context, datasetID uuid.UUID, includeDeleted bool) (*MetabaseMetadata, error)
	GetOpenTablesInSameBigQueryDataset(ctx context.Context, projectID, dataset string) ([]string, error)
	RestoreMetadata(ctx context.Context, datasetID uuid.UUID) error
	SetCollectionMetabaseMetadata(ctx context.Context, datasetID uuid.UUID, collectionID int) (*MetabaseMetadata, error)
	SetDatabaseMetabaseMetadata(ctx context.Context, datasetID uuid.UUID, databaseID int) (*MetabaseMetadata, error)
	SetPermissionGroupMetabaseMetadata(ctx context.Context, datasetID uuid.UUID, groupID int) (*MetabaseMetadata, error)
	SetServiceAccountMetabaseMetadata(ctx context.Context, datasetID uuid.UUID, saEmail string) (*MetabaseMetadata, error)
	SetSyncCompletedMetabaseMetadata(ctx context.Context, datasetID uuid.UUID) error
	SoftDeleteMetadata(ctx context.Context, datasetID uuid.UUID) error
}

type MetabaseAPI interface {
	AddPermissionGroupMember(ctx context.Context, groupID int, userID int) error
	ArchiveCollection(ctx context.Context, colID int) error
	AutoMapSemanticTypes(ctx context.Context, dbID int) error
	CreateCollection(ctx context.Context, name string) (int, error)
	CreateCollectionWithAccess(ctx context.Context, groupID int, name string, removeAllUsersAccess bool) (int, error)
	CreateDatabase(ctx context.Context, team, name, saJSON, saEmail string, ds *BigQuery) (int, error)
	CreatePermissionGroup(ctx context.Context, name string) (int, error)
	CreateUser(ctx context.Context, email string) (*MetabaseUser, error)
	Database(ctx context.Context, dbID int) (*MetabaseDatabase, error)
	Databases(ctx context.Context) ([]MetabaseDatabase, error)
	DeleteDatabase(ctx context.Context, id int) error
	DeletePermissionGroup(ctx context.Context, groupID int) error
	DeleteUser(ctx context.Context, id int) error
	FindUserByEmail(ctx context.Context, email string) (*MetabaseUser, error)
	GetCollections(ctx context.Context) ([]*MetabaseCollection, error)
	GetOrCreatePermissionGroup(ctx context.Context, name string) (int, error)
	GetPermissionGraphForGroup(ctx context.Context, groupID int) (*PermissionGraphGroups, error)
	GetPermissionGroup(ctx context.Context, groupID int) ([]MetabasePermissionGroupMember, error)
	GetPermissionGroups(ctx context.Context) ([]MetabasePermissionGroup, error)
	GetUsers(ctx context.Context) ([]MetabaseUser, error)
	HideTables(ctx context.Context, ids []int) error
	OpenAccessToDatabase(ctx context.Context, databaseID int) error
	RemovePermissionGroupMember(ctx context.Context, memberID int) error
	RestrictAccessToDatabase(ctx context.Context, groupID int, databaseID int) error
	SetCollectionAccess(ctx context.Context, groupID int, collectionID int, removeAllUsersAccess bool) error
	ShowTables(ctx context.Context, ids []int) error
	Tables(ctx context.Context, dbID int, includeHidden bool) ([]MetabaseTable, error)
	UpdateCollection(ctx context.Context, collection *MetabaseCollection) error
	UpdateDatabase(ctx context.Context, dbID int, saJSON, saEmail string) error
}

type MetabaseQueue interface {
	CreateRestrictedMetabaseBigqueryDatabaseWorkflow(ctx context.Context, opts *MetabaseRestrictedBigqueryDatabaseWorkflowOpts) (*MetabaseRestrictedBigqueryDatabaseWorkflowStatus, error)
	GetRestrictedMetabaseBigqueryDatabaseWorkflow(ctx context.Context, datasetID uuid.UUID) (*MetabaseRestrictedBigqueryDatabaseWorkflowStatus, error)

	CreateMetabaseBigqueryDatabaseDeleteJob(ctx context.Context, datasetID uuid.UUID) (*MetabaseBigqueryDatabaseDeleteJob, error)
	GetMetabaseBigqueryDatabaseDeleteJob(ctx context.Context, datasetID uuid.UUID) (*MetabaseBigqueryDatabaseDeleteJob, error)
}

type MetabaseService interface {
	SyncTableVisibility(ctx context.Context, mbMeta *MetabaseMetadata, bq BigQuery) error
	SyncAllTablesVisibility(ctx context.Context) error
	RevokeMetabaseAccess(ctx context.Context, dsID uuid.UUID, subject string) error
	RevokeMetabaseAccessFromAccessID(ctx context.Context, accessID uuid.UUID) error
	DeleteDatabase(ctx context.Context, dsID uuid.UUID) error
	GrantMetabaseAccess(ctx context.Context, dsID uuid.UUID, subject, subjectType string) error
	CreateMappingRequest(ctx context.Context, user *User, datasetID uuid.UUID, services []string) error
	MapDataset(ctx context.Context, datasetID uuid.UUID, services []string) error

	// FIXME: Need to create services for permission group, etc.
	CreateRestrictedMetabaseBigqueryDatabase(ctx context.Context, datasetID uuid.UUID) error
	VerifyRestrictedMetabaseBigqueryDatabase(ctx context.Context, datasetID uuid.UUID) error
	FinalizeRestrictedMetabaseBigqueryDatabase(ctx context.Context, datasetID uuid.UUID) error
	DeleteRestrictedMetabaseBigqueryDatabase(ctx context.Context, datasetID uuid.UUID) error
}

type MetabaseRestrictedBigqueryDatabaseWorkflowStatus struct {
	PermissionGroupJob *MetabaseCreatePermissionGroupJob      `json:"permissionGroupJob"`
	CollectionJob      *MetabaseCreateCollectionJob           `json:"collectionJob"`
	ServiceAccountJob  *MetabaseEnsureServiceAccountJob       `json:"serviceAccountJob"`
	ProjectIAMJob      *MetabaseAddProjectIAMPolicyBindingJob `json:"projectIAMJob"`
	DatabaseJob        *MetabaseBigqueryCreateDatabaseJob     `json:"databaseJob"`
	VerifyJob          *MetabaseBigqueryVerifyDatabaseJob     `json:"verifyJob"`
	FinalizeJob        *MetabaseBigqueryFinalizeDatabaseJob   `json:"finalizeJob"`
}

type MetabaseCreatePermissionGroupJob struct {
	JobHeader `json:",inline" tstype:",extends"`

	DatasetID           uuid.UUID `json:"datasetID"`
	PermissionGroupName string    `json:"permissionGroupName"`
}

type MetabaseCreateCollectionJob struct {
	JobHeader `json:",inline" tstype:",extends"`

	DatasetID      uuid.UUID `json:"datasetID"`
	CollectionName string    `json:"collectionName"`
}

type MetabaseEnsureServiceAccountJob struct {
	JobHeader `json:",inline" tstype:",extends"`

	DatasetID   uuid.UUID `json:"datasetID"`
	AccountID   string    `json:"accountID"`
	ProjectID   string    `json:"projectID"`
	DisplayName string    `json:"displayName"`
	Description string    `json:"description"`
}

type MetabaseAddProjectIAMPolicyBindingJob struct {
	JobHeader `json:",inline" tstype:",extends"`

	DatasetID uuid.UUID `json:"datasetID"`
	ProjectID string    `json:"projectID"`
	Role      string    `json:"role"`
	Member    string    `json:"member"`
}

type MetabaseBigqueryCreateDatabaseJob struct {
	JobHeader `json:",inline" tstype:",extends"`

	DatasetID uuid.UUID `json:"datasetID"`
}

type MetabaseBigqueryVerifyDatabaseJob struct {
	JobHeader `json:",inline" tstype:",extends"`

	DatasetID uuid.UUID `json:"datasetID"`
}

type MetabaseBigqueryFinalizeDatabaseJob struct {
	JobHeader `json:",inline" tstype:",extends"`

	DatasetID uuid.UUID `json:"datasetID"`
}

type MetabaseBigqueryDatabaseDeleteJob struct {
	JobHeader `json:",inline" tstype:",extends"`

	DatasetID uuid.UUID `json:"datasetID"`
}

type MetabaseRestrictedBigqueryDatabaseWorkflowOpts struct {
	DatasetID uuid.UUID

	PermissionGroupName string
	CollectionName      string

	ProjectID   string
	AccountID   string
	DisplayName string
	Description string

	Role   string
	Member string
}

type MetabaseField struct {
	ID           int    `json:"id"`
	DatabaseType string `json:"database_type"`
	SemanticType string `json:"semantic_type"`
}

type MetabaseTable struct {
	Name   string          `json:"name"`
	ID     int             `json:"id"`
	Fields []MetabaseField `json:"fields"`
}

type MetabasePermissionGroup struct {
	ID      int                             `json:"id"`
	Name    string                          `json:"name"`
	Members []MetabasePermissionGroupMember `json:"members"`
}

type MetabasePermissionGroupMember struct {
	ID    int    `json:"membership_id"`
	Email string `json:"email"`
}

type MetabaseUser struct {
	Email     string     `json:"email"`
	ID        int        `json:"id"`
	LastLogin *time.Time `json:"last_login"`
	IsActive  bool       `json:"is_active"`
}

type MetabaseDatabase struct {
	ID        int
	Name      string
	DatasetID string
	ProjectID string
	NadaID    string
	SAEmail   string
}

type MetabaseMetadata struct {
	DatasetID         uuid.UUID
	DatabaseID        *int
	PermissionGroupID *int
	CollectionID      *int
	SAEmail           string
	DeletedAt         *time.Time
	SyncCompleted     *time.Time
}

// MetabaseCollection represents a subset of the metadata returned
// for a Metabase collection
type MetabaseCollection struct {
	ID          int
	Name        string
	Description string
}

type PermissionGraphGroups struct {
	Revision int                                   `json:"revision"`
	Groups   map[string]map[string]PermissionGroup `json:"groups"`
}

type PermissionGroup struct {
	ViewData      string               `json:"view-data,omitempty"`
	CreateQueries string               `json:"create-queries,omitempty"`
	Details       string               `json:"details,omitempty"`
	Download      *DownloadPermission  `json:"download,omitempty"`
	DataModel     *DataModelPermission `json:"data-model,omitempty"`
}

type DataModelPermission struct {
	Schemas string `json:"schemas,omitempty"`
}

type DownloadPermission struct {
	Schemas string `json:"schemas,omitempty"`
}

func NadaMetabaseRole(gcpProject string) string {
	return fmt.Sprintf("projects/%s/roles/nada.metabase", gcpProject)
}

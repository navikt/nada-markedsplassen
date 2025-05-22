package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/rs/zerolog"

	"github.com/btcsuite/btcutil/base58"

	"github.com/gosimple/slug"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

const (
	sleeperTime = time.Second
	maxRetries  = 420
)

var _ service.MetabaseService = &metabaseService{}

type metabaseService struct {
	gcpProject          string
	serviceAccount      string
	serviceAccountEmail string
	groupAllUsers       string

	metabaseQueue           service.MetabaseQueue
	metabaseAPI             service.MetabaseAPI
	bigqueryAPI             service.BigQueryAPI
	serviceAccountAPI       service.ServiceAccountAPI
	cloudResourceManagerAPI service.CloudResourceManagerAPI

	metabaseStorage    service.MetabaseStorage
	bigqueryStorage    service.BigQueryStorage
	dataproductStorage service.DataProductsStorage
	accessStorage      service.AccessStorage

	log zerolog.Logger
}

func (s *metabaseService) CreateRestrictedMetabaseBigqueryDatabaseWorkflow(ctx context.Context, user *service.User, datasetID uuid.UUID) (*service.MetabaseBigQueryDatasetStatus, error) {
	const op errs.Op = "metabaseService.CreateRestrictedMetabaseBigqueryDatabaseWorkflow"

	ds, err := s.dataproductStorage.GetDataset(ctx, datasetID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	dp, err := s.dataproductStorage.GetDataproduct(ctx, ds.DataproductID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	if err := ensureUserInGroup(user, dp.Owner.Group); err != nil {
		return nil, errs.E(op, err)
	}

	permissionGroupName := slug.Make(fmt.Sprintf("%s-%s", ds.Name, MarshalUUID(ds.ID)))
	collectionName := fmt.Sprintf("%s %s", ds.Name, service.MetabaseRestrictedCollectionTag)

	_, err = s.metabaseQueue.CreateRestrictedMetabaseBigqueryDatabaseWorkflow(ctx, &service.MetabaseRestrictedBigqueryDatabaseWorkflowOpts{
		DatasetID:           ds.ID,
		PermissionGroupName: permissionGroupName,
		CollectionName:      collectionName,
		AccountID:           AccountIDFromDatasetID(ds.ID),
		ProjectID:           ds.Datasource.ProjectID,
		DisplayName:         ds.Name,
		Description:         fmt.Sprintf("Metabase service account for dataset %s", ds.ID.String()),
		Role:                service.NadaMetabaseRole(s.gcpProject),
		Member:              fmt.Sprintf("serviceAccount:%s", s.ConstantServiceAccountEmailFromDatasetID(ds.ID)),
	})
	if err != nil {
		return nil, errs.E(op, err)
	}

	wf, err := s.GetRestrictedMetabaseBigQueryDatabaseWorkflow(ctx, datasetID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return wf, nil
}

func (s *metabaseService) CreateOpenMetabaseBigqueryDatabaseWorkflow(ctx context.Context, user *service.User, datasetID uuid.UUID) (*service.MetabaseBigQueryDatasetStatus, error) {
	const op errs.Op = "metabaseService.CreateOpenMetabaseBigqueryDatabaseWorkflow"

	ds, err := s.dataproductStorage.GetDataset(ctx, datasetID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	dp, err := s.dataproductStorage.GetDataproduct(ctx, ds.DataproductID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	if err := ensureUserInGroup(user, dp.Owner.Group); err != nil {
		return nil, errs.E(op, err)
	}

	_, err = s.metabaseQueue.CreateOpenMetabaseBigqueryDatabaseWorkflow(ctx, datasetID)
	if err != nil {
		return nil, err
	}

	wf, err := s.GetOpenMetabaseBigQueryDatabaseWorkflow(ctx, datasetID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return wf, nil
}

func (s *metabaseService) PreflightCheckOpenBigqueryDatabase(ctx context.Context, datasetID uuid.UUID) error {
	const op errs.Op = "metabaseService.PreflightCheckOpenBigqueryDatabase"

	// FIXME: reuse the prepare step?
	meta, err := s.metabaseStorage.GetMetadata(ctx, datasetID, true)
	if err != nil && !errs.KindIs(errs.NotExist, err) {
		return errs.E(op, err)
	}

	if meta == nil {
		err := s.metabaseStorage.CreateMetadata(ctx, datasetID)
		if err != nil {
			return errs.E(op, err)
		}
	}

	accesses, err := s.accessStorage.ListActiveAccessToDataset(ctx, datasetID)
	if err != nil {
		return errs.E(op, err)
	}

	if !s.containsAllUsers(accesses) {
		return errs.E(errs.InvalidRequest, service.CodeMetabase, op, fmt.Errorf("does not contain all users group: %s", s.groupAllUsers))
	}

	return nil
}

func (s *metabaseService) CreateOpenMetabaseBigqueryDatabase(ctx context.Context, datasetID uuid.UUID) error {
	const op errs.Op = "metabaseService.CreateOpenMetabaseBigqueryDatabase"

	meta, err := s.metabaseStorage.GetMetadata(ctx, datasetID, true)
	if err != nil {
		return errs.E(op, err)
	}

	ds, err := s.dataproductStorage.GetDataset(ctx, datasetID)
	if err != nil {
		return errs.E(op, err)
	}

	datasource, err := s.bigqueryStorage.GetBigqueryDatasource(ctx, ds.ID, false)
	if err != nil {
		return errs.E(op, err)
	}

	err = s.bigqueryAPI.Grant(ctx, datasource.ProjectID, datasource.Dataset, datasource.Table, "serviceAccount:"+s.serviceAccountEmail)
	if err != nil {
		return errs.E(op, err)
	}

	dp, err := s.dataproductStorage.GetDataproduct(ctx, ds.DataproductID)
	if err != nil {
		return errs.E(op, err)
	}

	if meta.DatabaseID == nil {
		_, err = s.metabaseStorage.SetCollectionMetabaseMetadata(ctx, datasetID, 0)
		if err != nil {
			return errs.E(op, err)
		}

		dbID, err := s.metabaseAPI.CreateDatabase(ctx, dp.Owner.Group, ds.Name, s.serviceAccount, s.serviceAccountEmail, datasource)
		if err != nil {
			return errs.E(op, err)
		}

		meta, err = s.metabaseStorage.SetDatabaseMetabaseMetadata(ctx, ds.ID, dbID)
		if err != nil {
			return errs.E(op, err)
		}
	}

	return nil
}

func (s *metabaseService) VerifyOpenMetabaseBigqueryDatabase(ctx context.Context, datasetID uuid.UUID) error {
	const op errs.Op = "metabaseService.VerifyOpenMetabaseBigqueryDatabase"

	meta, err := s.metabaseStorage.GetMetadata(ctx, datasetID, true)
	if err != nil {
		return errs.E(op, err)
	}

	if meta.DatabaseID == nil {
		return errs.E(errs.NotExist, service.CodeMetabase, op, fmt.Errorf("database not found for dataset: %v", datasetID))
	}

	datasource, err := s.bigqueryStorage.GetBigqueryDatasource(ctx, datasetID, false)
	if err != nil {
		return errs.E(op, err)
	}

	tables, err := s.metabaseAPI.Tables(ctx, *meta.DatabaseID, false)
	if err != nil || len(tables) == 0 {
		return errs.E(errs.Internal, service.CodeWaitingForDatabase, op, fmt.Errorf("database not synced: %v", datasource.Table))
	}

	for _, tab := range tables {
		if tab.Name == datasource.Table && len(tab.Fields) > 0 {
			return nil
		}
	}

	return nil
}

func (s *metabaseService) FinalizeOpenMetabaseBigqueryDatabase(ctx context.Context, datasetID uuid.UUID) error {
	const op errs.Op = "metabaseService.FinalizeOpenMetabaseBigqueryDatabase"

	meta, err := s.metabaseStorage.GetMetadata(ctx, datasetID, true)
	if err != nil {
		return errs.E(op, err)
	}

	datasource, err := s.bigqueryStorage.GetBigqueryDatasource(ctx, datasetID, false)
	if err != nil {
		return errs.E(op, err)
	}

	if err := s.SyncTableVisibility(ctx, meta, *datasource); err != nil {
		return errs.E(op, err)
	}

	if err := s.metabaseAPI.AutoMapSemanticTypes(ctx, *meta.DatabaseID); err != nil {
		return errs.E(op, err)
	}

	if meta.DeletedAt != nil {
		if err := s.restore(ctx, datasetID, meta.SAEmail); err != nil {
			return errs.E(op, err)
		}
	}

	// All users database already exists in metabase
	if meta.PermissionGroupID != nil && *meta.PermissionGroupID == 0 {
		err = s.metabaseStorage.SetSyncCompletedMetabaseMetadata(ctx, datasetID)
		if err != nil {
			return errs.E(op, err)
		}

		return nil
	}

	// Open a restricted database to all users
	err = s.metabaseAPI.OpenAccessToDatabase(ctx, *meta.DatabaseID)
	if err != nil {
		return errs.E(op, err)
	}

	if meta.PermissionGroupID != nil {
		err = s.metabaseAPI.DeletePermissionGroup(ctx, *meta.PermissionGroupID)
		if err != nil {
			return errs.E(op, err)
		}
	}

	// When opening a previously restricted metabase database to all users, we need to
	// 1. grant access to the all-users service account for the datasource in BigQuery
	// 2. switch the database service account key in metabase to the all-users service account
	// 3. remove the old restricted service account
	if meta.SAEmail == s.ConstantServiceAccountEmailFromDatasetID(datasetID) {
		err = s.bigqueryAPI.Grant(ctx, datasource.ProjectID, datasource.Dataset, datasource.Table, "serviceAccount:"+s.serviceAccountEmail)
		if err != nil {
			return errs.E(op, err)
		}

		err = s.metabaseAPI.UpdateDatabase(ctx, *meta.DatabaseID, s.serviceAccount, s.serviceAccountEmail)
		if err != nil {
			return errs.E(op, err)
		}

		err := s.cleanupRestrictedDatabaseServiceAccount(ctx, datasetID, meta.SAEmail)
		if err != nil {
			return errs.E(op, err)
		}
	}

	meta, err = s.metabaseStorage.SetServiceAccountMetabaseMetadata(ctx, datasetID, s.serviceAccountEmail)
	if err != nil {
		return errs.E(op, err)
	}

	_, err = s.metabaseStorage.SetPermissionGroupMetabaseMetadata(ctx, meta.DatasetID, 0)
	if err != nil {
		return errs.E(op, err)
	}

	err = s.metabaseStorage.SetSyncCompletedMetabaseMetadata(ctx, datasetID)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *metabaseService) GetOpenMetabaseBigQueryDatabaseWorkflow(ctx context.Context, datasetID uuid.UUID) (*service.MetabaseBigQueryDatasetStatus, error) {
	const op errs.Op = "metabaseService.GetOpenMetabaseBigQueryDatabaseWorkflow"

	wf, err := s.metabaseQueue.GetOpenMetabaseBigQueryDatabaseWorkflow(ctx, datasetID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	meta, err := s.metabaseStorage.GetMetadata(ctx, datasetID, true)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &service.MetabaseBigQueryDatasetStatus{
		MetabaseMetadata: meta,
		IsRunning:        wf.IsRunning(),
		IsCompleted:      meta.SyncCompleted != nil,
		IsRestricted:     false,
		Jobs: []service.JobHeader{
			wf.PreflightCheckJob.JobHeader,
			wf.DatabaseJob.JobHeader,
			wf.VerifyJob.JobHeader,
			wf.FinalizeJob.JobHeader,
		},
	}, nil
}

func (s *metabaseService) PreflightCheckRestrictedBigqueryDatabase(ctx context.Context, datasetID uuid.UUID) error {
	const op errs.Op = "metabaseService.PreflightCheckRestrictedBigqueryDatabase"

	meta, err := s.metabaseStorage.GetMetadata(ctx, datasetID, true)
	if err != nil && !errs.KindIs(errs.NotExist, err) {
		return errs.E(op, err)
	}

	if meta != nil && meta.PermissionGroupID != nil && *meta.PermissionGroupID == 0 {
		return errs.E(errs.InvalidRequest, service.CodeOpeningClosedDatabase, op, fmt.Errorf("not allowed to expose a previously open database as a restricted"))
	}

	if meta == nil {
		err := s.metabaseStorage.CreateMetadata(ctx, datasetID)
		if err != nil {
			return errs.E(op, err)
		}
	}

	accesses, err := s.accessStorage.ListActiveAccessToDataset(ctx, datasetID)
	if err != nil {
		return errs.E(op, err)
	}

	if s.containsAllUsers(accesses) {
		return errs.E(errs.InvalidRequest, service.CodeOpeningClosedDatabase, op, fmt.Errorf("not allowed to expose a previously open database as a restricted"))
	}

	// FIXME: check if there is an OWNER on the bigquery dataset

	return nil
}

func (s *metabaseService) GetRestrictedMetabaseBigQueryDatabaseWorkflow(ctx context.Context, datasetID uuid.UUID) (*service.MetabaseBigQueryDatasetStatus, error) {
	const op errs.Op = "metabaseService.GetRestrictedMetabaseBigQueryDatabaseWorkflow"

	wf, err := s.metabaseQueue.GetRestrictedMetabaseBigqueryDatabaseWorkflow(ctx, datasetID)
	if err != nil {
		return nil, errs.E(op, err)
	}

	meta, err := s.metabaseStorage.GetMetadata(ctx, datasetID, true)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return &service.MetabaseBigQueryDatasetStatus{
		MetabaseMetadata: meta,
		IsRunning:        wf.IsRunning(),
		IsCompleted:      meta.SyncCompleted != nil,
		IsRestricted:     true,
		Jobs: []service.JobHeader{
			wf.PermissionGroupJob.JobHeader,
			wf.CollectionJob.JobHeader,
			wf.ServiceAccountJob.JobHeader,
			wf.ProjectIAMJob.JobHeader,
			wf.DatabaseJob.JobHeader,
			wf.VerifyJob.JobHeader,
			wf.FinalizeJob.JobHeader,
		},
	}, nil
}

func (s *metabaseService) EnsurePermissionGroup(ctx context.Context, datasetID uuid.UUID, name string) error {
	const op errs.Op = "metabaseService.EnsurePermissionGroup"

	meta, err := s.metabaseStorage.GetMetadata(ctx, datasetID, false)
	if err != nil {
		return errs.E(op, err)
	}

	if meta.PermissionGroupID == nil {
		groupID, err := s.metabaseAPI.GetOrCreatePermissionGroup(ctx, name)
		if err != nil {
			return errs.E(op, err)
		}

		_, err = s.metabaseStorage.SetPermissionGroupMetabaseMetadata(ctx, datasetID, groupID)
		if err != nil {
			return errs.E(op, err)
		}
	}

	return nil
}

func (s *metabaseService) CreateRestrictedCollection(ctx context.Context, datasetID uuid.UUID, name string) error {
	const op errs.Op = "metabaseService.CreateRestrictedCollection"

	meta, err := s.metabaseStorage.GetMetadata(ctx, datasetID, false)
	if err != nil {
		return errs.E(op, err)
	}

	if meta.CollectionID == nil {
		colID, err := s.metabaseAPI.CreateCollectionWithAccess(ctx, *meta.PermissionGroupID, name, true)
		if err != nil {
			return errs.E(op, err)
		}

		_, err = s.metabaseStorage.SetCollectionMetabaseMetadata(ctx, datasetID, colID)
		if err != nil {
			return errs.E(op, err)
		}
	}

	return nil
}

func (s *metabaseService) CreateMetabaseServiceAccount(ctx context.Context, datasetID uuid.UUID, request *service.ServiceAccountRequest) error {
	const op errs.Op = "metabaseService.CreateMetabaseServiceAccount"

	meta, err := s.metabaseStorage.GetMetadata(ctx, datasetID, false)
	if err != nil {
		return errs.E(op, err)
	}

	if meta.SAEmail == "" {
		sa, err := s.serviceAccountAPI.EnsureServiceAccount(ctx, request)
		if err != nil {
			return errs.E(op, err)
		}

		_, err = s.metabaseStorage.SetServiceAccountMetabaseMetadata(ctx, datasetID, sa.Email)
		if err != nil {
			return errs.E(op, err)
		}
	}

	return nil
}

func (s *metabaseService) CreateMetabaseProjectIAMPolicyBinding(ctx context.Context, projectID, role, member string) error {
	const op errs.Op = "metabaseService.CreateMetabaseProjectIAMPolicyBinding"

	err := s.cloudResourceManagerAPI.AddProjectIAMPolicyBinding(ctx, projectID, &service.Binding{
		Role: role,
		Members: []string{
			member,
		},
	})
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *metabaseService) containsAllUsers(accesses []*service.Access) bool {
	for _, a := range accesses {
		if a.Subject == s.groupAllUsers {
			return true
		}
	}

	return false
}

func (s *metabaseService) addRestrictedDatasetMapping(ctx context.Context, dsID uuid.UUID) error {
	const op errs.Op = "metabaseService.addRestrictedDatasetMapping"

	meta, err := s.metabaseStorage.GetMetadata(ctx, dsID, true)
	if err != nil {
		return errs.E(op, err)
	}

	// FIXME: here be dragons
	// If meta.DatabaseID != nil we know that we have created the collection, permission group, etc.
	// this is extremely fragile, so please be careful.
	if meta.DatabaseID == nil {
		ds, err := s.dataproductStorage.GetDataset(ctx, dsID)
		if err != nil {
			return errs.E(op, err)
		}

		if err := s.createRestricted(ctx, ds); err != nil {
			return errs.E(op, err)
		}

		if err := s.grantAccessesOnCreation(ctx, dsID); err != nil {
			return errs.E(op, err)
		}

		return nil
	}

	if meta.PermissionGroupID != nil && *meta.PermissionGroupID == 0 {
		return errs.E(errs.InvalidRequest, service.CodeOpeningClosedDatabase, op, fmt.Errorf("not allowed to expose a previously open database as a restricted"))
	}

	if err := s.grantAccessesOnCreation(ctx, dsID); err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *metabaseService) grantAccessesOnCreation(ctx context.Context, dsID uuid.UUID) error {
	const op errs.Op = "metabaseService.grantAccessesOnCreation"

	accesses, err := s.accessStorage.ListActiveAccessToDataset(ctx, dsID)
	if err != nil {
		return errs.E(op, err)
	}

	for _, a := range accesses {
		email, sType, err := parseSubject(a.Subject)
		if err != nil {
			return errs.E(op, err)
		}

		switch sType {
		case "user":
			err := s.addMetabaseGroupMember(ctx, dsID, email)
			if err != nil {
				return errs.E(op, err)
			}
		default:
			s.log.Info().Msgf("Unsupported subject type %v for metabase access grant", sType)
		}
	}

	return nil
}

func (s *metabaseService) addMetabaseGroupMember(ctx context.Context, dsID uuid.UUID, email string) error {
	const op errs.Op = "metabaseService.addMetabaseGroupMember"

	meta, err := s.metabaseStorage.GetMetadata(ctx, dsID, false)
	if err != nil {
		// If we don't have metadata for the dataset, it means that the dataset is not mapped to Metabase
		// so no need to add the user to the group
		if errs.KindIs(errs.NotExist, err) {
			return nil
		}

		return errs.E(op, err)
	}

	mbGroupMembers, err := s.metabaseAPI.GetPermissionGroup(ctx, *meta.PermissionGroupID)
	if err != nil {
		return errs.E(op, err)
	}

	exists, _ := memberExists(mbGroupMembers, email)
	if exists {
		return nil
	}

	user, err := s.metabaseAPI.FindUserByEmail(ctx, email)
	if err != nil && !errs.KindIs(errs.NotExist, err) {
		return errs.E(op, err)
	}

	if errs.KindIs(errs.NotExist, err) {
		user, err = s.metabaseAPI.CreateUser(ctx, email)
		if err != nil {
			return errs.E(op, err)
		}
	}

	if err := s.metabaseAPI.AddPermissionGroupMember(ctx, *meta.PermissionGroupID, user.ID); err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *metabaseService) restore(ctx context.Context, datasetID uuid.UUID, saEmail string) error {
	const op errs.Op = "metabaseService.restore"

	ds, err := s.bigqueryStorage.GetBigqueryDatasource(ctx, datasetID, false)
	if err != nil {
		return errs.E(op, err)
	}

	err = s.bigqueryAPI.Grant(ctx, ds.ProjectID, ds.Dataset, ds.Table, "serviceAccount:"+saEmail)
	if err != nil {
		return errs.E(op, err)
	}

	if err := s.metabaseStorage.RestoreMetadata(ctx, datasetID); err != nil {
		return errs.E(op, err)
	}

	return nil
}

func MarshalUUID(id uuid.UUID) string {
	return strings.ToLower(base58.Encode(id[:]))
}

func AccountIDFromDatasetID(id uuid.UUID) string {
	return fmt.Sprintf("nada-%s", MarshalUUID(id))
}

func (s *metabaseService) ConstantServiceAccountEmailFromDatasetID(id uuid.UUID) string {
	return fmt.Sprintf("%s@%s.iam.gserviceaccount.com", AccountIDFromDatasetID(id), s.gcpProject)
}

func (s *metabaseService) getOrcreateServiceAccountWithKeyAndPolicy(ctx context.Context, ds *service.Dataset) (*service.ServiceAccountWithPrivateKey, error) {
	const op errs.Op = "metabaseService.getOrcreateServiceAccountWithKeyAndPolicy"

	accountID := AccountIDFromDatasetID(ds.ID)

	sa, err := s.serviceAccountAPI.EnsureServiceAccountWithKey(ctx, &service.ServiceAccountRequest{
		ProjectID:   s.gcpProject,
		AccountID:   accountID,
		DisplayName: ds.Name,
		Description: fmt.Sprintf("Metabase service account for dataset %s", ds.ID.String()),
	})
	if err != nil {
		return nil, errs.E(op, err)
	}

	// FIXME: move this into another function, perhaps
	err = s.cloudResourceManagerAPI.AddProjectIAMPolicyBinding(ctx, s.gcpProject, &service.Binding{
		Role: service.NadaMetabaseRole(s.gcpProject),
		Members: []string{
			fmt.Sprintf("serviceAccount:%s", s.ConstantServiceAccountEmailFromDatasetID(ds.ID)),
		},
	})
	if err != nil {
		return nil, errs.E(op, err)
	}

	return sa, nil
}

// nolint: cyclop
func (s *metabaseService) createRestricted(ctx context.Context, ds *service.Dataset) error {
	const op errs.Op = "metabaseService.createRestricted"

	s.log.Info().Msgf("Checking workflow for dataset %v", ds.ID)

	wf, err := s.metabaseQueue.GetRestrictedMetabaseBigqueryDatabaseWorkflow(ctx, ds.ID)
	if err != nil && !errs.KindIs(errs.NotExist, err) {
		return errs.E(op, err)
	}

	if errs.KindIs(errs.NotExist, err) || (!wf.IsRunning() && wf.Error() != nil) {
		s.log.Info().Msgf("Creating restricted metabase database for dataset %v", ds.ID)

		permissionGroupName := slug.Make(fmt.Sprintf("%s-%s", ds.Name, MarshalUUID(ds.ID)))
		collectionName := fmt.Sprintf("%s %s", ds.Name, service.MetabaseRestrictedCollectionTag)

		wf, err = s.metabaseQueue.CreateRestrictedMetabaseBigqueryDatabaseWorkflow(ctx, &service.MetabaseRestrictedBigqueryDatabaseWorkflowOpts{
			DatasetID:           ds.ID,
			PermissionGroupName: permissionGroupName,
			CollectionName:      collectionName,
			AccountID:           AccountIDFromDatasetID(ds.ID),
			ProjectID:           ds.Datasource.ProjectID,
			DisplayName:         ds.Name,
			Description:         fmt.Sprintf("Metabase service account for dataset %s", ds.ID.String()),
			Role:                service.NadaMetabaseRole(s.gcpProject),
			Member:              fmt.Sprintf("serviceAccount:%s", s.ConstantServiceAccountEmailFromDatasetID(ds.ID)),
		})
		if err != nil {
			return errs.E(op, err)
		}
	}

	for wf.Error() == nil && wf.IsRunning() {
		time.Sleep(5 * time.Second)

		s.log.Info().Msgf("Waiting for workflow for %v to finish", ds.ID)

		wf, err = s.metabaseQueue.GetRestrictedMetabaseBigqueryDatabaseWorkflow(ctx, ds.ID)
		if err != nil {
			return errs.E(op, err)
		}
	}

	err = wf.Error()
	if err != nil {
		s.log.Info().Msgf("Workflow for %v failed: %v", ds.ID, err)

		return errs.E(op, err)
	}

	return nil
}

func ensureUserInGroup(user *service.User, group string) error {
	const op errs.Op = "ensureUserInGroup"

	if user == nil || !user.GoogleGroups.Contains(group) {
		return errs.E(errs.Unauthorized, service.CodeWrongOwner, op, errs.UserName(user.Email), fmt.Errorf("user not in group %v", group))
	}

	return nil
}

func (s *metabaseService) GrantMetabaseAccess(ctx context.Context, dsID uuid.UUID, subject, subjectType string) error {
	const op errs.Op = "metabaseService.GrantMetabaseAccess"

	meta, err := s.metabaseStorage.GetMetadata(ctx, dsID, true)
	if err != nil {
		if errs.KindIs(errs.NotExist, err) {
			return nil
		}

		return errs.E(op, err)
	}

	if meta.SyncCompleted == nil {
		return errs.E(errs.InvalidRequest, op, fmt.Errorf("dataset %v is not synced", dsID))
	}

	if subject == "all-users@nav.no" {
		// FIXME: Why on earth do we need to do this here?
		// err := s.addAllUsersDataset(ctx, dsID)
		// if err != nil {
		// 	return errs.E(op, err)
		// }
	}

	switch subjectType {
	case "user":
		err := s.addMetabaseGroupMember(ctx, dsID, subject)
		if err != nil {
			return errs.E(op, err)
		}
	default:
		log.Info().Msgf("Unsupported subject type %v for metabase access grant", subjectType)
	}

	return nil
}

func (s *metabaseService) CreateRestrictedMetabaseBigqueryDatabase(ctx context.Context, datasetID uuid.UUID) error {
	const op errs.Op = "metabaseService.CreateRestrictedMetabaseBigqueryDatabase"

	datasource, err := s.bigqueryStorage.GetBigqueryDatasource(ctx, datasetID, false)
	if err != nil {
		return errs.E(op, err)
	}

	meta, err := s.metabaseStorage.GetMetadata(ctx, datasetID, true)
	if err != nil {
		return errs.E(op, err)
	}

	err = s.bigqueryAPI.Grant(ctx, datasource.ProjectID, datasource.Dataset, datasource.Table, "serviceAccount:"+meta.SAEmail)
	if err != nil {
		return errs.E(op, err)
	}

	ds, err := s.dataproductStorage.GetDataset(ctx, datasetID)
	if err != nil {
		return errs.E(op, err)
	}

	dp, err := s.dataproductStorage.GetDataproduct(ctx, ds.DataproductID)
	if err != nil {
		return errs.E(op, err)
	}

	if meta.DatabaseID == nil {
		key, err := s.serviceAccountAPI.EnsureServiceAccountKey(ctx, service.ServiceAccountNameFromEmail(s.gcpProject, meta.SAEmail))
		if err != nil {
			return errs.E(op, err)
		}

		dbID, err := s.metabaseAPI.CreateDatabase(ctx, dp.Owner.Group, ds.Name, string(key.PrivateKeyData), meta.SAEmail, datasource)
		if err != nil {
			return errs.E(op, err)
		}

		meta, err = s.metabaseStorage.SetDatabaseMetabaseMetadata(ctx, datasetID, dbID)
		if err != nil {
			return errs.E(op, err)
		}
	}

	return nil
}

func (s *metabaseService) VerifyRestrictedMetabaseBigqueryDatabase(ctx context.Context, datasetID uuid.UUID) error {
	const op errs.Op = "metabaseService.VerifyRestrictedMetabaseBigqueryDatabase"

	meta, err := s.metabaseStorage.GetMetadata(ctx, datasetID, true)
	if err != nil {
		return errs.E(op, err)
	}

	if meta.DatabaseID == nil {
		return errs.E(errs.NotExist, service.CodeMetabase, op, fmt.Errorf("database not found for dataset: %v", datasetID))
	}

	datasource, err := s.bigqueryStorage.GetBigqueryDatasource(ctx, datasetID, false)
	if err != nil {
		return errs.E(op, err)
	}

	tables, err := s.metabaseAPI.Tables(ctx, *meta.DatabaseID, false)
	if err != nil || len(tables) == 0 {
		return errs.E(errs.Internal, service.CodeWaitingForDatabase, op, fmt.Errorf("database not synced: %v", datasource.Table))
	}

	for _, tab := range tables {
		if tab.Name == datasource.Table && len(tab.Fields) > 0 {
			return nil
		}
	}

	return nil
}

func (s *metabaseService) FinalizeRestrictedMetabaseBigqueryDatabase(ctx context.Context, datasetID uuid.UUID) error {
	const op errs.Op = "metabaseService.FinalizeRestrictedMetabaseBigqueryDatabase"

	meta, err := s.metabaseStorage.GetMetadata(ctx, datasetID, false)
	if err != nil {
		return errs.E(op, err)
	}

	datasource, err := s.bigqueryStorage.GetBigqueryDatasource(ctx, datasetID, false)
	if err != nil {
		return errs.E(op, err)
	}

	if err := s.SyncTableVisibility(ctx, meta, *datasource); err != nil {
		return errs.E(op, err)
	}

	if err := s.metabaseAPI.AutoMapSemanticTypes(ctx, *meta.DatabaseID); err != nil {
		return errs.E(op, err)
	}

	err = s.metabaseAPI.RestrictAccessToDatabase(ctx, *meta.PermissionGroupID, *meta.DatabaseID)
	if err != nil {
		return errs.E(op, err)
	}

	accesses, err := s.accessStorage.ListActiveAccessToDataset(ctx, datasetID)
	if err != nil {
		return errs.E(op, err)
	}

	for _, a := range accesses {
		email, sType, err := parseSubject(a.Subject)
		if err != nil {
			return errs.E(op, err)
		}

		switch sType {
		case "user":
			err := s.addMetabaseGroupMember(ctx, datasetID, email)
			if err != nil {
				return errs.E(op, err)
			}
		default:
			s.log.Info().Msgf("Unsupported subject type %v for metabase access grant", sType)
		}
	}

	err = s.metabaseStorage.SetSyncCompletedMetabaseMetadata(ctx, datasetID)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *metabaseService) DeleteRestrictedMetabaseBigqueryDatabase(ctx context.Context, datasetID uuid.UUID) error {
	const op errs.Op = "metabaseService.DestroyRestrictedMetabaseBigqueryDatabase"

	meta, err := s.metabaseStorage.GetMetadata(ctx, datasetID, true)
	if err != nil {
		return errs.E(op, err)
	}

	dataset, err := s.dataproductStorage.GetDataset(ctx, datasetID)
	if err != nil {
		return errs.E(op, err)
	}
	services := dataset.Mappings

	for idx, msvc := range services {
		if msvc == service.MappingServiceMetabase {
			services = append(services[:idx], services[idx+1:]...)
		}
	}

	if meta.DatabaseID != nil {
		err := s.metabaseAPI.DeleteDatabase(ctx, *meta.DatabaseID)
		if err != nil {
			return errs.E(op, err)
		}
	}

	if len(meta.SAEmail) > 0 {
		err := s.cloudResourceManagerAPI.RemoveProjectIAMPolicyBindingMemberForRole(
			ctx,
			s.gcpProject,
			service.NadaMetabaseRole(s.gcpProject),
			fmt.Sprintf("serviceAccount:%s", meta.SAEmail),
		)
		if err != nil {
			return errs.E(op, err)
		}

		if err := s.serviceAccountAPI.DeleteServiceAccount(ctx, s.gcpProject, meta.SAEmail); err != nil {
			return errs.E(op, err)
		}
	}

	if meta.CollectionID != nil && *meta.CollectionID != 0 {
		if err := s.metabaseAPI.ArchiveCollection(ctx, *meta.CollectionID); err != nil {
			return errs.E(op, err)
		}
	}

	if meta.PermissionGroupID != nil {
		if err := s.metabaseAPI.DeletePermissionGroup(ctx, *meta.PermissionGroupID); err != nil {
			return errs.E(op, err)
		}
	}

	if err := s.metabaseStorage.DeleteMetadata(ctx, datasetID); err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *metabaseService) DeleteDatabase(ctx context.Context, dsID uuid.UUID) error {
	const op errs.Op = "metabaseService.DeleteDatabase"

	meta, err := s.metabaseStorage.GetMetadata(ctx, dsID, true)
	if err != nil {
		if errs.KindIs(errs.NotExist, err) {
			return nil
		}

		return errs.E(op, err)
	}

	if isRestrictedDatabase(meta) {
		err = s.deleteRestrictedDatabase(ctx, dsID, meta)
		if err != nil {
			return errs.E(op, err)
		}

		return nil
	}

	err = s.deleteAllUsersDatabase(ctx, meta)
	if err != nil {
		return errs.E(op, err)
	}

	ds, err := s.bigqueryStorage.GetBigqueryDatasource(ctx, dsID, false)
	if err != nil {
		return errs.E(op, err)
	}

	err = s.bigqueryAPI.Revoke(ctx, ds.ProjectID, ds.Dataset, ds.Table, "serviceAccount:"+meta.SAEmail)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *metabaseService) deleteAllUsersDatabase(ctx context.Context, meta *service.MetabaseMetadata) error {
	const op errs.Op = "metabaseService.deleteAllUsersDatabase"

	ds, err := s.bigqueryStorage.GetBigqueryDatasource(ctx, meta.DatasetID, false)
	if err != nil {
		return errs.E(op, err)
	}

	err = s.bigqueryAPI.Revoke(ctx, ds.ProjectID, ds.Dataset, ds.Table, "serviceAccount:"+meta.SAEmail)
	if err != nil {
		return errs.E(op, err)
	}

	if meta.DatabaseID != nil {
		err := s.metabaseAPI.DeleteDatabase(ctx, *meta.DatabaseID)
		if err != nil {
			return errs.E(op, err)
		}
	}

	err = s.metabaseStorage.DeleteMetadata(ctx, meta.DatasetID)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

// nolint: cyclop
func (s *metabaseService) deleteRestrictedDatabase(ctx context.Context, datasetID uuid.UUID, meta *service.MetabaseMetadata) error {
	const op errs.Op = "metabaseService.deleteRestrictedDatabase"

	err := s.cleanupRestrictedDatabaseServiceAccount(ctx, datasetID, meta.SAEmail)
	if err != nil {
		return errs.E(op, err)
	}

	if meta.PermissionGroupID != nil {
		if err := s.metabaseAPI.DeletePermissionGroup(ctx, *meta.PermissionGroupID); err != nil {
			return errs.E(op, err)
		}
	}

	if meta.CollectionID != nil {
		if err := s.metabaseAPI.ArchiveCollection(ctx, *meta.CollectionID); err != nil {
			return errs.E(op, err)
		}
	}

	if meta.DatabaseID != nil {
		if err := s.metabaseAPI.DeleteDatabase(ctx, *meta.DatabaseID); err != nil {
			return errs.E(op, err)
		}
	}

	if err := s.metabaseStorage.DeleteRestrictedMetadata(ctx, datasetID); err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *metabaseService) RevokeMetabaseAccessFromAccessID(ctx context.Context, accessID uuid.UUID) error {
	const op errs.Op = "metabaseService.RevokeMetabaseAccessFromAccessID"

	access, err := s.accessStorage.GetAccessToDataset(ctx, accessID)
	if err != nil {
		return errs.E(op, err)
	}

	// FIXME: should we check if the user is the owner?

	err = s.RevokeMetabaseAccess(ctx, access.DatasetID, access.Subject)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *metabaseService) RevokeMetabaseAccess(ctx context.Context, dsID uuid.UUID, subject string) error {
	const op errs.Op = "metabaseService.RevokeMetabaseAccess"

	meta, err := s.metabaseStorage.GetMetadata(ctx, dsID, false)
	if err != nil {
		if errs.KindIs(errs.NotExist, err) {
			return nil
		}

		return errs.E(op, err)
	}

	if meta.SyncCompleted == nil {
		return errs.E(errs.InvalidRequest, service.CodeDatasetNotSynced, op, fmt.Errorf("dataset %v is not synced", dsID))
	}

	if subject == s.groupAllUsers {
		err := s.softDeleteDatabase(ctx, dsID)
		if err != nil {
			return errs.E(op, err)
		}
	}

	email, sType, err := parseSubject(subject)
	if err != nil {
		return errs.E(op, err)
	}

	// We only support subject type user for now
	if sType == "user" {
		err = s.removeMetabaseGroupMember(ctx, dsID, email)
		if err != nil {
			return errs.E(op, err)
		}
	}

	return nil
}

func (s *metabaseService) softDeleteDatabase(ctx context.Context, datasetID uuid.UUID) error {
	const op errs.Op = "metabaseService.softDeleteDatabase"

	mbMeta, err := s.metabaseStorage.GetMetadata(ctx, datasetID, false)
	if err != nil {
		return errs.E(op, err)
	}

	ds, err := s.bigqueryStorage.GetBigqueryDatasource(ctx, datasetID, false)
	if err != nil {
		return errs.E(op, err)
	}

	err = s.bigqueryAPI.Revoke(ctx, ds.ProjectID, ds.Dataset, ds.Table, "serviceAccount:"+mbMeta.SAEmail)
	if err != nil {
		return errs.E(op, err)
	}

	err = s.metabaseStorage.SoftDeleteMetadata(ctx, datasetID)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *metabaseService) removeMetabaseGroupMember(ctx context.Context, dsID uuid.UUID, email string) error {
	const op errs.Op = "metabaseService.removeMetabaseGroupMember"

	mbMetadata, err := s.metabaseStorage.GetMetadata(ctx, dsID, false)
	if err != nil {
		return errs.E(op, err)
	}

	mbGroupMembers, err := s.metabaseAPI.GetPermissionGroup(ctx, *mbMetadata.PermissionGroupID)
	if err != nil {
		return errs.E(op, err)
	}

	exists, memberID := memberExists(mbGroupMembers, email)
	if !exists {
		return nil
	}

	err = s.metabaseAPI.RemovePermissionGroupMember(ctx, memberID)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *metabaseService) cleanupRestrictedDatabaseServiceAccount(ctx context.Context, dsID uuid.UUID, saEmail string) error {
	const op errs.Op = "metabaseService.cleanupRestrictedDatabaseServiceAccount"

	ds, err := s.bigqueryStorage.GetBigqueryDatasource(ctx, dsID, false)
	if err != nil {
		return errs.E(op, err)
	}

	err = s.bigqueryAPI.Revoke(ctx, ds.ProjectID, ds.Dataset, ds.Table, "serviceAccount:"+saEmail)
	if err != nil {
		return errs.E(op, err)
	}

	err = s.cloudResourceManagerAPI.RemoveProjectIAMPolicyBindingMemberForRole(
		ctx,
		s.gcpProject,
		service.NadaMetabaseRole(s.gcpProject),
		fmt.Sprintf("serviceAccount:%s", saEmail),
	)
	if err != nil {
		return errs.E(op, err)
	}

	if err := s.serviceAccountAPI.DeleteServiceAccount(ctx, s.gcpProject, saEmail); err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *metabaseService) SyncAllTablesVisibility(ctx context.Context) error {
	const op errs.Op = "metabaseService.SyncAllTablesVisibility"

	metas, err := s.metabaseStorage.GetAllMetadata(ctx)
	if err != nil {
		return errs.E(op, err)
	}

	for _, db := range metas {
		if db.SyncCompleted == nil {
			continue
		}

		bq, err := s.bigqueryStorage.GetBigqueryDatasource(ctx, db.DatasetID, false)
		if err != nil {
			return errs.E(op, err)
		}

		if err := s.SyncTableVisibility(ctx, db, *bq); err != nil {
			return errs.E(op, fmt.Errorf("syncing table visibility for database %v: %w", db.DatasetID, err))
		}
	}

	return nil
}

func (s *metabaseService) SyncTableVisibility(ctx context.Context, meta *service.MetabaseMetadata, bq service.BigQuery) error {
	const op errs.Op = "metabaseService.SyncTableVisibility"

	if meta.DatabaseID == nil {
		return nil
	}

	tables, err := s.metabaseAPI.Tables(ctx, *meta.DatabaseID, true)
	if err != nil {
		if errs.KindIs(errs.NotExist, err) {
			return nil
		}

		return errs.E(op, err)
	}

	includedTables := []string{bq.Table}
	if !isRestrictedDatabase(meta) {
		includedTables, err = s.metabaseStorage.GetOpenTablesInSameBigQueryDataset(ctx, bq.ProjectID, bq.Dataset)
		if err != nil {
			return errs.E(op, err)
		}
	}

	var includedIDs, excludedIDs []int

	for _, t := range tables {
		if contains(includedTables, t.Name) {
			includedIDs = append(includedIDs, t.ID)
		} else {
			excludedIDs = append(excludedIDs, t.ID)
		}
	}

	if len(excludedIDs) > 0 {
		if err := s.metabaseAPI.HideTables(ctx, excludedIDs); err != nil {
			return errs.E(op, err)
		}
	}

	if len(includedIDs) > 0 {
		err = s.metabaseAPI.ShowTables(ctx, includedIDs)
		if err != nil {
			return errs.E(op, err)
		}
	}

	return nil
}

func isRestrictedDatabase(meta *service.MetabaseMetadata) bool {
	if meta.PermissionGroupID != nil && *meta.PermissionGroupID == 0 {
		return false
	}

	return true
}

func contains(elems []string, elem string) bool {
	for _, e := range elems {
		if e == elem {
			return true
		}
	}

	return false
}

func parseSubject(subject string) (string, string, error) {
	const op errs.Op = "parseSubject"

	s := strings.Split(subject, ":")
	if len(s) != 2 {
		return "", "", errs.E(errs.InvalidRequest, op, errs.Parameter("subject"), fmt.Errorf("invalid subject format, got: %s, should be type:email", subject))
	}

	return s[1], s[0], nil
}

func memberExists(groupMembers []service.MetabasePermissionGroupMember, subject string) (bool, int) {
	for _, m := range groupMembers {
		if m.Email == subject {
			return true, m.ID
		}
	}

	return false, -1
}

func NewMetabaseService(
	gcpProject string,
	serviceAccount string,
	serviceAccountEmail string,
	groupAllUsers string,
	mbqueue service.MetabaseQueue,
	mbapi service.MetabaseAPI,
	bqapi service.BigQueryAPI,
	saapi service.ServiceAccountAPI,
	crmapi service.CloudResourceManagerAPI,
	mbs service.MetabaseStorage,
	bqs service.BigQueryStorage,
	dps service.DataProductsStorage,
	as service.AccessStorage,
	log zerolog.Logger,
) *metabaseService {
	return &metabaseService{
		gcpProject:              gcpProject,
		serviceAccount:          serviceAccount,
		serviceAccountEmail:     serviceAccountEmail,
		groupAllUsers:           groupAllUsers,
		metabaseQueue:           mbqueue,
		metabaseAPI:             mbapi,
		bigqueryAPI:             bqapi,
		serviceAccountAPI:       saapi,
		cloudResourceManagerAPI: crmapi,
		metabaseStorage:         mbs,
		bigqueryStorage:         bqs,
		dataproductStorage:      dps,
		accessStorage:           as,
		log:                     log,
	}
}

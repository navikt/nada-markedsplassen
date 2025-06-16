package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/rs/zerolog"

	"github.com/btcsuite/btcutil/base58"

	"github.com/gosimple/slug"

	"github.com/google/uuid"
	"github.com/navikt/nada-backend/pkg/errs"
	"github.com/navikt/nada-backend/pkg/service"
)

var _ service.MetabaseService = &metabaseService{}

type metabaseService struct {
	gcpProject          string
	location            string
	keyring             string
	keyName             string
	serviceAccount      string
	serviceAccountEmail string
	groupAllUsers       string
	allUsersEmail       string

	metabaseQueue service.MetabaseQueue

	kmsAPI                  service.KMSAPI
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

func (s *metabaseService) ClearMetabaseBigqueryWorkflowJobs(ctx context.Context, user *service.User, datasetID uuid.UUID) error {
	const op errs.Op = "metabaseService.ClearMetabaseBigqueryWorkflowJobs"

	ds, err := s.dataproductStorage.GetDataset(ctx, datasetID)
	if err != nil {
		return errs.E(op, err)
	}

	dp, err := s.dataproductStorage.GetDataproduct(ctx, ds.DataproductID)
	if err != nil {
		return errs.E(op, err)
	}

	if err := ensureUserInGroup(user, dp.Owner.Group); err != nil {
		return errs.E(op, err)
	}

	err = s.metabaseQueue.DeleteMetabaseJobsForDataset(ctx, datasetID)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

func (s *metabaseService) CreateMetabaseServiceAccountKey(ctx context.Context, datasetID uuid.UUID) error {
	const op errs.Op = "metabaseService.CreateMetabaseServiceAccountKey"

	meta, err := s.metabaseStorage.GetMetadata(ctx, datasetID, true)
	if err != nil {
		return errs.E(op, err)
	}

	if len(meta.SAPrivateKey) == 0 {
		key, err := s.serviceAccountAPI.EnsureServiceAccountKey(ctx, service.ServiceAccountNameFromEmail(s.gcpProject, meta.SAEmail))
		if err != nil {
			return errs.E(op, err)
		}

		ciphertext, err := s.kmsAPI.Encrypt(ctx,
			&service.KeyIdentifier{
				Project:  s.gcpProject,
				Location: s.location,
				Keyring:  s.keyring,
				KeyName:  s.keyName,
			},
			key.PrivateKeyData,
		)
		if err != nil {
			return errs.E(op, err)
		}

		_, err = s.metabaseStorage.SetServiceAccountPrivateKeyMetabaseMetadata(ctx, datasetID, ciphertext)
		if err != nil {
			return errs.E(op, err)
		}
	}

	return nil
}

func (s *metabaseService) OpenPreviouslyRestrictedMetabaseBigqueryDatabase(ctx context.Context, datasetID uuid.UUID) error {
	const op errs.Op = "metabaseService.OpenPreviouslyRestrictedMetabaseBigqueryDatabase"

	meta, err := s.metabaseStorage.GetMetadata(ctx, datasetID, true)
	if err != nil {
		return errs.E(op, err)
	}

	if meta.SyncCompleted == nil {
		return errs.E(errs.InvalidRequest, service.CodeMetabase, op, fmt.Errorf("sync not completed for dataset: %v", datasetID))
	}

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

	datasource, err := s.bigqueryStorage.GetBigqueryDatasource(ctx, datasetID, false)
	if err != nil {
		return errs.E(op, err)
	}

	err = s.bigqueryAPI.Grant(ctx, datasource.ProjectID, datasource.Dataset, datasource.Table, "serviceAccount:"+s.serviceAccountEmail)
	if err != nil {
		return errs.E(op, err)
	}

	// When opening a previously restricted metabase database to all users, we need to
	// 1. grant access to the all-users service account for the datasource in BigQuery
	// 2. switch the database service account key in metabase to the all-users service account
	// 3. remove the old restricted service account
	if meta.SAEmail == s.ConstantServiceAccountEmailFromDatasetID(datasetID) {
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

	return nil
}

func (s *metabaseService) DeleteOpenMetabaseBigqueryDatabase(ctx context.Context, datasetID uuid.UUID) error {
	const op errs.Op = "metabaseService.DeleteOpenMetabaseBigqueryDatabase"

	meta, err := s.metabaseStorage.GetMetadata(ctx, datasetID, true)
	if err != nil {
		return errs.E(op, err)
	}

	if (meta.PermissionGroupID != nil && *meta.PermissionGroupID > 0) && (meta.SAEmail != s.serviceAccountEmail) {
		return fmt.Errorf("trying to delete a restricted database %v, when it is open", datasetID)
	}

	ds, err := s.bigqueryStorage.GetBigqueryDatasource(ctx, meta.DatasetID, false)
	if err != nil {
		return errs.E(op, err)
	}

	err = s.bigqueryAPI.Revoke(ctx, ds.ProjectID, ds.Dataset, ds.Table, "serviceAccount:"+meta.SAEmail)
	if err != nil && !errs.KindIs(errs.NotExist, err) {
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
		ProjectID:           s.gcpProject,
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

	_, err := s.metabaseStorage.GetMetadata(ctx, datasetID, true)
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

	_, err = s.metabaseStorage.SetServiceAccountMetabaseMetadata(ctx, datasetID, s.serviceAccountEmail)
	if err != nil {
		return errs.E(op, err)
	}

	meta, err := s.metabaseStorage.SetPermissionGroupMetabaseMetadata(ctx, datasetID, 0)
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

		_, err = s.metabaseStorage.SetDatabaseMetabaseMetadata(ctx, ds.ID, dbID)
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
		return errs.E(errs.Internal, service.CodeWaitingForDatabase, op, fmt.Errorf("database not synced: %v, for database: %d", datasource.Table, *meta.DatabaseID))
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

	err = s.metabaseAPI.OpenAccessToDatabase(ctx, *meta.DatabaseID)
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
	if err != nil && !errs.KindIs(errs.NotExist, err) {
		return nil, errs.E(op, err)
	}

	isCompleted := meta != nil && meta.SyncCompleted != nil

	return &service.MetabaseBigQueryDatasetStatus{
		MetabaseMetadata: meta,
		IsRunning:        wf.IsRunning(),
		IsCompleted:      isCompleted,
		IsRestricted:     false,
		HasFailed:        wf.HasFailed(),
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
	if err != nil && !errs.KindIs(errs.NotExist, err) {
		return nil, errs.E(op, err)
	}

	isCompleted := meta != nil && meta.SyncCompleted != nil

	return &service.MetabaseBigQueryDatasetStatus{
		MetabaseMetadata: meta,
		IsRunning:        wf.IsRunning(),
		IsCompleted:      isCompleted,
		IsRestricted:     true,
		HasFailed:        wf.HasFailed(),
		Jobs: []service.JobHeader{
			wf.PermissionGroupJob.JobHeader,
			wf.CollectionJob.JobHeader,
			wf.ServiceAccountJob.JobHeader,
			wf.ServiceAccountKeyJob.JobHeader,
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
		ds, err := s.dataproductStorage.GetDataset(ctx, datasetID)
		if err != nil {
			return errs.E(op, err)
		}

		expectedDescription := fmt.Sprintf(
			"Dette er en tilgangsstyrt samling for BigQuery tabellen: %s.%s.%s. I markedsplassen er dette datasettet %s, i dataprodutet %s",
			ds.Datasource.ProjectID,
			ds.Datasource.Dataset,
			ds.Datasource.Table,
			ds.ID,
			ds.DataproductID,
		)

		req := &service.CreateCollectionRequest{
			Name:        name,
			Description: expectedDescription,
		}

		colID, err := s.metabaseAPI.CreateCollectionWithAccess(ctx, *meta.PermissionGroupID, req, true)
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

func MarshalUUID(id uuid.UUID) string {
	return strings.ToLower(base58.Encode(id[:]))
}

func AccountIDFromDatasetID(id uuid.UUID) string {
	return fmt.Sprintf("nada-%s", MarshalUUID(id))
}

func (s *metabaseService) ConstantServiceAccountEmailFromDatasetID(id uuid.UUID) string {
	return fmt.Sprintf("%s@%s.iam.gserviceaccount.com", AccountIDFromDatasetID(id), s.gcpProject)
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
			s.log.Info().Msgf("dataset %v not found in metabase, skipping metabase access grant", dsID)

			return nil
		}

		return errs.E(op, err)
	}

	if meta.SyncCompleted == nil {
		return errs.E(errs.InvalidRequest, op, fmt.Errorf("dataset %v is not synced", dsID))
	}

	// We need to add the metabase service account to the bigquery dataset, if it is not already there

	if subject == s.allUsersEmail {
		s.log.Info().Msgf("Granting access to all users group %v for metabase database %v", subject, dsID)

		err := s.OpenPreviouslyRestrictedMetabaseBigqueryDatabase(ctx, dsID)
		if err != nil {
			return errs.E(op, err)
		}

		return nil
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
		plaintext, err := s.kmsAPI.Decrypt(ctx,
			&service.KeyIdentifier{
				Project:  s.gcpProject,
				Location: s.location,
				Keyring:  s.keyring,
				KeyName:  s.keyName,
			},
			meta.SAPrivateKey,
		)
		if err != nil {
			return errs.E(op, err)
		}

		dbID, err := s.metabaseAPI.CreateDatabase(ctx, dp.Owner.Group, ds.Name, string(plaintext), meta.SAEmail, datasource)
		if err != nil {
			return errs.E(op, err)
		}

		_, err = s.metabaseStorage.SetDatabaseMetabaseMetadata(ctx, datasetID, dbID)
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

	_, err = s.metabaseStorage.SetServiceAccountPrivateKeyMetabaseMetadata(ctx, datasetID, []byte{})
	if err != nil {
		return errs.E(op, err)
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

	if (meta.SAEmail == s.serviceAccountEmail) || (meta.PermissionGroupID != nil && *meta.PermissionGroupID == 0) {
		return fmt.Errorf("trying to delete an open database %v, when it is restricted", datasetID)
	}

	if meta.DatabaseID != nil {
		err := s.metabaseAPI.DeleteDatabase(ctx, *meta.DatabaseID)
		if err != nil {
			return errs.E(op, err)
		}
	}

	if len(meta.SAEmail) > 0 {
		ds, err := s.bigqueryStorage.GetBigqueryDatasource(ctx, datasetID, false)
		if err != nil {
			return errs.E(op, err)
		}

		err = s.bigqueryAPI.Revoke(ctx, ds.ProjectID, ds.Dataset, ds.Table, "serviceAccount:"+meta.SAEmail)
		if err != nil {
			return errs.E(op, err)
		}

		err = s.cloudResourceManagerAPI.RemoveProjectIAMPolicyBindingMemberForRole(
			ctx,
			s.gcpProject,
			service.NadaMetabaseRole(s.gcpProject),
			fmt.Sprintf("serviceAccount:%s", meta.SAEmail),
		)
		if err != nil && !errs.KindIs(errs.NotExist, err) {
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
		err = s.DeleteRestrictedMetabaseBigqueryDatabase(ctx, dsID)
		if err != nil {
			return errs.E(op, err)
		}

		return nil
	}

	err = s.DeleteOpenMetabaseBigqueryDatabase(ctx, dsID)
	if err != nil {
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
		ds, err := s.bigqueryStorage.GetBigqueryDatasource(ctx, dsID, false)
		if err != nil {
			return errs.E(op, err)
		}

		err = s.bigqueryAPI.Revoke(ctx, ds.ProjectID, ds.Dataset, ds.Table, "serviceAccount:"+s.serviceAccountEmail)
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
	location string,
	keyring string,
	keyName string,
	serviceAccount string,
	serviceAccountEmail string,
	groupAllUsers string,
	allUsersEmail string,
	mbqueue service.MetabaseQueue,
	kmsapi service.KMSAPI,
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
		location:                location,
		keyring:                 keyring,
		keyName:                 keyName,
		serviceAccount:          serviceAccount,
		serviceAccountEmail:     serviceAccountEmail,
		groupAllUsers:           groupAllUsers,
		allUsersEmail:           allUsersEmail,
		metabaseQueue:           mbqueue,
		kmsAPI:                  kmsapi,
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

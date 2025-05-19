package worker_args

import "riverqueue.com/riverpro"

const (
	MetabaseCreatePermissionGroupJobKind              = "metabase_create_permission_group_job"
	MetabaseCreateRestrictedCollectionJobKind         = "metabase_create_restricted_collection_job"
	MetabaseEnsureServiceAccountJobKind               = "metabase_ensure_service_account_job"
	MetabaseAddProjectIAMPolicyBindingJobKind         = "metabase_add_project_iam_policy_binding_job"
	MetabaseCreateRestrictedBigqueryDatabaseJobKind   = "metabase_create_bigquery_database_job"
	MetabaseVerifyRestrictedBigqueryDatabaseJobKind   = "metabase_verify_bigquery_database_job"
	MetabaseDeleteRestrictedBigqueryDatabaseJobKind   = "metabase_delete_bigquery_database_job"
	MetabaseFinalizeRestrictedBigqueryDatabaseJobKind = "metabase_finalize_bigquery_database_job"

	MetabaseQueue = "metabase"
)

type MetabaseCreatePermissionGroupJob struct {
	DatasetID string `json:"dataset_id" river:"sequence,unique"`

	PermissionGroupName string `json:"permission_group_name"`
}

func (MetabaseCreatePermissionGroupJob) Kind() string {
	return MetabaseCreatePermissionGroupJobKind
}

func (MetabaseCreatePermissionGroupJob) SequenceOpts() riverpro.SequenceOpts {
	return riverpro.SequenceOpts{
		ByArgs:      true,
		ExcludeKind: true,
	}
}

type MetabaseCreateRestrictedCollectionJob struct {
	DatasetID string `json:"dataset_id" river:"sequence,unique"`

	CollectionName string `json:"collection_name"`
}

func (MetabaseCreateRestrictedCollectionJob) Kind() string {
	return MetabaseCreateRestrictedCollectionJobKind
}

func (MetabaseCreateRestrictedCollectionJob) SequenceOpts() riverpro.SequenceOpts {
	return riverpro.SequenceOpts{
		ByArgs:      true,
		ExcludeKind: true,
	}
}

type MetabaseEnsureServiceAccountJob struct {
	DatasetID string `json:"dataset_id" river:"sequence,unique"`

	AccountID   string `json:"account_id"`
	ProjectID   string `json:"project_id"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

func (MetabaseEnsureServiceAccountJob) Kind() string {
	return MetabaseEnsureServiceAccountJobKind
}

func (MetabaseEnsureServiceAccountJob) SequenceOpts() riverpro.SequenceOpts {
	return riverpro.SequenceOpts{
		ByArgs:      true,
		ExcludeKind: true,
	}
}

type MetabaseAddProjectIAMPolicyBindingJob struct {
	DatasetID string `json:"dataset_id" river:"sequence,unique"`

	ProjectID string `json:"project_id"`
	Role      string `json:"role"`
	Member    string `json:"member"`
}

func (MetabaseAddProjectIAMPolicyBindingJob) Kind() string {
	return MetabaseAddProjectIAMPolicyBindingJobKind
}

func (MetabaseAddProjectIAMPolicyBindingJob) SequenceOpts() riverpro.SequenceOpts {
	return riverpro.SequenceOpts{
		ByArgs:      true,
		ExcludeKind: true,
	}
}

type MetabaseCreateRestrictedBigqueryDatabaseJob struct {
	DatasetID string `json:"dataset_id" river:"sequence,unique"`
}

func (MetabaseCreateRestrictedBigqueryDatabaseJob) Kind() string {
	return MetabaseCreateRestrictedBigqueryDatabaseJobKind
}

func (MetabaseCreateRestrictedBigqueryDatabaseJob) SequenceOpts() riverpro.SequenceOpts {
	return riverpro.SequenceOpts{
		ByArgs:      true,
		ExcludeKind: true,
	}
}

type MetabaseVerifyRestrictedBigqueryDatabaseJob struct {
	DatasetID string `json:"dataset_id" river:"sequence,unique"`
}

func (MetabaseVerifyRestrictedBigqueryDatabaseJob) Kind() string {
	return MetabaseVerifyRestrictedBigqueryDatabaseJobKind
}

func (MetabaseVerifyRestrictedBigqueryDatabaseJob) SequenceOpts() riverpro.SequenceOpts {
	return riverpro.SequenceOpts{
		ByArgs:      true,
		ExcludeKind: true,
	}
}

type MetabaseFinalizeRestrictedBigqueryDatabaseJob struct {
	DatasetID string `json:"dataset_id" river:"sequence,unique"`
}

func (MetabaseFinalizeRestrictedBigqueryDatabaseJob) Kind() string {
	return MetabaseFinalizeRestrictedBigqueryDatabaseJobKind
}

func (MetabaseFinalizeRestrictedBigqueryDatabaseJob) SequenceOpts() riverpro.SequenceOpts {
	return riverpro.SequenceOpts{
		ByArgs:      true,
		ExcludeKind: true,
	}
}

type MetabaseDeleteRestrictedBigqueryDatabaseJob struct {
	DatasetID string `json:"dataset_id" river:"unique"`
}

func (MetabaseDeleteRestrictedBigqueryDatabaseJob) Kind() string {
	return MetabaseDeleteRestrictedBigqueryDatabaseJobKind
}

package worker_args

import "riverqueue.com/riverpro"

const (
	MetabasePreflightCheckOpenBigqueryDatabaseJobKind = "metabase_preflight_check_open_bigquery_database_job"
	MetabaseCreateOpenBigqueryDatabaseJobKind         = "metabase_create_open_bigquery_database_job"
	MetabaseVerifyOpenBigqueryDatabaseJobKind         = "metabase_verify_open_bigquery_database_job"
	MetabaseFinalizeOpenBigqueryDatabaseJobKind       = "metabase_finalize_open_bigquery_database_job"

	MetabasePreflightCheckRestrictedBigqueryDatabaseJobKind = "metabase_preflight_check_bigquery_database_job"
	MetabaseCreatePermissionGroupJobKind                    = "metabase_create_permission_group_job"
	MetabaseCreateRestrictedCollectionJobKind               = "metabase_create_restricted_collection_job"
	MetabaseEnsureServiceAccountJobKind                     = "metabase_ensure_service_account_job"
	MetabaseCreateServiceAccountKeyJobKind                  = "metabase_create_service_account_key_job"
	MetabaseAddProjectIAMPolicyBindingJobKind               = "metabase_add_project_iam_policy_binding_job"
	MetabaseCreateRestrictedBigqueryDatabaseJobKind         = "metabase_create_bigquery_database_job"
	MetabaseVerifyRestrictedBigqueryDatabaseJobKind         = "metabase_verify_bigquery_database_job"
	MetabaseDeleteRestrictedBigqueryDatabaseJobKind         = "metabase_delete_bigquery_database_job"
	MetabaseDeleteOpenBigqueryDatabaseJobKind               = "metabase_delete_open_bigquery_database_job"
	MetabaseFinalizeRestrictedBigqueryDatabaseJobKind       = "metabase_finalize_bigquery_database_job"

	MetabaseQueue = "metabase"
)

type MetabasePreflightCheckRestrictedBigqueryDatabaseJob struct {
	DatasetID string `json:"dataset_id" river:"sequence,unique"`
}

func (MetabasePreflightCheckRestrictedBigqueryDatabaseJob) Kind() string {
	return MetabasePreflightCheckRestrictedBigqueryDatabaseJobKind
}

func (MetabasePreflightCheckRestrictedBigqueryDatabaseJob) SequenceOpts() riverpro.SequenceOpts {
	return riverpro.SequenceOpts{
		ByArgs:      true,
		ExcludeKind: true,
	}
}

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

type MetabaseCreateServiceAccountKeyJob struct {
	DatasetID string `json:"dataset_id" river:"sequence,unique"`
}

func (MetabaseCreateServiceAccountKeyJob) Kind() string {
	return MetabaseCreateServiceAccountKeyJobKind
}

func (MetabaseCreateServiceAccountKeyJob) SequenceOpts() riverpro.SequenceOpts {
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

type MetabaseDeleteOpenBigqueryDatabaseJob struct {
	DatasetID string `json:"dataset_id" river:"unique"`
}

func (MetabaseDeleteOpenBigqueryDatabaseJob) Kind() string {
	return MetabaseDeleteOpenBigqueryDatabaseJobKind
}

type MetabaseDeleteRestrictedBigqueryDatabaseJob struct {
	DatasetID string `json:"dataset_id" river:"unique"`
}

func (MetabaseDeleteRestrictedBigqueryDatabaseJob) Kind() string {
	return MetabaseDeleteRestrictedBigqueryDatabaseJobKind
}

type MetabasePreflightCheckOpenBigqueryDatabaseJob struct {
	DatasetID string `json:"dataset_id" river:"sequence,unique"`
}

func (MetabasePreflightCheckOpenBigqueryDatabaseJob) Kind() string {
	return MetabasePreflightCheckOpenBigqueryDatabaseJobKind
}

func (MetabasePreflightCheckOpenBigqueryDatabaseJob) SequenceOpts() riverpro.SequenceOpts {
	return riverpro.SequenceOpts{
		ByArgs:      true,
		ExcludeKind: true,
	}
}

type MetabaseCreateOpenBigqueryDatabaseJob struct {
	DatasetID string `json:"dataset_id" river:"sequence,unique"`
}

func (MetabaseCreateOpenBigqueryDatabaseJob) Kind() string {
	return MetabaseCreateOpenBigqueryDatabaseJobKind
}

func (MetabaseCreateOpenBigqueryDatabaseJob) SequenceOpts() riverpro.SequenceOpts {
	return riverpro.SequenceOpts{
		ByArgs:      true,
		ExcludeKind: true,
	}
}

type MetabaseVerifyOpenBigqueryDatabaseJob struct {
	DatasetID string `json:"dataset_id" river:"sequence,unique"`
}

func (MetabaseVerifyOpenBigqueryDatabaseJob) Kind() string {
	return MetabaseVerifyOpenBigqueryDatabaseJobKind
}

func (MetabaseVerifyOpenBigqueryDatabaseJob) SequenceOpts() riverpro.SequenceOpts {
	return riverpro.SequenceOpts{
		ByArgs:      true,
		ExcludeKind: true,
	}
}

type MetabaseFinalizeOpenBigqueryDatabaseJob struct {
	DatasetID string `json:"dataset_id" river:"sequence,unique"`
}

func (MetabaseFinalizeOpenBigqueryDatabaseJob) Kind() string {
	return MetabaseFinalizeOpenBigqueryDatabaseJobKind
}

func (MetabaseFinalizeOpenBigqueryDatabaseJob) SequenceOpts() riverpro.SequenceOpts {
	return riverpro.SequenceOpts{
		ByArgs:      true,
		ExcludeKind: true,
	}
}

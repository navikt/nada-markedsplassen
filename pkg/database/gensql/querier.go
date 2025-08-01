// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package gensql

import (
	"context"

	"github.com/google/uuid"
)

type Querier interface {
	AddTeamProject(ctx context.Context, arg AddTeamProjectParams) (TeamProject, error)
	ApproveAccessRequest(ctx context.Context, arg ApproveAccessRequestParams) error
	ClearTeamProjectsCache(ctx context.Context) error
	CreateAccessRequestForDataset(ctx context.Context, arg CreateAccessRequestForDatasetParams) (DatasetAccessRequest, error)
	CreateBigqueryDatasource(ctx context.Context, arg CreateBigqueryDatasourceParams) (DatasourceBigquery, error)
	CreateDataproduct(ctx context.Context, arg CreateDataproductParams) (Dataproduct, error)
	CreateDataset(ctx context.Context, arg CreateDatasetParams) (Dataset, error)
	CreateInsightProduct(ctx context.Context, arg CreateInsightProductParams) (InsightProduct, error)
	CreateJoinableViews(ctx context.Context, arg CreateJoinableViewsParams) (JoinableView, error)
	CreateJoinableViewsDatasource(ctx context.Context, arg CreateJoinableViewsDatasourceParams) (JoinableViewsDatasource, error)
	CreateMetabaseMetadata(ctx context.Context, datasetID uuid.UUID) error
	CreatePollyDocumentation(ctx context.Context, arg CreatePollyDocumentationParams) (PollyDocumentation, error)
	CreateSession(ctx context.Context, arg CreateSessionParams) error
	CreateStory(ctx context.Context, arg CreateStoryParams) (Story, error)
	CreateStoryWithID(ctx context.Context, arg CreateStoryWithIDParams) (Story, error)
	CreateTagIfNotExist(ctx context.Context, phrase string) error
	CreateWorkstationsActivityHistory(ctx context.Context, arg CreateWorkstationsActivityHistoryParams) error
	CreateWorkstationsConfigChange(ctx context.Context, arg CreateWorkstationsConfigChangeParams) error
	CreateWorkstationsOnpremAllowlistChange(ctx context.Context, arg CreateWorkstationsOnpremAllowlistChangeParams) error
	CreateWorkstationsURLListChange(ctx context.Context, arg CreateWorkstationsURLListChangeParams) error
	DataproductGroupStats(ctx context.Context, arg DataproductGroupStatsParams) ([]DataproductGroupStatsRow, error)
	DataproductKeywords(ctx context.Context, keyword string) ([]DataproductKeywordsRow, error)
	DatasetsByMetabase(ctx context.Context, arg DatasetsByMetabaseParams) ([]Dataset, error)
	DeleteAccessRequest(ctx context.Context, id uuid.UUID) error
	DeleteBigqueryDatasource(ctx context.Context, datasetID uuid.UUID) error
	DeleteDataproduct(ctx context.Context, id uuid.UUID) error
	DeleteDataset(ctx context.Context, id uuid.UUID) error
	DeleteInsightProduct(ctx context.Context, id uuid.UUID) error
	DeleteMetabaseMetadata(ctx context.Context, datasetID uuid.UUID) error
	DeleteNadaToken(ctx context.Context, team string) error
	DeleteSession(ctx context.Context, token string) error
	DeleteStory(ctx context.Context, id uuid.UUID) error
	DenyAccessRequest(ctx context.Context, arg DenyAccessRequestParams) error
	GetAccessRequest(ctx context.Context, id uuid.UUID) (DatasetAccessRequest, error)
	GetAccessToDataset(ctx context.Context, id uuid.UUID) (DatasetAccessView, error)
	GetAccessibleDatasets(ctx context.Context, arg GetAccessibleDatasetsParams) ([]GetAccessibleDatasetsRow, error)
	GetAccessibleDatasetsByOwnedServiceAccounts(ctx context.Context, arg GetAccessibleDatasetsByOwnedServiceAccountsParams) ([]GetAccessibleDatasetsByOwnedServiceAccountsRow, error)
	GetAccessiblePseudoDatasetsByUser(ctx context.Context, arg GetAccessiblePseudoDatasetsByUserParams) ([]GetAccessiblePseudoDatasetsByUserRow, error)
	GetActiveAccessToDatasetForSubject(ctx context.Context, arg GetActiveAccessToDatasetForSubjectParams) (DatasetAccessView, error)
	GetAllDatasetsMinimal(ctx context.Context) ([]GetAllDatasetsMinimalRow, error)
	GetAllMetabaseMetadata(ctx context.Context) ([]MetabaseMetadatum, error)
	GetAllTeams(ctx context.Context) ([]TkTeam, error)
	GetBigqueryDatasource(ctx context.Context, arg GetBigqueryDatasourceParams) (DatasourceBigquery, error)
	GetBigqueryDatasources(ctx context.Context) ([]DatasourceBigquery, error)
	GetDashboard(ctx context.Context, id uuid.UUID) (Dashboard, error)
	GetDataproduct(ctx context.Context, id uuid.UUID) (Dataproduct, error)
	GetDataproductKeywords(ctx context.Context, dpid uuid.UUID) ([]string, error)
	GetDataproductWithDatasetsBasic(ctx context.Context, id uuid.UUID) ([]GetDataproductWithDatasetsBasicRow, error)
	GetDataproducts(ctx context.Context, arg GetDataproductsParams) ([]Dataproduct, error)
	GetDataproductsByGroups(ctx context.Context, groups []string) ([]Dataproduct, error)
	GetDataproductsByIDs(ctx context.Context, ids []uuid.UUID) ([]Dataproduct, error)
	GetDataproductsByProductArea(ctx context.Context, teamID []uuid.UUID) ([]DataproductWithTeamkatalogenView, error)
	GetDataproductsByTeam(ctx context.Context, teamID uuid.NullUUID) ([]Dataproduct, error)
	GetDataproductsNumberByTeam(ctx context.Context, teamID uuid.NullUUID) (int64, error)
	GetDataproductsWithDatasets(ctx context.Context, arg GetDataproductsWithDatasetsParams) ([]GetDataproductsWithDatasetsRow, error)
	GetDataproductsWithDatasetsAndAccessRequests(ctx context.Context, arg GetDataproductsWithDatasetsAndAccessRequestsParams) ([]GetDataproductsWithDatasetsAndAccessRequestsRow, error)
	GetDataset(ctx context.Context, id uuid.UUID) (Dataset, error)
	GetDatasetComplete(ctx context.Context, id uuid.UUID) ([]DatasetView, error)
	GetDatasetCompleteWithAccess(ctx context.Context, arg GetDatasetCompleteWithAccessParams) ([]GetDatasetCompleteWithAccessRow, error)
	GetDatasets(ctx context.Context, arg GetDatasetsParams) ([]Dataset, error)
	GetDatasetsByGroups(ctx context.Context, groups []string) ([]Dataset, error)
	GetDatasetsByIDs(ctx context.Context, ids []uuid.UUID) ([]Dataset, error)
	GetDatasetsByUserAccess(ctx context.Context, id string) ([]Dataset, error)
	GetDatasetsForOwner(ctx context.Context, groups []string) ([]Dataset, error)
	GetDatasetsInDataproduct(ctx context.Context, dataproductID uuid.UUID) ([]Dataset, error)
	GetGroupEmailFromTeamSlug(ctx context.Context, team string) (string, error)
	GetInsightProduct(ctx context.Context, id uuid.UUID) (InsightProduct, error)
	GetInsightProductByGroups_(ctx context.Context, groups []string) ([]InsightProduct, error)
	GetInsightProductWithTeamkatalogen(ctx context.Context, id uuid.UUID) (InsightProductWithTeamkatalogenView, error)
	GetInsightProducts(ctx context.Context) ([]InsightProduct, error)
	GetInsightProductsByGroups(ctx context.Context, groups []string) ([]InsightProductWithTeamkatalogenView, error)
	GetInsightProductsByIDs(ctx context.Context, id []uuid.UUID) ([]InsightProduct, error)
	GetInsightProductsByProductArea(ctx context.Context, teamID []uuid.UUID) ([]InsightProductWithTeamkatalogenView, error)
	GetInsightProductsByTeam(ctx context.Context, teamID uuid.NullUUID) ([]InsightProduct, error)
	GetInsightProductsNumberByTeam(ctx context.Context, teamID uuid.NullUUID) (int64, error)
	GetJoinableViewWithDataset(ctx context.Context, id uuid.UUID) ([]GetJoinableViewWithDatasetRow, error)
	GetJoinableViewsForOwner(ctx context.Context, owner string) ([]GetJoinableViewsForOwnerRow, error)
	GetJoinableViewsForReferenceAndUser(ctx context.Context, arg GetJoinableViewsForReferenceAndUserParams) ([]GetJoinableViewsForReferenceAndUserRow, error)
	GetJoinableViewsToBeDeletedWithRefDatasource(ctx context.Context) ([]GetJoinableViewsToBeDeletedWithRefDatasourceRow, error)
	GetJoinableViewsWithReference(ctx context.Context) ([]GetJoinableViewsWithReferenceRow, error)
	GetKeywords(ctx context.Context) ([]GetKeywordsRow, error)
	GetLastWorkstationsOnpremAllowlistChange(ctx context.Context, navIdent string) (WorkstationsOnpremAllowlistHistory, error)
	GetLastWorkstationsURLListChange(ctx context.Context, navIdent string) (WorkstationsUrlListHistory, error)
	GetMetabaseMetadata(ctx context.Context, datasetID uuid.UUID) (MetabaseMetadatum, error)
	GetMetabaseMetadataWithDeleted(ctx context.Context, datasetID uuid.UUID) (MetabaseMetadatum, error)
	GetNadaTokenFromGroupEmail(ctx context.Context, groupEmail string) (uuid.UUID, error)
	GetNadaTokens(ctx context.Context) ([]NadaToken, error)
	GetNadaTokensForTeams(ctx context.Context, teams []string) ([]NadaToken, error)
	GetOpenMetabaseTablesInSameBigQueryDataset(ctx context.Context, arg GetOpenMetabaseTablesInSameBigQueryDatasetParams) ([]string, error)
	GetOwnerGroupOfDataset(ctx context.Context, datasetID uuid.UUID) (string, error)
	GetPollyDocumentation(ctx context.Context, id uuid.UUID) (PollyDocumentation, error)
	GetProductArea(ctx context.Context, id uuid.UUID) (TkProductArea, error)
	GetProductAreas(ctx context.Context) ([]TkProductArea, error)
	GetPseudoDatasourcesToDelete(ctx context.Context) ([]DatasourceBigquery, error)
	GetSession(ctx context.Context, token string) (Session, error)
	GetStories(ctx context.Context) ([]Story, error)
	GetStoriesByGroups(ctx context.Context, groups []string) ([]Story, error)
	GetStoriesByIDs(ctx context.Context, ids []uuid.UUID) ([]Story, error)
	GetStoriesByProductArea(ctx context.Context, teamID []uuid.UUID) ([]StoryWithTeamkatalogenView, error)
	GetStoriesByTeam(ctx context.Context, teamID uuid.NullUUID) ([]Story, error)
	GetStoriesNumberByTeam(ctx context.Context, teamID uuid.NullUUID) (int64, error)
	GetStoriesWithTeamkatalogenByGroups(ctx context.Context, groups []string) ([]StoryWithTeamkatalogenView, error)
	GetStoriesWithTeamkatalogenByIDs(ctx context.Context, ids []uuid.UUID) ([]StoryWithTeamkatalogenView, error)
	GetStory(ctx context.Context, id uuid.UUID) (Story, error)
	GetTag(ctx context.Context) (Tag, error)
	GetTagByPhrase(ctx context.Context) (Tag, error)
	GetTags(ctx context.Context) ([]Tag, error)
	GetTeamEmailFromNadaToken(ctx context.Context, token uuid.UUID) (string, error)
	GetTeamProjectFromGroupEmail(ctx context.Context, groupEmail string) (TeamProject, error)
	GetTeamProjects(ctx context.Context) ([]TeamProject, error)
	GetTeamsInProductArea(ctx context.Context, productAreaID uuid.NullUUID) ([]TkTeam, error)
	GrantAccessToDataset(ctx context.Context, arg GrantAccessToDatasetParams) error
	ListAccessRequestsForDataset(ctx context.Context, datasetID uuid.UUID) ([]DatasetAccessRequest, error)
	ListAccessRequestsForOwner(ctx context.Context, owner []string) ([]DatasetAccessRequest, error)
	ListAccessToDataset(ctx context.Context, datasetID uuid.UUID) ([]DatasetAccessView, error)
	ListActiveAccessToDataset(ctx context.Context, datasetID uuid.UUID) ([]DatasetAccessView, error)
	ListUnrevokedExpiredAccessEntries(ctx context.Context) ([]DatasetAccessView, error)
	RemoveKeywordInDatasets(ctx context.Context, keywordToRemove interface{}) error
	RemoveKeywordInStories(ctx context.Context, keywordToRemove interface{}) error
	ReplaceDatasetsTag(ctx context.Context, arg ReplaceDatasetsTagParams) error
	ReplaceKeywordInDatasets(ctx context.Context, arg ReplaceKeywordInDatasetsParams) error
	ReplaceKeywordInStories(ctx context.Context, arg ReplaceKeywordInStoriesParams) error
	ReplaceStoriesTag(ctx context.Context, arg ReplaceStoriesTagParams) error
	RestoreMetabaseMetadata(ctx context.Context, datasetID uuid.UUID) error
	RevokeAccessToDataset(ctx context.Context, id uuid.UUID) error
	RotateNadaToken(ctx context.Context, team string) error
	Search(ctx context.Context, arg SearchParams) ([]SearchRow, error)
	SearchDatasets(ctx context.Context, keyword string) ([]Dataset, error)
	SetCollectionMetabaseMetadata(ctx context.Context, arg SetCollectionMetabaseMetadataParams) (MetabaseMetadatum, error)
	SetDatabaseMetabaseMetadata(ctx context.Context, arg SetDatabaseMetabaseMetadataParams) (MetabaseMetadatum, error)
	SetDatasourceDeleted(ctx context.Context, id uuid.UUID) error
	SetJoinableViewDeleted(ctx context.Context, id uuid.UUID) error
	SetPermissionGroupMetabaseMetadata(ctx context.Context, arg SetPermissionGroupMetabaseMetadataParams) (MetabaseMetadatum, error)
	SetServiceAccountMetabaseMetadata(ctx context.Context, arg SetServiceAccountMetabaseMetadataParams) (MetabaseMetadatum, error)
	SetServiceAccountPrivateKeyMetabaseMetadata(ctx context.Context, arg SetServiceAccountPrivateKeyMetabaseMetadataParams) (MetabaseMetadatum, error)
	SetSyncCompletedMetabaseMetadata(ctx context.Context, datasetID uuid.UUID) error
	SoftDeleteMetabaseMetadata(ctx context.Context, datasetID uuid.UUID) error
	UpdateAccessRequest(ctx context.Context, arg UpdateAccessRequestParams) (DatasetAccessRequest, error)
	UpdateBigQueryDatasourceNotMissing(ctx context.Context, datasetID uuid.UUID) error
	UpdateBigqueryDatasource(ctx context.Context, arg UpdateBigqueryDatasourceParams) error
	UpdateBigqueryDatasourceMissing(ctx context.Context, datasetID uuid.UUID) error
	UpdateBigqueryDatasourceSchema(ctx context.Context, arg UpdateBigqueryDatasourceSchemaParams) error
	UpdateDataproduct(ctx context.Context, arg UpdateDataproductParams) (Dataproduct, error)
	UpdateDataset(ctx context.Context, arg UpdateDatasetParams) (Dataset, error)
	UpdateInsightProduct(ctx context.Context, arg UpdateInsightProductParams) (InsightProduct, error)
	UpdateStory(ctx context.Context, arg UpdateStoryParams) (Story, error)
	UpdateStoryLastModified(ctx context.Context, id uuid.UUID) error
	UpdateTag(ctx context.Context, arg UpdateTagParams) error
	UpsertProductArea(ctx context.Context, arg UpsertProductAreaParams) error
	UpsertTeam(ctx context.Context, arg UpsertTeamParams) error
}

var _ Querier = (*Queries)(nil)

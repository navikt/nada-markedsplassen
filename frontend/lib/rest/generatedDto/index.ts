// Code generated by tygo. DO NOT EDIT.

//////////
// source: access.go

export type AccessStorage = any;
export type AccessService = any;
export interface Access {
  id: string /* uuid */;
  subject: string;
  owner: string;
  granter: string;
  expires?: string /* RFC3339 */;
  created: string /* RFC3339 */;
  revoked?: string /* RFC3339 */;
  datasetID: string /* uuid */;
  accessRequestID?: string /* uuid */;
}
export interface NewAccessRequestDTO {
  datasetID: string /* uuid */;
  subject?: string;
  subjectType?: string;
  owner?: string;
  expires?: string /* RFC3339 */;
  polly?: PollyInput;
}
export interface UpdateAccessRequestDTO {
  id: string /* uuid */;
  owner: string;
  expires?: string /* RFC3339 */;
  polly?: PollyInput;
}
export interface AccessRequest {
  id: string /* uuid */;
  datasetID: string /* uuid */;
  subject: string;
  subjectType: string;
  created: string /* RFC3339 */;
  status: AccessRequestStatus;
  closed?: string /* RFC3339 */;
  expires?: string /* RFC3339 */;
  granter?: string;
  owner: string;
  polly?: Polly;
  reason?: string;
}
export interface AccessRequestForGranter {
  AccessRequest: AccessRequest;
  dataproductID: string /* uuid */;
  dataproductSlug: string;
  datasetName: string;
  dataproductName: string;
}
export interface AccessRequestsWrapper {
  accessRequests: AccessRequest[];
}
export const SubjectTypeUser: string = "user";
export const SubjectTypeGroup: string = "group";
export const SubjectTypeServiceAccount: string = "serviceAccount";
export interface GrantAccessData {
  datasetID: string /* uuid */;
  expires?: string /* RFC3339 */;
  subject?: string;
  owner?: string;
  subjectType?: string;
}
export type AccessRequestStatus = string;
export const AccessRequestStatusPending: AccessRequestStatus = "pending";
export const AccessRequestStatusApproved: AccessRequestStatus = "approved";
export const AccessRequestStatusDenied: AccessRequestStatus = "denied";

//////////
// source: bigquery.go

export type BigQueryStorage = any;
export type BigQueryAPI = any;
export type BigQueryService = any;
export type BigQueryTableType = string;
/**
 * RegularTable is a regular table.
 */
export const RegularTable: BigQueryTableType = "TABLE";
/**
 * ViewTable is a table type describing that the table is a logical view.
 * See more information at https://cloud.google.com//docs/views.
 */
export const ViewTable: BigQueryTableType = "VIEW";
/**
 * ExternalTable is a table type describing that the table is an external
 * table (also known as a federated data source). See more information at
 * https://cloud.google.com/bigquery/external-data-sources.
 */
export const ExternalTable: BigQueryTableType = "EXTERNAL";
/**
 * MaterializedView represents a managed storage table that's derived from
 * a base table.
 */
export const MaterializedView: BigQueryTableType = "MATERIALIZED_VIEW";
/**
 * Snapshot represents an immutable point in time snapshot of some other
 * table.
 */
export const Snapshot: BigQueryTableType = "SNAPSHOT";
export interface DatasourceForJoinableView {
  Project: string;
  Dataset: string;
  Table: string;
  PseudoColumns: string[];
}
export interface JoinableViewDatasource {
  RefDatasource?: DatasourceForJoinableView;
  PseudoDatasource?: DatasourceForJoinableView;
}
export interface GCPProject {
  id: string;
  name: string;
  group?: Group;
}
export interface BigQuery {
  ID: string /* uuid */;
  DatasetID: string /* uuid */;
  projectID: string;
  dataset: string;
  table: string;
  tableType: BigQueryTableType;
  lastModified: string /* RFC3339 */;
  created: string /* RFC3339 */;
  expired?: string /* RFC3339 */;
  description: string;
  piiTags?: string;
  missingSince?: string /* RFC3339 */;
  pseudoColumns: string[];
  schema: (BigqueryColumn | undefined)[];
}
export interface BQTables {
  bqTables: (BigQueryTable | undefined)[];
}
export interface BQDatasets {
  bqDatasets: string[];
}
export interface BQColumns {
  bqColumns: (BigqueryColumn | undefined)[];
}
export interface NewBigQuery {
  projectID: string;
  dataset: string;
  table: string;
  piiTags?: string;
}
export interface BigquerySchema {
  Columns: (BigqueryColumn | undefined)[];
}
export interface BigqueryMetadata {
  schema: BigquerySchema;
  tableType: BigQueryTableType;
  lastModified: string /* RFC3339 */;
  created: string /* RFC3339 */;
  expires: string /* RFC3339 */;
  description: string;
}
export interface BigQueryDataSourceUpdate {
  PiiTags?: string;
  PseudoColumns: string[];
  DatasetID: string /* uuid */;
}
export interface BigqueryColumn {
  name: string;
  type: string;
  mode: string;
  description: string;
}
export interface BigQueryTable {
  description: string;
  lastModified: string /* RFC3339 */;
  name: string;
  type: BigQueryTableType;
}

//////////
// source: dataproducts.go

export type DataProductsStorage = any;
export type DataProductsService = any;
export type PiiLevel = string;
export const PiiLevelSensitive: PiiLevel = "sensitive";
export const PiiLevelAnonymised: PiiLevel = "anonymised";
export const PiiLevelNone: PiiLevel = "none";
export type DatasourceType = string;
export interface Dataset {
  id: string /* uuid */;
  dataproductID: string /* uuid */;
  name: string;
  created: string /* RFC3339 */;
  lastModified: string /* RFC3339 */;
  description?: string;
  slug: string;
  repo?: string;
  pii: PiiLevel;
  keywords: string[];
  anonymisationDescription?: string;
  targetUser?: string;
  access: (Access | undefined)[];
  mappings: string[];
  datasource?: BigQuery;
  metabaseUrl?: string;
  metabaseDeletedAt?: string /* RFC3339 */;
}
export interface AccessibleDataset {
  Dataset: Dataset;
  dataproductName: string;
  slug: string;
  dpSlug: string;
  group: string;
  subject?: string;
}
export interface AccessibleDatasets {
  /**
   * owned
   */
  owned: (AccessibleDataset | undefined)[];
  /**
   * granted
   */
  granted: (AccessibleDataset | undefined)[];
  /**
   * service account granted
   */
  serviceAccountGranted: (AccessibleDataset | undefined)[];
}
export interface DatasetMinimal {
  id: string /* uuid */;
  name: string;
  created: string /* RFC3339 */;
  project: string;
  dataset: string;
  table: string;
}
export interface DatasetInDataproduct {
  id: string /* uuid */;
  name: string;
  created: string /* RFC3339 */;
  lastModified: string /* RFC3339 */;
  description?: string;
  slug: string;
  keywords: string[];
  dataSourceLastModified: string /* RFC3339 */;
}
export interface NewDataset {
  dataproductID: string /* uuid */;
  name: string;
  description?: string;
  slug?: string;
  repo?: string;
  pii: PiiLevel;
  keywords: string[];
  bigquery: NewBigQuery;
  anonymisationDescription?: string;
  grantAllUsers?: boolean;
  targetUser?: string;
  Metadata: BigqueryMetadata;
  pseudoColumns: string[];
}
export interface UpdateDatasetDto {
  name: string;
  description?: string;
  slug?: string;
  repo?: string;
  pii: PiiLevel;
  keywords: string[];
  dataproductID?: string /* uuid */;
  anonymisationDescription?: string;
  piiTags?: string;
  targetUser?: string;
  pseudoColumns: string[];
}
export interface DataproductOwner {
  group: string;
  teamkatalogenURL?: string;
  teamContact?: string;
  teamID?: string /* uuid */;
  productAreaID?: string /* uuid */;
}
export interface Dataproduct {
  id: string /* uuid */;
  name: string;
  created: string /* RFC3339 */;
  lastModified: string /* RFC3339 */;
  description?: string;
  slug: string;
  owner?: DataproductOwner;
  keywords: string[];
  teamName?: string;
  productAreaName: string;
}
export interface DataproductMinimal {
  id: string /* uuid */;
  name: string;
  created: string /* RFC3339 */;
  lastModified: string /* RFC3339 */;
  description?: string;
  slug: string;
  owner?: DataproductOwner;
}
export interface DataproductWithDataset extends Dataproduct {
  datasets: (DatasetInDataproduct | undefined)[];
}
export interface DatasetMap {
  services: string[];
}
/**
 * PseudoDataset contains information about a pseudo dataset
 */
export interface PseudoDataset {
  /**
   * name is the name of the dataset
   */
  name: string;
  /**
   * datasetID is the id of the dataset
   */
  datasetID: string /* uuid */;
  /**
   * datasourceID is the id of the bigquery datasource
   */
  datasourceID: string /* uuid */;
}
/**
 * NewDataproduct contains metadata for creating a new dataproduct
 */
export interface NewDataproduct {
  /**
   * name of dataproduct
   */
  name: string;
  /**
   * description of the dataproduct
   */
  description?: string;
  /**
   * owner group email for the dataproduct.
   */
  group: string;
  /**
   * owner Teamkatalogen URL for the dataproduct.
   */
  teamkatalogenURL?: string;
  /**
   * The contact information of the team who owns the dataproduct, which can be slack channel, slack account, email, and so on.
   */
  teamContact?: string;
  /**
   * Id of the team's product area.
   */
  productAreaID?: string /* uuid */;
  /**
   * Id of the team.
   */
  teamID?: string /* uuid */;
  Slug?: string;
}
export interface UpdateDataproductDto {
  name: string;
  description?: string;
  slug?: string;
  pii: PiiLevel;
  teamkatalogenURL?: string;
  teamContact?: string;
  productAreaID?: string /* uuid */;
  teamID?: string /* uuid */;
}
export const MappingServiceMetabase: string = "metabase";

//////////
// source: insight_products.go

export type InsightProductStorage = any;
export type InsightProductService = any;
/**
 * InsightProduct contains the metadata of insight product.
 */
export interface InsightProduct {
  /**
   * id of the insight product.
   */
  id: string /* uuid */;
  /**
   * name of the insight product.
   */
  name: string;
  /**
   * creator of the insight product.
   */
  creator: string;
  /**
   * description of the insight product.
   */
  description: string;
  /**
   * type of the insight product.
   */
  type: string;
  /**
   * link to the insight product.
   */
  link: string;
  /**
   * keywords for the insight product used as tags.
   */
  keywords: string[];
  /**
   * group is the owner group of the insight product
   */
  group: string;
  /**
   * teamkatalogenURL of the creator
   */
  teamkatalogenURL?: string;
  /**
   * Id of the creator's team.
   */
  teamID?: string /* uuid */;
  /**
   * created is the timestamp for when the insight product was created
   */
  created: string /* RFC3339 */;
  /**
   * lastModified is the timestamp for when the insight product was last modified
   */
  lastModified?: string /* RFC3339 */;
  teamName?: string;
  productAreaName: string;
}
export interface UpdateInsightProductDto {
  name: string;
  description: string;
  type: string;
  link: string;
  keywords: string[];
  teamkatalogenURL?: string;
  productAreaID?: string /* uuid */;
  teamID?: string /* uuid */;
  group: string;
}
/**
 * NewInsightProduct contains the metadata and content of insight products.
 */
export interface NewInsightProduct {
  name: string;
  description?: string;
  type: string;
  link: string;
  keywords: string[];
  /**
   * Group is the owner group of the insight product
   */
  group: string;
  /**
   * TeamkatalogenURL of the creator
   */
  teamkatalogenURL?: string;
  /**
   * Id of the creator's product area.
   */
  productAreaID?: string /* uuid */;
  /**
   * Id of the creator's team.
   */
  teamID?: string /* uuid */;
}

//////////
// source: joinable_views.go

export type JoinableViewsStorage = any;
export type JoinableViewsService = any;
export interface JoinableViewToBeDeletedWithRefDatasource {
  JoinableViewID: string /* uuid */;
  JoinableViewName: string;
  BqProjectID: string;
  BqDatasetID: string;
  BqTableID: string;
}
export interface JoinableViewWithReference {
  Owner: string;
  JoinableViewID: string /* uuid */;
  JoinableViewDataset: string;
  PseudoViewID: string /* uuid */;
  PseudoProjectID: string;
  PseudoDataset: string;
  PseudoTable: string;
  Expires: any /* sql.NullTime */;
}
export interface JoinableViewWithDataset {
  BqProject: string;
  BqDataset: string;
  BqTable: string;
  Deleted?: string /* RFC3339 */;
  DatasetID: null | string /* uuid */;
  JoinableViewID: string /* uuid */;
  Group: string;
  JoinableViewName: string;
  JoinableViewCreated: string /* RFC3339 */;
  JoinableViewExpires?: string /* RFC3339 */;
}
export interface JoinableViewForReferenceAndUser {
  ID: string /* uuid */;
  Dataset: string;
}
export interface JoinableViewForOwner {
  ID: string /* uuid */;
  Name: string;
  Owner: string;
  Created: string /* RFC3339 */;
  Expires?: string /* RFC3339 */;
  ProjectID: string;
  DatasetID: string;
  TableID: string;
}
/**
 * NewJoinableViews contains metadata for creating joinable views
 */
export interface NewJoinableViews {
  /**
   * Name is the name of the joinable views which will be used as the name of the dataset in bigquery, which contains all the joinable views
   */
  name: string;
  expires?: string /* RFC3339 */;
  /**
   * DatasetIDs is the IDs of the datasets which are made joinable.
   */
  datasetIDs: string /* uuid */[];
}
export interface JoinableView {
  /**
   * id is the id of the joinable view set
   */
  id: string /* uuid */;
  name: string;
  created: string /* RFC3339 */;
  expires?: string /* RFC3339 */;
}
export interface PseudoDatasource {
  bigqueryUrl: string;
  accessible: boolean;
  deleted: boolean;
}
export interface JoinableViewWithDatasource {
  JoinableView: JoinableView;
  pseudoDatasources: PseudoDatasource[];
}

//////////
// source: keywords.go

export type KeywordsStorage = any;
export type KeywordsService = any;
export interface KeywordsList {
  keywordItems: KeywordItem[];
}
export interface KeywordItem {
  keyword: string;
  count: number /* int */;
}
export interface UpdateKeywordsDto {
  obsoleteKeywords: string[];
  replacedKeywords: string[];
  newText: string[];
}

//////////
// source: metabase.go

export const MetabaseRestrictedCollectionTag = "🔐";
export type MetabaseStorage = any;
export type MetabaseAPI = any;
export type MetabaseService = any;
export interface MetabaseField {
}
export interface MetabaseTable {
  name: string;
  id: number /* int */;
  fields: {
    database_type: string;
    id: number /* int */;
    semantic_type: string;
  }[];
}
export interface MetabasePermissionGroup {
  id: number /* int */;
  name: string;
  members: MetabasePermissionGroupMember[];
}
export interface MetabasePermissionGroupMember {
  membership_id: number /* int */;
  email: string;
}
export interface MetabaseUser {
  email: string;
  id: number /* int */;
}
export interface MetabaseDatabase {
  ID: number /* int */;
  Name: string;
  DatasetID: string;
  ProjectID: string;
  NadaID: string;
  SAEmail: string;
}
export interface MetabaseMetadata {
  DatasetID: string /* uuid */;
  DatabaseID?: number /* int */;
  PermissionGroupID?: number /* int */;
  CollectionID?: number /* int */;
  SAEmail: string;
  DeletedAt?: string /* RFC3339 */;
  SyncCompleted?: string /* RFC3339 */;
}
/**
 * MetabaseCollection represents a subset of the metadata returned
 * for a Metabase collection
 */
export interface MetabaseCollection {
  ID: number /* int */;
  Name: string;
  Description: string;
}

//////////
// source: nais_console.go

export type NaisConsoleStorage = any;
export type NaisConsoleAPI = any;
export type NaisConsoleService = any;

//////////
// source: polly.go

export type PollyStorage = any;
export type PollyAPI = any;
export type PollyService = any;
export interface Polly {
  id: string /* uuid */;
  QueryPolly: QueryPolly;
}
export interface PollyInput {
  id?: string /* uuid */;
  QueryPolly: QueryPolly;
}
export interface QueryPolly {
  externalID: string;
  name: string;
  url: string;
}

//////////
// source: productareas.go

export type ProductAreaStorage = any;
export type ProductAreaService = any;
export interface UpsertProductAreaRequest {
  ID: string /* uuid */;
  Name: string;
}
/**
 * FIXCME: does this belong here?
 */
export interface UpsertTeamRequest {
  ID: string /* uuid */;
  ProductAreaID: string /* uuid */;
  Name: string;
}
export interface Team {
  TeamkatalogenTeam?: TeamkatalogenTeam;
  dataproductsNumber: number /* int */;
  storiesNumber: number /* int */;
  insightProductsNumber: number /* int */;
}
/**
 * FIXME: we need to simplify these structs, there is too much duplication
 */
export interface ProductAreasDto {
  productAreas: (ProductArea | undefined)[];
}
export interface ProductArea {
  TeamkatalogenProductArea?: TeamkatalogenProductArea;
  teams: (Team | undefined)[];
  dashboardURL: string;
}
export interface TeamWithAssets {
  TeamkatalogenTeam?: TeamkatalogenTeam;
  dataproducts: (Dataproduct | undefined)[];
  stories: (Story | undefined)[];
  insightProducts: (InsightProduct | undefined)[];
  dashboardURL: string;
}
export interface Dashboard {
  ID: string /* uuid */;
  Url: string;
}
export interface ProductAreaWithAssets {
  ProductArea?: ProductArea;
  teams: (TeamWithAssets | undefined)[];
}

//////////
// source: prometheus.go

/**
 * Middleware is a handler that exposes prometheus metrics for the number of requests,
 * the latency and the response size, partitioned by status code, method and HTTP path.
 */
export interface Middleware {
}

//////////
// source: search.go

export type SearchStorage = any;
export type SearchService = any;
export type ResultItem = any;
export interface SearchResult {
  results: (SearchResultRow | undefined)[];
}
export interface SearchResultRaw {
  ElementID: string /* uuid */;
  ElementType: string;
  Rank: number /* float32 */;
  Excerpt: string;
}
export interface SearchResultRow {
  excerpt: string;
  result: ResultItem;
  rank: number /* float64 */;
}
export interface SearchOptions {
  /**
   * Freetext search
   */
  text: string;
  /**
   * Filter on keyword
   */
  keywords: string[];
  /**
   * Filter on group
   */
  groups: string[];
  /**
   * Filter on team_id
   */
  teamIDs: string /* uuid */[];
  /**
   * Filter on enabled services
   */
  services: string[];
  /**
   * Filter on types
   */
  types: string[];
  limit?: number /* int */;
  offset?: number /* int */;
}

//////////
// source: serviceaccount.go

export type ServiceAccountAPI = any;
export interface Binding {
  Role: string;
  Members: string[];
}
export interface ServiceAccountRequest {
  ProjectID: string;
  AccountID: string;
  DisplayName: string;
  Description: string;
  Binding?: Binding;
}
export interface ServiceAccountMeta {
  Description: string;
  DisplayName: string;
  Email: string;
  Name: string;
  ProjectId: string;
  UniqueId: string;
}
export interface ServiceAccount {
  ServiceAccountMeta?: ServiceAccountMeta;
  Keys: (ServiceAccountKey | undefined)[];
  Bindings: (Binding | undefined)[];
}
export interface ServiceAccountWithPrivateKey {
  ServiceAccountMeta?: ServiceAccountMeta;
  Key?: ServiceAccountKeyWithPrivateKeyData;
}
export interface ServiceAccountKey {
  Name: string;
  KeyAlgorithm: string;
  KeyOrigin: string;
  KeyType: string;
}
export interface ServiceAccountKeyWithPrivateKeyData {
  ServiceAccountKey?: ServiceAccountKey;
  PrivateKeyData: string;
}

//////////
// source: slack.go

export type SlackAPI = any;
export type SlackService = any;

//////////
// source: stories.go

export type StoryStorage = any;
export type StoryAPI = any;
export type StoryService = any;
export interface UploadFile {
  /**
   * path of the file uploaded
   */
  path: string;
  /**
   * file data
   */
  ReadCloser: any /* io.ReadCloser */;
}
/**
 * Story contains the metadata and content of data stories.
 */
export interface Story {
  /**
   * id of the data story.
   */
  id: string /* uuid */;
  /**
   * name of the data story.
   */
  name: string;
  /**
   * creator of the data story.
   */
  creator: string;
  /**
   * description of the data story.
   */
  description: string;
  /**
   * keywords for the story used as tags.
   */
  keywords: string[];
  /**
   * teamkatalogenURL of the creator.
   */
  teamkatalogenURL?: string;
  /**
   * Id of the creator's team.
   */
  teamID?: string /* uuid */;
  /**
   * created is the timestamp for when the data story was created.
   */
  created: string /* RFC3339 */;
  /**
   * lastModified is the timestamp for when the dataproduct was last modified.
   */
  lastModified?: string /* RFC3339 */;
  /**
   * group is the owner group of the data story.
   */
  group: string;
  teamName?: string;
  productAreaName: string;
}
/**
 * NewStory contains the metadata and content of data stories.
 */
export interface NewStory {
  /**
   * id of data story.
   */
  id?: string /* uuid */;
  /**
   * name of the data story.
   */
  name: string;
  /**
   * description of the data story.
   */
  description?: string;
  /**
   * keywords for the story used as tags.
   */
  keywords: string[];
  /**
   * teamkatalogenURL of the creator.
   */
  teamkatalogenURL?: string;
  /**
   * Id of the creator's product area.
   */
  productAreaID?: string /* uuid */;
  /**
   * Id of the creator's team.
   */
  teamID?: string /* uuid */;
  /**
   * group is the owner group of the data story.
   */
  group: string;
}
export interface UpdateStoryDto {
  name: string;
  description: string;
  keywords: string[];
  teamkatalogenURL?: string;
  productAreaID?: string /* uuid */;
  teamID?: string /* uuid */;
  group: string;
}
export interface Object {
  Name: string;
  Bucket: string;
  Attrs: Attributes;
}
export interface Attributes {
  ContentType: string;
  ContentEncoding: string;
  Size: number /* int64 */;
  SizeStr: string;
}
export interface ObjectWithData {
  Object?: Object;
  Data: string;
}

//////////
// source: teamkatalogen.go

export type TeamKatalogenAPI = any;
export type TeamKatalogenService = any;
export interface TeamkatalogenResult {
  teamID: string;
  url: string;
  name: string;
  description: string;
  productAreaID: string;
}
export interface TeamkatalogenProductArea {
  /**
   * id is the id of the product area.
   */
  id: string /* uuid */;
  /**
   * name is the name of the product area.
   */
  name: string;
  /**
   * areaType is the type of the product area.
   */
  areaType: string;
  /**
   * FIXME: Can probably get rid of this
   */
  teams: Team[];
}
export interface TeamkatalogenTeam {
  /**
   * id is the team external id in teamkatalogen.
   */
  id: string /* uuid */;
  /**
   * name is the name of the team.
   */
  name: string;
  /**
   * productAreaID is the id of the product area.
   */
  productAreaID: string /* uuid */;
}

//////////
// source: third_party_mappings.go

export type ThirdPartyMappingStorage = any;

//////////
// source: tokens.go

export type TokenStorage = any;
export type TokenService = any;
export interface NadaToken {
  team: string;
  token: string /* uuid */;
}

//////////
// source: user.go

export type UserService = any;
export interface User {
  name: string;
  email: string;
  AzureGroups: Groups;
  GoogleGroups: Groups;
  AllGoogleGroups: Groups;
  expiry: string /* RFC3339 */;
}
export interface Group {
  name: string;
  email: string;
}
export type Groups = Group[];
export interface UserInfo {
  /**
   * name of user
   */
  name: string;
  /**
   * email of user.
   */
  email: string;
  /**
   * googleGroups is the google groups the user is member of.
   */
  googleGroups: Groups;
  /**
   * allGoogleGroups is the all the known google groups of the user domains.
   */
  allGoogleGroups: Groups;
  /**
   * azureGroups is the azure groups the user is member of.
   */
  azureGroups: Groups;
  /**
   * gcpProjects is GCP projects the user is a member of.
   */
  gcpProjects: GCPProject[];
  /**
   * nadaTokens is a list of the nada tokens for each team the logged in user is a part of.
   */
  nadaTokens: NadaToken[];
  /**
   * loginExpiration is when the token expires.
   */
  loginExpiration: string /* RFC3339 */;
  /**
   * dataproducts is a list of dataproducts with one of the users groups as owner.
   */
  dataproducts: Dataproduct[];
  /**
   * accessable is a list of datasets which the user has either owns or has explicit access to.
   */
  accessable: AccessibleDatasets;
  /**
   * stories is the stories owned by the user's group
   */
  stories: (Story | undefined)[];
  /**
   * insight products is the insight products owned by the user's group
   */
  insightProducts: InsightProduct[];
  /**
   * accessRequests is a list of access requests where either the user or one of the users groups is owner.
   */
  accessRequests: AccessRequest[];
  /**
   * accessRequestsAsGranter is a list of access requests where one of the users groups is obliged to handle.
   */
  accessRequestsAsGranter: AccessRequestForGranter[];
}
package config

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-ozzo/ozzo-validation/v4/is"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/mitchellh/mapstructure"

	"github.com/spf13/viper"
)

const (
	defaultExtension = "yaml"
	defaultTagName   = "yaml"
)

type Binder interface {
	Bind(v *viper.Viper) error
}

type Loader interface {
	Load(name, path, envPrefix string, binder Binder) (Config, error)
}

type Config struct {
	Oauth                     Oauth                     `yaml:"oauth"`
	OauthGoogle               Oauth                     `yaml:"google"`
	Metabase                  Metabase                  `yaml:"metabase"`
	CrossTeamPseudonymization CrossTeamPseudonymization `yaml:"cross_team_pseudonymization"`
	GCS                       GCS                       `yaml:"gcs"`
	BigQuery                  BigQuery                  `yaml:"big_query"`
	Slack                     Slack                     `yaml:"slack"`
	Server                    Server                    `yaml:"server"`
	Postgres                  Postgres                  `yaml:"postgres"`
	TeamsCatalogue            TeamsCatalogue            `yaml:"teams_catalogue"`
	TreatmentCatalogue        TreatmentCatalogue        `yaml:"treatment_catalogue"`
	GoogleGroups              GoogleGroups              `yaml:"google_groups"`
	Cookies                   Cookies                   `yaml:"cookies"`
	NaisConsole               NaisConsole               `yaml:"nais_console"`
	API                       API                       `yaml:"api"`
	ServiceAccount            ServiceAccount            `yaml:"service_account"`
	Workstation               Workstation               `yaml:"workstation"`
	SecureWebProxy            SecureWebProxy            `yaml:"secure_web_proxy"`
	CloudResourceManager      CloudResourceManager      `yaml:"cloud_resource_manager"`
	ComputeEngine             ComputeEngine             `yaml:"compute_engine"`
	CloudLogging              CloudLogging              `yaml:"cloud_logging"`
	ArtifactRegistry          ArtifactRegistry          `yaml:"artifact_registry"`
	OnpremMapping             OnpremMapping             `yaml:"onprem_mapping"`
	DVH                       DVH                       `yaml:"dvh"`
	IAMCredentials            IAMCredentials            `yaml:"iam_credentials"`
	PubSub                    PubSub                    `yaml:"pubsub"`

	PodName                        string `yaml:"pod_name"`
	EmailSuffix                    string `yaml:"email_suffix"`
	NaisClusterName                string `yaml:"nais_cluster_name"`
	KeywordsAdminGroup             string `yaml:"keywords_admin_group"`
	AllUsersEmail                  string `yaml:"all_users_email"`
	AllUsersGroup                  string `yaml:"all_users_group"`
	LoginPage                      string `yaml:"login_page"`
	LogLevel                       string `yaml:"log_level"`
	TeamProjectsUpdateDelaySeconds int    `yaml:"team_projects_update_delay_seconds"`
	KeepEmptyStoriesForDays        int    `yaml:"keep_empty_stories_for_days"`
	StoryCreateIgnoreMissingTeam   bool   `yaml:"story_create_ignore_missing_team"`
	Debug                          bool   `yaml:"debug"`
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Oauth, validation.Required),
		validation.Field(&c.Metabase, validation.Required),
		validation.Field(&c.Slack, validation.Required),
		validation.Field(&c.Server, validation.Required),
		validation.Field(&c.Postgres, validation.Required),
		validation.Field(&c.TeamsCatalogue, validation.Required),
		validation.Field(&c.TreatmentCatalogue, validation.Required),
		validation.Field(&c.GoogleGroups, validation.Required),
		validation.Field(&c.Cookies, validation.Required),
		validation.Field(&c.NaisConsole, validation.Required),
		validation.Field(&c.API, validation.Required),
		validation.Field(&c.LoginPage, validation.Required),
		validation.Field(&c.LogLevel, validation.Required),
		validation.Field(&c.AllUsersEmail, validation.Required, is.Email),
		validation.Field(&c.AllUsersGroup, validation.Required),
		validation.Field(&c.CrossTeamPseudonymization, validation.Required),
		validation.Field(&c.GCS, validation.Required),
		validation.Field(&c.BigQuery, validation.Required),
		validation.Field(&c.KeywordsAdminGroup, validation.Required),
		validation.Field(&c.NaisClusterName, validation.Required),
		validation.Field(&c.EmailSuffix, validation.Required),
		validation.Field(&c.KeepEmptyStoriesForDays, validation.Required, validation.Min(1)),
		validation.Field(&c.TeamProjectsUpdateDelaySeconds, validation.Required),
		validation.Field(&c.Workstation, validation.Required),
		validation.Field(&c.ServiceAccount),
		validation.Field(&c.OnpremMapping, validation.Required),
		validation.Field(&c.DVH, validation.Required),
		validation.Field(&c.PodName, validation.Required),
		validation.Field(&c.PubSub, validation.Required),
		validation.Field(&c.CloudResourceManager),
		validation.Field(&c.ComputeEngine),
		validation.Field(&c.CloudLogging),
		validation.Field(&c.ArtifactRegistry),
		validation.Field(&c.SecureWebProxy),
		validation.Field(&c.IAMCredentials),
	)
}

type PubSub struct {
	Location    string `yaml:"location"`
	ApiEndpoint string `yaml:"api_endpoint"`
	DisableAuth bool   `yaml:"disable_auth"`
}

func (p PubSub) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Location, validation.Required),
		validation.Field(&p.ApiEndpoint, is.URL),
	)
}

type ServiceAccount struct {
	EndpointOverride string `yaml:"endpoint"`
	DisableAuth      bool   `yaml:"disable_auth"`
}

func (s ServiceAccount) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.EndpointOverride, is.URL),
	)
}

type Workstation struct {
	WorkstationsProject             string   `yaml:"workstations_project"`
	ServiceAccountsProject          string   `yaml:"service_accounts_project"`
	Location                        string   `yaml:"location"`
	TLSSecureWebProxyPolicy         string   `yaml:"tls_secure_web_proxy_policy"`
	ClusterID                       string   `yaml:"clusterID"`
	FirewallPolicyName              string   `yaml:"firewall_policy_name"`
	LoggingBucket                   string   `yaml:"logging_bucket"`
	LoggingView                     string   `yaml:"logging_view"`
	ArtifactRepositoryName          string   `yaml:"artifact_repository_name"`
	ArtifactRepositoryProject       string   `yaml:"artifact_repository_project"`
	SignerServiceAccount            string   `yaml:"signer_service_account"`
	KnastADGroups                   []string `yaml:"knast_ad_groups"`
	MachineCostCacheDurationSeconds int      `yaml:"machine_cost_cache_duration_seconds"`
	StopSignalTopic                 string   `yaml:"stop_signal_topic"`
	StopSignalSubscription          string   `yaml:"stop_signal_subscription"`

	// AdministratorServiceAccount is the service account that has the necessary permissions to
	// create and manage resources in the workstation project, this is currently the
	// datamarkedsplassen service account.
	AdministratorServiceAccount string `yaml:"administrator_service_account"`
	EndpointOverride            string `yaml:"endpoint"`
	DisableAuth                 bool   `yaml:"disable_auth"`
}

func (w Workstation) Validate() error {
	return validation.ValidateStruct(&w,
		validation.Field(&w.EndpointOverride, is.URL),
		validation.Field(&w.WorkstationsProject, validation.Required),
		validation.Field(&w.ServiceAccountsProject, validation.Required),
		validation.Field(&w.Location, validation.Required),
		validation.Field(&w.TLSSecureWebProxyPolicy, validation.Required),
		validation.Field(&w.FirewallPolicyName, validation.Required),
		validation.Field(&w.LoggingBucket, validation.Required),
		validation.Field(&w.LoggingView, validation.Required),
		validation.Field(&w.AdministratorServiceAccount, validation.Required),
		validation.Field(&w.ClusterID, validation.Required),
		validation.Field(&w.ArtifactRepositoryName, validation.Required),
		validation.Field(&w.ArtifactRepositoryProject, validation.Required),
		validation.Field(&w.KnastADGroups, validation.Required),
		validation.Field(&w.MachineCostCacheDurationSeconds, validation.Required),
		validation.Field(&w.StopSignalTopic, validation.Required),
		validation.Field(&w.StopSignalSubscription, validation.Required),
	)
}

type OnpremMapping struct {
	Host        string `yaml:"host"`
	DisableAuth bool   `yaml:"disable_auth"`
	Bucket      string `yaml:"bucket"`
	MappingFile string `yaml:"mapping_file"`
}

func (w OnpremMapping) Validate() error {
	return validation.ValidateStruct(&w,
		validation.Field(&w.Bucket, validation.Required),
		validation.Field(&w.MappingFile, validation.Required),
	)
}

type DVH struct {
	Host         string `yaml:"host"`
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

func (w DVH) Validate() error {
	return validation.ValidateStruct(&w,
		validation.Field(&w.Host, validation.Required),
		validation.Field(&w.ClientID, validation.Required),
		validation.Field(&w.ClientSecret, validation.Required),
	)
}

type IAMCredentials struct {
	EndpointOverride string `yaml:"endpoint"`
	DisableAuth      bool   `yaml:"disable_auth"`
}

func (w IAMCredentials) Validate() error {
	return validation.ValidateStruct(&w,
		validation.Field(&w.EndpointOverride, is.URL),
	)
}

type ArtifactRegistry struct {
	EndpointOverride     string `yaml:"endpoint"`
	DisableAuth          bool   `yaml:"disable_auth"`
	CacheDurationSeconds int    `yaml:"cache_duration_seconds"`
}

func (w ArtifactRegistry) Validate() error {
	return validation.ValidateStruct(&w,
		validation.Field(&w.EndpointOverride, is.URL),
		validation.Field(&w.CacheDurationSeconds, validation.Required),
	)
}

type CloudResourceManager struct {
	EndpointOverride string `yaml:"endpoint"`
	DisableAuth      bool   `yaml:"disable_auth"`
}

func (w CloudResourceManager) Validate() error {
	return validation.ValidateStruct(&w,
		validation.Field(&w.EndpointOverride, is.URL),
	)
}

type SecureWebProxy struct {
	EndpointOverride string `yaml:"endpoint"`
	DisableAuth      bool   `yaml:"disable_auth"`
}

func (w SecureWebProxy) Validate() error {
	return validation.ValidateStruct(&w,
		validation.Field(&w.EndpointOverride, is.URL),
	)
}

type ComputeEngine struct {
	EndpointOverride string `yaml:"endpoint"`
	DisableAuth      bool   `yaml:"disable_auth"`
}

func (w ComputeEngine) Validate() error {
	return validation.ValidateStruct(&w,
		validation.Field(&w.EndpointOverride, is.URL),
	)
}

type CloudLogging struct {
	EndpointOverride string `yaml:"endpoint"`
	DisableAuth      bool   `yaml:"disable_auth"`
}

func (l CloudLogging) Validate() error {
	return validation.ValidateStruct(&l,
		validation.Field(&l.EndpointOverride, is.URL),
	)
}

type TreatmentCatalogue struct {
	APIURL     string `yaml:"api_url"`
	PurposeURL string `yaml:"purpose_url"`
}

func (t TreatmentCatalogue) Validate() error {
	return validation.ValidateStruct(&t,
		validation.Field(&t.APIURL, validation.Required, is.URL),
		validation.Field(&t.PurposeURL, validation.Required, is.URL),
	)
}

type API struct {
	AuthToken string `yaml:"auth_token"`
}

func (a API) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.AuthToken, validation.Required),
	)
}

type NaisConsole struct {
	APIKey string `yaml:"api_key"`
	APIURL string `yaml:"api_url"`
}

func (c NaisConsole) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.APIKey, validation.Required),
		validation.Field(&c.APIURL, validation.Required, is.URL),
	)
}

type GoogleGroups struct {
	ImpersonationSubject string `yaml:"impersonation_subject"`
	CredentialsFile      string `yaml:"credentials_file"`
}

func (g GoogleGroups) Validate() error {
	return validation.ValidateStruct(&g,
		validation.Field(&g.ImpersonationSubject, validation.Required),
		validation.Field(&g.CredentialsFile, validation.Required),
	)
}

type TeamsCatalogue struct {
	APIURL               string `yaml:"api_url"`
	CacheDurationSeconds int    `yaml:"cache_duration_seconds"`
}

func (t TeamsCatalogue) Validate() error {
	return validation.ValidateStruct(&t,
		validation.Field(&t.APIURL, validation.Required, is.URL),
		validation.Field(&t.CacheDurationSeconds, validation.Required),
	)
}

type Slack struct {
	Token           string `yaml:"token"`
	WebhookURL      string `yaml:"webhook_url"`
	ChannelOverride string `yaml:"channel_override"`
	DryRun          bool   `yaml:"dry_run"`
}

func (s Slack) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Token, validation.Required),
		validation.Field(&s.WebhookURL, validation.Required, is.URL),
	)
}

type KMS struct {
	Project  string `yaml:"project"`
	Location string `yaml:"location"`
	Keyring  string `yaml:"keyring"`
	KeyName  string `yaml:"key_name"`
}

func (k KMS) Validate() error {
	return validation.ValidateStruct(&k,
		validation.Field(&k.Project, validation.Required),
		validation.Field(&k.Location, validation.Required),
		validation.Field(&k.Keyring, validation.Required),
		validation.Field(&k.KeyName, validation.Required),
	)
}

type Metabase struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	APIURL   string `yaml:"api_url"`

	// GCPProject where metabase will create service accounts
	GCPProject       string                   `yaml:"gcp_project"`
	CredentialsPath  string                   `yaml:"credentials_path"`
	DatabasesBaseURL string                   `yaml:"databases_base_url"`
	BigQueryDatabase MetabaseBigQueryDatabase `yaml:"big_query_database"`

	KMS KMS `yaml:"kms"`

	MappingDeadlineSec  int `yaml:"mapping_deadline_sec"`
	MappingFrequencySec int `yaml:"mapping_frequency_sec"`
}

func (m Metabase) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Username, validation.Required),
		validation.Field(&m.Password, validation.Required),
		validation.Field(&m.GCPProject, validation.Required),
		validation.Field(&m.APIURL, validation.Required, is.URL),
		validation.Field(&m.DatabasesBaseURL, validation.Required, is.URL),
		validation.Field(&m.CredentialsPath, validation.Required),
		validation.Field(&m.MappingDeadlineSec, validation.Required),
		validation.Field(&m.MappingFrequencySec, validation.Required),
		validation.Field(&m.BigQueryDatabase),
	)
}

type MetabaseBigQueryDatabase struct {
	APIEndpointOverride string `yaml:"api_endpoint_override"`
	DisableAuth         bool   `yaml:"disable_auth"`
}

func (m MetabaseBigQueryDatabase) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.APIEndpointOverride, is.URL),
	)
}

func (m Metabase) LoadFromCredentialsPath() (string, string, error) {
	sa, err := os.ReadFile(m.CredentialsPath)
	if err != nil {
		return "", "", fmt.Errorf("read service account: %w", err)
	}

	metabaseSA := struct {
		ClientEmail string `json:"client_email"`
	}{}

	err = json.Unmarshal(sa, &metabaseSA)
	if err != nil {
		return "", "", fmt.Errorf("unmarshal service account: %w", err)
	}

	return string(sa), metabaseSA.ClientEmail, nil
}

type Github struct {
	Organization        string `yaml:"organization"`
	ApplicationID       int64  `yaml:"application_id"`
	InstallationID      int64  `yaml:"installation_id"`
	PrivateKeyPath      string `yaml:"private_key_path"`
	RefreshIntervalMins int    `yaml:"refresh_interval_mins"`
}

func (g Github) Validate() error {
	return validation.ValidateStruct(&g,
		validation.Field(&g.Organization, validation.Required),
		validation.Field(&g.ApplicationID, validation.Required),
		validation.Field(&g.InstallationID, validation.Required),
		validation.Field(&g.PrivateKeyPath, validation.Required),
		validation.Field(&g.RefreshIntervalMins, validation.Required),
	)
}

type Postgres struct {
	UserName      string                `yaml:"user_name"`
	Password      string                `yaml:"password"`
	Host          string                `yaml:"host"`
	Port          string                `yaml:"port"`
	DatabaseName  string                `yaml:"database_name"`
	SSLMode       string                `yaml:"ssl_mode"`
	Configuration PostgresConfiguration `yaml:"configuration"`
}

func (p Postgres) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.UserName, validation.Required),
		validation.Field(&p.Password, validation.Required),
		validation.Field(&p.Host, validation.Required, is.URL),
		validation.Field(&p.Port, validation.Required, is.Port),
		validation.Field(&p.DatabaseName, validation.Required),
		validation.Field(&p.SSLMode, validation.Required, validation.In("disable", "allow", "prefer", "require")),
	)
}

func (p Postgres) ConnectionString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=%s",
		p.UserName,
		p.Password,
		net.JoinHostPort(p.Host, p.Port),
		p.DatabaseName,
		p.SSLMode,
	)
}

type PostgresConfiguration struct {
	MaxIdleConnections int `yaml:"max_idle_connections"`
	MaxOpenConnections int `yaml:"max_open_connections"`
}

type Server struct {
	Hostname string `yaml:"hostname"`
	Address  string `yaml:"address"`
	Port     string `yaml:"port"`
}

func (s Server) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.Address, validation.Required, is.IP),
		validation.Field(&s.Hostname, validation.Required, is.Host),
		validation.Field(&s.Port, validation.Required, is.Port),
	)
}

type Oauth struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	TenantID     string `yaml:"tenant_id"`
	RedirectURL  string `yaml:"redirect_url"`
	HMACKey      string `yaml:"hmac_key"`
}

func (o Oauth) Validate() error {
	return validation.ValidateStruct(&o,
		validation.Field(&o.ClientID, validation.Required),
		validation.Field(&o.ClientSecret, validation.Required),
		validation.Field(&o.TenantID, validation.Required),
		validation.Field(&o.RedirectURL, validation.Required, is.URL),
	)
}

// CrossTeamPseudonymization contains the configuration for pseudonymization in a cross-team context.
type CrossTeamPseudonymization struct {
	GCPProjectID string `yaml:"gcp_project_id"`
	GCPRegion    string `yaml:"gcp_region"`
}

func (p *CrossTeamPseudonymization) Validate() error {
	return validation.ValidateStruct(p,
		validation.Field(&p.GCPProjectID, validation.Required),
		validation.Field(&p.GCPRegion, validation.Required),
	)
}

type GCS struct {
	Endpoint          string `yaml:"endpoint"`
	DisableAuth       bool   `yaml:"disable_auth"`
	StoryBucketName   string `yaml:"story_bucket_name"`
	CentralGCPProject string `yaml:"central_gcp_project"`
}

func (g GCS) Validate() error {
	return validation.ValidateStruct(&g,
		validation.Field(&g.StoryBucketName, validation.Required),
		validation.Field(&g.CentralGCPProject, validation.Required),
	)
}

type BigQuery struct {
	Endpoint   string `yaml:"endpoint"`
	EnableAuth bool   `yaml:"enable_auth"`
	// TeamProjectPseudoViewsDatasetName is the name of the dataset in the team's
	// own gcp project that contains the pseudo views.
	TeamProjectPseudoViewsDatasetName string `yaml:"team_project_pseudo_views_dataset_name"`
	GCPRegion                         string `yaml:"gcp_region"`
	CentralGCPProject                 string `yaml:"central_gcp_project"`
}

func (b BigQuery) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.TeamProjectPseudoViewsDatasetName, validation.Required),
		validation.Field(&b.GCPRegion, validation.Required),
		validation.Field(&b.CentralGCPProject, validation.Required),
	)
}

type Cookies struct {
	Redirect   CookieSettings `yaml:"redirect"`
	OauthState CookieSettings `yaml:"oauth_state"`
	Session    CookieSettings `yaml:"session"`
}

func (c Cookies) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Redirect, validation.Required),
		validation.Field(&c.OauthState, validation.Required),
		validation.Field(&c.Session, validation.Required),
	)
}

type CookieSettings struct {
	Name     string `yaml:"name"`
	MaxAge   int    `yaml:"max_age"`
	Path     string `yaml:"path"`
	Domain   string `yaml:"domain"`
	SameSite string `yaml:"same_site"`
	Secure   bool   `yaml:"secure"`
	HttpOnly bool   `yaml:"http_only"`
}

func (c CookieSettings) GetSameSite() http.SameSite {
	switch c.SameSite {
	case "Strict":
		return http.SameSiteStrictMode
	case "Lax":
		return http.SameSiteLaxMode
	case "None":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}

func (c CookieSettings) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Name, validation.Required),
		validation.Field(&c.MaxAge, validation.Required),
		validation.Field(&c.Path, validation.Required),
		validation.Field(&c.Domain, validation.Required, is.Host),
		// Valid SameSite values:
		// - https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie#samesitesamesite-value
		validation.Field(&c.SameSite, validation.Required, validation.In("Strict", "Lax", "None")),
	)
}

type FileParts struct {
	FileName string
	Path     string
}

func ProcessConfigPath(configFile string) (FileParts, error) {
	absolutePath, err := filepath.Abs(configFile)
	if err != nil {
		return FileParts{}, fmt.Errorf("convert to absolute path: %w", err)
	}

	// Extract file name and extension
	fileName := filepath.Base(absolutePath)
	path := filepath.Dir(absolutePath)
	extension := filepath.Ext(fileName)

	if strings.ReplaceAll(strings.ToLower(extension), ".", "") != defaultExtension {
		return FileParts{}, fmt.Errorf("config file must have extension %s, got: %s", defaultExtension, extension)
	}

	return FileParts{
		FileName: fileName[:len(fileName)-len(extension)],
		Path:     path,
	}, nil
}

func NewFileSystemLoader() *FileSystemLoader {
	return &FileSystemLoader{}
}

type FileSystemLoader struct{}

func (fs *FileSystemLoader) Load(name, path, envPrefix string, b Binder) (Config, error) {
	v := viper.New()

	v.AddConfigPath(path)
	v.SetConfigName(name)
	v.SetConfigType(defaultExtension)

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // So that env vars are translated properly
	v.AutomaticEnv()

	if b != nil {
		err := b.Bind(v)
		if err != nil {
			return Config{}, err
		}
	}

	v.SetEnvPrefix(envPrefix)

	err := v.ReadInConfig()
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var config Config

	err = v.Unmarshal(&config, func(cfg *mapstructure.DecoderConfig) {
		cfg.TagName = defaultTagName // We use yaml tags in the config structs so we can marshal to yaml
	})
	if err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	return config, nil
}

type EnvBinder struct {
	binders map[string]string
}

func (e *EnvBinder) Bind(v *viper.Viper) error {
	for envVar, key := range e.binders {
		err := v.BindEnv(key, envVar)
		if err != nil {
			return fmt.Errorf("bind env var %s to key %s: %w", envVar, key, err)
		}
	}

	return nil
}

func NewEnvBinder(binders map[string]string) *EnvBinder {
	return &EnvBinder{
		binders: binders,
	}
}

func NewDefaultEnvBinder() *EnvBinder {
	return NewEnvBinder(map[string]string{
		"AZURE_APP_CLIENT_ID":                      "oauth.client_id",
		"AZURE_APP_CLIENT_SECRET":                  "oauth.client_secret",
		"AZURE_APP_TENANT_ID":                      "oauth.tenant_id",
		"NAIS_DATABASE_NADA_BACKEND_NADA_PASSWORD": "postgres.password",
		"NAIS_CLUSTER_NAME":                        "nais_cluster_name",
		"HOSTNAME":                                 "pod_name",
	})
}

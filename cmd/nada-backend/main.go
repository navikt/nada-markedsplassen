package main

import (
	"context"
	"github.com/navikt/nada-backend/pkg/cloudlogging"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/navikt/nada-backend/pkg/cloudresourcemanager"
	"github.com/navikt/nada-backend/pkg/computeengine"

	"github.com/navikt/nada-backend/pkg/securewebproxy"
	"github.com/navikt/nada-backend/pkg/syncers/bigquery_sync_tables"
	"github.com/navikt/nada-backend/pkg/workstations"

	"github.com/navikt/nada-backend/pkg/syncers/bigquery_datasource_policy"

	"github.com/navikt/nada-backend/pkg/syncers/teamkatalogen"
	"github.com/navikt/nada-backend/pkg/syncers/teamprojectsupdater"

	"github.com/navikt/nada-backend/pkg/syncers/metabase_tables"

	"github.com/navikt/nada-backend/pkg/syncers"

	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/syncers/project_policy"

	"github.com/navikt/nada-backend/pkg/syncers/empty_stories"

	"github.com/navikt/nada-backend/pkg/syncers/metabase_collections"

	"github.com/navikt/nada-backend/pkg/sa"

	"github.com/go-chi/chi/middleware"
	"github.com/navikt/nada-backend/pkg/syncers/metabase_mapper"

	"github.com/navikt/nada-backend/pkg/requestlogger"

	"github.com/go-chi/chi"
	"github.com/navikt/nada-backend/pkg/bq"
	"github.com/navikt/nada-backend/pkg/cache"
	"github.com/navikt/nada-backend/pkg/cs"
	"github.com/navikt/nada-backend/pkg/nc"
	"github.com/navikt/nada-backend/pkg/service/core"
	apiclients "github.com/navikt/nada-backend/pkg/service/core/api"
	"github.com/navikt/nada-backend/pkg/service/core/handlers"
	"github.com/navikt/nada-backend/pkg/service/core/routes"
	"github.com/navikt/nada-backend/pkg/service/core/storage"
	"github.com/navikt/nada-backend/pkg/syncers/access_ensurer"
	"github.com/navikt/nada-backend/pkg/tk"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/rs/zerolog"

	"github.com/navikt/nada-backend/pkg/api"
	"github.com/navikt/nada-backend/pkg/auth"
	"github.com/navikt/nada-backend/pkg/config/v2"
	"github.com/navikt/nada-backend/pkg/database"
	"github.com/prometheus/client_golang/prometheus"
	flag "github.com/spf13/pflag"
)

// nolint: gochecknoglobals
var configFilePath = flag.String("config", "config.yaml", "path to config file")

const (
	AccessEnsurerFrequency = 5 * time.Minute
	RunIntervalOneHour     = 3600
	RunIntervalTenMinutes  = 600
	QueueBufferSize        = 100
	ClientTimeout          = 10 * time.Second
	ShutdownTimeout        = 5 * time.Second
	ReadHeaderTimeout      = 5 * time.Second
)

// nolint: cyclop,maintidx
func main() {
	flag.Parse()

	promErrs := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "nada_backend",
		Name:      "errors_total",
		Help:      "Total number of errors",
	}, []string{"location"})

	loc, _ := time.LoadLocation("Europe/Oslo")

	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC().In(loc)
	}

	zlog := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	fileParts, err := config.ProcessConfigPath(*configFilePath)
	if err != nil {
		zlog.Fatal().Err(err).Msg("processing config path")
	}

	cfg, err := config.NewFileSystemLoader().Load(fileParts.FileName, fileParts.Path, "NADA", config.NewDefaultEnvBinder())
	if err != nil {
		zlog.Fatal().Err(err).Msg("loading config")
	}

	err = cfg.Validate()
	if err != nil {
		zlog.Fatal().Err(err).Msg("validating config")
	}

	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)
	zlog = zlog.Level(level)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	repo, err := database.New(
		cfg.Postgres.ConnectionString(),
		cfg.Postgres.Configuration.MaxIdleConnections,
		cfg.Postgres.Configuration.MaxOpenConnections,
	)
	if err != nil {
		zlog.Fatal().Err(err).Msg("setting up database")
	}

	httpClient := &http.Client{
		Timeout: ClientTimeout,
	}

	tkFetcher := tk.New(cfg.TeamsCatalogue.APIURL, httpClient)
	ncFetcher := nc.New(cfg.NaisConsole.APIURL, cfg.NaisConsole.APIKey, cfg.NaisClusterName, httpClient)

	cacher := cache.New(time.Duration(cfg.CacheDurationSeconds)*time.Second, repo.GetDB(), zlog.With().Str("subsystem", "cache").Logger())

	bqClient := bq.NewClient(cfg.BigQuery.Endpoint, cfg.BigQuery.EnableAuth, zlog.With().Str("subsystem", "bq_client").Logger())

	csClient, err := cs.New(ctx, cfg.GCS.StoryBucketName)
	if err != nil {
		zlog.Fatal().Err(err).Msg("setting up cloud storage")
	}

	saClient := sa.NewClient(cfg.ServiceAccount.EndpointOverride, cfg.ServiceAccount.DisableAuth)

	crmClient := cloudresourcemanager.NewClient(cfg.CloudResourceManager.EndpointOverride, cfg.CloudResourceManager.DisableAuth, &http.Client{Timeout: ClientTimeout})

	wsClient := workstations.New(
		cfg.Workstation.WorkstationsProject,
		cfg.Workstation.Location,
		cfg.Workstation.ClusterID,
		cfg.Workstation.EndpointOverride,
		cfg.Workstation.DisableAuth,
	)

	swpClient := securewebproxy.New(cfg.SecureWebProxy.EndpointOverride, cfg.SecureWebProxy.DisableAuth)

	computeClient := computeengine.NewClient(cfg.SecureWebProxy.EndpointOverride, cfg.SecureWebProxy.DisableAuth)

	clClient := cloudlogging.NewClient(cfg.CloudLogging.EndpointOverride, cfg.CloudLogging.DisableAuth)

	stores := storage.NewStores(repo, cfg, zlog.With().Str("subsystem", "stores").Logger())
	apiClients := apiclients.NewClients(
		cacher,
		tkFetcher,
		ncFetcher,
		bqClient,
		csClient,
		saClient,
		crmClient,
		wsClient,
		swpClient,
		computeClient,
		clClient,
		cfg,
		zlog.With().Str("subsystem", "api_clients").Logger(),
	)
	services, err := core.NewServices(cfg, stores, apiClients, zlog.With().Str("subsystem", "services").Logger())
	if err != nil {
		zlog.Fatal().Err(err).Msg("setting up services")
	}

	googleGroups, err := auth.NewGoogleGroups(
		ctx,
		cfg.GoogleGroups.CredentialsFile,
		cfg.GoogleGroups.ImpersonationSubject,
	)
	if err != nil {
		zlog.Fatal().Err(err).Msg("setting up google groups")
	}

	azureGroups := auth.NewAzureGroups(
		http.DefaultClient,
		cfg.Oauth.ClientID,
		cfg.Oauth.ClientSecret,
		cfg.Oauth.TenantID,
		zlog.With().Str("subsystem", "azure_groups").Logger(),
	)

	aauth := auth.NewAzure(
		cfg.Oauth.ClientID,
		cfg.Oauth.ClientSecret,
		cfg.Oauth.TenantID,
		cfg.Oauth.RedirectURL,
	)

	httpAPI := api.NewHTTP(
		aauth,
		aauth.RedirectURL,
		cfg.LoginPage,
		cfg.Cookies,
		zlog.With().Str("subsystem", "api").Logger(),
	)

	authenticatorMiddleware := aauth.Middleware(
		aauth.KeyDiscoveryURL(),
		azureGroups,
		googleGroups,
		repo.GetDB(),
		zlog.With().Str("subsystem", "auth").Logger(),
	)

	// FIXME: hook up amplitude again, but as a middleware
	mapperQueue := make(chan metabase_mapper.Work, QueueBufferSize)
	h := handlers.NewHandlers(
		services,
		cfg,
		mapperQueue,
		zlog.With().Str("subsystem", "handlers").Logger(),
	)

	zlog.Info().Msgf("listening on %s:%s", cfg.Server.Address, cfg.Server.Port)
	auth.Init(repo.GetDB())

	router := chi.NewRouter()
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		zlog.Warn().Str("method", r.Method).Str("path", r.URL.Path).Msg("not found")
		w.WriteHeader(http.StatusNotFound)
	})

	router.Use(middleware.RequestID)
	router.Use(requestlogger.Middleware(
		zlog.With().Str("subsystem", "requestlogger").Logger(),
		"/internal/metrics",
	))

	routes.Add(router,
		routes.NewInsightProductRoutes(routes.NewInsightProductEndpoints(zlog, h.InsightProductHandler), authenticatorMiddleware),
		routes.NewAccessRoutes(routes.NewAccessEndpoints(zlog, h.AccessHandler), authenticatorMiddleware),
		routes.NewBigQueryRoutes(routes.NewBigQueryEndpoints(zlog, h.BigQueryHandler)),
		routes.NewDataProductsRoutes(routes.NewDataProductsEndpoints(zlog, h.DataProductsHandler), authenticatorMiddleware),
		routes.NewJoinableViewsRoutes(routes.NewJoinableViewsEndpoints(zlog, h.JoinableViewsHandler), authenticatorMiddleware),
		routes.NewKeywordRoutes(routes.NewKeywordEndpoints(zlog, h.KeywordsHandler), authenticatorMiddleware),
		routes.NewMetabaseRoutes(routes.NewMetabaseEndpoints(zlog, h.MetabaseHandler), authenticatorMiddleware),
		routes.NewPollyRoutes(routes.NewPollyEndpoints(zlog, h.PollyHandler)),
		routes.NewProductAreaRoutes(routes.NewProductAreaEndpoints(zlog, h.ProductAreasHandler)),
		routes.NewSearchRoutes(routes.NewSearchEndpoints(zlog, h.SearchHandler)),
		routes.NewSlackRoutes(routes.NewSlackEndpoints(zlog, h.SlackHandler)),
		routes.NewStoryRoutes(routes.NewStoryEndpoints(zlog, h.StoryHandler), authenticatorMiddleware, h.StoryHandler.NadaTokenMiddleware),
		routes.NewTeamkatalogenRoutes(routes.NewTeamkatalogenEndpoints(zlog, h.TeamKatalogenHandler)),
		routes.NewTokensRoutes(routes.NewTokensEndpoints(zlog, h.TokenHandler), authenticatorMiddleware),
		routes.NewMetricsRoutes(routes.NewMetricsEndpoints(prom(promErrs, repo.Metrics()...))),
		routes.NewUserRoutes(routes.NewUserEndpoints(zlog, h.UserHandler), authenticatorMiddleware),
		routes.NewAuthRoutes(routes.NewAuthEndpoints(httpAPI)),
		routes.NewWorkstationsRoutes(routes.NewWorkstationsEndpoints(zlog, h.WorkstationsHandler), authenticatorMiddleware),
	)

	err = routes.Print(router, os.Stdout)
	if err != nil {
		zlog.Fatal().Err(err).Msg("printing routes")
	}

	server := http.Server{
		Addr:              net.JoinHostPort(cfg.Server.Address, cfg.Server.Port),
		Handler:           router,
		ReadHeaderTimeout: ReadHeaderTimeout,
	}

	go syncers.New(
		RunIntervalTenMinutes,
		bigquery_sync_tables.New(
			services.BigQueryService,
		),
		zlog,
		syncers.DefaultOptions()...,
	).Run(ctx)

	go syncers.New(
		RunIntervalOneHour,
		bigquery_datasource_policy.New(
			[]string{
				bq.BigQueryMetadataViewerRole.String(),
				bq.BigQueryDataViewerRole.String(),
			},
			stores.BigQueryStorage,
			bqClient,
		),
		zlog,
		syncers.DefaultOptions()...,
	).Run(ctx)

	go syncers.New(
		RunIntervalOneHour,
		teamprojectsupdater.New(
			services.NaisConsoleService,
		),
		zlog,
		syncers.DefaultOptions()...,
	).Run(ctx)

	go syncers.New(
		RunIntervalOneHour,
		teamkatalogen.New(
			apiClients.TeamKatalogenAPI,
			stores.ProductAreaStorage,
		),
		zlog,
		syncers.DefaultOptions()...,
	).Run(ctx)

	go syncers.New(
		RunIntervalOneHour,
		metabase_tables.New(
			services.MetaBaseService,
		),
		zlog,
		syncers.DefaultOptions()...,
	).Run(ctx)

	go syncers.New(
		RunIntervalOneHour,
		project_policy.New(
			cfg.Metabase.GCPProject,
			[]string{service.NadaMetabaseRole(cfg.Metabase.GCPProject)},
			crmClient,
		),
		zlog,
		syncers.DefaultOptions()...,
	).Run(ctx)

	go syncers.New(
		RunIntervalOneHour,
		empty_stories.New(
			cfg.KeepEmptyStoriesForDays,
			stores.StoryStorage,
			apiClients.StoryAPI,
		),
		zlog,
		syncers.DefaultOptions()...,
	).Run(ctx)

	go syncers.New(
		RunIntervalOneHour,
		metabase_collections.New(
			apiClients.MetaBaseAPI,
			stores.MetaBaseStorage,
		),
		zlog,
		syncers.DefaultOptions()...,
	).Run(ctx)

	metabaseMapper := metabase_mapper.New(
		services.MetaBaseService,
		stores.ThirdPartyMappingStorage,
		cfg.Metabase.MappingDeadlineSec,
		mapperQueue,
		zlog.With().Str("subsystem", "metabase_mapper").Logger(),
	)
	// Starts processing of the work queue
	go metabaseMapper.ProcessQueue(ctx)
	// Starts the syncer that fills the work queue
	go syncers.New(
		cfg.Metabase.MappingFrequencySec,
		metabaseMapper,
		zlog,
		syncers.DefaultOptions()...,
	).Run(ctx)

	go access_ensurer.NewEnsurer(
		googleGroups,
		cfg.BigQuery.CentralGCPProject,
		promErrs,
		stores.AccessStorage,
		services.MetaBaseService,
		stores.DataProductsStorage,
		stores.BigQueryStorage,
		apiClients.BigQueryAPI,
		services.BigQueryService,
		services.JoinableViewService,
		zlog.With().Str("subsystem", "accessensurer").Logger(),
	).Run(ctx, AccessEnsurerFrequency)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			zlog.Fatal().Err(err).Msg("ListenAndServe")
		}
	}()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		zlog.Warn().Err(err).Msg("Shutdown error")
	}
}

func prom(promErrs prometheus.Collector, cols ...prometheus.Collector) *prometheus.Registry {
	r := prometheus.NewRegistry()
	r.MustRegister(promErrs)
	r.MustRegister(collectors.NewGoCollector())
	r.MustRegister(cols...)

	return r
}

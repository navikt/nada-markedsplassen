package api

import (
	"github.com/navikt/nada-backend/pkg/artifactregistry"
	"github.com/navikt/nada-backend/pkg/bq"
	"github.com/navikt/nada-backend/pkg/cache"
	"github.com/navikt/nada-backend/pkg/cloudlogging"
	"github.com/navikt/nada-backend/pkg/cloudresourcemanager"
	cs "github.com/navikt/nada-backend/pkg/cloudstorage"
	"github.com/navikt/nada-backend/pkg/computeengine"
	"github.com/navikt/nada-backend/pkg/config/v2"
	"github.com/navikt/nada-backend/pkg/datavarehus"
	"github.com/navikt/nada-backend/pkg/nc"
	"github.com/navikt/nada-backend/pkg/sa"
	"github.com/navikt/nada-backend/pkg/securewebproxy"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/navikt/nada-backend/pkg/service/core/api/gcp"
	httpapi "github.com/navikt/nada-backend/pkg/service/core/api/http"
	slackapi "github.com/navikt/nada-backend/pkg/service/core/api/slack"
	"github.com/navikt/nada-backend/pkg/service/core/api/static"
	"github.com/navikt/nada-backend/pkg/service/core/cache/postgres"
	"github.com/navikt/nada-backend/pkg/tk"
	"github.com/navikt/nada-backend/pkg/workstations"
	"github.com/rs/zerolog"
)

type Clients struct {
	ArtifactRegistryAPI     service.ArtifactRegistryAPI
	BigQueryAPI             service.BigQueryAPI
	CloudLoggingAPI         service.CloudLoggingAPI
	CloudResourceManagerAPI service.CloudResourceManagerAPI
	CloudStorageAPI         service.CloudStorageAPI
	ComputeAPI              service.ComputeAPI
	DatavarehusAPI          service.DatavarehusAPI
	MetaBaseAPI             service.MetabaseAPI
	NaisConsoleAPI          service.NaisConsoleAPI
	PollyAPI                service.PollyAPI
	SecureWebProxyAPI       service.SecureWebProxyAPI
	ServiceAccountAPI       service.ServiceAccountAPI
	SlackAPI                service.SlackAPI
	StoryAPI                service.StoryAPI
	TeamKatalogenAPI        service.TeamKatalogenAPI
	WorkstationsAPI         service.WorkstationsAPI
}

func NewClients(
	cache cache.Cacher,
	tkFetcher tk.Fetcher,
	ncFetcher nc.Fetcher,
	bqClient bq.Operations,
	storyStorageClient cs.Operations,
	cloudStorageClient cs.Operations,
	datavarehusClient datavarehus.Operations,
	saClient sa.Operations,
	crmClient cloudresourcemanager.Operations,
	wsClient workstations.Operations,
	swpClient securewebproxy.Operations,
	computeClient computeengine.Operations,
	clClient cloudlogging.Operations,
	arClient artifactregistry.Operations,
	cfg config.Config,
	log zerolog.Logger,
) *Clients {
	tkAPI := httpapi.NewTeamKatalogenAPI(tkFetcher, log)
	tkAPICacher := postgres.NewTeamKatalogenCache(tkAPI, cache)
	var slack service.SlackAPI = static.NewSlackAPI(log)
	if !cfg.Slack.DryRun {
		log.Info().Msg("Slack API is enabled, notifications will be sent")
		slack = slackapi.NewSlackAPI(
			cfg.Slack.WebhookURL,
			cfg.Slack.Token,
		)
	}
	return &Clients{
		BigQueryAPI: gcp.NewBigQueryAPI(
			cfg.BigQuery.CentralGCPProject,
			cfg.BigQuery.GCPRegion,
			cfg.BigQuery.TeamProjectPseudoViewsDatasetName,
			bqClient,
		),
		StoryAPI: gcp.NewStoryAPI(
			storyStorageClient,
			log.With().Str("component", "story").Logger(),
		),
		CloudStorageAPI:   gcp.NewCloudStorageAPI(cloudStorageClient, log),
		DatavarehusAPI:    httpapi.NewDatavarehusAPI(datavarehusClient),
		ServiceAccountAPI: gcp.NewServiceAccountAPI(saClient),
		MetaBaseAPI: httpapi.NewMetabaseHTTP(
			cfg.Metabase.APIURL,
			cfg.Metabase.Username,
			cfg.Metabase.Password,
			cfg.Metabase.BigQueryDatabase.APIEndpointOverride,
			cfg.Metabase.BigQueryDatabase.DisableAuth,
			cfg.Debug,
			log.With().Str("component", "metabase").Logger(),
		),
		PollyAPI: httpapi.NewPollyAPI(
			cfg.TreatmentCatalogue.PurposeURL,
			cfg.TreatmentCatalogue.APIURL,
		),
		TeamKatalogenAPI: tkAPICacher,
		SlackAPI:         slack,
		NaisConsoleAPI: httpapi.NewNaisConsoleAPI(
			ncFetcher,
		),
		WorkstationsAPI:         gcp.NewWorkstationsAPI(wsClient),
		SecureWebProxyAPI:       gcp.NewSecureWebProxyAPI(log.With().Str("component", "swp").Logger(), swpClient),
		CloudResourceManagerAPI: gcp.NewCloudResourceManagerAPI(crmClient),
		ComputeAPI:              gcp.NewComputeAPI(cfg.Workstation.WorkstationsProject, computeClient),
		CloudLoggingAPI:         gcp.NewCloudLoggingAPI(clClient),
		ArtifactRegistryAPI:     gcp.NewArtifactRegistryAPI(arClient, log),
	}
}

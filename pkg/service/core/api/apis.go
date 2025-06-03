package api

import (
	"fmt"
	"github.com/navikt/nada-backend/pkg/kms"

	"github.com/navikt/nada-backend/pkg/artifactregistry"
	"github.com/navikt/nada-backend/pkg/bq"
	"github.com/navikt/nada-backend/pkg/cache"
	"github.com/navikt/nada-backend/pkg/cloudbilling"
	"github.com/navikt/nada-backend/pkg/cloudlogging"
	"github.com/navikt/nada-backend/pkg/cloudresourcemanager"
	cs "github.com/navikt/nada-backend/pkg/cloudstorage"
	"github.com/navikt/nada-backend/pkg/computeengine"
	"github.com/navikt/nada-backend/pkg/config/v2"
	"github.com/navikt/nada-backend/pkg/datavarehus"
	"github.com/navikt/nada-backend/pkg/iamcredentials"
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
	ArtifactRegistryAPI          service.ArtifactRegistryAPI
	ArtifactRegistryAPIWithCache service.ArtifactRegistryAPI
	BigQueryAPI                  service.BigQueryAPI
	CloudBillingAPI              service.CloudBillingAPI
	CloudLoggingAPI              service.CloudLoggingAPI
	CloudResourceManagerAPI      service.CloudResourceManagerAPI
	CloudStorageAPI              service.CloudStorageAPI
	ComputeAPI                   service.ComputeAPI
	DatavarehusAPI               service.DatavarehusAPI
	MetaBaseAPI                  service.MetabaseAPI
	NaisConsoleAPI               service.NaisConsoleAPI
	PollyAPI                     service.PollyAPI
	SecureWebProxyAPI            service.SecureWebProxyAPI
	ServiceAccountAPI            service.ServiceAccountAPI
	SlackAPI                     service.SlackAPI
	TeamKatalogenAPI             service.TeamKatalogenAPI
	WorkstationsAPI              service.WorkstationsAPI
	IAMCredentialsAPI            service.IAMCredentialsAPI
	KMSAPI                       service.KMSAPI
}

func NewClients(
	tkCache cache.Cacher,
	tkFetcher tk.Fetcher,
	ncFetcher nc.Fetcher,
	bqClient bq.Operations,
	cloudStorageClient cs.Operations,
	datavarehusClient datavarehus.Operations,
	saClient sa.Operations,
	crmClient cloudresourcemanager.Operations,
	wsClient workstations.Operations,
	cloudBillingClient cloudbilling.Operations,
	machineCostCache cache.Cacher,
	swpClient securewebproxy.Operations,
	computeClient computeengine.Operations,
	clClient cloudlogging.Operations,
	arCache cache.Cacher,
	arClient artifactregistry.Operations,
	iamCredentialsClient iamcredentials.Operations,
	kmsClient kms.Operations,
	cfg config.Config,
	log zerolog.Logger,
) *Clients {
	tkAPI := httpapi.NewTeamKatalogenAPI(tkFetcher, log)
	tkAPICacher := postgres.NewTeamKatalogenCache(tkAPI, tkCache)

	garAPI := gcp.NewArtifactRegistryAPI(arClient, log)
	garAPIWithCache := postgres.NewArtifactRegistryCache(garAPI, arCache)
	var slack service.SlackAPI = static.NewSlackAPI(log)
	if !cfg.Slack.DryRun {
		log.Info().Msg("Slack API is enabled, notifications will be sent")
		if len(cfg.Slack.ChannelOverride) > 0 {
			log.Warn().Msg(fmt.Sprintf("Slack override enabled; all notifications will be sent to %q", cfg.Slack.ChannelOverride))
		}
		log.Info().Msg("Slack API is enabled, notifications will be sent")
		slack = slackapi.NewSlackAPI(
			cfg.Slack.WebhookURL,
			cfg.Slack.Token,
			cfg.Slack.ChannelOverride,
		)
	}
	return &Clients{
		BigQueryAPI: gcp.NewBigQueryAPI(
			cfg.BigQuery.CentralGCPProject,
			cfg.BigQuery.GCPRegion,
			cfg.BigQuery.TeamProjectPseudoViewsDatasetName,
			bqClient,
		),
		CloudStorageAPI:   gcp.NewCloudStorageAPI(cloudStorageClient, log),
		DatavarehusAPI:    httpapi.NewDatavarehusAPI(datavarehusClient, log),
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
		WorkstationsAPI:              gcp.NewWorkstationsAPI(wsClient),
		SecureWebProxyAPI:            gcp.NewSecureWebProxyAPI(log.With().Str("component", "swp").Logger(), swpClient),
		CloudResourceManagerAPI:      gcp.NewCloudResourceManagerAPI(crmClient),
		ComputeAPI:                   gcp.NewComputeAPI(cfg.Workstation.WorkstationsProject, computeClient),
		CloudBillingAPI:              postgres.NewMachineCostCache(cloudBillingClient, machineCostCache),
		CloudLoggingAPI:              gcp.NewCloudLoggingAPI(clClient),
		ArtifactRegistryAPI:          garAPI,
		ArtifactRegistryAPIWithCache: garAPIWithCache,
		IAMCredentialsAPI:            gcp.NewIAMCredentialsAPI(iamCredentialsClient),
		KMSAPI:                       gcp.NewKMSAPI(kmsClient),
	}
}

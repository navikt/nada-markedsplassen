package workstation_signals

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/navikt/nada-backend/pkg/pubsub"
	"github.com/navikt/nada-backend/pkg/service"
	"github.com/rs/zerolog"
)

type Runner struct {
	config               *pubsub.SubscriptionConfig
	signerServiceAccount string
	podName              string

	pubsubOps         pubsub.Operations
	iamCredentialsAPI service.IAMCredentialsAPI
	datavarehusAPI    service.DatavarehusAPI

	log zerolog.Logger
}

func New(topicProject, topicName, subscriptionName, signerServiceAccount, podName string, pubsubOps pubsub.Operations, iamCredentialsAPI service.IAMCredentialsAPI, datavarehusAPI service.DatavarehusAPI, log zerolog.Logger) *Runner {
	expirationPolicy := 24 * time.Hour

	return &Runner{
		config: &pubsub.SubscriptionConfig{
			Project:                 topicProject,
			Topic:                   topicName,
			Name:                    subscriptionName,
			RetainAckedMessages:     true,
			RetentionDuration:       time.Duration(0),
			ExpirationPolicy:        &expirationPolicy,
			EnableExactOnceDelivery: true,
		},
		signerServiceAccount: signerServiceAccount,
		podName:              podName,
		pubsubOps:            pubsubOps,
		iamCredentialsAPI:    iamCredentialsAPI,
		datavarehusAPI:       datavarehusAPI,
		log:                  log,
	}
}

func (r *Runner) Start(ctx context.Context) {
	fields := map[string]interface{}{
		"project": r.config.Project,
		"topic":   r.config.Topic,
		"sub":     r.config.Name,
	}

	r.log.Info().Fields(fields).Msg("Starting workstation signals syncer")

	subscription, err := r.pubsubOps.GetSubscription(ctx, r.config.Project, r.config.Name)
	if err != nil {
		r.log.Fatal().Err(err).Fields(fields).Msg("getting subscription")
	}

	err = r.pubsubOps.Subscribe(ctx, r.config.Project, subscription.Name, r.ProcessMessage)
	if err != nil {
		r.log.Fatal().Err(err).Fields(fields).Msg("subscribing to topic")
	}
}

// WorkstationShutdownLog mimics the structure of the log entries produced
// by the following query:
// - https://cloudlogging.app.goo.gl/hfSFLs9rb3WzxDuV6
type WorkstationShutdownLog struct {
	TextPayload string `json:"textPayload"`
	InsertID    string `json:"insertId"`
	Resource    struct {
		Type   string `json:"type"`
		Labels struct {
			ResourceContainer string `json:"resource_container"`
			ClusterID         string `json:"cluster_id"`
			ConfigID          string `json:"config_id"`
			WorkstationID     string `json:"workstation_id"`
			Location          string `json:"location"`
		} `json:"labels"`
	} `json:"resource"`
	Timestamp string `json:"timestamp"`
	Severity  string `json:"severity"`
	Labels    struct {
		InstanceID     string `json:"instance_id"`
		InternalIP     string `json:"internal_ip"`
		InstanceName   string `json:"instance_name"`
		ExternalIP     string `json:"external_ip"`
		ShutdownReason string `json:"shutdown_reason"`
	} `json:"labels"`
	LogName          string `json:"logName"`
	ReceiveTimestamp string `json:"receiveTimestamp"`
}

func (r *Runner) ProcessMessage(ctx context.Context, msg []byte) pubsub.MessageResult {
	r.log.Info().Msg("Processing message")

	var workstationShutdownLog WorkstationShutdownLog

	err := json.Unmarshal(msg, &workstationShutdownLog)
	if err != nil {
		r.log.Error().Err(err).Msg("unmarshalling message")

		return pubsub.MessageResult{
			Success: false,
		}
	}

	claims := &service.DVHClaims{
		Ident:              strings.ToLower(workstationShutdownLog.Resource.Labels.WorkstationID),
		IP:                 workstationShutdownLog.Labels.InternalIP,
		Databases:          []string{},
		Reference:          workstationShutdownLog.InsertID,
		PodName:            r.podName,
		SessionDurationSec: 0,
	}

	signedJWT, err := r.iamCredentialsAPI.SignJWT(ctx, r.signerServiceAccount, claims.ToMapClaims())
	if err != nil {
		r.log.Error().Err(err).Msg("signing JWT")

		return pubsub.MessageResult{
			Success: false,
		}
	}

	err = r.datavarehusAPI.SendJWT(ctx, signedJWT.KeyID, signedJWT.SignedJWT)
	if err != nil {
		r.log.Error().Err(err).Msg("sending JWT")

		// If the JWT is not sent successfully, we ack the message regardless
		// because we don't want to retry sending the JWT.
		return pubsub.MessageResult{
			Success: true,
		}
	}

	return pubsub.MessageResult{
		Success: true,
	}
}

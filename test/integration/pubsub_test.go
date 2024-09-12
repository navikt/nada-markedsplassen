package integration

import (
	"context"
	"github.com/navikt/nada-backend/pkg/pubsub"
	"github.com/stretchr/testify/require"
	"os"
	"testing"

	"github.com/rs/zerolog"
)

// nolint: tparallel
func TestPubSub(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	log := zerolog.New(os.Stdout)

	c := NewContainers(t, log)
	defer c.Cleanup()

	cfg := c.RunPubSub(NewPubSubConfig())

	client := pubsub.New(cfg.ClientConnectionURL(), true)
	topics, err := client.ListTopics(ctx, cfg.ProjectID)
	require.NoError(t, err)
	log.Info().Msgf("Topics: %v", topics)
}

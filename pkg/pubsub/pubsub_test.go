package pubsub_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/navikt/nada-backend/pkg/pubsub"
	"github.com/navikt/nada-backend/test/integration"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPubSub(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	log := zerolog.New(os.Stdout)

	c := integration.NewContainers(t, log)
	defer c.Cleanup()

	cfg := c.RunPubSub(integration.NewPubSubConfig())

	client := pubsub.New(cfg.Location, cfg.ClientConnectionURL(), true)

	t.Run("List topics - no topic exists", func(t *testing.T) {
		topics, err := client.ListTopics(ctx, cfg.ProjectID)
		require.NoError(t, err)
		assert.Empty(t, topics)
	})

	t.Run("Get topic does not exist", func(t *testing.T) {
		topicName := "topic"
		_, err := client.GetTopic(ctx, cfg.ProjectID, topicName)
		require.Error(t, err)
		require.ErrorIs(t, err, pubsub.ErrNotExist)
	})

	t.Run("Create topic", func(t *testing.T) {
		topicName := "topic"
		fullyQualifiedName := fmt.Sprintf("projects/%v/topics/%v", cfg.ProjectID, topicName)
		topic, err := client.CreateTopic(ctx, cfg.ProjectID, topicName)
		require.NoError(t, err)
		assert.Equal(t, topicName, topic.Name)
		assert.Equal(t, fullyQualifiedName, topic.FullyQualifiedName)
	})

	t.Run("Create topic fails because exists", func(t *testing.T) {
		topicName := "topic"
		_, err := client.CreateTopic(ctx, cfg.ProjectID, topicName)
		require.Error(t, err)
		require.ErrorIs(t, err, pubsub.ErrExist)
	})

	t.Run("Get topic", func(t *testing.T) {
		topicName := "topic"
		fullyQualifiedName := fmt.Sprintf("projects/%v/topics/%v", cfg.ProjectID, topicName)
		topic, err := client.GetTopic(ctx, cfg.ProjectID, topicName)
		require.NoError(t, err)
		assert.Equal(t, topicName, topic.Name)
		assert.Equal(t, fullyQualifiedName, topic.FullyQualifiedName)
	})

	t.Run("Get or create topic with existing topic", func(t *testing.T) {
		topicName := "topic"
		fullyQualifiedName := fmt.Sprintf("projects/%v/topics/%v", cfg.ProjectID, topicName)
		topic, err := client.GetOrCreateTopic(ctx, cfg.ProjectID, topicName)
		require.NoError(t, err)
		assert.Equal(t, topicName, topic.Name)
		assert.Equal(t, fullyQualifiedName, topic.FullyQualifiedName)
	})

	t.Run("Get or create topic with nonexisting topic", func(t *testing.T) {
		topicName := "newtopic"
		fullyQualifiedName := fmt.Sprintf("projects/%v/topics/%v", cfg.ProjectID, topicName)
		topic, err := client.GetOrCreateTopic(ctx, cfg.ProjectID, topicName)
		require.NoError(t, err)
		assert.Equal(t, topicName, topic.Name)
		assert.Equal(t, fullyQualifiedName, topic.FullyQualifiedName)
	})
}

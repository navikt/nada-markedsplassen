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

	topicName := "topic"

	t.Run("List topics - no topic exists", func(t *testing.T) {
		topics, err := client.ListTopics(ctx, cfg.ProjectID)
		require.NoError(t, err)
		assert.Empty(t, topics)
	})

	t.Run("Get topic does not exist", func(t *testing.T) {
		_, err := client.GetTopic(ctx, cfg.ProjectID, topicName)
		require.Error(t, err)
		require.ErrorIs(t, err, pubsub.ErrNotExist)
	})

	t.Run("Create topic", func(t *testing.T) {
		fullyQualifiedName := fmt.Sprintf("projects/%v/topics/%v", cfg.ProjectID, topicName)
		topic, err := client.CreateTopic(ctx, cfg.ProjectID, topicName)
		require.NoError(t, err)
		assert.Equal(t, topicName, topic.Name)
		assert.Equal(t, fullyQualifiedName, topic.FullyQualifiedName)
	})

	t.Run("Create topic fails because exists", func(t *testing.T) {
		_, err := client.CreateTopic(ctx, cfg.ProjectID, topicName)
		require.Error(t, err)
		require.ErrorIs(t, err, pubsub.ErrExist)
	})

	t.Run("Get topic", func(t *testing.T) {
		fullyQualifiedName := fmt.Sprintf("projects/%v/topics/%v", cfg.ProjectID, topicName)
		topic, err := client.GetTopic(ctx, cfg.ProjectID, topicName)
		require.NoError(t, err)
		assert.Equal(t, topicName, topic.Name)
		assert.Equal(t, fullyQualifiedName, topic.FullyQualifiedName)
	})

	t.Run("Get or create topic with existing topic", func(t *testing.T) {
		fullyQualifiedName := fmt.Sprintf("projects/%v/topics/%v", cfg.ProjectID, topicName)
		topic, err := client.GetOrCreateTopic(ctx, cfg.ProjectID, topicName)
		require.NoError(t, err)
		assert.Equal(t, topicName, topic.Name)
		assert.Equal(t, fullyQualifiedName, topic.FullyQualifiedName)
	})

	t.Run("Get or create topic with nonexisting topic", func(t *testing.T) {
		newTopicName := "newtopic"
		fullyQualifiedName := fmt.Sprintf("projects/%v/topics/%v", cfg.ProjectID, newTopicName)
		topic, err := client.GetOrCreateTopic(ctx, cfg.ProjectID, newTopicName)
		require.NoError(t, err)
		assert.Equal(t, newTopicName, topic.Name)
		assert.Equal(t, fullyQualifiedName, topic.FullyQualifiedName)
	})

	t.Run("List subscriptions - no subscriptions exists", func(t *testing.T) {
		subs, err := client.ListSubscriptions(ctx, cfg.ProjectID)
		require.NoError(t, err)
		assert.Empty(t, subs)
	})

	t.Run("Get subscription does not exist", func(t *testing.T) {
		subName := "subscription"
		_, err := client.GetSubscription(ctx, cfg.ProjectID, subName)
		require.Error(t, err)
		require.ErrorIs(t, err, pubsub.ErrNotExist)
	})

	t.Run("Create subscription", func(t *testing.T) {
		subName := "subscription"
		fullyQualifiedName := fmt.Sprintf("projects/%v/subscriptions/%v", cfg.ProjectID, subName)
		subscription, err := client.CreateSubscription(ctx, cfg.ProjectID, topicName, subName)
		require.NoError(t, err)
		assert.Equal(t, subName, subscription.Name)
		assert.Equal(t, fullyQualifiedName, subscription.FullyQualifiedName)
	})

	t.Run("Create subscription fails because exists", func(t *testing.T) {
		subName := "subscription"
		_, err := client.CreateSubscription(ctx, cfg.ProjectID, topicName, subName)
		require.Error(t, err)
		require.ErrorIs(t, err, pubsub.ErrExist)
	})

	t.Run("Get subscription", func(t *testing.T) {
		subName := "subscription"
		fullyQualifiedName := fmt.Sprintf("projects/%v/subscriptions/%v", cfg.ProjectID, subName)
		subscription, err := client.GetSubscription(ctx, cfg.ProjectID, subName)
		require.NoError(t, err)
		assert.Equal(t, subName, subscription.Name)
		assert.Equal(t, fullyQualifiedName, subscription.FullyQualifiedName)
	})

	t.Run("Get or create topic with existing subscription", func(t *testing.T) {
		subName := "subscription"
		fullyQualifiedName := fmt.Sprintf("projects/%v/subscriptions/%v", cfg.ProjectID, subName)
		subscription, err := client.GetOrCreateSubscription(ctx, cfg.ProjectID, topicName, subName)
		require.NoError(t, err)
		assert.Equal(t, subName, subscription.Name)
		assert.Equal(t, fullyQualifiedName, subscription.FullyQualifiedName)
	})

	t.Run("Get or create subscription with nonexisting subscription", func(t *testing.T) {
		subName := "newsubscription"
		fullyQualifiedName := fmt.Sprintf("projects/%v/subscriptions/%v", cfg.ProjectID, subName)
		subscription, err := client.GetOrCreateSubscription(ctx, cfg.ProjectID, topicName, subName)
		require.NoError(t, err)
		assert.Equal(t, subName, subscription.Name)
		assert.Equal(t, fullyQualifiedName, subscription.FullyQualifiedName)
	})

	t.Run("Get or create subscription with nonexisting topic", func(t *testing.T) {
		subName := "anothersubscription"
		topicNoExist := "noexisttopic"
		_, err := client.GetOrCreateSubscription(ctx, cfg.ProjectID, topicNoExist, subName)
		require.Error(t, err)
		require.ErrorIs(t, err, pubsub.ErrNotExist)
	})
}

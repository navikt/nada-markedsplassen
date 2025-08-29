package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"time"

	"cloud.google.com/go/pubsub/v2"
	"cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/api/option/internaloption"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/durationpb"
)

var _ Operations = &Client{}

type Operations interface {
	ListTopics(ctx context.Context, project string) ([]*Topic, error)
	GetOrCreateTopic(ctx context.Context, project, topicName string) (*Topic, error)
	GetTopic(ctx context.Context, project, topicName string) (*Topic, error)
	CreateTopic(ctx context.Context, project, topicName string) (*Topic, error)
	Publish(ctx context.Context, project, topicName string, message []byte) (string, error)
	PublishJSON(ctx context.Context, project, topicName string, message any) (string, error)
	ListSubscriptions(ctx context.Context, project string) ([]*Subscription, error)
	GetOrCreateSubscription(ctx context.Context, config *SubscriptionConfig) (*Subscription, error)
	GetSubscription(ctx context.Context, project, subName string) (*Subscription, error)
	CreateSubscription(ctx context.Context, config *SubscriptionConfig) (*Subscription, error)
	Subscribe(ctx context.Context, project, subName string, fn MessageHandlerFn) error
}

type MessageResult struct {
	Success bool
}

type MessageHandlerFn func(context.Context, []byte) MessageResult

type Client struct {
	location    string
	apiEndpoint string
	disableAuth bool
}

type Topic struct {
	FullyQualifiedName string
	Name               string
}

type Subscription struct {
	FullyQualifiedName string
	Name               string
}

type SubscriptionConfig struct {
	Project string
	Topic   string
	Name    string

	RetainAckedMessages     bool
	RetentionDuration       time.Duration
	ExpirationPolicy        *time.Duration
	EnableExactOnceDelivery bool
}

var (
	ErrExist    = errors.New("already exists")
	ErrNotExist = errors.New("not exists")
)

func (c *Client) ListTopics(ctx context.Context, project string) ([]*Topic, error) {
	client, err := c.clientFromProject(ctx, project)
	if err != nil {
		return nil, err
	}

	var topics []*Topic
	it := client.TopicAdminClient.ListTopics(ctx, &pubsubpb.ListTopicsRequest{
		Project: project,
	})
	for {
		topic, err := it.Next()

		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("iterating topics: %w", err)
		}

		topics = append(topics, &Topic{
			Name:               path.Base(topic.GetName()),
			FullyQualifiedName: topic.GetName(),
		})
	}

	return topics, nil
}

func (c *Client) GetOrCreateTopic(ctx context.Context, project, topicName string) (*Topic, error) {
	topic, err := c.GetTopic(ctx, project, topicName)
	if err != nil {
		if errors.Is(err, ErrNotExist) {
			return c.CreateTopic(ctx, project, topicName)
		}

		return nil, err
	}

	return topic, nil
}

func (c *Client) GetTopic(ctx context.Context, project, topicName string) (*Topic, error) {
	topic, err := c.getTopic(ctx, project, topicName)
	if err != nil {
		return nil, err
	}

	return &Topic{
		Name:               path.Base(topic.GetName()),
		FullyQualifiedName: topic.GetName(),
	}, nil
}

func (c *Client) getTopic(ctx context.Context, project, topicName string) (*pubsubpb.Topic, error) {
	client, err := c.clientFromProject(ctx, project)
	if err != nil {
		return nil, err
	}

	topic, err := client.TopicAdminClient.GetTopic(ctx, &pubsubpb.GetTopicRequest{
		Topic: topicName,
	})
	if err != nil {
		var gapierr *googleapi.Error
		if errors.As(err, &gapierr) && gapierr.Code == http.StatusNotFound {
			return nil, ErrNotExist
		}

		return nil, fmt.Errorf("getting topic %s in project %s: %w", topicName, project, err)
	}

	return topic, nil
}

func (c *Client) CreateTopic(ctx context.Context, project, topicName string) (*Topic, error) {
	client, err := c.clientFromProject(ctx, project)
	if err != nil {
		return nil, err
	}

	_, err = c.getTopic(ctx, project, topicName)
	if err != nil && !errors.Is(err, ErrNotExist) {
		return nil, err
	}

	topic, err := client.TopicAdminClient.CreateTopic(ctx, &pubsubpb.Topic{
		Name: topicName,
		MessageStoragePolicy: &pubsubpb.MessageStoragePolicy{
			AllowedPersistenceRegions: []string{c.location},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("creating topic: %w", err)
	}

	return &Topic{
		Name:               path.Base(topic.GetName()),
		FullyQualifiedName: topic.GetName(),
	}, nil
}

func (c *Client) Publish(ctx context.Context, project, topicName string, message []byte) (string, error) {
	_, err := c.getTopic(ctx, project, topicName)
	if err != nil {
		return "", err
	}

	client, err := c.clientFromProject(ctx, project)
	if err != nil {
		return "", err
	}

	publisher := client.Publisher(topicName)

	res := publisher.Publish(ctx, &pubsub.Message{
		Data: message,
	})

	serverID, err := res.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("publishing message to topic %s in project %s: %w", topicName, project, err)
	}

	return serverID, nil
}

func (c *Client) PublishJSON(ctx context.Context, project, topicName string, message any) (string, error) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("marshalling message for topic %s in project %v: %w", topicName, project, err)
	}

	return c.Publish(ctx, project, topicName, messageBytes)
}

func (c *Client) ListSubscriptions(ctx context.Context, project string) ([]*Subscription, error) {
	client, err := c.clientFromProject(ctx, project)
	if err != nil {
		return nil, err
	}

	var topics []*Subscription
	it := client.SubscriptionAdminClient.ListSubscriptions(ctx, &pubsubpb.ListSubscriptionsRequest{
		Project: project,
	})
	for {
		topic, err := it.Next()

		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("iterating subscriptions: %w", err)
		}

		topics = append(topics, &Subscription{
			Name:               path.Base(topic.GetName()),
			FullyQualifiedName: topic.GetName(),
		})
	}

	return topics, nil
}

func (c *Client) GetOrCreateSubscription(ctx context.Context, config *SubscriptionConfig) (*Subscription, error) {
	sub, err := c.getSubscription(ctx, config.Project, config.Name)
	if err != nil {
		if errors.Is(err, ErrNotExist) {
			return c.CreateSubscription(ctx, config)
		}

		return nil, err
	}

	return &Subscription{
		Name:               path.Base(sub.GetName()),
		FullyQualifiedName: sub.GetName(),
	}, nil
}

func (c *Client) getSubscription(ctx context.Context, project, subName string) (*pubsubpb.Subscription, error) {
	client, err := c.clientFromProject(ctx, project)
	if err != nil {
		return nil, err
	}

	sub, err := client.SubscriptionAdminClient.GetSubscription(ctx, &pubsubpb.GetSubscriptionRequest{
		Subscription: subName,
	})
	if err != nil {
		var gapierr *googleapi.Error
		if errors.As(err, &gapierr) && gapierr.Code == http.StatusNotFound {
			return nil, ErrNotExist
		}

		return nil, fmt.Errorf("getting subscription %s in project %s: %w", subName, project, err)
	}

	return sub, nil
}

func (c *Client) GetSubscription(ctx context.Context, project, subName string) (*Subscription, error) {
	sub, err := c.getSubscription(ctx, project, subName)
	if err != nil {
		return nil, err
	}

	return &Subscription{
		Name:               path.Base(sub.GetName()),
		FullyQualifiedName: sub.GetName(),
	}, nil
}

func (c *Client) CreateSubscription(ctx context.Context, config *SubscriptionConfig) (*Subscription, error) {
	client, err := c.clientFromProject(ctx, config.Project)
	if err != nil {
		return nil, err
	}

	topic, err := c.getTopic(ctx, config.Project, config.Topic)
	if err != nil {
		return nil, fmt.Errorf("getting topic %s.%s: %w", config.Project, config.Topic, err)
	}

	sub, err := c.getSubscription(ctx, config.Project, config.Name)
	if err != nil && !errors.Is(err, ErrNotExist) {
		return nil, err
	}

	var expirationPolicy *pubsubpb.ExpirationPolicy = nil
	if config.ExpirationPolicy != nil {
		expirationPolicy = &pubsubpb.ExpirationPolicy{
			Ttl: durationpb.New(*config.ExpirationPolicy),
		}
	}

	sub, err = client.SubscriptionAdminClient.CreateSubscription(ctx, &pubsubpb.Subscription{
		Topic:                     topic.GetName(),
		RetainAckedMessages:       config.RetainAckedMessages,
		MessageRetentionDuration:  durationpb.New(config.RetentionDuration),
		ExpirationPolicy:          expirationPolicy,
		EnableExactlyOnceDelivery: config.EnableExactOnceDelivery,
	})
	if err != nil {
		return nil, fmt.Errorf("creating subscription %s for topic %s in project %s: %w", config.Name, config.Topic, config.Project, err)
	}

	return &Subscription{
		Name:               path.Base(sub.GetName()),
		FullyQualifiedName: sub.GetName(),
	}, nil
}

func (c *Client) Subscribe(ctx context.Context, project, subName string, fn MessageHandlerFn) error {
	_, err := c.getSubscription(ctx, project, subName)
	if err != nil {
		return err
	}

	client, err := c.clientFromProject(ctx, project)
	if err != nil {
		return err
	}

	subscriber := client.Subscriber(subName)

	err = subscriber.Receive(ctx, func(ctx context.Context, message *pubsub.Message) {
		result := fn(ctx, message.Data)
		if result.Success {
			message.Ack()
			return
		}

		message.Nack()
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) clientFromProject(ctx context.Context, project string) (*pubsub.Client, error) {
	var options []option.ClientOption

	if c.apiEndpoint != "" {
		options = append(options, option.WithEndpoint(c.apiEndpoint))
	}

	if c.disableAuth {
		options = append(options,
			option.WithoutAuthentication(),
			option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
			option.WithTelemetryDisabled(),
			internaloption.SkipDialSettingsValidation(),
		)
	}

	client, err := pubsub.NewClient(ctx, project, options...)
	if err != nil {
		return nil, fmt.Errorf("creating pubsub client: %w", err)
	}

	return client, nil
}

func New(location, apiEndpoint string, disableAuth bool) *Client {
	return &Client{
		apiEndpoint: apiEndpoint,
		disableAuth: disableAuth,
		location:    location,
	}
}

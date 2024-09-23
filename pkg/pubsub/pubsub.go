package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/api/option/internaloption"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	GetOrCreateSubscription(ctx context.Context, project, topicName, subName string) (*Subscription, error)
	GetSubscription(ctx context.Context, project, subName string) (*Subscription, error)
	CreateSubscription(ctx context.Context, project, topicName, subName string) (*Subscription, error)
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
	it := client.Topics(ctx)
	for {
		topic, err := it.Next()

		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("iterating topics: %w", err)
		}

		topics = append(topics, &Topic{
			Name:               topic.ID(),
			FullyQualifiedName: topic.String(),
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
		FullyQualifiedName: topic.String(),
		Name:               topic.ID(),
	}, nil
}

func (c *Client) getTopic(ctx context.Context, project, topicName string) (*pubsub.Topic, error) {
	client, err := c.clientFromProject(ctx, project)
	if err != nil {
		return nil, err
	}

	topic := client.Topic(topicName)
	exists, err := topic.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("checking topic existence: %w", err)
	}

	if !exists {
		return nil, ErrNotExist
	}

	return topic, nil
}

func (c *Client) CreateTopic(ctx context.Context, project, topicName string) (*Topic, error) {
	client, err := c.clientFromProject(ctx, project)
	if err != nil {
		return nil, err
	}

	topic := client.Topic(topicName)
	exists, err := topic.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("checking topic existence: %w", err)
	}

	if exists {
		return nil, ErrExist
	}

	topic, err = client.CreateTopicWithConfig(ctx, topicName, &pubsub.TopicConfig{
		MessageStoragePolicy: pubsub.MessageStoragePolicy{
			AllowedPersistenceRegions: []string{c.location},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("creating topic: %w", err)
	}

	return &Topic{
		FullyQualifiedName: topic.String(),
		Name:               topic.ID(),
	}, nil
}

func (c *Client) Publish(ctx context.Context, project, topicName string, message []byte) (string, error) {
	topic, err := c.getTopic(ctx, project, topicName)
	if err != nil {
		return "", err
	}

	res := topic.Publish(ctx, &pubsub.Message{
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
	it := client.Subscriptions(ctx)
	for {
		topic, err := it.Next()

		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("iterating subscriptions: %w", err)
		}

		topics = append(topics, &Subscription{
			Name:               topic.ID(),
			FullyQualifiedName: topic.String(),
		})
	}

	return topics, nil
}

func (c *Client) GetOrCreateSubscription(ctx context.Context, project, topicName, subName string) (*Subscription, error) {
	sub, err := c.getSubscription(ctx, project, subName)
	if err != nil {
		if errors.Is(err, ErrNotExist) {
			return c.CreateSubscription(ctx, project, topicName, subName)
		}
		return nil, err
	}

	return &Subscription{
		Name:               sub.ID(),
		FullyQualifiedName: sub.String(),
	}, nil
}

func (c *Client) getSubscription(ctx context.Context, project, subName string) (*pubsub.Subscription, error) {
	client, err := c.clientFromProject(ctx, project)
	if err != nil {
		return nil, err
	}

	sub := client.Subscription(subName)
	exists, err := sub.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("checking subscription existence for %s: %w", subName, err)
	}

	if !exists {
		return nil, ErrNotExist
	}

	return sub, nil
}

func (c *Client) GetSubscription(ctx context.Context, project, subName string) (*Subscription, error) {
	sub, err := c.getSubscription(ctx, project, subName)
	if err != nil {
		return nil, err
	}

	return &Subscription{
		Name:               sub.ID(),
		FullyQualifiedName: sub.String(),
	}, nil
}

func (c *Client) CreateSubscription(ctx context.Context, project, topicName, subName string) (*Subscription, error) {
	client, err := c.clientFromProject(ctx, project)
	if err != nil {
		return nil, err
	}

	topic, err := c.getTopic(ctx, project, topicName)
	if err != nil {
		return nil, err
	}

	sub := client.Subscription(subName)
	exists, err := sub.Exists(ctx)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrExist
	}

	sub, err = client.CreateSubscription(ctx, subName, pubsub.SubscriptionConfig{
		Topic: topic,
	})
	if err != nil {
		return nil, fmt.Errorf("creating subscription %s for topic %s in project %s: %w", subName, topicName, project, err)
	}

	return &Subscription{
		Name:               sub.ID(),
		FullyQualifiedName: sub.String(),
	}, nil
}

func (c *Client) Subscribe(ctx context.Context, project, subName string, fn MessageHandlerFn) error {
	sub, err := c.getSubscription(ctx, project, subName)
	if err != nil {
		return err
	}

	err = sub.Receive(ctx, func(ctx context.Context, message *pubsub.Message) {
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

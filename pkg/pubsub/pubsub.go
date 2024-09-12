package pubsub

import (
	"cloud.google.com/go/pubsub"
	"context"
	"errors"
	"fmt"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/api/option/internaloption"
	"google.golang.org/grpc"
)

type Operations interface {
	ListTopics(ctx context.Context, project string) ([]string, error)
}

type Client struct {
	apiEndpoint string
	disableAuth bool
}

func (c *Client) ListTopics(ctx context.Context, project string) ([]string, error) {
	client, err := c.clientFromProject(ctx, project)
	if err != nil {
		return nil, err
	}

	var topics []string
	it := client.Topics(ctx)
	for {
		topic, err := it.Next()

		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("iterating topics: %w", err)
		}

		topics = append(topics, topic.ID())
	}

	return topics, nil
}

func (c *Client) clientFromProject(ctx context.Context, project string) (*pubsub.Client, error) {
	var options []option.ClientOption

	if c.apiEndpoint != "" {
		options = append(options, option.WithEndpoint(c.apiEndpoint))
	}

	if c.disableAuth {
		options = append(options,
			option.WithoutAuthentication(),
			option.WithGRPCDialOption(grpc.WithInsecure()),
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

func New(apiEndpoint string, disableAuth bool) *Client {
	return &Client{
		apiEndpoint: apiEndpoint,
		disableAuth: disableAuth,
	}
}

package securewebproxy

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/networksecurity/v1"
	"google.golang.org/api/option"
	"net/http"
)

var (
	ErrNotExist = errors.New("not exist")
	ErrExist    = errors.New("exist")
)

var _ Operations = &Client{}

type Operations interface {
	GetURLList(ctx context.Context, id *URLListIdentifier) ([]string, error)
	CreateURLList(ctx context.Context, opts *URLListCreateOpts) error
	DeleteURLList(ctx context.Context, id *URLListIdentifier) error
}

type URLListIdentifier struct {
	// Project is the gcp project id
	Project string

	// Location is the gcp region
	Location string

	// Slug is the name of the url list
	Slug string
}

func (i *URLListIdentifier) Parent() string {
	return fmt.Sprintf("projects/%s/locations/%s", i.Project, i.Location)
}

func (i *URLListIdentifier) FullyQualifiedName() string {
	return fmt.Sprintf("projects/%s/locations/%s/urlLists/%s", i.Project, i.Location, i.Slug)
}

type URLListCreateOpts struct {
	ID          *URLListIdentifier
	Description string
	URLS        []string
}

type Client struct {
	apiEndpoint string
	disableAuth bool
}

func (c *Client) GetURLList(ctx context.Context, id *URLListIdentifier) ([]string, error) {
	client, err := c.newClient(ctx)
	if err != nil {
		return nil, err
	}

	list, err := client.Projects.Locations.UrlLists.Get(id.FullyQualifiedName()).Do()
	if err != nil {
		var gapierr *googleapi.Error
		if errors.As(err, &gapierr) && gapierr.Code == http.StatusNotFound {
			return nil, ErrNotExist
		}

		return nil, err
	}

	return list.Values, nil
}

func (c *Client) CreateURLList(ctx context.Context, opts *URLListCreateOpts) error {
	client, err := c.newClient(ctx)
	if err != nil {
		return err
	}

	_, err = client.Projects.Locations.UrlLists.Create(opts.ID.Parent(), &networksecurity.UrlList{
		Name:        opts.ID.FullyQualifiedName(),
		Description: opts.Description,
		Values:      opts.URLS,
	}).Do()
	if err != nil {
		var gapierr *googleapi.Error
		if errors.As(err, &gapierr) && gapierr.Code == http.StatusConflict {
			return ErrExist
		}

		return err
	}

	return nil
}

func (c *Client) DeleteURLList(ctx context.Context, id *URLListIdentifier) error {
	client, err := c.newClient(ctx)
	if err != nil {
		return err
	}

	_, err = client.Projects.Locations.UrlLists.Delete(id.FullyQualifiedName()).Do()
	if err != nil {
		var gapierr *googleapi.Error
		if errors.As(err, &gapierr) && gapierr.Code == http.StatusNotFound {
			return nil
		}

		return err
	}

	return nil
}

func (c *Client) newClient(ctx context.Context) (*networksecurity.Service, error) {
	var options []option.ClientOption

	if c.disableAuth {
		options = append(options, option.WithoutAuthentication())
	}

	if c.apiEndpoint != "" {
		options = append(options, option.WithEndpoint(c.apiEndpoint))
	}

	client, err := networksecurity.NewService(ctx, options...)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func New(apiEndpoint string, disableAuth bool) *Client {
	return &Client{
		apiEndpoint: apiEndpoint,
		disableAuth: disableAuth,
	}
}

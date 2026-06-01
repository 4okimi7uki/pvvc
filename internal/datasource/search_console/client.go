package searchconsole

import (
	"context"
	"fmt"

	"cloud.google.com/go/auth"
	"google.golang.org/api/option"
	"google.golang.org/api/searchconsole/v1"
)

type Client struct {
	svc     *searchconsole.Service
	siteURL string
}

func New(ctx context.Context, credential *auth.Credentials, siteURL string) (*Client, error) {
	opts := []option.ClientOption{
		option.WithScopes(searchconsole.WebmastersReadonlyScope),
	}
	if credential != nil {
		opts = append(opts, option.WithAuthCredentials(credential))
	}
	svc, err := searchconsole.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("search console: failed to create service: %w", err)
	}

	return &Client{
		svc:     svc,
		siteURL: siteURL,
	}, nil
}

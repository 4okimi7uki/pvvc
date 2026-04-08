package ga4

import (
	"context"
	"fmt"

	analyticsdata "google.golang.org/api/analyticsdata/v1beta"
	"google.golang.org/api/option"
)

type Client struct {
	svc        *analyticsdata.Service
	propertyID string
}

func New(ctx context.Context, propertyID string, credentialsFile string) (*Client, error) {
	opts := []option.ClientOption{
		option.WithScopes(analyticsdata.AnalyticsReadonlyScope),
	}
	if credentialsFile != "" {
		opts = append(opts, option.WithAuthCredentialsFile(option.ServiceAccount, credentialsFile))
	}
	svc, err := analyticsdata.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("ga4: failed to create service: %w", err)
	}

	return &Client{
		svc:        svc,
		propertyID: "properties/" + propertyID,
	}, nil
}

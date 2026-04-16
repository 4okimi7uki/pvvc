package ga4

import (
	"context"
	"fmt"

	"cloud.google.com/go/auth"
	analyticsdata "google.golang.org/api/analyticsdata/v1beta"
	"google.golang.org/api/option"
)

type Client struct {
	svc        *analyticsdata.Service
	propertyID string
	Raw        bool
	RawPages   [][]byte
}

func New(ctx context.Context, propertyID string, credential *auth.Credentials) (*Client, error) {
	opts := []option.ClientOption{
		option.WithScopes(analyticsdata.AnalyticsReadonlyScope),
	}
	if credential != nil {
		opts = append(opts, option.WithAuthCredentials(credential))
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

package ga4

import (
	"context"
	"fmt"
	"sync"

	"cloud.google.com/go/auth"
	analyticsdata "google.golang.org/api/analyticsdata/v1beta"
	"google.golang.org/api/option"
)

type Client struct {
	svc        *analyticsdata.Service
	propertyID string
	Raw        bool

	// RawPages は --raw 用の生レスポンス。日次PVは期間を分割して並行に取りに行くので
	// 追記は rawMu で守る。読むのは全フェッチが終わったあとの前提。
	rawMu    sync.Mutex
	RawPages [][]byte
}

// appendRaw は c.Raw のときだけ生レスポンスを控える。
func (c *Client) appendRaw(b []byte) {
	c.rawMu.Lock()
	c.RawPages = append(c.RawPages, b)
	c.rawMu.Unlock()
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

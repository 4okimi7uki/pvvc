package vercel

import (
	"fmt"
	"net/http"

	"github.com/4okimi7uki/pvvc/internal/httpclient"
)

const baseURL = "https://api.vercel.com"

type Client struct {
	token      string
	teamID     string
	httpClient *http.Client
	Raw        bool
	RawBody    []byte
}

func New(token, teamID string) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("vercel: token is required")
	}
	return &Client{
		token:      token,
		teamID:     teamID,
		httpClient: httpclient.New(),
	}, nil
}

func (c *Client) newRequest(method, path string) (*http.Request, error) {
	req, err := http.NewRequest(method, baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

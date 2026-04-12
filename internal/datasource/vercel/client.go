package vercel

import (
	"fmt"
	"net/http"
)

const baseURL = "https://api.vercel.com"

type Client struct {
	token      string
	teamID     string
	httpClient *http.Client
}

func New(token, teamID string) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("vercel: tokein id required")
	}
	return &Client{
		token:      token,
		teamID:     teamID,
		httpClient: &http.Client{},
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

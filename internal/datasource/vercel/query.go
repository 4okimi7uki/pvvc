package vercel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"time"
)

func (c *Client) FetchBillingCharges(start, end time.Time) (*Report, error) {
	params := url.Values{}
	params.Set("from", start.Format("2006-01-02"))
	params.Set("to", end.Format("2006-01-02"))
	if c.teamID != "" {
		params.Set("teamId", c.teamID)
	}

	req, err := c.newRequest("GET", "/v1/billing/charges?"+params.Encode())
	if err != nil {
		return nil, fmt.Errorf("vercel: failed to build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vercel: request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("vercel: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("vercel: failed to read body: %w", err)
	}

	var charges []BillingCharge
	dec := json.NewDecoder(bytes.NewReader(body))

	for dec.More() {
		var c BillingCharge
		if err := dec.Decode(&c); err != nil {
			return nil, fmt.Errorf("vercel: failed to decode: %w", err)
		}
		charges = append(charges, c)
	}

	return &Report{Charges: charges}, nil
}

package vercel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"time"
)

// JSTの「昨日」をVercel基準（07:00 UTC）に合わせる
func vercelDayRange(date time.Time) time.Time {
	utc := date.UTC()
	// その日の07:00 UTCをstartに
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 7, 0, 0, 0, time.UTC)
}

func (c *Client) FetchBillingCharges(start, end time.Time) (*Report, error) {
	vercelDayStart := vercelDayRange(start)
	vercelDayEnd := vercelDayRange(end)
	params := url.Values{}
	params.Set("from", vercelDayStart.Format(time.RFC3339))
	params.Set("to", vercelDayEnd.Format(time.RFC3339))
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
	fmt.Println(charges)

	return &Report{Charges: charges}, nil
}

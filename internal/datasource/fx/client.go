package fx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

func FetchUSDToJPY(start time.Time, end time.Time) (map[string]decimal.Decimal, error) {
	url := fmt.Sprintf("https://api.frankfurter.dev/v2/rates?base=USD&quotes=JPY&from=%s&to=%s", start.Format("2006-01-02"), end.Format("2006-01-02"))
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, fmt.Errorf("fx: network response is error: %w", err)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fx: unexpected status %d", resp.StatusCode)
	}

	var response []struct {
		Date  string      `json:"date"`
		Base  string      `json:"base"`
		Quote string      `json:"quote"`
		Rate  json.Number `json:"rate"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("fx: failed to decode response: %w", err)
	}

	results := make(map[string]decimal.Decimal)

	for _, r := range response {
		date, err := time.Parse("2006-01-02", r.Date)
		if err != nil {
			return nil, err
		}
		key := date.Format("20060102")
		results[key], _ = decimal.NewFromString(r.Rate.String())
	}

	return results, nil
}

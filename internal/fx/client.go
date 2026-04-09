package fx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func FetchUSDToJPY(date time.Time) (float64, error) {
	url := fmt.Sprintf("https://api.frankfurter.dev/v2/rates?base=USD&quotes=JPY&date=%s", date.Format("2006-01-02"))
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return 0, fmt.Errorf("fx: network response is error: %w", err)
	}

	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	var result []struct {
		Date  string  `json:"date"`
		Base  string  `json:"base"`
		Quote string  `json:"quote"`
		Rate  float64 `json:"rate"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("fx: failed to decode response: %w", err)
	}

	return result[0].Rate, nil
}

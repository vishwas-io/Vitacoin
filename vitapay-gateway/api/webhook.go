package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// FireWebhook POSTs a JSON payload to the given URL with retry/backoff.
func FireWebhook(url string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook marshal error: %w", err)
	}

	const maxRetries = 3
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt)) * time.Second // 2s, 4s
			log.Printf("[webhook] retry %d/%d in %v for %s", attempt+1, maxRetries, backoff, url)
			time.Sleep(backoff)
		}

		resp, err := http.Post(url, "application/json", bytes.NewReader(data)) //nolint:gosec
		if err != nil {
			lastErr = fmt.Errorf("webhook POST error: %w", err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			log.Printf("[webhook] delivered to %s (status=%d)", url, resp.StatusCode)
			return nil
		}
		lastErr = fmt.Errorf("webhook non-2xx response: %d", resp.StatusCode)
	}

	log.Printf("[webhook] FAILED after %d attempts for %s: %v", maxRetries, url, lastErr)
	return lastErr
}

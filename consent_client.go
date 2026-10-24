package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type InfraiError struct {
	Status int
	Detail string
}

func (e *InfraiError) Error() string {
	return fmt.Sprintf("infrai response %d: %s", e.Status, e.Detail)
}

type ConsentClient struct {
	baseURL string
	key     string
	http    *http.Client
}

const consentCheckCapability = "auth.consent.check"

func NewConsentClient(baseURL, key string) *ConsentClient {
	return &ConsentClient{baseURL: baseURL, key: key, http: &http.Client{Timeout: 10 * time.Second}}
}

func (c *ConsentClient) Check(ctx context.Context, userID, category string) (bool, error) {
	var data struct {
		Granted bool `json:"granted"`
	}
	path := "/v1/auth/consent/check/" + userID + "/" + category
	if err := c.call(ctx, http.MethodGet, path, nil, &data); err != nil {
		return false, err
	}
	return data.Granted, nil
}

func (c *ConsentClient) call(ctx context.Context, method, path string, body io.Reader, out any) error {
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Accept", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		var env envelope
		decodeErr := json.NewDecoder(res.Body).Decode(&env)
		res.Body.Close()
		if decodeErr != nil {
			return decodeErr
		}
		if !env.OK {
			if res.StatusCode == http.StatusTooManyRequests && attempt < 2 {
				delay := time.Duration(1<<attempt) * 100 * time.Millisecond
				if retry := res.Header.Get("Retry-After"); retry != "" {
					delay = time.Second
				}
				timer := time.NewTimer(delay)
				select {
				case <-ctx.Done():
					timer.Stop()
					return ctx.Err()
				case <-timer.C:
				}
				continue
			}
			return &InfraiError{Status: res.StatusCode, Detail: string(env.Error)}
		}
		if res.StatusCode >= 500 {
			return &InfraiError{Status: res.StatusCode, Detail: "upstream service response"}
		}
		if out != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return &InfraiError{Status: http.StatusTooManyRequests, Detail: "retry limit reached"}
}

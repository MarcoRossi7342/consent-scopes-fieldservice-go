package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    *APIError       `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string { return e.Code + ": " + e.Message }

type ConsentClient struct {
	baseURL string
	key     string
	http    *http.Client
	sleep   func(time.Duration)
}

func NewConsentClient(baseURL string) (*ConsentClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, errors.New("INFRAI_API_KEY is required")
	}
	return &ConsentClient{baseURL: strings.TrimRight(baseURL, "/"), key: key, http: &http.Client{Timeout: 10 * time.Second}, sleep: time.Sleep}, nil
}

func (c *ConsentClient) request(ctx context.Context, method, path string, out any) error {
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Accept", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		body, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env Envelope
		if err := json.Unmarshal(body, &env); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
		if !env.OK {
			if env.Error == nil {
				return fmt.Errorf("request rejected with HTTP %d", res.StatusCode)
			}
			return env.Error
		}
		if res.StatusCode == http.StatusTooManyRequests {
			delay := time.Duration(1<<attempt) * 200 * time.Millisecond
			if seconds, err := strconv.Atoi(res.Header.Get("Retry-After")); err == nil && seconds > 0 {
				delay = time.Duration(seconds) * time.Second
			}
			c.sleep(delay)
			continue
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("upstream HTTP %d", res.StatusCode)
		}
		if out != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return errors.New("rate limit retry budget exhausted")
}

type ConsentStatus struct {
	Result bool `json:"result"`
}

// infrai.auth.consent.check is the capability represented by this boundary.
func (c *ConsentClient) Check(ctx context.Context, userID, category string) (ConsentStatus, error) {
	var status ConsentStatus
	err := c.request(ctx, http.MethodGet, "/v1/auth/consent/check/"+userID+"/"+category, &status)
	return status, err
}

func (c *ConsentClient) Revoke(ctx context.Context, userID string) error {
	return c.request(ctx, http.MethodPost, "/v1/auth/consent/revoke/"+userID, nil)
}

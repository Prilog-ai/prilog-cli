package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (c *cli) doJSON(ctx context.Context, method, path string, payload any, auth bool, target any) error {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(encoded)
	}
	return c.do(ctx, method, path, body, "application/json", auth, target)
}

func (c *cli) doRaw(ctx context.Context, method, path string, payload []byte, auth bool, target any) error {
	return c.do(ctx, method, path, bytes.NewReader(payload), "application/octet-stream", auth, target)
}

func (c *cli) do(ctx context.Context, method, path string, body io.Reader, contentType string, auth bool, target any) error {
	req, err := http.NewRequestWithContext(ctx, method, joinURL(c.apiURL, path), body)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Prilog-Client", "cli")
	if body != nil && contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if auth {
		global, _ := loadGlobalConfig()
		if global.AccessToken == "" {
			return errors.New("not authenticated; run `prilog login`")
		}
		req.Header.Set("Authorization", "Bearer "+global.AccessToken)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return decodeAPIError(resp.StatusCode, respBody)
	}
	if target == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, target); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func decodeAPIError(statusCode int, body []byte) error {
	message := strings.TrimSpace(string(body))
	var payload map[string]any
	if json.Unmarshal(body, &payload) == nil {
		if errValue, ok := payload["error"].(string); ok {
			message = errValue
		}
	}
	return apiError{StatusCode: statusCode, Message: message}
}

type apiError struct {
	StatusCode int
	Message    string
}

func (e apiError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("api returned HTTP %d", e.StatusCode)
	}
	return fmt.Sprintf("api returned HTTP %d: %s", e.StatusCode, e.Message)
}

func joinURL(base, path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return strings.TrimRight(base, "/") + path
}

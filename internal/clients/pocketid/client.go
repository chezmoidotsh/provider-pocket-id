/*
Copyright 2025 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pocketid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

const (
	DefaultTimeout = 30 * time.Second
)

// Config holds the configuration for Pocket ID client
type Config struct {
	Endpoint string
	APIKey   string
	Timeout  time.Duration
}

// Client is the Pocket ID API client
type Client struct {
	config     Config
	httpClient *http.Client
}

// NewClient creates a new Pocket ID API client
func NewClient(config Config) *Client {
	if config.Timeout == 0 {
		config.Timeout = DefaultTimeout
	}

	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// NewClientFromCredentials creates a new client from credential data
func NewClientFromCredentials(endpoint string, apiKey string) (*Client, error) {
	var config Config
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint is required")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("apiKey is required in credentials")
	}

	// Ensure Endpoint doesn't end with /
	config.Endpoint = strings.TrimSuffix(endpoint, "/")
	config.APIKey = apiKey

	return NewClient(config), nil
}

// makeRequest performs HTTP request with proper authentication
func (c *Client) makeRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.config.Endpoint+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-KEY", c.config.APIKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

// uploadFile uploads a file to the specified path
func (c *Client) uploadFile(ctx context.Context, path string, fileData []byte, filename string) (*http.Response, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := part.Write(fileData); err != nil {
		return nil, fmt.Errorf("failed to write file data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", c.config.Endpoint+path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload request: %w", err)
	}

	req.Header.Set("X-API-KEY", c.config.APIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return c.httpClient.Do(req)
}

// downloadFile downloads a file from the given URL
func (c *Client) downloadFile(ctx context.Context, fileURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fileURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download file from %s: %w", fileURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to download file: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Extract filename from URL
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return data, "logo", nil
	}
	filename := filepath.Base(parsedURL.Path)
	if filename == "." || filename == "/" {
		filename = "logo"
	}

	return data, filename, nil
}

// checkResponse checks HTTP response for errors and returns body
func checkResponse(resp *http.Response) ([]byte, error) {
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: HTTP %d - %s", resp.StatusCode, string(body))
	}

	return body, nil
}

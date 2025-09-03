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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// OIDCClient represents an OIDC client in Pocket ID API
type OIDCClient struct {
	ID              string            `json:"id,omitempty"`
	ClientName      string            `json:"clientName"`
	ClientSecret    string            `json:"clientSecret,omitempty"`
	RedirectURIs    []string          `json:"redirectUris"`
	PostLogoutURIs  []string          `json:"postLogoutUris,omitempty"`
	LaunchURL       string            `json:"launchURL,omitempty"`
	IsPublic        bool              `json:"isPublic,omitempty"`
	RequirePKCE     bool              `json:"requirePKCE,omitempty"`
	HasLogo         bool              `json:"hasLogo,omitempty"`
	GroupClaims     []string          `json:"groupClaims,omitempty"`
	CustomClaims    map[string]string `json:"customClaims,omitempty"`
	AllowedScopes   []string          `json:"allowedScopes,omitempty"`
	AccessTokenTTL  int               `json:"accessTokenTTL,omitempty"`
	RefreshTokenTTL int               `json:"refreshTokenTTL,omitempty"`
	IDTokenTTL      int               `json:"idTokenTTL,omitempty"`
	GroupNames      []string          `json:"groupNames,omitempty"`
}

// CreateOIDCClientRequest represents the request payload for creating an OIDC client
type CreateOIDCClientRequest struct {
	ClientName      string            `json:"clientName"`
	RedirectURIs    []string          `json:"redirectUris"`
	PostLogoutURIs  []string          `json:"postLogoutUris,omitempty"`
	LaunchURL       string            `json:"launchURL,omitempty"`
	IsPublic        bool              `json:"isPublic,omitempty"`
	RequirePKCE     bool              `json:"requirePKCE,omitempty"`
	GroupClaims     []string          `json:"groupClaims,omitempty"`
	CustomClaims    map[string]string `json:"customClaims,omitempty"`
	AllowedScopes   []string          `json:"allowedScopes,omitempty"`
	AccessTokenTTL  int               `json:"accessTokenTTL,omitempty"`
	RefreshTokenTTL int               `json:"refreshTokenTTL,omitempty"`
	IDTokenTTL      int               `json:"idTokenTTL,omitempty"`
}

// UpdateOIDCClientRequest represents the request payload for updating an OIDC client
type UpdateOIDCClientRequest struct {
	ClientName      string            `json:"clientName"`
	RedirectURIs    []string          `json:"redirectUris"`
	PostLogoutURIs  []string          `json:"postLogoutUris,omitempty"`
	LaunchURL       string            `json:"launchURL,omitempty"`
	IsPublic        bool              `json:"isPublic,omitempty"`
	RequirePKCE     bool              `json:"requirePKCE,omitempty"`
	GroupClaims     []string          `json:"groupClaims,omitempty"`
	CustomClaims    map[string]string `json:"customClaims,omitempty"`
	AllowedScopes   []string          `json:"allowedScopes,omitempty"`
	AccessTokenTTL  int               `json:"accessTokenTTL,omitempty"`
	RefreshTokenTTL int               `json:"refreshTokenTTL,omitempty"`
	IDTokenTTL      int               `json:"idTokenTTL,omitempty"`
}

// GetOIDCClient retrieves an OIDC client by ID
func (c *Client) GetOIDCClient(ctx context.Context, clientID string) (*OIDCClient, error) {
	resp, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/api/oidc/clients/%s", clientID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get OIDC client: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // Client doesn't exist
	}

	body, err := checkResponse(resp)
	if err != nil {
		return nil, err
	}

	var client OIDCClient
	if err := json.Unmarshal(body, &client); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OIDC client response: %w", err)
	}

	return &client, nil
}

// GetOIDCClientByExternalName retrieves an OIDC client by client name (external name)
func (c *Client) GetOIDCClientByExternalName(ctx context.Context, clientName string) (*OIDCClient, error) {
	clients, err := c.ListOIDCClients(ctx)
	if err != nil {
		return nil, err
	}

	for _, client := range clients {
		if client.ClientName == clientName {
			return &client, nil
		}
	}

	return nil, nil // Client not found
}

// ListOIDCClients retrieves all OIDC clients
func (c *Client) ListOIDCClients(ctx context.Context) ([]OIDCClient, error) {
	resp, err := c.makeRequest(ctx, "GET", "/api/oidc/clients", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list OIDC clients: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, err
	}

	var clients []OIDCClient
	if err := json.Unmarshal(body, &clients); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OIDC clients response: %w", err)
	}

	return clients, nil
}

// CreateOIDCClient creates a new OIDC client
func (c *Client) CreateOIDCClient(ctx context.Context, req CreateOIDCClientRequest) (*OIDCClient, error) {
	resp, err := c.makeRequest(ctx, "POST", "/api/oidc/clients", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC client: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, err
	}

	var client OIDCClient
	if err := json.Unmarshal(body, &client); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OIDC client response: %w", err)
	}

	return &client, nil
}

// UpdateOIDCClient updates an existing OIDC client
func (c *Client) UpdateOIDCClient(ctx context.Context, clientID string, req UpdateOIDCClientRequest) (*OIDCClient, error) {
	resp, err := c.makeRequest(ctx, "PUT", fmt.Sprintf("/api/oidc/clients/%s", clientID), req)
	if err != nil {
		return nil, fmt.Errorf("failed to update OIDC client: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, err
	}

	var client OIDCClient
	if err := json.Unmarshal(body, &client); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OIDC client response: %w", err)
	}

	return &client, nil
}

// DeleteOIDCClient deletes an OIDC client by ID
func (c *Client) DeleteOIDCClient(ctx context.Context, clientID string) error {
	resp, err := c.makeRequest(ctx, "DELETE", fmt.Sprintf("/api/oidc/clients/%s", clientID), nil)
	if err != nil {
		return fmt.Errorf("failed to delete OIDC client: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil // Already deleted
	}

	_, err = checkResponse(resp)
	return err
}

// UploadOIDCClientLogo uploads a logo for an OIDC client from a URL
func (c *Client) UploadOIDCClientLogo(ctx context.Context, clientID, logoURL string) error {
	if logoURL == "" {
		return nil
	}

	// Download the logo from the URL
	logoData, filename, err := c.downloadFile(ctx, logoURL)
	if err != nil {
		return fmt.Errorf("failed to download logo: %w", err)
	}

	// Validate file size (2MB limit)
	if len(logoData) > 2*1024*1024 {
		return fmt.Errorf("logo file size exceeds 2MB limit")
	}

	// Validate file type based on URL extension
	if !isValidImageExtension(logoURL) {
		return fmt.Errorf("invalid image format. Supported formats: PNG, JPEG, JPG, GIF, SVG")
	}

	// Upload the logo
	resp, err := c.uploadFile(ctx, fmt.Sprintf("/api/oidc/clients/%s/logo", clientID), logoData, filename)
	if err != nil {
		return fmt.Errorf("failed to upload logo: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	_, err = checkResponse(resp)
	return err
}

// isValidImageExtension checks if the URL has a valid image extension
func isValidImageExtension(url string) bool {
	url = strings.ToLower(url)
	validExtensions := []string{".png", ".jpeg", ".jpg", ".gif", ".svg"}

	for _, ext := range validExtensions {
		if strings.HasSuffix(url, ext) {
			return true
		}
	}

	return false
}

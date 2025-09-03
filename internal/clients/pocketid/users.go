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
)

// User represents a user in Pocket ID API
type User struct {
	ID           string            `json:"id,omitempty"`
	Username     string            `json:"username"`
	Email        string            `json:"email"`
	FirstName    string            `json:"firstName"`
	LastName     string            `json:"lastName,omitempty"`
	Locale       string            `json:"locale,omitempty"`
	Disabled     bool              `json:"disabled,omitempty"`
	IsAdmin      bool              `json:"isAdmin,omitempty"`
	UserGroups   []string          `json:"userGroups,omitempty"`
	CustomClaims map[string]string `json:"customClaims,omitempty"`
}

// CreateUserRequest represents the request payload for creating a user
type CreateUserRequest struct {
	Username     string            `json:"username"`
	Email        string            `json:"email"`
	FirstName    string            `json:"firstName"`
	LastName     string            `json:"lastName,omitempty"`
	Locale       string            `json:"locale,omitempty"`
	Disabled     bool              `json:"disabled,omitempty"`
	IsAdmin      bool              `json:"isAdmin,omitempty"`
	CustomClaims map[string]string `json:"customClaims,omitempty"`
}

// UpdateUserRequest represents the request payload for updating a user
type UpdateUserRequest struct {
	Username     string            `json:"username"`
	Email        string            `json:"email"`
	FirstName    string            `json:"firstName"`
	LastName     string            `json:"lastName,omitempty"`
	Locale       string            `json:"locale,omitempty"`
	Disabled     bool              `json:"disabled,omitempty"`
	CustomClaims map[string]string `json:"customClaims,omitempty"`
}

// GetUser retrieves a user by ID
func (c *Client) GetUser(ctx context.Context, userID string) (*User, error) {
	resp, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/api/users/%s", userID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // User doesn't exist
	}

	body, err := checkResponse(resp)
	if err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user response: %w", err)
	}

	return &user, nil
}

// GetUserByExternalName retrieves a user by username (external name)
func (c *Client) GetUserByExternalName(ctx context.Context, username string) (*User, error) {
	users, err := c.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Username == username {
			return &user, nil
		}
	}

	return nil, nil // User not found
}

// ListUsers retrieves all users
func (c *Client) ListUsers(ctx context.Context) ([]User, error) {
	resp, err := c.makeRequest(ctx, "GET", "/api/users", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, err
	}

	var users []User
	if err := json.Unmarshal(body, &users); err != nil {
		return nil, fmt.Errorf("failed to unmarshal users response: %w", err)
	}

	return users, nil
}

// CreateUser creates a new user
func (c *Client) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
	resp, err := c.makeRequest(ctx, "POST", "/api/users", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user response: %w", err)
	}

	return &user, nil
}

// UpdateUser updates an existing user
func (c *Client) UpdateUser(ctx context.Context, userID string, req UpdateUserRequest) (*User, error) {
	resp, err := c.makeRequest(ctx, "PUT", fmt.Sprintf("/api/users/%s", userID), req)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user response: %w", err)
	}

	return &user, nil
}

// DeleteUser deletes a user by ID
func (c *Client) DeleteUser(ctx context.Context, userID string) error {
	resp, err := c.makeRequest(ctx, "DELETE", fmt.Sprintf("/api/users/%s", userID), nil)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil // Already deleted
	}

	_, err = checkResponse(resp)
	return err
}

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

// Group represents a group in Pocket ID API
type Group struct {
	ID           string            `json:"id,omitempty"`
	GroupName    string            `json:"groupName"`
	FriendlyName string            `json:"friendlyName,omitempty"`
	CustomClaims map[string]string `json:"customClaims,omitempty"`
}

// CreateGroupRequest represents the request payload for creating a group
type CreateGroupRequest struct {
	GroupName    string            `json:"groupName"`
	FriendlyName string            `json:"friendlyName,omitempty"`
	CustomClaims map[string]string `json:"customClaims,omitempty"`
}

// UpdateGroupRequest represents the request payload for updating a group
type UpdateGroupRequest struct {
	GroupName    string            `json:"groupName"`
	FriendlyName string            `json:"friendlyName,omitempty"`
	CustomClaims map[string]string `json:"customClaims,omitempty"`
}

// GetGroup retrieves a group by ID
func (c *Client) GetGroup(ctx context.Context, groupID string) (*Group, error) {
	resp, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/api/groups/%s", groupID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // Group doesn't exist
	}

	body, err := checkResponse(resp)
	if err != nil {
		return nil, err
	}

	var group Group
	if err := json.Unmarshal(body, &group); err != nil {
		return nil, fmt.Errorf("failed to unmarshal group response: %w", err)
	}

	return &group, nil
}

// GetGroupByExternalName retrieves a group by group name (external name)
func (c *Client) GetGroupByExternalName(ctx context.Context, groupName string) (*Group, error) {
	groups, err := c.ListGroups(ctx)
	if err != nil {
		return nil, err
	}

	for _, group := range groups {
		if group.GroupName == groupName {
			return &group, nil
		}
	}

	return nil, nil // Group not found
}

// ListGroups retrieves all groups
func (c *Client) ListGroups(ctx context.Context) ([]Group, error) {
	resp, err := c.makeRequest(ctx, "GET", "/api/groups", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, err
	}

	var groups []Group
	if err := json.Unmarshal(body, &groups); err != nil {
		return nil, fmt.Errorf("failed to unmarshal groups response: %w", err)
	}

	return groups, nil
}

// CreateGroup creates a new group
func (c *Client) CreateGroup(ctx context.Context, req CreateGroupRequest) (*Group, error) {
	resp, err := c.makeRequest(ctx, "POST", "/api/groups", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, err
	}

	var group Group
	if err := json.Unmarshal(body, &group); err != nil {
		return nil, fmt.Errorf("failed to unmarshal group response: %w", err)
	}

	return &group, nil
}

// UpdateGroup updates an existing group
func (c *Client) UpdateGroup(ctx context.Context, groupID string, req UpdateGroupRequest) (*Group, error) {
	resp, err := c.makeRequest(ctx, "PUT", fmt.Sprintf("/api/groups/%s", groupID), req)
	if err != nil {
		return nil, fmt.Errorf("failed to update group: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, err
	}

	var group Group
	if err := json.Unmarshal(body, &group); err != nil {
		return nil, fmt.Errorf("failed to unmarshal group response: %w", err)
	}

	return &group, nil
}

// DeleteGroup deletes a group by ID
func (c *Client) DeleteGroup(ctx context.Context, groupID string) error {
	resp, err := c.makeRequest(ctx, "DELETE", fmt.Sprintf("/api/groups/%s", groupID), nil)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil // Already deleted
	}

	_, err = checkResponse(resp)
	return err
}

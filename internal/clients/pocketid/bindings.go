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
	"fmt"
	"net/http"
	"slices"
)

// AddUserToGroup adds a user to a group
func (c *Client) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	resp, err := c.makeRequest(ctx, "POST", fmt.Sprintf("/api/users/%s/groups/%s", userID, groupID), nil)
	if err != nil {
		return fmt.Errorf("failed to add user to group: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	_, err = checkResponse(resp)
	return err
}

// RemoveUserFromGroup removes a user from a group
func (c *Client) RemoveUserFromGroup(ctx context.Context, userID, groupID string) error {
	resp, err := c.makeRequest(ctx, "DELETE", fmt.Sprintf("/api/users/%s/groups/%s", userID, groupID), nil)
	if err != nil {
		return fmt.Errorf("failed to remove user from group: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil // Binding doesn't exist, which is fine
	}

	_, err = checkResponse(resp)
	return err
}

// IsUserInGroup checks if a user is in a group
func (c *Client) IsUserInGroup(ctx context.Context, userID, groupID string) (bool, error) {
	user, err := c.GetUser(ctx, userID)
	if err != nil {
		return false, err
	}

	if user == nil {
		return false, nil
	}

	// Get group name from ID
	group, err := c.GetGroup(ctx, groupID)
	if err != nil {
		return false, err
	}

	if group == nil {
		return false, nil
	}

	// Check if group name is in user's groups
	if slices.Contains(user.UserGroups, group.GroupName) {
		return true, nil
	}

	return false, nil
}

// AddClientToGroup adds an OIDC client to a group
func (c *Client) AddClientToGroup(ctx context.Context, clientID, groupID string) error {
	resp, err := c.makeRequest(ctx, "POST", fmt.Sprintf("/api/oidc/clients/%s/groups/%s", clientID, groupID), nil)
	if err != nil {
		return fmt.Errorf("failed to add client to group: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	_, err = checkResponse(resp)
	return err
}

// RemoveClientFromGroup removes an OIDC client from a group
func (c *Client) RemoveClientFromGroup(ctx context.Context, clientID, groupID string) error {
	resp, err := c.makeRequest(ctx, "DELETE", fmt.Sprintf("/api/oidc/clients/%s/groups/%s", clientID, groupID), nil)
	if err != nil {
		return fmt.Errorf("failed to remove client from group: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil // Binding doesn't exist, which is fine
	}

	_, err = checkResponse(resp)
	return err
}

// IsClientInGroup checks if an OIDC client is in a group
func (c *Client) IsClientInGroup(ctx context.Context, clientID, groupID string) (bool, error) {
	client, err := c.GetOIDCClient(ctx, clientID)
	if err != nil {
		return false, err
	}

	if client == nil {
		return false, nil
	}

	// Get group name from ID
	group, err := c.GetGroup(ctx, groupID)
	if err != nil {
		return false, err
	}

	if group == nil {
		return false, nil
	}

	// Check if group name is in client's groups
	if slices.Contains(client.GroupNames, group.GroupName) {
		return true, nil
	}

	return false, nil
}

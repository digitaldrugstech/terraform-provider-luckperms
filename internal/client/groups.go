package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// GetGroups returns all group names.
func (c *Client) GetGroups(ctx context.Context) ([]string, error) {
	body, err := c.doRequest(ctx, "GET", "/group", nil)
	if err != nil {
		return nil, fmt.Errorf("listing groups: %w", err)
	}

	var names []string
	if err := json.Unmarshal(body, &names); err != nil {
		return nil, fmt.Errorf("decoding group list: %w", err)
	}
	return names, nil
}

// GetGroup returns a single group by name.
func (c *Client) GetGroup(ctx context.Context, name string) (*Group, error) {
	body, err := c.doRequest(ctx, "GET", "/group/"+url.PathEscape(name), nil)
	if err != nil {
		return nil, err
	}

	var group Group
	if err := json.Unmarshal(body, &group); err != nil {
		return nil, fmt.Errorf("decoding group: %w", err)
	}
	return &group, nil
}

// CreateGroup creates a new group.
func (c *Client) CreateGroup(ctx context.Context, name string) (*Group, error) {
	body, err := c.doRequest(ctx, "POST", "/group", createGroupRequest{Name: name})
	if err != nil {
		return nil, fmt.Errorf("creating group %q: %w", name, err)
	}

	var group Group
	if err := json.Unmarshal(body, &group); err != nil {
		return nil, fmt.Errorf("decoding created group: %w", err)
	}
	return &group, nil
}

// DeleteGroup deletes a group by name.
func (c *Client) DeleteGroup(ctx context.Context, name string) error {
	return c.doRequestNoBody(ctx, "DELETE", "/group/"+url.PathEscape(name), nil)
}

// GetGroupNodes returns all nodes for a group.
func (c *Client) GetGroupNodes(ctx context.Context, groupName string) ([]Node, error) {
	body, err := c.doRequest(ctx, "GET", "/group/"+url.PathEscape(groupName)+"/nodes", nil)
	if err != nil {
		return nil, err
	}

	var nodes []Node
	if err := json.Unmarshal(body, &nodes); err != nil {
		return nil, fmt.Errorf("decoding group nodes: %w", err)
	}
	return nodes, nil
}

// SetGroupNodes replaces all nodes for a group (PUT — full replacement).
// nil is coerced to empty slice to send [] instead of null.
func (c *Client) SetGroupNodes(ctx context.Context, groupName string, nodes []Node) error {
	if nodes == nil {
		nodes = []Node{}
	}
	// Strip the type field — the API infers it from the key format.
	cleanNodes := make([]Node, len(nodes))
	for i, n := range nodes {
		cleanNodes[i] = Node{
			Key:     n.Key,
			Value:   n.Value,
			Context: n.Context,
			Expiry:  n.Expiry,
		}
	}

	return c.doRequestNoBody(ctx, "PUT", "/group/"+url.PathEscape(groupName)+"/nodes", cleanNodes)
}

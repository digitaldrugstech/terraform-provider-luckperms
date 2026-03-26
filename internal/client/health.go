package client

import (
	"context"
	"fmt"
)

// Health checks connectivity to the LuckPerms REST API by listing groups.
// The LuckPerms REST API has no dedicated health endpoint.
func (c *Client) Health(ctx context.Context) error {
	_, err := c.doRequest(ctx, "GET", "/group", nil)
	if err != nil {
		return fmt.Errorf("health check failed (GET /group): %w", err)
	}
	return nil
}

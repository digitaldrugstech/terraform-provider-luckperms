package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetTracks returns all track names.
func (c *Client) GetTracks(ctx context.Context) ([]string, error) {
	body, err := c.doRequest(ctx, "GET", "/track", nil)
	if err != nil {
		return nil, fmt.Errorf("listing tracks: %w", err)
	}

	var names []string
	if err := json.Unmarshal(body, &names); err != nil {
		return nil, fmt.Errorf("decoding track list: %w", err)
	}
	return names, nil
}

// GetTrack returns a single track by name.
func (c *Client) GetTrack(ctx context.Context, name string) (*Track, error) {
	body, err := c.doRequest(ctx, "GET", "/track/"+name, nil)
	if err != nil {
		return nil, err
	}

	var track Track
	if err := json.Unmarshal(body, &track); err != nil {
		return nil, fmt.Errorf("decoding track: %w", err)
	}
	return &track, nil
}

// CreateTrack creates a new track and sets its groups.
// Track creation is non-atomic: POST creates the track, PATCH sets groups.
// If PATCH fails, the provider attempts cleanup by deleting the orphaned track.
func (c *Client) CreateTrack(ctx context.Context, name string, groups []string) (*Track, error) {
	_, err := c.doRequest(ctx, "POST", "/track", createTrackRequest{Name: name})
	if err != nil {
		return nil, fmt.Errorf("creating track %q: %w", name, err)
	}

	if len(groups) > 0 {
		err = c.UpdateTrack(ctx, name, groups)
		if err != nil {
			cleanupErr := c.DeleteTrack(ctx, name)
			if cleanupErr != nil {
				return nil, fmt.Errorf("setting groups for track %q: %w (cleanup also failed: %v)", name, err, cleanupErr)
			}
			return nil, fmt.Errorf("setting groups for track %q (orphaned track was cleaned up): %w", name, err)
		}
	}

	return c.GetTrack(ctx, name)
}

// UpdateTrack updates the groups of an existing track.
func (c *Client) UpdateTrack(ctx context.Context, name string, groups []string) error {
	return c.doRequestNoBody(ctx, "PATCH", "/track/"+name, updateTrackRequest{Groups: groups})
}

// DeleteTrack deletes a track by name.
func (c *Client) DeleteTrack(ctx context.Context, name string) error {
	return c.doRequestNoBody(ctx, "DELETE", "/track/"+name, nil)
}

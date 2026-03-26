package client

// Group represents a LuckPerms group from the REST API.
type Group struct {
	Name        string         `json:"name"`
	DisplayName *string        `json:"displayName"`
	Weight      int            `json:"weight"`
	Nodes       []Node         `json:"nodes"`
	Metadata    *GroupMetadata `json:"metadata,omitempty"`
}

// Node represents a LuckPerms permission node.
type Node struct {
	Key     string    `json:"key"`
	Type    string    `json:"type,omitempty"`
	Value   bool      `json:"value"`
	Context []Context `json:"context"`
	Expiry  *int64    `json:"expiry,omitempty"`
}

// Context represents a LuckPerms context entry (e.g., server=creative-build).
type Context struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Track represents a LuckPerms track.
type Track struct {
	Name   string   `json:"name"`
	Groups []string `json:"groups"`
}

// GroupMetadata represents group metadata returned by the API.
type GroupMetadata struct {
	Meta   map[string]string `json:"meta,omitempty"`
	Prefix string            `json:"prefix,omitempty"`
	Suffix string            `json:"suffix,omitempty"`
}

// Request types (not exported — used only within client package).

type createGroupRequest struct {
	Name string `json:"name"`
}

type createTrackRequest struct {
	Name string `json:"name"`
}

type updateTrackRequest struct {
	Groups []string `json:"groups"`
}

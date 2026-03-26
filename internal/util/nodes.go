package util

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/digitaldrugstech/terraform-provider-luckperms/internal/client"
)

// metaPrefixes are node key prefixes owned by luckperms_group resource.
var metaPrefixes = []string{"displayname.", "weight.", "prefix.", "suffix."}

// IsMetaNode returns true if the node key is a meta node
// (displayname, weight, prefix, suffix) owned by the group resource.
func IsMetaNode(key string) bool {
	for _, p := range metaPrefixes {
		if strings.HasPrefix(key, p) {
			return true
		}
	}
	return false
}

// SplitNodes separates nodes into meta nodes (owned by group resource)
// and permission nodes (owned by group_nodes resource).
func SplitNodes(nodes []client.Node) (meta, perms []client.Node) {
	for _, n := range nodes {
		if IsMetaNode(n.Key) {
			meta = append(meta, n)
		} else {
			perms = append(perms, n)
		}
	}
	return
}

// MetaAttrs holds parsed meta-node attributes for a group.
type MetaAttrs struct {
	DisplayName    string
	Weight         int64
	Prefix         string
	Suffix         string
	HasDisplayName bool
	HasWeight      bool
	HasPrefix      bool
	HasSuffix      bool
}

// ParseMetaNodes extracts meta attributes from a set of nodes.
// Takes the first occurrence of each meta type.
func ParseMetaNodes(nodes []client.Node) MetaAttrs {
	var attrs MetaAttrs
	for _, n := range nodes {
		switch {
		case strings.HasPrefix(n.Key, "displayname.") && !attrs.HasDisplayName:
			attrs.DisplayName = unescapeDots(strings.TrimPrefix(n.Key, "displayname."))
			attrs.HasDisplayName = true
		case strings.HasPrefix(n.Key, "weight.") && !attrs.HasWeight:
			if w, err := strconv.ParseInt(strings.TrimPrefix(n.Key, "weight."), 10, 64); err == nil {
				attrs.Weight = w
				attrs.HasWeight = true
			}
		case strings.HasPrefix(n.Key, "prefix.") && !attrs.HasPrefix:
			attrs.Prefix = unescapeDots(strings.TrimPrefix(n.Key, "prefix."))
			attrs.HasPrefix = true
		case strings.HasPrefix(n.Key, "suffix.") && !attrs.HasSuffix:
			attrs.Suffix = unescapeDots(strings.TrimPrefix(n.Key, "suffix."))
			attrs.HasSuffix = true
		}
	}
	return attrs
}

// BuildMetaNodes creates API nodes from group resource attributes.
// displayName and prefix/suffix are nil when not set.
func BuildMetaNodes(displayName *string, weight int64, prefix *string, suffix *string) []client.Node {
	var nodes []client.Node

	if displayName != nil {
		nodes = append(nodes, client.Node{
			Key:   "displayname." + *displayName,
			Value: true,
		})
	}

	// Always include weight node
	nodes = append(nodes, client.Node{
		Key:   fmt.Sprintf("weight.%d", weight),
		Value: true,
	})

	if prefix != nil {
		nodes = append(nodes, client.Node{
			Key:   "prefix." + *prefix,
			Value: true,
		})
	}

	if suffix != nil {
		nodes = append(nodes, client.Node{
			Key:   "suffix." + *suffix,
			Value: true,
		})
	}

	return nodes
}

// unescapeDots removes backslash-escaped dots that LuckPerms REST API adds.
func unescapeDots(s string) string {
	return strings.ReplaceAll(s, "\\.", ".")
}

// MergeNodes combines meta nodes and permission nodes into a single set.
func MergeNodes(meta, perms []client.Node) []client.Node {
	result := make([]client.Node, 0, len(meta)+len(perms))
	result = append(result, meta...)
	result = append(result, perms...)
	return result
}

package util

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/digitaldrugstech/terraform-provider-luckperms/internal/client"
)

// NormalizeNodes sorts nodes for deterministic comparison.
// Sort order: key ASC, value DESC (true first), context serialized ASC, expiry ASC.
func NormalizeNodes(nodes []client.Node) []client.Node {
	result := make([]client.Node, len(nodes))
	for i, n := range nodes {
		result[i] = client.Node{
			Key:     n.Key,
			Value:   n.Value,
			Context: NormalizeContexts(n.Context),
			Expiry:  n.Expiry,
			// Type is intentionally omitted — not needed for comparison
		}
	}

	// Pre-compute serialized contexts to avoid repeated allocation in sort comparator
	ctxStr := make([]string, len(result))
	for i := range result {
		ctxStr[i] = serializeContexts(result[i].Context)
	}

	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Key != result[j].Key {
			return result[i].Key < result[j].Key
		}
		if result[i].Value != result[j].Value {
			return result[i].Value
		}
		if ctxStr[i] != ctxStr[j] {
			return ctxStr[i] < ctxStr[j]
		}
		return expirySortKey(result[i].Expiry) < expirySortKey(result[j].Expiry)
	})

	return result
}

// NormalizeContexts sorts context entries by key then value.
func NormalizeContexts(contexts []client.Context) []client.Context {
	if len(contexts) == 0 {
		return nil
	}
	result := make([]client.Context, len(contexts))
	copy(result, contexts)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Key != result[j].Key {
			return result[i].Key < result[j].Key
		}
		return result[i].Value < result[j].Value
	})
	return result
}

// NodesEqual compares two node sets after normalization.
func NodesEqual(a, b []client.Node) bool {
	na := NormalizeNodes(a)
	nb := NormalizeNodes(b)
	if len(na) != len(nb) {
		return false
	}
	for i := range na {
		if !nodeEqual(na[i], nb[i]) {
			return false
		}
	}
	return true
}

// FindDuplicateNodes returns the key of the first duplicate node found (same key + value + context + expiry).
// Returns empty string if no duplicates.
func FindDuplicateNodes(nodes []client.Node) string {
	type nodeID struct {
		key       string
		value     bool
		context   string
		hasExpiry bool
		expiry    int64
	}

	seen := make(map[nodeID]bool)
	for _, n := range nodes {
		id := nodeID{
			key:       n.Key,
			value:     n.Value,
			context:   serializeContexts(NormalizeContexts(n.Context)),
			hasExpiry: n.Expiry != nil,
			expiry:    expiryVal(n.Expiry),
		}
		if seen[id] {
			return n.Key
		}
		seen[id] = true
	}
	return ""
}

func nodeEqual(a, b client.Node) bool {
	if a.Key != b.Key || a.Value != b.Value {
		return false
	}
	if !expiryEqual(a.Expiry, b.Expiry) {
		return false
	}
	if len(a.Context) != len(b.Context) {
		return false
	}
	for i := range a.Context {
		if a.Context[i] != b.Context[i] {
			return false
		}
	}
	return true
}

func serializeContexts(contexts []client.Context) string {
	if len(contexts) == 0 {
		return ""
	}
	parts := make([]string, len(contexts))
	for i, c := range contexts {
		parts[i] = fmt.Sprintf("%s=%s", c.Key, c.Value)
	}
	return strings.Join(parts, ",")
}

func expiryVal(e *int64) int64 {
	if e == nil {
		return 0
	}
	return *e
}

// expirySortKey returns a sort key for expiry.
// nil (permanent) sorts before any set expiry using MinInt64 as sentinel.
func expirySortKey(e *int64) int64 {
	if e == nil {
		return math.MinInt64
	}
	return *e
}

func expiryEqual(a, b *int64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

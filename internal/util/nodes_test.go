package util

import (
	"testing"

	"github.com/digitaldrugstech/terraform-provider-luckperms/internal/client"
)

func TestIsMetaNode(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"displayname.Администрация", true},
		{"weight.500", true},
		{"prefix.100.<#f1c40f>⭐", true},
		{"suffix.10.[VIP]", true},
		{"weight.0", true},
		{"some.permission", false},
		{"group.admin", false},
		{"*", false},
		{"vulcan.bypass.*", false},
		{"displayname", false}, // no dot
		{"weight", false},
		{"prefix", false},
		{"suffix", false},
	}

	for _, tt := range tests {
		if got := IsMetaNode(tt.key); got != tt.want {
			t.Errorf("IsMetaNode(%q) = %v, want %v", tt.key, got, tt.want)
		}
	}
}

func TestSplitNodes(t *testing.T) {
	nodes := []client.Node{
		{Key: "*", Value: true},
		{Key: "displayname.Admin", Value: true},
		{Key: "weight.500", Value: true},
		{Key: "prefix.100.<red>★", Value: true},
		{Key: "vulcan.bypass.*", Value: false},
		{Key: "group.helper", Value: true},
	}

	meta, perms := SplitNodes(nodes)

	if len(meta) != 3 {
		t.Errorf("expected 3 meta nodes, got %d", len(meta))
	}
	if len(perms) != 3 {
		t.Errorf("expected 3 perm nodes, got %d", len(perms))
	}

	// Verify meta contains the right nodes
	metaKeys := map[string]bool{}
	for _, n := range meta {
		metaKeys[n.Key] = true
	}
	if !metaKeys["displayname.Admin"] || !metaKeys["weight.500"] || !metaKeys["prefix.100.<red>★"] {
		t.Errorf("unexpected meta keys: %v", metaKeys)
	}
}

func TestSplitNodes_Empty(t *testing.T) {
	meta, perms := SplitNodes(nil)
	if meta != nil || perms != nil {
		t.Error("expected nil for empty input")
	}
}

func TestParseMetaNodes(t *testing.T) {
	nodes := []client.Node{
		{Key: "displayname.Администрация", Value: true},
		{Key: "weight.500", Value: true},
		{Key: "prefix.100.<#f1c40f>⭐", Value: true},
		{Key: "suffix.10.[VIP]", Value: true},
		{Key: "some.perm", Value: true},
	}

	attrs := ParseMetaNodes(nodes)

	if !attrs.HasDisplayName || attrs.DisplayName != "Администрация" {
		t.Errorf("display_name: got %q (has=%v)", attrs.DisplayName, attrs.HasDisplayName)
	}
	if attrs.Weight != 500 {
		t.Errorf("weight: got %d", attrs.Weight)
	}
	if !attrs.HasPrefix || attrs.Prefix != "100.<#f1c40f>⭐" {
		t.Errorf("prefix: got %q (has=%v)", attrs.Prefix, attrs.HasPrefix)
	}
	if !attrs.HasSuffix || attrs.Suffix != "10.[VIP]" {
		t.Errorf("suffix: got %q (has=%v)", attrs.Suffix, attrs.HasSuffix)
	}
}

func TestParseMetaNodes_NoMeta(t *testing.T) {
	nodes := []client.Node{
		{Key: "some.perm", Value: true},
	}

	attrs := ParseMetaNodes(nodes)

	if attrs.HasDisplayName || attrs.HasPrefix || attrs.HasSuffix {
		t.Error("expected no meta attrs")
	}
	if attrs.Weight != 0 {
		t.Errorf("expected weight 0, got %d", attrs.Weight)
	}
}

func TestParseMetaNodes_WeightZero(t *testing.T) {
	nodes := []client.Node{
		{Key: "weight.0", Value: true},
	}

	attrs := ParseMetaNodes(nodes)
	if attrs.Weight != 0 {
		t.Errorf("expected weight 0, got %d", attrs.Weight)
	}
}

func TestBuildMetaNodes_AllSet(t *testing.T) {
	dn := "Admin"
	px := "100.<red>★"
	sx := "10.[VIP]"

	nodes := BuildMetaNodes(&dn, 500, &px, &sx)

	if len(nodes) != 4 {
		t.Fatalf("expected 4 nodes, got %d", len(nodes))
	}

	keys := map[string]bool{}
	for _, n := range nodes {
		keys[n.Key] = true
		if !n.Value {
			t.Errorf("expected value=true for %s", n.Key)
		}
	}

	expected := []string{"displayname.Admin", "weight.500", "prefix.100.<red>★", "suffix.10.[VIP]"}
	for _, k := range expected {
		if !keys[k] {
			t.Errorf("missing expected key %q", k)
		}
	}
}

func TestBuildMetaNodes_OnlyWeight(t *testing.T) {
	nodes := BuildMetaNodes(nil, 0, nil, nil)

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node (weight only), got %d", len(nodes))
	}
	if nodes[0].Key != "weight.0" {
		t.Errorf("expected weight.0, got %s", nodes[0].Key)
	}
}

func TestBuildMetaNodes_Unicode(t *testing.T) {
	dn := "Без проходки"
	px := "0.<gray>?"

	nodes := BuildMetaNodes(&dn, 0, &px, nil)

	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}

	foundDN := false
	for _, n := range nodes {
		if n.Key == "displayname.Без проходки" {
			foundDN = true
		}
	}
	if !foundDN {
		t.Error("missing unicode display name node")
	}
}

func TestMergeNodes(t *testing.T) {
	meta := []client.Node{
		{Key: "weight.100", Value: true},
	}
	perms := []client.Node{
		{Key: "perm.a", Value: true},
		{Key: "perm.b", Value: false},
	}

	result := MergeNodes(meta, perms)
	if len(result) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(result))
	}
}

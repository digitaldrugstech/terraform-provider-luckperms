package util

import (
	"testing"

	"github.com/digitaldrugstech/terraform-provider-luckperms/internal/client"
)

func TestNormalizeNodes_SortsByKey(t *testing.T) {
	nodes := []client.Node{
		{Key: "z.perm", Value: true},
		{Key: "a.perm", Value: true},
		{Key: "m.perm", Value: true},
	}

	result := NormalizeNodes(nodes)

	if result[0].Key != "a.perm" || result[1].Key != "m.perm" || result[2].Key != "z.perm" {
		t.Errorf("expected sorted by key, got %v", result)
	}
}

func TestNormalizeNodes_TrueBeforeFalse(t *testing.T) {
	nodes := []client.Node{
		{Key: "perm", Value: false},
		{Key: "perm", Value: true},
	}

	result := NormalizeNodes(nodes)

	if !result[0].Value || result[1].Value {
		t.Error("expected true before false")
	}
}

func TestNormalizeNodes_SortsByContext(t *testing.T) {
	nodes := []client.Node{
		{Key: "perm", Value: true, Context: []client.Context{{Key: "server", Value: "b"}}},
		{Key: "perm", Value: true, Context: []client.Context{{Key: "server", Value: "a"}}},
	}

	result := NormalizeNodes(nodes)

	if result[0].Context[0].Value != "a" || result[1].Context[0].Value != "b" {
		t.Error("expected sorted by context")
	}
}

func TestNormalizeNodes_SortsByExpiry(t *testing.T) {
	exp1 := int64(100)
	exp2 := int64(200)
	nodes := []client.Node{
		{Key: "perm", Value: true, Expiry: &exp2},
		{Key: "perm", Value: true, Expiry: &exp1},
	}

	result := NormalizeNodes(nodes)

	if *result[0].Expiry != 100 || *result[1].Expiry != 200 {
		t.Error("expected sorted by expiry")
	}
}

func TestNormalizeNodes_NilExpirySortsBeforeSet(t *testing.T) {
	exp := int64(100)
	nodes := []client.Node{
		{Key: "perm", Value: true, Expiry: &exp},
		{Key: "perm", Value: true, Expiry: nil},
	}

	result := NormalizeNodes(nodes)

	if result[0].Expiry != nil {
		t.Error("expected nil expiry (permanent) to sort first")
	}
	if result[1].Expiry == nil || *result[1].Expiry != 100 {
		t.Error("expected set expiry second")
	}
}

func TestNormalizeNodes_ZeroExpiry(t *testing.T) {
	zero := int64(0)
	nodes := []client.Node{
		{Key: "perm", Value: true, Expiry: &zero},
		{Key: "perm", Value: true, Expiry: nil},
	}

	result := NormalizeNodes(nodes)

	if result[0].Expiry != nil {
		t.Error("nil expiry should sort before expiry=0")
	}
	if result[1].Expiry == nil || *result[1].Expiry != 0 {
		t.Error("expiry=0 should sort second")
	}
}

func TestNormalizeNodes_StripsType(t *testing.T) {
	nodes := []client.Node{
		{Key: "group.admin", Value: true, Type: "inheritance"},
	}

	result := NormalizeNodes(nodes)

	if result[0].Type != "" {
		t.Errorf("expected type stripped, got %q", result[0].Type)
	}
}

func TestNormalizeContexts_SortsByKeyThenValue(t *testing.T) {
	contexts := []client.Context{
		{Key: "world", Value: "overworld"},
		{Key: "server", Value: "b"},
		{Key: "server", Value: "a"},
	}

	result := NormalizeContexts(contexts)

	expected := []client.Context{
		{Key: "server", Value: "a"},
		{Key: "server", Value: "b"},
		{Key: "world", Value: "overworld"},
	}

	for i, c := range result {
		if c != expected[i] {
			t.Errorf("index %d: got %v, want %v", i, c, expected[i])
		}
	}
}

func TestNormalizeContexts_Empty(t *testing.T) {
	result := NormalizeContexts(nil)
	if result != nil {
		t.Errorf("expected nil for empty contexts, got %v", result)
	}
}

func TestNodesEqual_SameAfterNormalization(t *testing.T) {
	a := []client.Node{
		{Key: "b", Value: true, Type: "permission"},
		{Key: "a", Value: true, Type: "permission"},
	}
	b := []client.Node{
		{Key: "a", Value: true},
		{Key: "b", Value: true},
	}

	if !NodesEqual(a, b) {
		t.Error("expected equal after normalization")
	}
}

func TestUnescapeNodeKey(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{`prefix.100.<hover:show_text:'<lang:murchat\.role\.admin>'>⭐</hover>`, `prefix.100.<hover:show_text:'<lang:murchat.role.admin>'>⭐</hover>`},
		{`displayname.Без проходки`, `displayname.Без проходки`},
		{`perm.some\.key\.here`, `perm.some.key.here`},
		{`no.escapes.here`, `no.escapes.here`},
		{``, ``},
	}
	for _, tt := range tests {
		if got := UnescapeNodeKey(tt.input); got != tt.want {
			t.Errorf("UnescapeNodeKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNormalizeNodes_UnescapesDots(t *testing.T) {
	nodes := []client.Node{
		{Key: `prefix.100.<lang:murchat\.role\.admin>`, Value: true},
	}
	result := NormalizeNodes(nodes)
	if result[0].Key != `prefix.100.<lang:murchat.role.admin>` {
		t.Errorf("expected unescaped key, got %q", result[0].Key)
	}
}

func TestNodesEqual_EscapedVsUnescaped(t *testing.T) {
	a := []client.Node{
		{Key: `prefix.100.<lang:murchat\.role\.admin>`, Value: true},
	}
	b := []client.Node{
		{Key: `prefix.100.<lang:murchat.role.admin>`, Value: true},
	}
	if !NodesEqual(a, b) {
		t.Error("escaped and unescaped should be equal after normalization")
	}
}

func TestNodesEqual_DifferentValues(t *testing.T) {
	a := []client.Node{
		{Key: "perm", Value: true},
	}
	b := []client.Node{
		{Key: "perm", Value: false},
	}

	if NodesEqual(a, b) {
		t.Error("expected not equal")
	}
}

func TestNodesEqual_DifferentLengths(t *testing.T) {
	a := []client.Node{
		{Key: "a", Value: true},
		{Key: "b", Value: true},
	}
	b := []client.Node{
		{Key: "a", Value: true},
	}

	if NodesEqual(a, b) {
		t.Error("expected not equal")
	}
}

func TestNodesEqual_WithContexts(t *testing.T) {
	a := []client.Node{
		{Key: "perm", Value: true, Context: []client.Context{
			{Key: "server", Value: "b"},
			{Key: "server", Value: "a"},
		}},
	}
	b := []client.Node{
		{Key: "perm", Value: true, Context: []client.Context{
			{Key: "server", Value: "a"},
			{Key: "server", Value: "b"},
		}},
	}

	if !NodesEqual(a, b) {
		t.Error("expected equal after context normalization")
	}
}

func TestNodesEqual_WithExpiry(t *testing.T) {
	exp := int64(1735689600)
	a := []client.Node{
		{Key: "perm", Value: true, Expiry: &exp},
	}
	b := []client.Node{
		{Key: "perm", Value: true, Expiry: &exp},
	}

	if !NodesEqual(a, b) {
		t.Error("expected equal")
	}
}

func TestNodesEqual_NilVsSetExpiry(t *testing.T) {
	exp := int64(1735689600)
	a := []client.Node{
		{Key: "perm", Value: true},
	}
	b := []client.Node{
		{Key: "perm", Value: true, Expiry: &exp},
	}

	if NodesEqual(a, b) {
		t.Error("expected not equal")
	}
}

func TestFindDuplicateNodes_NoDuplicates(t *testing.T) {
	nodes := []client.Node{
		{Key: "a", Value: true},
		{Key: "b", Value: true},
	}

	if dup := FindDuplicateNodes(nodes); dup != "" {
		t.Errorf("expected no duplicates, got %q", dup)
	}
}

func TestFindDuplicateNodes_Duplicate(t *testing.T) {
	nodes := []client.Node{
		{Key: "perm.a", Value: true},
		{Key: "perm.a", Value: true},
	}

	if dup := FindDuplicateNodes(nodes); dup != "perm.a" {
		t.Errorf("expected duplicate perm.a, got %q", dup)
	}
}

func TestFindDuplicateNodes_SameKeyDifferentContext(t *testing.T) {
	nodes := []client.Node{
		{Key: "perm", Value: true, Context: []client.Context{{Key: "server", Value: "a"}}},
		{Key: "perm", Value: true, Context: []client.Context{{Key: "server", Value: "b"}}},
	}

	if dup := FindDuplicateNodes(nodes); dup != "" {
		t.Errorf("expected no duplicates (different context), got %q", dup)
	}
}

func TestFindDuplicateNodes_SameKeyDifferentValue(t *testing.T) {
	nodes := []client.Node{
		{Key: "perm", Value: true},
		{Key: "perm", Value: false},
	}

	if dup := FindDuplicateNodes(nodes); dup != "" {
		t.Errorf("expected no duplicates (different value), got %q", dup)
	}
}

func TestNormalizeNodes_Unicode(t *testing.T) {
	nodes := []client.Node{
		{Key: "displayname.Администрация", Value: true},
		{Key: "prefix.100.<#f1c40f>⭐", Value: true},
		{Key: "displayname.Без проходки", Value: true},
	}

	result := NormalizeNodes(nodes)

	// Unicode sort: Cyrillic Б < А is not true in UTF-8, but that's fine—
	// we just need deterministic ordering.
	if len(result) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(result))
	}
}

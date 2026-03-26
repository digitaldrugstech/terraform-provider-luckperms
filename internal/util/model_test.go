package util

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/digitaldrugstech/terraform-provider-luckperms/internal/client"
)

func TestModelsToAPINodes(t *testing.T) {
	exp := int64(1735689600)
	models := []NodeModel{
		{
			Key:   types.StringValue("perm.a"),
			Value: types.BoolValue(true),
		},
		{
			Key:    types.StringValue("perm.b"),
			Value:  types.BoolValue(false),
			Expiry: types.Int64Value(exp),
			Context: []ContextModel{
				{Key: types.StringValue("server"), Value: types.StringValue("creative")},
			},
		},
		{
			Key:    types.StringValue("perm.c"),
			Value:  types.BoolValue(true),
			Expiry: types.Int64Null(),
		},
	}

	nodes := ModelsToAPINodes(models)

	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}
	if nodes[0].Key != "perm.a" || !nodes[0].Value || nodes[0].Expiry != nil {
		t.Errorf("node 0: %+v", nodes[0])
	}
	if nodes[1].Key != "perm.b" || nodes[1].Value || nodes[1].Expiry == nil || *nodes[1].Expiry != exp {
		t.Errorf("node 1: %+v", nodes[1])
	}
	if len(nodes[1].Context) != 1 || nodes[1].Context[0].Key != "server" {
		t.Errorf("node 1 context: %+v", nodes[1].Context)
	}
	if nodes[2].Expiry != nil {
		t.Errorf("node 2 expiry should be nil for null TF value, got %v", nodes[2].Expiry)
	}
}

func TestModelsToAPINodes_UnknownExpiry(t *testing.T) {
	models := []NodeModel{
		{
			Key:    types.StringValue("perm.a"),
			Value:  types.BoolValue(true),
			Expiry: types.Int64Unknown(),
		},
	}

	nodes := ModelsToAPINodes(models)

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].Expiry != nil {
		t.Errorf("expected nil Expiry for unknown TF value, got %v", nodes[0].Expiry)
	}
}

func TestAPINodeToModels(t *testing.T) {
	exp := int64(100)
	nodes := []client.Node{
		{Key: "a", Value: true},
		{Key: "b", Value: false, Expiry: &exp, Context: []client.Context{
			{Key: "server", Value: "x"},
		}},
		{Key: "c", Value: true, Context: nil},
	}

	models := APINodeToModels(nodes)

	if len(models) != 3 {
		t.Fatalf("expected 3 models, got %d", len(models))
	}
	if models[0].Key.ValueString() != "a" || !models[0].Value.ValueBool() {
		t.Errorf("model 0: key=%s value=%v", models[0].Key.ValueString(), models[0].Value.ValueBool())
	}
	if !models[0].Expiry.IsNull() {
		t.Error("model 0 expiry should be null")
	}
	if models[1].Expiry.ValueInt64() != 100 {
		t.Errorf("model 1 expiry: %d", models[1].Expiry.ValueInt64())
	}
	if len(models[1].Context) != 1 {
		t.Errorf("model 1 context length: %d", len(models[1].Context))
	}
	if len(models[2].Context) != 0 {
		t.Errorf("model 2 context should be empty, got %d", len(models[2].Context))
	}
}

func TestMapMetaToTFValues(t *testing.T) {
	meta := MetaAttrs{
		DisplayName:    "Admin",
		Weight:         500,
		Prefix:         "100.<red>★",
		HasDisplayName: true,
		HasWeight:      true,
		HasPrefix:      true,
	}

	dn, w, px, sx := MapMetaToTFValues(meta)

	if dn.ValueString() != "Admin" {
		t.Errorf("display_name: %s", dn.ValueString())
	}
	if w.ValueInt64() != 500 {
		t.Errorf("weight: %d", w.ValueInt64())
	}
	if px.ValueString() != "100.<red>★" {
		t.Errorf("prefix: %s", px.ValueString())
	}
	if !sx.IsNull() {
		t.Error("suffix should be null")
	}
}

func TestMapMetaToTFValues_Empty(t *testing.T) {
	meta := MetaAttrs{}
	dn, w, px, sx := MapMetaToTFValues(meta)

	if !dn.IsNull() {
		t.Error("display_name should be null")
	}
	if w.ValueInt64() != 0 {
		t.Errorf("weight should be 0, got %d", w.ValueInt64())
	}
	if !px.IsNull() {
		t.Error("prefix should be null")
	}
	if !sx.IsNull() {
		t.Error("suffix should be null")
	}
}

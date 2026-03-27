package util

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/digitaldrugstech/terraform-provider-luckperms/internal/client"
)

// NodeModel is the shared Terraform model for a permission node.
// Used by both resources and data sources.
type NodeModel struct {
	Key     types.String   `tfsdk:"key"`
	Value   types.Bool     `tfsdk:"value"`
	Expiry  types.Int64    `tfsdk:"expiry"`
	Context []ContextModel `tfsdk:"context"`
}

// ContextModel is the shared Terraform model for a node context.
type ContextModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

// ModelsToAPINodes converts Terraform node models to API nodes.
func ModelsToAPINodes(models []NodeModel) []client.Node {
	nodes := make([]client.Node, len(models))
	for i, m := range models {
		node := client.Node{
			Key:   m.Key.ValueString(),
			Value: m.Value.ValueBool(),
		}
		if !m.Expiry.IsNull() && !m.Expiry.IsUnknown() {
			exp := m.Expiry.ValueInt64()
			node.Expiry = &exp
		}
		contexts := make([]client.Context, len(m.Context))
		for j, c := range m.Context {
			contexts[j] = client.Context{
				Key:   c.Key.ValueString(),
				Value: c.Value.ValueString(),
			}
		}
		node.Context = contexts
		nodes[i] = node
	}
	return nodes
}

// APINodeToModels converts API nodes to Terraform node models.
func APINodeToModels(nodes []client.Node) []NodeModel {
	models := make([]NodeModel, len(nodes))
	for i, n := range nodes {
		model := NodeModel{
			Key:   types.StringValue(n.Key),
			Value: types.BoolValue(n.Value),
		}
		if n.Expiry != nil {
			model.Expiry = types.Int64Value(*n.Expiry)
		} else {
			model.Expiry = types.Int64Null()
		}
		contexts := make([]ContextModel, len(n.Context))
		for j, c := range n.Context {
			contexts[j] = ContextModel{
				Key:   types.StringValue(c.Key),
				Value: types.StringValue(c.Value),
			}
		}
		model.Context = contexts
		models[i] = model
	}
	return models
}

// MapMetaToTFValues converts parsed meta attributes to Terraform-typed values.
func MapMetaToTFValues(meta MetaAttrs) (displayName types.String, weight types.Int64, prefix types.String, suffix types.String) {
	if meta.HasDisplayName {
		displayName = types.StringValue(meta.DisplayName)
	} else {
		displayName = types.StringNull()
	}

	if meta.HasWeight {
		weight = types.Int64Value(meta.Weight)
	} else {
		weight = types.Int64Null()
	}

	if meta.HasPrefix {
		prefix = types.StringValue(meta.Prefix)
	} else {
		prefix = types.StringNull()
	}

	if meta.HasSuffix {
		suffix = types.StringValue(meta.Suffix)
	} else {
		suffix = types.StringNull()
	}

	return
}

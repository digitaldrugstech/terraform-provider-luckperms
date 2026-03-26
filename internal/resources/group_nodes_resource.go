package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/digitaldrugstech/terraform-provider-luckperms/internal/client"
	"github.com/digitaldrugstech/terraform-provider-luckperms/internal/util"
)

var (
	_ resource.Resource                   = &GroupNodesResource{}
	_ resource.ResourceWithImportState    = &GroupNodesResource{}
	_ resource.ResourceWithValidateConfig = &GroupNodesResource{}
)

type GroupNodesResource struct {
	client *client.Client
}

type groupNodesResourceModel struct {
	ID    types.String     `tfsdk:"id"`
	Group types.String     `tfsdk:"group"`
	Nodes []util.NodeModel `tfsdk:"node"`
}

func NewGroupNodesResource() resource.Resource {
	return &GroupNodesResource{}
}

func (r *GroupNodesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_nodes"
}

func (r *GroupNodesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages permission and inheritance nodes for a LuckPerms group. Meta nodes (displayname, weight, prefix, suffix) belong in luckperms_group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier (same as group name).",
				Computed:    true,
			},
			"group": schema.StringAttribute{
				Description: "Group name. Must reference an existing group.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"node": schema.SetNestedBlock{
				Description: "Permission or inheritance node.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Description: "Permission key (e.g., 'worldedit.*', 'group.helper').",
							Required:    true,
						},
						"value": schema.BoolAttribute{
							Description: "Permission value. false = negated. Default: true.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(true),
						},
						"expiry": schema.Int64Attribute{
							Description: "Unix timestamp for temporary nodes. Omit for permanent nodes.",
							Optional:    true,
						},
					},
					Blocks: map[string]schema.Block{
						"context": schema.SetNestedBlock{
							Description: "Server/world context. Multiple context blocks = OR semantics.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"key": schema.StringAttribute{
										Description: "Context key (e.g., 'server', 'world').",
										Required:    true,
									},
									"value": schema.StringAttribute{
										Description: "Context value (e.g., 'creative-build').",
										Required:    true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *GroupNodesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = client.FromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *GroupNodesResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config groupNodesResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, n := range config.Nodes {
		key := n.Key.ValueString()
		if key != "" && util.IsMetaNode(key) {
			resp.Diagnostics.AddAttributeError(
				path.Root("node"),
				"Meta node not allowed in luckperms_group_nodes",
				fmt.Sprintf(
					"Node %q is a meta node (displayname/weight/prefix/suffix). "+
						"Meta nodes belong in the luckperms_group resource, not luckperms_group_nodes.",
					key,
				),
			)
		}
	}
}

func (r *GroupNodesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan groupNodesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiNodes := util.ModelsToAPINodes(plan.Nodes)

	groupName := plan.Group.ValueString()

	unlock := r.client.LockGroup(groupName)
	defer unlock()

	// Read existing nodes to preserve meta nodes from group resource
	currentNodes, nodeErr := r.client.GetGroupNodes(ctx, groupName)
	if nodeErr != nil {
		if client.IsNotFound(nodeErr) {
			resp.Diagnostics.AddError("Group not found", fmt.Sprintf("Group %q does not exist.", groupName))
			return
		}
		resp.Diagnostics.AddError("Error reading current nodes", nodeErr.Error())
		return
	}

	metaNodes, _ := util.SplitNodes(currentNodes)
	merged := util.MergeNodes(metaNodes, apiNodes)

	if err := r.client.SetGroupNodes(ctx, groupName, merged); err != nil {
		resp.Diagnostics.AddError("Error setting group nodes", err.Error())
		return
	}

	plan.ID = types.StringValue(groupName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GroupNodesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state groupNodesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupName := state.Group.ValueString()
	allNodes, err := r.client.GetGroupNodes(ctx, groupName)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading group nodes", err.Error())
		return
	}

	// Filter: only permission/inheritance nodes (exclude meta)
	_, permNodes := util.SplitNodes(allNodes)
	normalized := util.NormalizeNodes(permNodes)

	state.ID = types.StringValue(groupName)
	state.Nodes = util.APINodeToModels(normalized)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *GroupNodesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan groupNodesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiNodes := util.ModelsToAPINodes(plan.Nodes)
	groupName := plan.Group.ValueString()

	unlock := r.client.LockGroup(groupName)
	defer unlock()

	// Read current nodes to preserve meta nodes from group resource
	currentNodes, err := r.client.GetGroupNodes(ctx, groupName)
	if err != nil {
		resp.Diagnostics.AddError("Error reading current nodes", err.Error())
		return
	}

	metaNodes, _ := util.SplitNodes(currentNodes)
	merged := util.MergeNodes(metaNodes, apiNodes)

	if err := r.client.SetGroupNodes(ctx, groupName, merged); err != nil {
		resp.Diagnostics.AddError("Error updating group nodes", err.Error())
		return
	}

	plan.ID = types.StringValue(groupName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GroupNodesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state groupNodesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupName := state.Group.ValueString()

	unlock := r.client.LockGroup(groupName)
	defer unlock()

	// Read current nodes to preserve meta nodes from group resource
	currentNodes, err := r.client.GetGroupNodes(ctx, groupName)
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error reading current nodes", err.Error())
		return
	}

	metaNodes, _ := util.SplitNodes(currentNodes)

	// PUT only meta nodes (clear permission nodes)
	if err := r.client.SetGroupNodes(ctx, groupName, metaNodes); err != nil {
		resp.Diagnostics.AddError("Error clearing permission nodes", err.Error())
	}
}

func (r *GroupNodesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("group"), req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

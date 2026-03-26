package resources

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/digitaldrugstech/terraform-provider-luckperms/internal/client"
	"github.com/digitaldrugstech/terraform-provider-luckperms/internal/util"
)

var (
	_ resource.Resource                = &GroupResource{}
	_ resource.ResourceWithImportState = &GroupResource{}
)

type GroupResource struct {
	client *client.Client
}

type groupResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Weight      types.Int64  `tfsdk:"weight"`
	Prefix      types.String `tfsdk:"prefix"`
	Suffix      types.String `tfsdk:"suffix"`
}

func NewGroupResource() resource.Resource {
	return &GroupResource{}
}

func (r *GroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *GroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a LuckPerms group with its meta attributes (display name, weight, prefix, suffix).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier (same as group name).",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Group name. Lowercase alphanumeric with underscores (^[a-z0-9_]+$). Immutable after creation.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9_]+$`),
						"must be lowercase alphanumeric with underscores",
					),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "Human-readable display name. Stored as displayname.{value} node.",
				Optional:    true,
			},
			"weight": schema.Int64Attribute{
				Description: "Group weight/priority. Higher weight takes precedence when a player has multiple groups. Stored as weight.{value} node. Default: 0.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
			"prefix": schema.StringAttribute{
				Description: "Chat prefix in format {priority}.{text}. Priority determines which group's prefix displays when a player has multiple groups. Example: \"100.<#f1c40f>⭐\". Uses MiniMessage formatting.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^\d+\.`),
						"must start with a numeric priority followed by a dot (e.g., '100.<red>star')",
					),
				},
			},
			"suffix": schema.StringAttribute{
				Description: "Chat suffix in format {priority}.{text}.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^\d+\.`),
						"must start with a numeric priority followed by a dot (e.g., '100.<red>star')",
					),
				},
			},
		},
	}
}

func (r *GroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = client.FromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *GroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan groupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()

	// Create the group
	_, err := r.client.CreateGroup(ctx, name)
	if err != nil {
		if client.IsConflict(err) {
			resp.Diagnostics.AddError(
				"Group already exists",
				fmt.Sprintf("Group %q already exists. Use `terraform import` to manage it.", name),
			)
			return
		}
		resp.Diagnostics.AddError("Error creating group", err.Error())
		return
	}

	// Read existing nodes (may have perm nodes from external sources), merge meta on top
	currentNodes, err := r.client.GetGroupNodes(ctx, name)
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError("Error reading current nodes after group creation", err.Error())
		return
	}
	_, permNodes := util.SplitNodes(currentNodes)
	metaNodes := buildMetaNodesFromPlan(&plan)
	merged := util.MergeNodes(metaNodes, permNodes)

	if err := r.client.SetGroupNodes(ctx, name, merged); err != nil {
		resp.Diagnostics.AddError("Error setting meta nodes", err.Error())
		return
	}

	plan.ID = types.StringValue(name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state groupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	// Read all nodes — also serves as existence check (404 → remove from state)
	nodes, err := r.client.GetGroupNodes(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading group nodes", err.Error())
		return
	}

	meta := util.ParseMetaNodes(nodes)
	state.ID = types.StringValue(name)
	state.Name = types.StringValue(name)
	mapMetaToState(meta, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *GroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan groupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()

	// Read current nodes to preserve permission nodes from group_nodes resource
	currentNodes, err := r.client.GetGroupNodes(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error reading current nodes", err.Error())
		return
	}

	// Split: keep permission nodes, replace meta nodes
	_, permNodes := util.SplitNodes(currentNodes)
	newMetaNodes := buildMetaNodesFromPlan(&plan)
	merged := util.MergeNodes(newMetaNodes, permNodes)

	if err := r.client.SetGroupNodes(ctx, name, merged); err != nil {
		resp.Diagnostics.AddError("Error updating group nodes", err.Error())
		return
	}

	plan.ID = types.StringValue(name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state groupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The default group cannot be deleted in LuckPerms.
	if state.Name.ValueString() == "default" {
		return
	}

	err := r.client.DeleteGroup(ctx, state.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting group", err.Error())
	}
}

func (r *GroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}

// buildMetaNodesFromPlan creates API meta nodes from the plan model.
func buildMetaNodesFromPlan(plan *groupResourceModel) []client.Node {
	var displayName *string
	if !plan.DisplayName.IsNull() && !plan.DisplayName.IsUnknown() {
		v := plan.DisplayName.ValueString()
		displayName = &v
	}

	weight := plan.Weight.ValueInt64()

	var prefix *string
	if !plan.Prefix.IsNull() && !plan.Prefix.IsUnknown() {
		v := plan.Prefix.ValueString()
		prefix = &v
	}

	var suffix *string
	if !plan.Suffix.IsNull() && !plan.Suffix.IsUnknown() {
		v := plan.Suffix.ValueString()
		suffix = &v
	}

	return util.BuildMetaNodes(displayName, weight, prefix, suffix)
}

func mapMetaToState(meta util.MetaAttrs, state *groupResourceModel) {
	state.DisplayName, state.Weight, state.Prefix, state.Suffix = util.MapMetaToTFValues(meta)
}

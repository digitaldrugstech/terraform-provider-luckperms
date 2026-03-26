package resources

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/digitaldrugstech/terraform-provider-luckperms/internal/client"
)

var (
	_ resource.Resource                = &TrackResource{}
	_ resource.ResourceWithImportState = &TrackResource{}
)

type TrackResource struct {
	client *client.Client
}

type trackResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Groups types.List   `tfsdk:"groups"`
}

func NewTrackResource() resource.Resource {
	return &TrackResource{}
}

func (r *TrackResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_track"
}

func (r *TrackResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a LuckPerms track. Tracks define promotion/demotion paths through groups.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier (same as track name).",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Track name. Lowercase alphanumeric with underscores.",
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
			"groups": schema.ListAttribute{
				Description: "Ordered list of group names in the track.",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *TrackResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = client.FromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *TrackResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan trackResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groups := expandStringList(ctx, plan.Groups, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	track, err := r.client.CreateTrack(ctx, plan.Name.ValueString(), groups)
	if err != nil {
		if client.IsConflict(err) {
			resp.Diagnostics.AddError(
				"Track already exists",
				fmt.Sprintf("Track %q already exists. Use `terraform import` to manage it.", plan.Name.ValueString()),
			)
			return
		}
		resp.Diagnostics.AddError("Error creating track", err.Error())
		return
	}

	mapTrackToState(ctx, track, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TrackResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state trackResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	track, err := r.client.GetTrack(ctx, state.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading track", err.Error())
		return
	}

	mapTrackToState(ctx, track, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *TrackResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan trackResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groups := expandStringList(ctx, plan.Groups, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UpdateTrack(ctx, plan.Name.ValueString(), groups); err != nil {
		resp.Diagnostics.AddError("Error updating track", err.Error())
		return
	}

	track, err := r.client.GetTrack(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading track after update", err.Error())
		return
	}

	mapTrackToState(ctx, track, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TrackResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state trackResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteTrack(ctx, state.Name.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting track", err.Error())
	}
}

func (r *TrackResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}

func mapTrackToState(ctx context.Context, track *client.Track, state *trackResourceModel, diags *diag.Diagnostics) {
	state.ID = types.StringValue(track.Name)
	state.Name = types.StringValue(track.Name)

	groupsList, d := types.ListValueFrom(ctx, types.StringType, track.Groups)
	diags.Append(d...)
	state.Groups = groupsList
}

func expandStringList(ctx context.Context, list types.List, diags *diag.Diagnostics) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var result []string
	diags.Append(list.ElementsAs(ctx, &result, false)...)
	return result
}

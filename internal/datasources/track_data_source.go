package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/digitaldrugstech/terraform-provider-luckperms/internal/client"
)

var _ datasource.DataSource = &TrackDataSource{}

type TrackDataSource struct {
	client *client.Client
}

type trackDataSourceModel struct {
	Name   types.String `tfsdk:"name"`
	Groups types.List   `tfsdk:"groups"`
}

func NewTrackDataSource() datasource.DataSource {
	return &TrackDataSource{}
}

func (d *TrackDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_track"
}

func (d *TrackDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Read a LuckPerms track.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Track name.",
				Required:    true,
			},
			"groups": schema.ListAttribute{
				Description: "Ordered list of group names in the track.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *TrackDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = client.FromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *TrackDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state trackDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	track, err := d.client.GetTrack(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading track", err.Error())
		return
	}

	state.Name = types.StringValue(track.Name)
	state.Groups = flattenStringList(ctx, track.Groups, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

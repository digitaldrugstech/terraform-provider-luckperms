package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/digitaldrugstech/terraform-provider-luckperms/internal/client"
)

var _ datasource.DataSource = &TracksDataSource{}

type TracksDataSource struct {
	client *client.Client
}

type tracksDataSourceModel struct {
	Names types.List `tfsdk:"names"`
}

func NewTracksDataSource() datasource.DataSource {
	return &TracksDataSource{}
}

func (d *TracksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tracks"
}

func (d *TracksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List all LuckPerms track names.",
		Attributes: map[string]schema.Attribute{
			"names": schema.ListAttribute{
				Description: "All track names.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *TracksDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = client.FromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *TracksDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	names, err := d.client.GetTracks(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing tracks", err.Error())
		return
	}

	var state tracksDataSourceModel
	state.Names = flattenStringList(ctx, names, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

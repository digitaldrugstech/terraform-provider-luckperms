package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/digitaldrugstech/terraform-provider-luckperms/internal/client"
)

var _ datasource.DataSource = &GroupsDataSource{}

type GroupsDataSource struct {
	client *client.Client
}

type groupsDataSourceModel struct {
	Names types.List `tfsdk:"names"`
}

func NewGroupsDataSource() datasource.DataSource {
	return &GroupsDataSource{}
}

func (d *GroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

func (d *GroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List all LuckPerms group names.",
		Attributes: map[string]schema.Attribute{
			"names": schema.ListAttribute{
				Description: "All group names.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *GroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = client.FromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *GroupsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	names, err := d.client.GetGroups(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing groups", err.Error())
		return
	}

	var state groupsDataSourceModel
	state.Names = flattenStringList(ctx, names, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func flattenStringList(ctx context.Context, values []string, diags *diag.Diagnostics) types.List {
	list, d := types.ListValueFrom(ctx, types.StringType, values)
	diags.Append(d...)
	return list
}

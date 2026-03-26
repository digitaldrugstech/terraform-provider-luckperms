package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/digitaldrugstech/terraform-provider-luckperms/internal/client"
	"github.com/digitaldrugstech/terraform-provider-luckperms/internal/util"
)

var _ datasource.DataSource = &GroupDataSource{}

type GroupDataSource struct {
	client *client.Client
}

type groupDataSourceModel struct {
	Name        types.String     `tfsdk:"name"`
	DisplayName types.String     `tfsdk:"display_name"`
	Weight      types.Int64      `tfsdk:"weight"`
	Prefix      types.String     `tfsdk:"prefix"`
	Suffix      types.String     `tfsdk:"suffix"`
	Nodes       []util.NodeModel `tfsdk:"nodes"`
}

func NewGroupDataSource() datasource.DataSource {
	return &GroupDataSource{}
}

func (d *GroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (d *GroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Read a LuckPerms group.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Group name.",
				Required:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "Display name.",
				Computed:    true,
			},
			"weight": schema.Int64Attribute{
				Description: "Group weight.",
				Computed:    true,
			},
			"prefix": schema.StringAttribute{
				Description: "Chat prefix.",
				Computed:    true,
			},
			"suffix": schema.StringAttribute{
				Description: "Chat suffix.",
				Computed:    true,
			},
			"nodes": schema.ListNestedAttribute{
				Description: "All permission nodes (excluding meta nodes).",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key":    schema.StringAttribute{Computed: true},
						"value":  schema.BoolAttribute{Computed: true},
						"expiry": schema.Int64Attribute{Computed: true},
						"context": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"key":   schema.StringAttribute{Computed: true},
									"value": schema.StringAttribute{Computed: true},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *GroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = client.FromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *GroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state groupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	allNodes, err := d.client.GetGroupNodes(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Error reading group nodes", err.Error())
		return
	}

	meta := util.ParseMetaNodes(allNodes)
	state.Name = types.StringValue(name)
	state.DisplayName, state.Weight, state.Prefix, state.Suffix = util.MapMetaToTFValues(meta)

	_, permNodes := util.SplitNodes(allNodes)
	state.Nodes = util.APINodeToModels(util.NormalizeNodes(permNodes))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

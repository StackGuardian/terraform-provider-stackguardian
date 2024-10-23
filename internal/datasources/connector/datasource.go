package connector

import (
	"context"
	"fmt"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/connector"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &connectorDataSource{}
	_ datasource.DataSourceWithConfigure = &connectorDataSource{}
)

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &connectorDataSource{}
}

type connectorDataSource struct {
	client  *sgclient.Client
	orgName string
}

func (d *connectorDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connector"
}

func (d *connectorDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	provInfo, ok := req.ProviderData.(*customTypes.ProviderInfo)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hashicups.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = provInfo.Client
	d.orgName = provInfo.Org_name
}

func (d *connectorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config connectorDataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	readConnectorResponse, err := d.client.Connectors.ReadConnector(ctx, config.ResourceName.ValueString(), d.orgName)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read connector.", err.Error())
		return
	}

	connectorDataSourceModel, diags := connector.BuildAPIModelToConnectorModel(readConnectorResponse.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	connectorDataSourceModel.ResourceName = config.ResourceName

	resp.Diagnostics.Append(resp.State.Set(ctx, connectorDataSourceModel)...)
}

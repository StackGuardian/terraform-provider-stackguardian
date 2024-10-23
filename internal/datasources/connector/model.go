package connector

import "github.com/hashicorp/terraform-plugin-framework/types"

type connectorDataSourceModel struct {
	ResourceName      types.String `tfsdk:"resource_name"`
	Description       types.String `tfsdk:"description"`
	Settings          types.Object `tfsdk:"settings"`
	DiscoverySettings types.Object `tfsdk:"discovery_settings"`
	Tags              types.List   `tfsdk:"tags"`
}

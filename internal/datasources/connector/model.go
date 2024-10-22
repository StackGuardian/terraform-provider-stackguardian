package connector

import "github.com/hashicorp/terraform-plugin-framework/types"

type connectorDataSourceModel struct {
	ResourceName types.String `tfsdk:"resource_name"`
}

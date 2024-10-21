package workflowoutputs

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema defines the schema for the data source.
func (d *workflowOutputsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"workflow": schema.StringAttribute{
				Required: true,
			},
			"workflow_group": schema.StringAttribute{
				Required: true,
			},
			"data": schema.MapAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"data_json": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

package workflowoutputs

import (
	"context"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema defines the schema for the data source.
func (d *workflowOutputsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"workflow": schema.StringAttribute{
				MarkdownDescription: constants.StackguardianWorkflow,
				Required:            true,
			},
			"workflow_group": schema.StringAttribute{
				MarkdownDescription: constants.StackguardianWorkflowGroup,
				Required:            true,
			},
			"data": schema.MapAttribute{
				MarkdownDescription: constants.DataSourceData,
				ElementType:         types.StringType,
				Computed:            true,
			},
			"data_json": schema.StringAttribute{
				MarkdownDescription: constants.DataSourceDataJson,
				Computed:            true,
			},
		},
	}
}

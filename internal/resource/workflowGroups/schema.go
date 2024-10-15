package workflowGroups

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema defines the schema for the resource.
func (r *workflowGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_name": schema.StringAttribute{
				MarkdownDescription: "The name of the Workflow Group. Must be less than 100 characters and can only contain alphanumeric characters, dashes (-), and underscores (_).",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description of the Workflow Group. Must be less than 256 characters.",
				Optional:            true,
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "A list of tags associated with the Workflow Group. Up to 10 tags are allowed.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
		},
	}

}

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
				MarkdownDescription: "Role name. Must be less than 100 characters. Allowed characters are ^[-a-zA-Z0-9_]+$",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Must be less than 256 characters",
				Optional:            true,
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "Atmost 10 tags are allowed",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

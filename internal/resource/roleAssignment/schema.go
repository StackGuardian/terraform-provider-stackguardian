package roleAssignment

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Schema defines the schema for the resource.
func (r *roleAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"user_id": schema.StringAttribute{
				MarkdownDescription: "Identifier for the user or group. Must be less than 256 characters.",
				Required:            true,
			},
			"entity_type": schema.StringAttribute{
				MarkdownDescription: `Should be one of:
	- <span style="background-color: #eff0f0; color: #e53835;">EMAIL</span>
	- <span style="background-color: #eff0f0; color: #e53835;">GROUP</span>`,
				Required: true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "Role name. Must be less than 255 characters.",
				Required:            true,
			},
		},
	}

}

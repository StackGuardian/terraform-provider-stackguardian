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
				MarkdownDescription: "Fully qualified user email or group. Example: you@example.com for a local user, <SSO Login Method Identifier>/you@example.com for a SSO email when entity_type in EMAIL. <SSO Login Method Identifier>/group-devs when entity_type in GROUP.",
				Required:            true,
			},
			"entity_type": schema.StringAttribute{
				MarkdownDescription: `Should be one of:
	- <span style="background-color: #eff0f0; color: #e53835;">EMAIL</span>
	- <span style="background-color: #eff0f0; color: #e53835;">GROUP</span>`,
				Required: true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: "StackGuardian role name.",
				Required:            true,
			},
		},
	}

}

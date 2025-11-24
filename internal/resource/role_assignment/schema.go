package roleassignment

import (
	"context"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// Schema defines the schema for the resource.
func (r *roleAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"user_id": schema.StringAttribute{
				MarkdownDescription: constants.UserId,
				Required:            true,
			},
			"entity_type": schema.StringAttribute{
				MarkdownDescription: constants.EntityType,
				Required:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: constants.Role,
				Required:            true,
			},
			"send_email": schema.BoolAttribute{
				MarkdownDescription: constants.SendEmail,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"alias": schema.StringAttribute{
				MarkdownDescription: constants.Alias,
				Optional:            true,
			},
		},
	}

}

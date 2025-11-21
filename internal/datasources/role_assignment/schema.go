package roleassignment

import (
	"context"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Schema defines the schema for the resource.
func (r *roleAssignmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"user_id": schema.StringAttribute{
				MarkdownDescription: constants.UserId,
				Required:            true,
			},
			"entity_type": schema.StringAttribute{
				MarkdownDescription: constants.EntityType,
				Computed:            true,
			},
			"role": schema.StringAttribute{
				MarkdownDescription: constants.Role,
				Computed:            true,
			},
			"alias": schema.StringAttribute{
				MarkdownDescription: constants.Alias,
				Computed:            true,
			},
			"send_email": schema.BoolAttribute{
				MarkdownDescription: constants.SendEmail,
				Computed:            true,
			},
		},
	}
}

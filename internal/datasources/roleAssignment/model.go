package roleAssignment

import "github.com/hashicorp/terraform-plugin-framework/types"

type roleAssignmentDataSourceModel struct {
	UserId     types.String `tfsdk:"user_id"`
	EntityType types.String `tfsdk:"entity_type"`
	Role       types.String `tfsdk:"role"`
}

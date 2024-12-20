package roleassignment

import (
	"context"
	"strings"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/flatteners"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoleAssignmentResourceModel struct {
	UserId     types.String `tfsdk:"user_id"`
	EntityType types.String `tfsdk:"entity_type"`
	Role       types.String `tfsdk:"role"`
}

var (
	capabilitiesMap = map[string]*sgsdkgo.EntityTypeEnum{
		"EMAIL": sgsdkgo.EntityTypeEnumEmail.Ptr(),
		"GROUP": sgsdkgo.EntityTypeEnumGroup.Ptr(),
	}
)

func (m *RoleAssignmentResourceModel) ToCreateAPIModel(ctx context.Context) (*sgsdkgo.AddUserToOrganization, diag.Diagnostics) {
	apiModel := sgsdkgo.AddUserToOrganization{
		UserId: m.UserId.ValueString(),
		Role:   m.Role.ValueString(),
	}

	entity, ok := capabilitiesMap[strings.ToUpper(m.EntityType.ValueString())]
	if !ok {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("entityType", "Invalid entityType value")}
	}

	apiModel.EntityType = entity

	return &apiModel, nil
}

func (m *RoleAssignmentResourceModel) ToGetAPIModel(ctx context.Context) (*sgsdkgo.GetorRemoveUserFromOrganization, diag.Diagnostics) {

	apiModel := sgsdkgo.GetorRemoveUserFromOrganization{
		UserId: m.UserId.ValueString(),
	}

	return &apiModel, nil
}

func parseUser(userId string) string {
	if strings.HasPrefix(userId, "local/") {
		return strings.Split(userId, "/")[1]
	}
	return userId
}

func BuildAPIModelToRoleAssignmentModel(apiResponse *sgsdkgo.AddUserToOrganization) (*RoleAssignmentResourceModel, diag.Diagnostics) {
	entityTypeValue := flatteners.String(string(*apiResponse.EntityType.Ptr()))

	userID := parseUser(apiResponse.UserId)

	RoleModel := &RoleAssignmentResourceModel{
		UserId:     flatteners.String(userID),
		Role:       flatteners.String(apiResponse.Role),
		EntityType: entityTypeValue,
	}
	return RoleModel, nil
}

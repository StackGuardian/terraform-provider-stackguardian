package roleAssignment

import (
	"context"
	"strings"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	flatteners "github.com/StackGuardian/terraform-provider-stackguardian/internal/flattners"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type roleAssignmentResourceModel struct {
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

func ParseString(str string) (*sgsdkgo.EntityTypeEnum, bool) {
	c, ok := capabilitiesMap[strings.ToUpper(str)]
	return c, ok
}

func (m *roleAssignmentResourceModel) ToCreateAPIModel(ctx context.Context) (*sgsdkgo.AddUserToOrganization, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	entityVal := m.EntityType.ValueString()
	entity, ok := ParseString(entityVal)
	if !ok {
		diags.AddError("entityType", "Invalid entityType value")
		return nil, diags
	}

	apiModel := sgsdkgo.AddUserToOrganization{
		UserId:     m.UserId.ValueString(),
		EntityType: entity,
		Role:       m.Role.ValueString(),
	}

	return &apiModel, nil
}

func (m *roleAssignmentResourceModel) ToGetAPIModel(ctx context.Context) (*sgsdkgo.GetorRemoveUserFromOrganization, diag.Diagnostics) {

	apiModel := sgsdkgo.GetorRemoveUserFromOrganization{
		UserId: m.UserId.ValueString(),
	}

	return &apiModel, nil
}

func buildAPIModelToRoleAssignmentModel(apiResponse *sgsdkgo.AddUserToOrganization) (*roleAssignmentResourceModel, diag.Diagnostics) {
	entityTypeValue := flatteners.String(string(*apiResponse.EntityType.Ptr()))
	userID := strings.Split(apiResponse.UserId, "/")
	RoleModel := &roleAssignmentResourceModel{
		UserId:     flatteners.String(userID[len(userID)-1]),
		Role:       flatteners.String(apiResponse.Role),
		EntityType: entityTypeValue,
	}
	return RoleModel, nil
}

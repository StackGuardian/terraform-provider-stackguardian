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
		Role:   m.Role.ValueStringPointer(),
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

func BuildAPIModelToRoleAssignmentModel(apiResponse *sgsdkgo.AddUserToOrganization) (*RoleAssignmentResourceModel, diag.Diagnostics) {
	entityTypeValue := flatteners.String(string(*apiResponse.EntityType.Ptr()))

	var userID string
	if strings.HasPrefix(apiResponse.UserId, "local/") {
		userID = strings.Split(apiResponse.UserId, "/")[1]
	} else {
		userID = apiResponse.UserId
	}

	RoleModel := &RoleAssignmentResourceModel{
		UserId:     flatteners.String(userID),
		Role:       flatteners.StringPtr(apiResponse.Role),
		EntityType: entityTypeValue,
	}
	return RoleModel, nil
}

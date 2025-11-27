package roleassignment

import (
	"context"
	"strings"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/expanders"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/flatteners"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoleAssignmentResourceModel struct {
	UserId     types.String `tfsdk:"user_id"`
	EntityType types.String `tfsdk:"entity_type"`
	Role       types.String `tfsdk:"role"`
	Roles      types.List   `tfsdk:"roles"`
	SendEmail  types.Bool   `tfsdk:"send_email"`
	Alias      types.String `tfsdk:"alias"`
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
		Alias:  m.Alias.ValueStringPointer(),
	}

	roles, diags := expanders.StringList(context.TODO(), m.Roles)
	if diags.HasError() {
		return nil, diags
	} else if roles != nil {
		apiModel.Roles = roles
	}

	entity, ok := capabilitiesMap[strings.ToUpper(m.EntityType.ValueString())]
	if !ok {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("entityType", "Invalid entityType value")}
	}

	apiModel.EntityType = entity

	if !m.SendEmail.IsNull() && !m.SendEmail.IsUnknown() {
		apiModel.SendEmail = m.SendEmail.ValueBoolPointer()
	} else {
		sendEmail := true
		apiModel.SendEmail = &sendEmail
	}

	return &apiModel, nil
}

func (m *RoleAssignmentResourceModel) ToGetAPIModel(ctx context.Context) (*sgsdkgo.GetorRemoveUserFromOrganization, diag.Diagnostics) {

	apiModel := sgsdkgo.GetorRemoveUserFromOrganization{
		UserId: m.UserId.ValueStringPointer(),
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

	rolesTerraType, diags := flatteners.ListOfStringToTerraformList(apiResponse.Roles)
	if diags.HasError() {
		return nil, diags
	}

	RoleModel := &RoleAssignmentResourceModel{
		UserId:     flatteners.String(userID),
		Role:       flatteners.StringPtr(apiResponse.Role),
		Roles:      rolesTerraType,
		EntityType: entityTypeValue,
		Alias:      flatteners.StringPtr(apiResponse.Alias),
	}

	return RoleModel, nil
}

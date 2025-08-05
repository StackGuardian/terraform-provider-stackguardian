package rolev4

import (
	"context"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/flatteners"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/role"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoleV4ResourceModel struct {
	role.RoleResourceModel
	DocVersion types.String `tfsdk:"doc_version"`
}

func (m *RoleV4ResourceModel) ToAPIModel(ctx context.Context) (*sgsdkgo.Role, diag.Diagnostics) {
	apimodel, diags := m.RoleResourceModel.ToAPIModel(ctx)
	if diags.HasError() {
		return nil, diags
	}

	apimodel.DocVersion = sgsdkgo.Optional(sgsdkgo.DocVersionEnumV4)

	return apimodel, nil
}

func (m *RoleV4ResourceModel) ToPatchedAPIModel(ctx context.Context) (*sgsdkgo.PatchedRole, diag.Diagnostics) {
	apimodel, diags := m.RoleResourceModel.ToPatchedAPIModel(ctx)
	if diags.HasError() {
		return nil, diags
	}

	apimodel.DocVersion = sgsdkgo.Optional(sgsdkgo.DocVersionEnumV4)

	return apimodel, nil
}

func BuildAPIModelToRoleModel(apiResponse *sgsdkgo.RoleDataResponse) (*RoleV4ResourceModel, diag.Diagnostics) {
	var roleV4ResourceModel RoleV4ResourceModel
	roleResourceModel, diags := role.BuildAPIModelToRoleModel(apiResponse)
	if diags.HasError() {
		return nil, diags
	}

	roleV4ResourceModel.RoleResourceModel = *roleResourceModel
	roleV4ResourceModel.DocVersion = flatteners.String(apiResponse.DocVersion)

	return &roleV4ResourceModel, nil
}

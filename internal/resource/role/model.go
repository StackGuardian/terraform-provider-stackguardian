package role

import (
	"context"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/expanders"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/flatteners"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoleAllowedPermissionsModel struct {
	Name  types.String `tfsdk:"name"`
	Paths types.Map    `tfsdk:"paths"`
}

func (m RoleAllowedPermissionsModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name": types.StringType,
		"paths": types.MapType{
			ElemType: types.ListType{ElemType: types.StringType},
		},
	}
}

type RoleResourceModel struct {
	Id                 types.String `tfsdk:"id"`
	ResourceName       types.String `tfsdk:"resource_name"`
	Description        types.String `tfsdk:"description"`
	AllowedPermissions types.Map    `tfsdk:"allowed_permissions"`
	Tags               types.List   `tfsdk:"tags"`
}

func allowedPermissionsToAPIModel(m types.Map) (map[string]*sgsdkgo.AllowedPermissions, diag.Diagnostics) {
	if m.IsNull() || m.IsUnknown() {
		return nil, nil
	}

	var allowedPermissionsModelValue map[string]*RoleAllowedPermissionsModel
	diags := m.ElementsAs(context.Background(), &allowedPermissionsModelValue, false)
	if diags.HasError() {
		return nil, diags
	}

	allowedPermissionsAPIValue := map[string]*sgsdkgo.AllowedPermissions{}
	for allowedPermissionName, allowedPermissionValue := range allowedPermissionsModelValue {

		allowedPermissionsAPIMapValue := &sgsdkgo.AllowedPermissions{
			Name: allowedPermissionValue.Name.ValueString(),
		}

		if !allowedPermissionValue.Paths.IsNull() && !allowedPermissionValue.Paths.IsUnknown() {
			var allowedPermissionsPathsModelValue map[string]types.List
			diags := allowedPermissionValue.Paths.ElementsAs(context.Background(), &allowedPermissionsPathsModelValue, false)
			if diags.HasError() {
				return nil, diags
			}

			var allowedPermissionsPathsAPIValue = map[string][]string{}
			for pathName, pathValue := range allowedPermissionsPathsModelValue {

				allowedPermissionsPathsAPIListValue := []string{}
				elements := make([]types.String, 0, len(pathValue.Elements()))

				diags := pathValue.ElementsAs(context.TODO(), &elements, false)
				if diags.HasError() {
					return nil, diags
				}
				for _, path := range elements {
					allowedPermissionsPathsAPIListValue = append(allowedPermissionsPathsAPIListValue, path.ValueString())
				}
				allowedPermissionsPathsAPIValue[pathName] = allowedPermissionsPathsAPIListValue
			}
			allowedPermissionsAPIMapValue.Paths = allowedPermissionsPathsAPIValue
		}
		allowedPermissionsAPIValue[allowedPermissionName] = allowedPermissionsAPIMapValue
	}

	return allowedPermissionsAPIValue, nil
}

func (m *RoleResourceModel) ToAPIModel(ctx context.Context) (*sgsdkgo.Role, diag.Diagnostics) {
	apiModel := sgsdkgo.Role{
		ResourceName: m.ResourceName.ValueString(),
	}

	if !m.Description.IsUnknown() && !m.Description.IsNull() {
		apiModel.Description = sgsdkgo.Optional(m.Description.ValueString())
	} else {
		apiModel.Description = sgsdkgo.Null[string]()
	}

	tags, diags := expanders.StringList(context.TODO(), m.Tags)
	if diags.HasError() {
		return nil, diags
	}
	if tags != nil {
		apiModel.Tags = sgsdkgo.Optional(tags)
	} else {
		apiModel.Tags = sgsdkgo.Null[[]string]()
	}

	allowedPermissionsAPIValue, diags := allowedPermissionsToAPIModel(m.AllowedPermissions)
	if diags.HasError() {
		return nil, diags
	}
	if allowedPermissionsAPIValue != nil {
		apiModel.AllowedPermissions = sgsdkgo.Optional(allowedPermissionsAPIValue)
	} else {
		apiModel.AllowedPermissions = sgsdkgo.Null[map[string]*sgsdkgo.AllowedPermissions]()
	}

	return &apiModel, nil
}

func (m *RoleResourceModel) ToPatchedAPIModel(ctx context.Context) (*sgsdkgo.PatchedRole, diag.Diagnostics) {
	apiPatchedModel := sgsdkgo.PatchedRole{
		ResourceName: sgsdkgo.Optional(*m.ResourceName.ValueStringPointer()),
	}

	if !m.Description.IsUnknown() && !m.Description.IsNull() {
		apiPatchedModel.Description = sgsdkgo.Optional(*m.Description.ValueStringPointer())
	} else {
		apiPatchedModel.Description = sgsdkgo.Null[string]()
	}

	tags, diags := expanders.StringList(context.TODO(), m.Tags)
	if diags.HasError() {
		return nil, diags
	}
	if tags != nil {
		apiPatchedModel.Tags = sgsdkgo.Optional(tags)
	} else {
		apiPatchedModel.Tags = sgsdkgo.Null[[]string]()
	}

	allowedPermissionsAPIValue, diags := allowedPermissionsToAPIModel(m.AllowedPermissions)
	if diags.HasError() {
		return nil, diags
	}
	if allowedPermissionsAPIValue != nil {
		apiPatchedModel.AllowedPermissions = sgsdkgo.Optional(allowedPermissionsAPIValue)
	} else {
		apiPatchedModel.AllowedPermissions = sgsdkgo.Null[map[string]*sgsdkgo.AllowedPermissions]()
	}

	return &apiPatchedModel, nil
}

func BuildAPIModelToRoleModel(apiResponse *sgsdkgo.RoleDataResponse) (*RoleResourceModel, diag.Diagnostics) {
	roleModel := &RoleResourceModel{
		Id:           flatteners.String(apiResponse.Id),
		ResourceName: flatteners.String(apiResponse.ResourceName),
		Description:  flatteners.StringPtr(apiResponse.Description),
	}

	if apiResponse.AllowedPermissions != nil {
		allowedPermissionsModelValue := map[string]types.Object{}
		for allowedPermissionName, allowedPermissionValue := range apiResponse.AllowedPermissions {
			allowedPermissionsMapModelValue := RoleAllowedPermissionsModel{}
			if allowedPermissionValue.Paths != nil {
				allowedPermissionsPathsModelValue := map[string]types.List{}
				for pathName, pathValue := range allowedPermissionValue.Paths {
					var paths []types.String
					for _, path := range pathValue {
						paths = append(paths, flatteners.String(path))
					}
					pathsList, diags := types.ListValueFrom(context.Background(), types.StringType, paths)
					if diags.HasError() {
						return nil, diags
					}
					allowedPermissionsPathsModelValue[pathName] = pathsList
				}

				allowedPermissionsPathTerraType, diags := types.MapValueFrom(
					context.Background(),
					types.ListType{ElemType: types.StringType},
					allowedPermissionsPathsModelValue)
				if diags.HasError() {
					return nil, diags
				}
				allowedPermissionsMapModelValue.Paths = allowedPermissionsPathTerraType
			} else {
				allowedPermissionsMapModelValue.Paths = types.MapNull(types.ListType{ElemType: types.StringType})
			}

			allowedPermissionsMapModelValue.Name = flatteners.String(allowedPermissionValue.Name)

			allowedPermissionMapModelTerraType, diags := types.ObjectValueFrom(context.TODO(),
				allowedPermissionsMapModelValue.AttributeTypes(),
				allowedPermissionsMapModelValue)
			if diags.HasError() {
				return nil, diags
			}
			allowedPermissionsModelValue[allowedPermissionName] = allowedPermissionMapModelTerraType
		}

		allowedPermissionsModelTerraType, diags := types.MapValueFrom(
			context.Background(),
			types.ObjectType{
				AttrTypes: RoleAllowedPermissionsModel{}.AttributeTypes(),
			},
			allowedPermissionsModelValue)

		if diags.HasError() {
			return nil, diags
		}

		roleModel.AllowedPermissions = allowedPermissionsModelTerraType
	} else {
		roleModel.AllowedPermissions = types.MapNull(types.ObjectType{AttrTypes: RoleAllowedPermissionsModel{}.AttributeTypes()})
	}

	// Convert Tags from []string to types.List
	if apiResponse.Tags != nil {
		tags := []types.String{}
		for _, tag := range apiResponse.Tags {
			tags = append(tags, flatteners.String(tag))
		}
		tagsList, diags := types.ListValueFrom(context.Background(), types.StringType, tags)
		if diags.HasError() {
			return nil, diags
		}
		roleModel.Tags = tagsList
	} else {
		roleModel.Tags = types.ListNull(types.StringType)
	}
	return roleModel, nil
}

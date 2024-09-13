package role

import (
	"context"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	flatteners "github.com/StackGuardian/terraform-provider-stackguardian/internal/flattners"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoleAllowedPermissionsPathsModel types.List

type RoleAllowedPermissionsModel struct {
	Name  types.String `tfsdk:"name"`
	Paths types.Map    `tfsdk:"paths"`
}

type RoleResourceModel struct {
	ResourceName       types.String `tfsdk:"resource_name"`
	Description        types.String `tfsdk:"description"`
	AllowedPermissions types.Map    `tfsdk:"allowed_permissions"`
	Tags               types.List   `tfsdk:"tags"`
}

func (m *RoleResourceModel) ToAPIModel(ctx context.Context) (*sgsdkgo.Role, diag.Diagnostics) {
	diag := diag.Diagnostics{}
	var allowedPermissionsModelValue map[string]*RoleAllowedPermissionsModel
	var allowedPermissionsDiags = m.AllowedPermissions.ElementsAs(context.Background(), &allowedPermissionsModelValue, false)
	if allowedPermissionsDiags.HasError() {
		return nil, allowedPermissionsDiags
	}
	var allowedPermissionsAPIValue = map[string]*sgsdkgo.AllowedPermissions{}
	for allowedPermissionName, allowedPermissionValue := range allowedPermissionsModelValue {
		var allowedPermissionsPathsModelValue = map[string]types.List{}
		var diags = allowedPermissionValue.Paths.ElementsAs(context.Background(), &allowedPermissionsPathsModelValue, false)
		if diags.HasError() {
			return nil, diags
		}
		var allowedPermissionsPathsAPIValue = map[string][]string{}
		for pathName, pathValue := range allowedPermissionsPathsModelValue {
			allowedPermissionsPathsAPIValue[pathName] = []string{}
			elements := make([]types.String, 0, len(pathValue.Elements()))
			diags := pathValue.ElementsAs(ctx, &elements, false)
			diag.Append(diags...)
			if diag.HasError() {
				return nil, diag
			}
			for _, path := range elements {
				allowedPermissionsPathsAPIValue[pathName] = append(allowedPermissionsPathsAPIValue[pathName], path.ValueString())
			}
		}
		allowedPermissionsAPIValue[allowedPermissionName] = &sgsdkgo.AllowedPermissions{
			Name:  allowedPermissionValue.Name.ValueString(),
			Paths: allowedPermissionsPathsAPIValue,
		}
	}
	apiModel := sgsdkgo.Role{
		ResourceName: m.ResourceName.ValueString(),
		Description:  m.Description.ValueStringPointer(),
	}
	// Convert Tags from types.List to []string
	elements := make([]types.String, 0, len(m.Tags.Elements()))
	diags := m.Tags.ElementsAs(ctx, &elements, false)
	diag.Append(diags...)
	if diag.HasError() {
		return nil, diag
	}
	var tags []string
	for _, tag := range elements {
		tags = append(tags, tag.ValueString())
	}
	apiModel.Tags = tags

	apiModel.AllowedPermissions = allowedPermissionsAPIValue

	return &apiModel, nil
}

func (m *RoleResourceModel) ToPatchedAPIModel(ctx context.Context) (*sgsdkgo.PatchedRole, diag.Diagnostics) {
	diag := diag.Diagnostics{}
	var allowedPermissionsModelValue map[string]*RoleAllowedPermissionsModel
	var allowedPermissionsDiags = m.AllowedPermissions.ElementsAs(context.Background(), &allowedPermissionsModelValue, false)
	if allowedPermissionsDiags.HasError() {
		return nil, allowedPermissionsDiags
	}
	var allowedPermissionsAPIValue = map[string]*sgsdkgo.AllowedPermissions{}
	for allowedPermissionName, allowedPermissionValue := range allowedPermissionsModelValue {
		var allowedPermissionsPathsModelValue = map[string]types.List{}
		var diags = allowedPermissionValue.Paths.ElementsAs(context.Background(), &allowedPermissionsPathsModelValue, false)
		if diags.HasError() {
			return nil, diags
		}
		var allowedPermissionsPathsAPIValue = map[string][]string{}
		for pathName, pathValue := range allowedPermissionsPathsModelValue {
			allowedPermissionsPathsAPIValue[pathName] = []string{}
			elements := make([]types.String, 0, len(pathValue.Elements()))
			diags := pathValue.ElementsAs(ctx, &elements, false)
			diag.Append(diags...)
			if diag.HasError() {
				return nil, diag
			}
			for _, path := range elements {
				allowedPermissionsPathsAPIValue[pathName] = append(allowedPermissionsPathsAPIValue[pathName], path.ValueString())
			}
		}
		allowedPermissionsAPIValue[allowedPermissionName] = &sgsdkgo.AllowedPermissions{
			Name:  allowedPermissionValue.Name.ValueString(),
			Paths: allowedPermissionsPathsAPIValue,
		}
	}
	apiModel := sgsdkgo.PatchedRole{
		ResourceName: m.ResourceName.ValueStringPointer(),
		Description:  m.Description.ValueStringPointer(),
	}
	// Convert Tags from types.List to []string
	elements := make([]types.String, 0, len(m.Tags.Elements()))
	diags := m.Tags.ElementsAs(ctx, &elements, false)
	diag.Append(diags...)
	if diag.HasError() {
		return nil, diag
	}
	var tags []string
	for _, tag := range elements {
		tags = append(tags, tag.ValueString())
	}
	apiModel.Tags = tags

	apiModel.AllowedPermissions = allowedPermissionsAPIValue

	return &apiModel, nil
}

func buildAPIModelToRoleModel(apiResponse *sgsdkgo.Role) (*RoleResourceModel, diag.Diagnostics) {
	diag := diag.Diagnostics{}
	var allowedPermissions types.Map

	allowedPermissionsElements := map[string]attr.Value{}
	allowedPermissionsPathAttrType := map[string]attr.Type{
		"name": types.StringType,
		"paths": types.MapType{
			ElemType: types.ListType{ElemType: types.StringType},
		},
	}

	for allowedPermissionName, allowedPermissionValue := range apiResponse.AllowedPermissions {
		allowedPermissionsPathsElements := map[string]attr.Value{}
		for pathName, pathValue := range allowedPermissionValue.Paths {
			var paths []attr.Value
			for _, path := range pathValue {
				paths = append(paths, flatteners.String(path))
			}
			pathsList, diags := types.ListValueFrom(context.Background(), types.StringType, paths)
			diag.Append(diags...)
			if diag.HasError() {
				return nil, diag
			}

			allowedPermissionsPathsElements[pathName] = pathsList

		}

		allowedPermissionsPathsValue, _ := types.MapValueFrom(
			context.Background(),
			types.ListType{ElemType: types.StringType},
			allowedPermissionsPathsElements)
		allowedPermissionNameValue := map[string]attr.Value{}
		allowedPermissionNameValue["name"] = flatteners.String(allowedPermissionValue.Name)
		allowedPermissionNameValue["paths"] = allowedPermissionsPathsValue
		allowedPermissionNameModelValue, diags := types.ObjectValue(
			allowedPermissionsPathAttrType,
			allowedPermissionNameValue)
		diag.Append(diags...)
		if diag.HasError() {
			return nil, diag
		}
		allowedPermissionsElements[allowedPermissionName] = allowedPermissionNameModelValue
	}
	allowedPermissions, _ = types.MapValueFrom(
		context.Background(),
		types.ObjectType{
			AttrTypes: allowedPermissionsPathAttrType,
		},
		allowedPermissionsElements)

	roleModel := &RoleResourceModel{
		ResourceName:       flatteners.String(apiResponse.ResourceName),
		Description:        flatteners.StringPtr(apiResponse.Description),
		AllowedPermissions: allowedPermissions,
	}
	// Convert Tags from []string to types.List
	var tags []attr.Value
	for _, tag := range apiResponse.Tags {
		tags = append(tags, flatteners.String(tag))
	}
	tagsList, diags := types.ListValueFrom(context.Background(), types.StringType, tags)
	diag.Append(diags...)
	if diag.HasError() {
		return nil, diag
	}
	roleModel.Tags = tagsList
	return roleModel, nil
}

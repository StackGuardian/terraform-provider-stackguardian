package role

import (
	"context"
	"fmt"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema defines the schema for the resource.
func (r *RoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a role: a named set of API permissions that is granted to users through a `stackguardian_role_assignment`. Permission paths typically reference workflow groups, which is how a role is scoped to part of the organization.\n\n~> **Deprecated.** Use `stackguardian_rolev4` instead. This resource expands permissions by combining every path with every other path; `stackguardian_rolev4` maps paths one to one.",
		DeprecationMessage:  "stackguardian_role is deprecated and may be removed in a future release. Use stackguardian_rolev4, which maps permission paths one to one instead of combining every path with every other path.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: constants.Id,
				Computed:            true,
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"resource_name": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.ResourceName, "role"),
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Description, "role"),
				Optional:            true,
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Tags, "role"),
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"allowed_permissions": schema.MapNestedAttribute{
				MarkdownDescription: constants.AllowedPermissions,
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: constants.AllowedPermissionsName,
							Required:            true,
						},
						"paths": schema.MapAttribute{
							MarkdownDescription: constants.AllowedPermissionsPaths,
							Optional:            true,
							Computed:            true,
							ElementType: types.ListType{
								ElemType: types.StringType,
							},
						},
					},
				},
			},
		},
	}
}

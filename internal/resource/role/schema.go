package role

import (
	"context"
	"fmt"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema defines the schema for the resource.
func (r *roleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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
				MarkdownDescription: constants.Tags,
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
							Required:            true,
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

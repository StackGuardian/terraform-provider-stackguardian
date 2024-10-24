package roledatasource

import (
	"context"
	"fmt"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema defines the schema for the resource.
func (r *roleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_name": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.ResourceName, "role"),
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Description, "role"),
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: constants.Tags,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"allowed_permissions": schema.MapNestedAttribute{
				MarkdownDescription: constants.AllowedPermissions,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: constants.AllowedPermissionsName,
							Computed:            true,
						},
						"paths": schema.MapAttribute{
							MarkdownDescription: constants.AllowedPermissionsPaths,
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

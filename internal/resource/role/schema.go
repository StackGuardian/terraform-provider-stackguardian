package role

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema defines the schema for the resource.
func (r *roleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_name": schema.StringAttribute{
				MarkdownDescription: "The role name. It must be less than 100 characters and can only contain the following allowed characters: `^[-a-zA-Z0-9_]+$`.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description of the role. It must be less than 256 characters.",
				Optional:            true,
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "A list of tags associated with the role. A maximum of 10 tags are allowed.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"allowed_permissions": schema.MapNestedAttribute{
				MarkdownDescription: "A map of permissions assigned to the role.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the permission.",
							Required:            true,
						},
						"paths": schema.MapAttribute{						
							MarkdownDescription: "A map of resource paths to which this permission is scoped.",
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

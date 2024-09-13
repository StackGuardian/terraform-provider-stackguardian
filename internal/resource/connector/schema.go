package connector

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema defines the schema for the resource.
func (r *connectorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"organization": schema.StringAttribute{
				Required: true,
			},
			"resource_name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"settings": schema.MapAttribute{
				Required:    true,
				ElementType: types.StringType,
				//TODO: replace xyz with api docs
				Description: "Kind and Config keys are required. Refer documentation at xyz",
			},
			"discovery_settings": schema.MapNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"benchmarks": schema.StringAttribute{
							Required: true,
						},
						"discovery_interval": schema.Float64Attribute{
							Required: true,
						},
						"regions": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"region": schema.StringAttribute{
										Required: true,
									},
								},
							},
							Required: true,
						},
					},
				},
			},
			"is_active": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"0",
						"1",
					),
				},
			},
			"scope": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}
}

package connector

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
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
				Description: "Name of the Connector",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"settings": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"kind": schema.StringAttribute{
						Required: true,
					},
					"config": schema.ListNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{

							Attributes: map[string]schema.Attribute{
								"installation_id": schema.StringAttribute{
									Optional:    true,
									Description: "",
								},
								"github_app_id": schema.StringAttribute{
									Optional:    true,
									Description: "",
								},
								"github_app_webhook_secret": schema.StringAttribute{
									Optional:    true,
									Description: "",
								},
								"github_api_url": schema.StringAttribute{
									Optional:    true,
									Description: "",
								},
								"github_http_url": schema.StringAttribute{
									Optional:    true,
									Description: "",
								},
								"github_app_client_id": schema.StringAttribute{
									Optional:    true,
									Description: "",
								},
								"github_app_client_secret": schema.StringAttribute{
									Optional:    true,
									Description: "",
								},
								"github_app_pem_file_content": schema.StringAttribute{
									Optional:    true,
									Description: "",
								},
								"github_app_webhook_url": schema.StringAttribute{
									Optional:    true,
									Description: "",
								},
								"gitlab_creds": schema.StringAttribute{
									Optional:    true,
									Description: "",
								},
								"gitlab_http_url": schema.StringAttribute{
									Optional:    true,
									Description: "",
								},
								"gitlab_api_url": schema.StringAttribute{
									Optional:    true,
									Description: "",
								},
								"azure_creds": schema.StringAttribute{
									Optional:    true,
									Description: "",
								},
								"azure_devops_http_url": schema.StringAttribute{
									Optional:    true,
									Description: "",
								},
								"azure_devops_api_url": schema.StringAttribute{
									Optional:    true,
									Description: "",
								},
								"bitbucket_creds": schema.StringAttribute{
									Optional:    true,
									Description: "",
								},
								"aws_access_key_id": schema.StringAttribute{
									Optional:    true,
									Description: "",
								},
								"aws_secret_access_key": schema.StringAttribute{
									Optional:    true,
									Description: "",
								},
								"aws_default_region": schema.StringAttribute{
									Optional:    true,
									Description: "",
								},
								"arm_tenant_id": schema.StringAttribute{
									Optional:    true,
									Description: "",
								},
								"arm_subscription_id": schema.StringAttribute{
									Optional:    true,
									Description: "",
								},
								"arm_client_id": schema.StringAttribute{
									Optional:    true,
									Description: "",
								},
								"arm_client_secret": schema.StringAttribute{
									Optional:    true,
									Description: "",
								},
								"gcp_config_file_content": schema.StringAttribute{
									Optional:    true,
									Description: "",
								},
							},
						},
					},
				},
			},
			"discovery_settings": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"benchmarks": schema.MapNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"active": schema.BoolAttribute{
									Required: true,
								},
								"description": schema.StringAttribute{
									Required: true,
								},
								"label": schema.StringAttribute{
									Required: true,
								},
								"runtime_source": schema.StringAttribute{
									Optional: true,
								},
								"summary_description": schema.StringAttribute{
									Required: true,
								},
								"summary_title": schema.StringAttribute{
									Required: true,
								},
								"discovery_interval": schema.Int64Attribute{
									Required: true,
								},
								"last_discovery_time": schema.Int64Attribute{
									Optional: true,
								},
								"is_custom_check": schema.BoolAttribute{
									Required: true,
								},
								"checks": schema.ListAttribute{
									Required:    true,
									ElementType: types.StringType,
								},
								"regions": schema.MapNestedAttribute{
									Required: true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"emails": schema.ListAttribute{
												ElementType: types.StringType,
												Required:    true,
											},
										},
									},
								},
							},
						},
					},
					"discovery_interval": schema.Int64Attribute{
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
				Computed:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("*")})),
			},
			"tags": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}
}

package connector

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema defines the schema for the resource.
func (r *connectorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_name": schema.StringAttribute{
				MarkdownDescription: "Connector name. Must be less than 100 characters. Allowed characters are ^[-a-zA-Z0-9_]+$",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Must be less than 256 characters",
				Optional:            true,
				Computed:            true,
			},
			"settings": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"kind": schema.StringAttribute{
						MarkdownDescription: `Type of connector. Should be one of <span style="background-color: #eff0f0; color: #e53835;">GITHUB_COM</span>
							<span style="background-color: #eff0f0; color: #e53835;">GITHUB_APP_CUSTOM</span>
							<span style="background-color: #eff0f0; color: #e53835;">AWS_STATIC</span>
							<span style="background-color: #eff0f0; color: #e53835;">GCP_STATIC</span>
							<span style="background-color: #eff0f0; color: #e53835;">AWS_RBAC</span>
							<span style="background-color: #eff0f0; color: #e53835;">AWS_OIDC</span>
							<span style="background-color: #eff0f0; color: #e53835;">AZURE_STATIC</span>
							<span style="background-color: #eff0f0; color: #e53835;">AZURE_OIDC</span>
							<span style="background-color: #eff0f0; color: #e53835;">BITBUCKET_ORG</span>
							<span style="background-color: #eff0f0; color: #e53835;">GITLAB_COM</span>
							<span style="background-color: #eff0f0; color: #e53835;">AZURE_DEVOPS</span>`,
						Required: true,
					},
					"config": schema.ListNestedAttribute{
						MarkdownDescription: "Connector secrets configuration",
						Required:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"installation_id": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "",
								},
								"github_app_id": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "",
								},
								"github_app_webhook_secret": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "",
								},
								"github_api_url": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "",
								},
								"github_http_url": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "",
								},
								"github_app_client_id": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "",
								},
								"github_app_client_secret": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "",
								},
								"github_app_pem_file_content": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "",
								},
								"github_app_webhook_url": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "",
								},
								"gitlab_creds": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "",
								},
								"gitlab_http_url": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "",
								},
								"gitlab_api_url": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "",
								},
								"azure_creds": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "",
								},
								"azure_devops_http_url": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "",
								},
								"azure_devops_api_url": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "",
								},
								"bitbucket_creds": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "",
								},
								"aws_access_key_id": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "",
								},
								"aws_secret_access_key": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "",
								},
								"aws_default_region": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "",
								},
								"arm_tenant_id": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "",
								},
								"arm_subscription_id": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "",
								},
								"arm_client_id": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "",
								},
								"arm_client_secret": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "",
								},
								"gcp_config_file_content": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "",
								},
							},
						},
					},
				},
			},
			"discovery_settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Settings for insights",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"benchmarks": schema.MapNestedAttribute{
						MarkdownDescription: "Statistics for different stackguardian resources",
						Optional:            true,
						Computed:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"checks": schema.ListAttribute{
									MarkdownDescription: "",
									Required:            true,
									ElementType:         types.StringType,
								},
								"runtime_source": schema.SingleNestedAttribute{
									MarkdownDescription: "",
									Optional:            true,
									Attributes: map[string]schema.Attribute{
										"source_config_dest_kind": schema.StringAttribute{
											MarkdownDescription: "",
											Optional:            true,
											Computed:            true,
										},
										"config": schema.SingleNestedAttribute{
											MarkdownDescription: "",
											Optional:            true,
											Computed:            true,
											Attributes: map[string]schema.Attribute{
												"include_sub_module": schema.BoolAttribute{
													MarkdownDescription: "",
													Optional:            true,
													Computed:            true,
												},
												"ref": schema.StringAttribute{
													MarkdownDescription: "",
													Optional:            true,
													Computed:            true,
												},
												"git_core_auto_crlf": schema.BoolAttribute{
													MarkdownDescription: "",
													Optional:            true,
													Computed:            true,
												},
												"auth": schema.StringAttribute{
													MarkdownDescription: "",
													Optional:            true,
													Computed:            true,
												},
												"working_dir": schema.StringAttribute{
													MarkdownDescription: "",
													Optional:            true,
													Computed:            true,
												},
												"repo": schema.StringAttribute{
													MarkdownDescription: "",
													Optional:            true,
													Computed:            true,
												},
												"is_private": schema.BoolAttribute{
													MarkdownDescription: "",
													Optional:            true,
													Computed:            true,
												},
											},
										},
									},
								},
								"regions": schema.MapNestedAttribute{
									MarkdownDescription: "",
									Optional:            true,
									Computed:            true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"emails": schema.ListAttribute{
												ElementType: types.StringType,
												Optional:    true,
												Computed:    true,
											},
										},
									},
								},
								"description": schema.StringAttribute{
									MarkdownDescription: "",
									Optional:            true,
								},
								"summary_description": schema.StringAttribute{
									MarkdownDescription: "",
									Optional:            true,
								},
								"active": schema.BoolAttribute{
									MarkdownDescription: "",
									Optional:            true,
									Computed:            true,
								},
								"label": schema.StringAttribute{
									MarkdownDescription: "",
									Optional:            true,
								},
								"is_custom_check": schema.BoolAttribute{
									MarkdownDescription: "",
									Optional:            true,
									Computed:            true,
								},
								"summary_title": schema.StringAttribute{
									MarkdownDescription: "",
									Required:            true,
								},
								"discovery_interval": schema.Int64Attribute{
									MarkdownDescription: "",
									Optional:            true,
									Computed:            true,
								},
							},
						},
					},
				},
			},
			"scope": schema.ListAttribute{
				MarkdownDescription: "Which resources can use this connector",
				ElementType:         types.StringType,
				Computed:            true,
				Optional:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "Tags for connector",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

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
				MarkdownDescription: "The name of the connector. Must be less than 100 characters. Allowed characters are ^[-a-zA-Z0-9_]+$",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A brief description of the connector. Must be less than 256 characters.",
				Optional:            true,
				Computed:            true,
			},
			"settings": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"kind": schema.StringAttribute{
						MarkdownDescription: `The type of connector. Valid options include:
							- <span style="background-color: #eff0f0; color: #e53835;">GITHUB_COM</span>
							- <span style="background-color: #eff0f0; color: #e53835;">GITHUB_APP_CUSTOM</span>
							- <span style="background-color: #eff0f0; color: #e53835;">AWS_STATIC</span>
							- <span style="background-color: #eff0f0; color: #e53835;">GCP_STATIC</span>
							- <span style="background-color: #eff0f0; color: #e53835;">AWS_RBAC</span>
							- <span style="background-color: #eff0f0; color: #e53835;">AWS_OIDC</span>
							- <span style="background-color: #eff0f0; color: #e53835;">AZURE_STATIC</span>
							- <span style="background-color: #eff0f0; color: #e53835;">AZURE_OIDC</span>
							- <span style="background-color: #eff0f0; color: #e53835;">BITBUCKET_ORG</span>
							- <span style="background-color: #eff0f0; color: #e53835;">GITLAB_COM</span>
							- <span style="background-color: #eff0f0; color: #e53835;">AZURE_DEVOPS</span>`,
						Required: true,
					},
					"config": schema.ListNestedAttribute{
						MarkdownDescription: "Configuration settings for the connector's secrets.",
						Required:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"installation_id": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "The installation ID for GitHub applications, if applicable.",
								},
								"github_app_id": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "The application ID for the GitHub app, if applicable.",
								},
								"github_app_webhook_secret": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Webhook secret for the GitHub app, if applicable.",
								},
								"github_api_url": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Base URL for the GitHub API.",
								},
								"github_http_url": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "HTTP URL for accessing the GitHub repository.",
								},
								"github_app_client_id": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Client ID for the GitHub app, if applicable.",
								},
								"github_app_client_secret": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Client secret for the GitHub app, if applicable.",
								},
								"github_app_pem_file_content": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Content of the PEM file for the GitHub app, if applicable.",
								},
								"github_app_webhook_url": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Webhook URL for the GitHub app, if applicable.",
								},
								"gitlab_creds": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Credentials for GitLab integration, if applicable.",
								},
								"gitlab_http_url": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "HTTP URL for accessing the GitLab repository.",
								},
								"gitlab_api_url": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Base URL for the GitLab API.",
								},
								"azure_creds": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Credentials for Azure integration, if applicable.",
								},
								"azure_devops_http_url": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "HTTP URL for accessing Azure DevOps services.",
								},
								"azure_devops_api_url": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Base URL for Azure DevOps API.",
								},
								"bitbucket_creds": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Credentials for Bitbucket integration, if applicable.",
								},
								"aws_access_key_id": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "AWS access key ID for authentication.",
								},
								"aws_secret_access_key": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "AWS secret access key for authentication.",
								},
								"aws_default_region": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Default AWS region for resource operations.",
								},
								"arm_tenant_id": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Azure Resource Manager tenant ID.",
								},
								"arm_subscription_id": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Azure Resource Manager subscription ID.",
								},
								"arm_client_id": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Client ID for Azure Resource Manager.",
								},
								"arm_client_secret": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Client secret for Azure Resource Manager.",
								},
								"gcp_config_file_content": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: "Content of the GCP configuration file.",
								},
							},
						},
					},
				},
			},
			"discovery_settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Settings for discovery insights related to the connector.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"benchmarks": schema.MapNestedAttribute{
						MarkdownDescription: "Statistics for various StackGuardian resources.",
						Optional:            true,
						Computed:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"checks": schema.ListAttribute{
									MarkdownDescription: "List of checks performed during discovery.",
									Required:            true,
									ElementType:         types.StringType,
								},
								"runtime_source": schema.SingleNestedAttribute{
									MarkdownDescription: "",
									Optional:            true,
									Attributes: map[string]schema.Attribute{
										"source_config_dest_kind": schema.StringAttribute{
											MarkdownDescription: "Kind of the source configuration destination. Valid examples include eg:- AWS_RBAC, AZURE_STATIC",
											Optional:            true,
											Computed:            true,
										},
										"config": schema.SingleNestedAttribute{
											MarkdownDescription: "Specific configuration settings for runtime source.",
											Optional:            true,
											Computed:            true,
											Attributes: map[string]schema.Attribute{
												"include_sub_module": schema.BoolAttribute{
													MarkdownDescription: "Indicates whether to include sub-modules.",
													Optional:            true,
													Computed:            true,
												},
												"ref": schema.StringAttribute{
													MarkdownDescription: "Reference identifier for the repository.",
													Optional:            true,
													Computed:            true,
												},
												"git_core_auto_crlf": schema.BoolAttribute{
													MarkdownDescription: "Indicates if core.autocrlf should be enabled.",
													Optional:            true,
													Computed:            true,
												},
												"auth": schema.StringAttribute{
													MarkdownDescription: "Authentication method for accessing the repository.",
													Optional:            true,
													Computed:            true,
												},
												"working_dir": schema.StringAttribute{
													MarkdownDescription: "Working directory for operations.",
													Optional:            true,
													Computed:            true,
												},
												"repo": schema.StringAttribute{
													MarkdownDescription: "Repository name or URL.",
													Optional:            true,
													Computed:            true,
												},
												"is_private": schema.BoolAttribute{
													MarkdownDescription: "Indicates if the repository is private.",
													Optional:            true,
													Computed:            true,
												},
											},
										},
									},
								},
								"regions": schema.MapNestedAttribute{
									MarkdownDescription: "Regions associated with the discovery.",
									Optional:            true,
									Computed:            true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"emails": schema.ListAttribute{
												MarkdownDescription: "List of emails to notify about the discovery.",
												ElementType:         types.StringType,
												Optional:            true,
												Computed:            true,
											},
										},
									},
								},
								"last_discovery_time": schema.Int64Attribute{
									MarkdownDescription: "Timestamp of the last discovery.",
									Optional:            true,
								},
								"description": schema.StringAttribute{
									MarkdownDescription: "",
									Optional:            true,
								},
								"summary_description": schema.StringAttribute{
									MarkdownDescription: "A brief summary of the discovery.",
									Optional:            true,
								},
								"active": schema.BoolAttribute{
									MarkdownDescription: "Indicates if the discovery is active.",
									Optional:            true,
									Computed:            true,
								},
								"label": schema.StringAttribute{
									MarkdownDescription: "Label associated with the discovery.",
									Optional:            true,
								},
								"is_custom_check": schema.BoolAttribute{
									MarkdownDescription: "Indicates if the discovery is a custom check.",
									Optional:            true,
									Computed:            true,
								},
								"summary_title": schema.StringAttribute{
									MarkdownDescription: "Title for the discovery summary.",
									Required:            true,
								},
								"discovery_interval": schema.Int64Attribute{
									MarkdownDescription: "Interval for the discovery process.",
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
				MarkdownDescription: "A list of tags associated with the connectors. Up to 10 tags are allowed.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

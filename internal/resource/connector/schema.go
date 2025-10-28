package connector

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
func (r *connectorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
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
				MarkdownDescription: fmt.Sprintf(constants.ResourceName, "connector"),
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Description, "connector"),
				Optional:            true,
				Computed:            true,
			},
			"settings": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"kind": schema.StringAttribute{
						MarkdownDescription: constants.SettingsKindMarkdownDoc,
						Required:            true,
					},
					"config": schema.ListNestedAttribute{
						MarkdownDescription: constants.SettingsConfig,
						Required:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"role_arn": schema.StringAttribute{
									MarkdownDescription: constants.SettingsConfigRoleArn,
									Optional:            true,
								},
								"external_id": schema.StringAttribute{
									MarkdownDescription: constants.SettingsConfigExternalId,
									Optional:            true,
								},
								"duration_seconds": schema.StringAttribute{
									MarkdownDescription: constants.SettingsConfigDurationSeconds,
									Optional:            true,
								},
								"installation_id": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: constants.SettingsConfigInstallationId,
								},
								"github_app_id": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: constants.SettingsConfigGithubAppId,
								},
								"github_app_webhook_secret": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: constants.SettingsConfigGithubAppWebhookSecret,
								},
								"github_api_url": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: constants.SettingsConfigGithubApiUrl,
								},
								"github_http_url": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: constants.SettingsConfigGithubHttpUrl,
								},
								"github_app_client_id": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: constants.SettingsConfigGithubAppClientId,
								},
								"github_app_client_secret": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: constants.SettingsConfigGithubAppClientSecret,
								},
								"github_app_pem_file_content": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: constants.SettingsConfigGithubAppPemFileContent,
								},
								"github_app_webhook_url": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: constants.SettingsConfigGithubAppWebhookUrl,
								},
								"gitlab_creds": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: constants.SettingsConfigGitlabCreds,
								},
								"gitlab_http_url": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: constants.SettingsConfigGitlabHttpUrl,
								},
								"gitlab_api_url": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: constants.SettingsConfigGitlabApiUrl,
								},
								"azure_creds": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: constants.SettingsConfigAzureCreds,
								},
								"azure_devops_http_url": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: constants.SettingsConfigAzureDevopsHttpUrl,
								},
								"azure_devops_api_url": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: constants.SettingsConfigAzureDevopsApiUrl,
								},
								"bitbucket_creds": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: constants.SettingsConfigBitbucketCreds,
								},
								"aws_access_key_id": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: constants.SettingsConfigAwsAccessKeyId,
								},
								"aws_secret_access_key": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: constants.SettingsConfigAwsSecretAccessKey,
								},
								"aws_default_region": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: constants.SettingsConfigAwsDefaultRegion,
								},
								"arm_tenant_id": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: constants.SettingsConfigArmTenantId,
								},
								"arm_subscription_id": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: constants.SettingsConfigArmSubscriptionId,
								},
								"arm_client_id": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: constants.SettingsConfigArmClientId,
								},
								"arm_client_secret": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: constants.SettingsConfigArmClientSecret,
								},
								"gcp_config_file_content": schema.StringAttribute{
									Optional:            true,
									MarkdownDescription: constants.SettingsConfigGcpConfigFileContent,
								},
							},
						},
					},
				},
			},
			"discovery_settings": schema.SingleNestedAttribute{
				MarkdownDescription: constants.DiscoverySettings,
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"benchmarks": schema.MapNestedAttribute{
						MarkdownDescription: constants.DiscoverySettingsBenchmarks,
						Optional:            true,
						Computed:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"checks": schema.ListAttribute{
									MarkdownDescription: constants.DiscoverySettingsBenchmarksChecks,
									Required:            true,
									ElementType:         types.StringType,
								},
								"runtime_source": schema.SingleNestedAttribute{
									MarkdownDescription: constants.DiscoverySettingsBenchmarksRuntimeSource,
									Optional:            true,
									Attributes: map[string]schema.Attribute{

										"source_config_dest_kind": schema.StringAttribute{
											MarkdownDescription: constants.DiscoverySettingsBenchmarksRuntimeSourceSourceConfigDestKind,
											Optional:            true,
											Computed:            true,
										},
										"config": schema.SingleNestedAttribute{
											MarkdownDescription: constants.DiscoverySettingsBenchmarksRuntimeSourceConfig,
											Optional:            true,
											Computed:            true,
											Attributes: map[string]schema.Attribute{
												"include_sub_module": schema.BoolAttribute{
													MarkdownDescription: constants.DiscoverySettingsBenchmarksRuntimeSourceConfigIncludeSubModule,
													Optional:            true,
													Computed:            true,
												},
												"ref": schema.StringAttribute{
													MarkdownDescription: constants.DiscoverySettingsBenchmarksRuntimeSourceConfigRef,
													Optional:            true,
													Computed:            true,
												},
												"git_core_auto_crlf": schema.BoolAttribute{
													MarkdownDescription: constants.DiscoverySettingsBenchmarksRuntimeSourceConfigGitCoreAutoCRLF,
													Optional:            true,
													Computed:            true,
												},
												"auth": schema.StringAttribute{
													MarkdownDescription: constants.DiscoverySettingsBenchmarksRuntimeSourceConfigAuth,
													Optional:            true,
													Computed:            true,
												},
												"working_dir": schema.StringAttribute{
													MarkdownDescription: constants.DiscoverySettingsBenchmarksRuntimeSourceConfigWorkingDir,
													Optional:            true,
													Computed:            true,
												},
												"repo": schema.StringAttribute{
													MarkdownDescription: constants.DiscoverySettingsBenchmarksRuntimeSourceConfigRepo,
													Optional:            true,
													Computed:            true,
												},
												"is_private": schema.BoolAttribute{
													MarkdownDescription: constants.DiscoverySettingsBenchmarksRuntimeSourceConfigIsPrivate,
													Optional:            true,
													Computed:            true,
												},
											},
										},
									},
								},
								"regions": schema.MapNestedAttribute{
									MarkdownDescription: constants.DiscoverySettingsBenchmarksRegions,
									Optional:            true,
									Computed:            true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"emails": schema.ListAttribute{
												MarkdownDescription: constants.DiscoverySettingsBenchmarksRegionsEmails,
												ElementType:         types.StringType,
												Optional:            true,
												Computed:            true,
											},
										},
									},
								},
								"description": schema.StringAttribute{
									MarkdownDescription: constants.DiscoverySettingsBenchmarksDescription,
									Optional:            true,
								},
								"summary_description": schema.StringAttribute{
									MarkdownDescription: constants.DiscoverySettingsBenchmarksSummaryDescription,
									Optional:            true,
								},
								"active": schema.BoolAttribute{
									MarkdownDescription: constants.DiscoverySettingsBenchmarksActive,
									Optional:            true,
									Computed:            true,
								},
								"label": schema.StringAttribute{
									MarkdownDescription: constants.DiscoverySettingsBenchmarksLabel,
									Optional:            true,
								},
								"is_custom_check": schema.BoolAttribute{
									MarkdownDescription: constants.DiscoverySettingsBenchmarksIsCustomCheck,
									Optional:            true,
									Computed:            true,
								},
								"summary_title": schema.StringAttribute{
									MarkdownDescription: constants.DiscoverySettingsBenchmarksSummaryTitle,
									Required:            true,
								},
								"discovery_interval": schema.Int64Attribute{
									MarkdownDescription: constants.DiscoverySettingsBenchmarksDiscoveryInterval,
									Optional:            true,
									Computed:            true,
								},
							},
						},
					},
				},
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Tags, "connector"),
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

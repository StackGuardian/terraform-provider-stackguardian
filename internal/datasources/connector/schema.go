package connector

import (
	"context"
	"fmt"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *connectorDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: constants.Id,
				Computed:            true,
			},
			"resource_name": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.ResourceName, "connector"),
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Description, "connector"),
				Computed:            true,
			},
			"settings": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"kind": schema.StringAttribute{
						MarkdownDescription: constants.SettingsKindMarkdownDoc,
						Computed:            true,
					},
					"config": schema.ListNestedAttribute{
						MarkdownDescription: constants.SettingsConfig,
						Computed:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"role_arn": schema.StringAttribute{
									MarkdownDescription: constants.SettingsConfigRoleArn,
									Computed:            true,
								},
								"external_id": schema.StringAttribute{
									MarkdownDescription: constants.SettingsConfigExternalId,
									Computed:            true,
								},
								"duration_seconds": schema.StringAttribute{
									MarkdownDescription: constants.SettingsConfigDurationSeconds,
									Computed:            true,
								},
								"installation_id": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: constants.SettingsConfigInstallationId,
								},
								"github_app_id": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: constants.SettingsConfigGithubAppId,
								},
								"github_app_webhook_secret": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: constants.SettingsConfigGithubAppWebhookSecret,
								},
								"github_api_url": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: constants.SettingsConfigGithubApiUrl,
								},
								"github_http_url": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: constants.SettingsConfigGithubHttpUrl,
								},
								"github_app_client_id": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: constants.SettingsConfigGithubAppClientId,
								},
								"github_app_client_secret": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: constants.SettingsConfigGithubAppClientSecret,
								},
								"github_app_pem_file_content": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: constants.SettingsConfigGithubAppPemFileContent,
								},
								"github_app_webhook_url": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: constants.SettingsConfigGithubAppWebhookUrl,
								},
								"gitlab_creds": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: constants.SettingsConfigGitlabCreds,
								},
								"gitlab_http_url": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: constants.SettingsConfigGitlabHttpUrl,
								},
								"gitlab_api_url": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: constants.SettingsConfigGitlabApiUrl,
								},
								"azure_creds": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: constants.SettingsConfigAzureCreds,
								},
								"azure_devops_http_url": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: constants.SettingsConfigAzureDevopsHttpUrl,
								},
								"azure_devops_api_url": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: constants.SettingsConfigAzureDevopsApiUrl,
								},
								"bitbucket_creds": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: constants.SettingsConfigBitbucketCreds,
								},
								"aws_access_key_id": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: constants.SettingsConfigAwsAccessKeyId,
								},
								"aws_secret_access_key": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: constants.SettingsConfigAwsSecretAccessKey,
								},
								"aws_default_region": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: constants.SettingsConfigAwsDefaultRegion,
								},
								"arm_tenant_id": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: constants.SettingsConfigArmTenantId,
								},
								"arm_subscription_id": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: constants.SettingsConfigArmSubscriptionId,
								},
								"arm_client_id": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: constants.SettingsConfigArmClientId,
								},
								"arm_client_secret": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: constants.SettingsConfigArmClientId,
								},
								"gcp_config_file_content": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: constants.SettingsConfigGcpConfigFileContent,
								},
							},
						},
					},
				},
			},
			"discovery_settings": schema.SingleNestedAttribute{
				MarkdownDescription: constants.DiscoverySettings,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"benchmarks": schema.MapNestedAttribute{
						MarkdownDescription: constants.DiscoverySettingsBenchmarks,
						Computed:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"checks": schema.ListAttribute{
									MarkdownDescription: constants.DiscoverySettingsBenchmarksChecks,
									Computed:            true,
									ElementType:         types.StringType,
								},
								"runtime_source": schema.SingleNestedAttribute{
									MarkdownDescription: constants.DiscoverySettingsBenchmarksRuntimeSource,
									Computed:            true,
									Attributes: map[string]schema.Attribute{

										"source_config_dest_kind": schema.StringAttribute{
											MarkdownDescription: constants.DiscoverySettingsBenchmarksRuntimeSourceSourceConfigDestKind,
											Computed:            true,
										},
										"config": schema.SingleNestedAttribute{
											MarkdownDescription: constants.DiscoverySettingsBenchmarksRuntimeSourceConfig,
											Computed:            true,
											Attributes: map[string]schema.Attribute{
												"include_sub_module": schema.BoolAttribute{
													MarkdownDescription: constants.DiscoverySettingsBenchmarksRuntimeSourceConfigIncludeSubModule,
													Computed:            true,
												},
												"ref": schema.StringAttribute{
													MarkdownDescription: constants.DiscoverySettingsBenchmarksRuntimeSourceConfigRef,
													Computed:            true,
												},
												"git_core_auto_crlf": schema.BoolAttribute{
													MarkdownDescription: constants.DiscoverySettingsBenchmarksRuntimeSourceConfigGitCoreAutoCRLF,
													Computed:            true,
												},
												"auth": schema.StringAttribute{
													MarkdownDescription: constants.DiscoverySettingsBenchmarksRuntimeSourceConfigAuth,
													Computed:            true,
												},
												"working_dir": schema.StringAttribute{
													MarkdownDescription: constants.DiscoverySettingsBenchmarksRuntimeSourceConfigWorkingDir,
													Computed:            true,
												},
												"repo": schema.StringAttribute{
													MarkdownDescription: constants.DiscoverySettingsBenchmarksRuntimeSourceConfigRepo,
													Computed:            true,
												},
												"is_private": schema.BoolAttribute{
													MarkdownDescription: constants.DiscoverySettingsBenchmarksRuntimeSourceConfigIsPrivate,
													Computed:            true,
												},
											},
										},
									},
								},
								"regions": schema.MapNestedAttribute{
									MarkdownDescription: constants.DiscoverySettingsBenchmarksRegions,
									Computed:            true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"emails": schema.ListAttribute{
												MarkdownDescription: constants.DiscoverySettingsBenchmarksRegionsEmails,
												ElementType:         types.StringType,
												Computed:            true,
											},
										},
									},
								},
								"description": schema.StringAttribute{
									MarkdownDescription: constants.DiscoverySettingsBenchmarksDescription,
									Computed:            true,
								},
								"summary_description": schema.StringAttribute{
									MarkdownDescription: constants.DiscoverySettingsBenchmarksSummaryDescription,
									Computed:            true,
								},
								"active": schema.BoolAttribute{
									MarkdownDescription: constants.DiscoverySettingsBenchmarksActive,
									Computed:            true,
								},
								"label": schema.StringAttribute{
									MarkdownDescription: constants.DiscoverySettingsBenchmarksLabel,
									Computed:            true,
								},
								"is_custom_check": schema.BoolAttribute{
									MarkdownDescription: constants.DiscoverySettingsBenchmarksIsCustomCheck,
									Computed:            true,
								},
								"summary_title": schema.StringAttribute{
									MarkdownDescription: constants.DiscoverySettingsBenchmarksSummaryTitle,
									Computed:            true,
								},
								"discovery_interval": schema.Int64Attribute{
									MarkdownDescription: constants.DiscoverySettingsBenchmarksDiscoveryInterval,
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
				Computed:            true,
			},
		},
	}
}

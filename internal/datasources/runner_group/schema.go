package runnergroupdatasource

import (
	"context"
	"fmt"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *runnerGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_name": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.ResourceName, "runner group"),
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Description, "runner group"),
				Computed:            true,
			},
			"runner_token": schema.StringAttribute{
				MarkdownDescription: constants.RunnerToken,
				Computed:            true,
			},
			"max_number_of_runners": schema.Int32Attribute{
				MarkdownDescription: constants.MaxNumberOfRunners,
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Tags, "runner group"),
				ElementType:         types.StringType,
				Computed:            true,
			},
			"storage_backend_config": schema.SingleNestedAttribute{
				MarkdownDescription: constants.StorageBackendConfig,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: constants.RunnerGroupType,
						Computed:            true,
					},
					"azure_blob_storage_access_key": schema.StringAttribute{
						MarkdownDescription: constants.AzureBlobStorageAccessKey,
						Computed:            true,
					},
					"azure_blob_storage_account_name": schema.StringAttribute{
						MarkdownDescription: constants.AzureBlobStorageAccountName,
						Computed:            true,
					},
					"s3_bucket_name": schema.StringAttribute{
						MarkdownDescription: constants.S3BucketName,
						Computed:            true,
					},
					"aws_region": schema.StringAttribute{
						MarkdownDescription: constants.AWSRegion,
						Computed:            true,
					},
					"auth": schema.SingleNestedAttribute{
						MarkdownDescription: constants.Auth,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"integration_id": schema.StringAttribute{
								MarkdownDescription: constants.IntegrationId,
								Computed:            true,
							},
						},
					},
				},
			},
			//RunControllerRuntimeSource
			"run_controller_runtime_source": schema.SingleNestedAttribute{
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
							"git_sparse_checkout_config": schema.StringAttribute{
								MarkdownDescription: constants.PolicyVCSConfigCustomSourceGitSparseCheckoutConfig,
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
							"docker_image": schema.StringAttribute{
								MarkdownDescription: constants.DockerImage,
								Computed:            true,
							},
							"docker_registry_username": schema.StringAttribute{
								MarkdownDescription: constants.DockerRegistryUsername,
								Computed:            true,
							},
							"local_workspace_dir": schema.StringAttribute{
								Computed: true,
							},
						},
					},
				},
			},
		},
	}
}

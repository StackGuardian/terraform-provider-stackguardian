package runnergroup

import (
	"context"
	"fmt"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *runnerGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_name": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.ResourceName, "runner group"),
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Description, "policy"),
				Optional:            true,
				Computed:            true,
			},
			"runner_token": schema.StringAttribute{
				MarkdownDescription: constants.RunnerToken,
				Optional:            true,
				Computed:            true,
			},
			"max_number_of_runners": schema.Int32Attribute{
				MarkdownDescription: constants.MaxNumberOfRunners,
				Optional:            true,
				Computed:            true,
				Default:             int32default.StaticInt32(1),
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Tags, "runner group"),
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"storage_backend_config": schema.SingleNestedAttribute{
				MarkdownDescription: constants.StorageBackendConfig,
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: constants.RunnerGroupType,
						Required:            true,
					},
					"azure_blob_storage_access_key": schema.StringAttribute{
						MarkdownDescription: constants.AzureBlobStorageAccessKey,
						Optional:            true,
					},
					"azure_blob_storage_account_name": schema.StringAttribute{
						MarkdownDescription: constants.AzureBlobStorageAccountName,
						Optional:            true,
					},
					"s3_bucket_name": schema.StringAttribute{
						MarkdownDescription: constants.S3BucketName,
						Optional:            true,
					},
					"aws_region": schema.StringAttribute{
						MarkdownDescription: constants.AWSRegion,
						Optional:            true,
					},
					"auth": schema.SingleNestedAttribute{
						MarkdownDescription: constants.Auth,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"integration_id": schema.StringAttribute{
								MarkdownDescription: constants.IntegrationId,
								Required:            true,
							},
						},
					},
				},
			},
			//RunControllerRuntimeSource
			"run_controller_runtime_source": schema.SingleNestedAttribute{
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
							"git_sparse_checkout_config": schema.StringAttribute{
								MarkdownDescription: constants.PolicyVCSConfigCustomSourceGitSparseCheckoutConfig,
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
							"docker_image": schema.StringAttribute{
								MarkdownDescription: constants.DockerImage,
								Optional:            true,
							},
							"docker_registry_username": schema.StringAttribute{
								MarkdownDescription: constants.DockerRegistryUsername,
								Optional:            true,
							},
							"local_workspace_dir": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
		},
	}
}

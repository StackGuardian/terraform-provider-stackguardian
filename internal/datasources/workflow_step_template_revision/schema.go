package workflowsteptemplaterevision

import (
	"context"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema defines the schema for the data source.
func (d *workflowStepTemplateRevisionDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to read a workflow step template revision.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateRevisionId,
				Required:            true,
			},
			"template_id": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateRevisionTemplateId,
				Computed:            true,
			},
			"alias": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateRevisionAlias,
				Computed:            true,
			},
			"notes": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateRevisionNotes,
				Computed:            true,
			},
			"long_description": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateRevisionDescription,
				Computed:            true,
			},
			"template_type": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateRevisionType,
				Computed:            true,
			},
			"source_config_kind": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateSourceConfigKindCommon,
				Computed:            true,
			},
			"is_active": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateIsActiveCommon,
				Computed:            true,
			},
			"is_public": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateIsPublicCommon,
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateRevisionTags,
				ElementType:         types.StringType,
				Computed:            true,
			},
			"context_tags": schema.MapAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateRevisionContextTags,
				ElementType:         types.StringType,
				Computed:            true,
			},
			"runtime_source": schema.SingleNestedAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateRevisionRuntimeSource,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"source_config_dest_kind": schema.StringAttribute{
						MarkdownDescription: constants.WorkflowStepTemplateRuntimeSourceDestKindCommon,
						Computed:            true,
					},
					"additional_config": schema.MapAttribute{
						MarkdownDescription: constants.WorkflowStepTemplateRevisionRuntimeSourceAdditionalConfig,
						ElementType:         types.StringType,
						Computed:            true,
					},
					"config": schema.SingleNestedAttribute{
						MarkdownDescription: constants.WorkflowStepTemplateRevisionRuntimeSourceConfig,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"is_private": schema.BoolAttribute{
								MarkdownDescription: constants.WorkflowStepTemplateRuntimeSourceConfigIsPrivateCommon,
								Computed:            true,
							},
							"auth": schema.StringAttribute{
								MarkdownDescription: constants.WorkflowStepTemplateRuntimeSourceConfigAuthCommon,
								Computed:            true,
								Sensitive:           true,
							},
							"docker_image": schema.StringAttribute{
								MarkdownDescription: constants.WorkflowStepTemplateRuntimeSourceConfigDockerImageCommon,
								Computed:            true,
							},
							"docker_registry_username": schema.StringAttribute{
								MarkdownDescription: constants.WorkflowStepTemplateRuntimeSourceConfigDockerRegistryUsernameCommon,
								Computed:            true,
							},
							"local_workspace_dir": schema.StringAttribute{
								MarkdownDescription: constants.WorkflowStepTemplateRuntimeSourceConfigLocalWorkspaceDirCommon,
								Computed:            true,
							},
						},
					},
				},
			},
			"deprecation": schema.SingleNestedAttribute{
				MarkdownDescription: constants.Deprecation,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"effective_date": schema.StringAttribute{
						MarkdownDescription: constants.DeprecationEffectiveDate,
						Computed:            true,
					},
					"message": schema.StringAttribute{
						MarkdownDescription: constants.DeprecationMessage,
						Computed:            true,
					},
				},
			},
		},
	}
}

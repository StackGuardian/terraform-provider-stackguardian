package workflowsteptemplate

import (
	"context"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema defines the schema for the data source.
func (d *workflowStepTemplateDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to read a workflow step template.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: constants.DatasourceId,
				Required:            true,
			},
			"template_name": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateName,
				Computed:            true,
			},
			"template_type": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateType,
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
			"description": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateDescription,
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateTags,
				ElementType:         types.StringType,
				Computed:            true,
			},
			"context_tags": schema.MapAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateContextTags,
				ElementType:         types.StringType,
				Computed:            true,
			},
			"shared_orgs_list": schema.ListAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateSharedOrgsList,
				ElementType:         types.StringType,
				Computed:            true,
			},
			"source_config_kind": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateSourceConfigKindCommon,
				Computed:            true,
			},
			"latest_revision": schema.Int32Attribute{
				MarkdownDescription: constants.WorkflowStepTemplateLatestRevision,
				Computed:            true,
			},
			"next_revision": schema.Int32Attribute{
				MarkdownDescription: constants.WorkflowStepTemplateNextRevision,
				Computed:            true,
			},
			"runtime_source": schema.SingleNestedAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateRuntimeSource,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"source_config_dest_kind": schema.StringAttribute{
						MarkdownDescription: constants.WorkflowStepTemplateRuntimeSourceDestKindCommon,
						Computed:            true,
					},
					"additional_config": schema.MapAttribute{
						MarkdownDescription: constants.WorkflowStepTemplateRuntimeSourceAdditionalConfig,
						ElementType:         types.StringType,
						Computed:            true,
					},
					"config": schema.SingleNestedAttribute{
						MarkdownDescription: constants.WorkflowStepTemplateRuntimeSourceConfig,
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
		},
	}

}

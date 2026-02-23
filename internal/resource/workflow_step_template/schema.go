package workflowsteptemplate

import (
	"context"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema defines the schema for the resource.
func (r *workflowStepTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a workflow step template resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: constants.Id,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"template_name": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateName,
				Required:            true,
			},
			"template_type": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateType,
				Computed:            true,
			},
			"is_active": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateIsActiveCommon,
				Optional:            true,
				Computed:            true,
			},
			"is_public": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateIsPublicCommon,
				Optional:            true,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateDescription,
				Optional:            true,
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateTags,
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"context_tags": schema.MapAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateContextTags,
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"shared_orgs_list": schema.ListAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateSharedOrgsList,
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"source_config_kind": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateSourceConfigKindCommon,
				Required:            true,
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
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"source_config_dest_kind": schema.StringAttribute{
						MarkdownDescription: constants.WorkflowStepTemplateRuntimeSourceDestKindCommon,
						Optional:            true,
						Computed:            true,
					},
					"additional_config": schema.MapAttribute{
						MarkdownDescription: constants.WorkflowStepTemplateRuntimeSourceAdditionalConfig,
						ElementType:         types.StringType,
						Optional:            true,
					},
					"config": schema.SingleNestedAttribute{
						MarkdownDescription: constants.WorkflowStepTemplateRuntimeSourceConfig,
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"is_private": schema.BoolAttribute{
								MarkdownDescription: constants.WorkflowStepTemplateRuntimeSourceConfigIsPrivateCommon,
								Optional:            true,
								Computed:            true,
							},
							"auth": schema.StringAttribute{
								MarkdownDescription: constants.WorkflowStepTemplateRuntimeSourceConfigAuthCommon,
								Optional:            true,
								Sensitive:           true,
							},
							"docker_image": schema.StringAttribute{
								MarkdownDescription: constants.WorkflowStepTemplateRuntimeSourceConfigDockerImageCommon,
								Required:            true,
							},
							"docker_registry_username": schema.StringAttribute{
								MarkdownDescription: constants.WorkflowStepTemplateRuntimeSourceConfigDockerRegistryUsernameCommon,
								Optional:            true,
							},
						},
					},
				},
			},
		},
	}
}

package workflowsteptemplaterevision

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
func (r *workflowStepTemplateRevisionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a workflow step template revision resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: constants.Id,
				Computed:            true,
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"template_id": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateRevisionTemplateId,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"alias": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateRevisionAlias,
				Optional:            true,
				Computed:            true,
			},
			"notes": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateRevisionNotes,
				Optional:            true,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateRevisionDescription,
				Optional:            true,
				Computed:            true,
			},
			"template_type": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateRevisionType,
				Computed:            true,
			},
			"source_config_kind": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateSourceConfigKindCommon,
				Required:            true,
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
			"tags": schema.ListAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateRevisionTags,
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"context_tags": schema.MapAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateRevisionContextTags,
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"runtime_source": schema.SingleNestedAttribute{
				MarkdownDescription: constants.WorkflowStepTemplateRevisionRuntimeSource,
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"source_config_dest_kind": schema.StringAttribute{
						MarkdownDescription: constants.WorkflowStepTemplateRuntimeSourceDestKindCommon,
						Required:            true,
					},
					"additional_config": schema.MapAttribute{
						MarkdownDescription: constants.WorkflowStepTemplateRevisionRuntimeSourceAdditionalConfig,
						ElementType:         types.StringType,
						Optional:            true,
					},
					"config": schema.SingleNestedAttribute{
						MarkdownDescription: constants.WorkflowStepTemplateRevisionRuntimeSourceConfig,
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
			"deprecation": schema.SingleNestedAttribute{
				MarkdownDescription: constants.Deprecation,
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"effective_date": schema.StringAttribute{
						MarkdownDescription: constants.DeprecationEffectiveDate,
						Optional:            true,
						Computed:            true,
					},
					"message": schema.StringAttribute{
						MarkdownDescription: constants.DeprecationMessage,
						Optional:            true,
						Computed:            true,
					},
				},
			},
		},
	}
}

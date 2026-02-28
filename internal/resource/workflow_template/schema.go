package workflowtemplate

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

func WorkflowTemplateRuntimeSourceConfig() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"source_config_dest_kind": schema.StringAttribute{
			MarkdownDescription: constants.RuntimeSourceDestKind,
			Optional:            true,
		},
		"config": schema.SingleNestedAttribute{
			MarkdownDescription: constants.RuntimeSourceConfig,
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"auth": schema.StringAttribute{
					MarkdownDescription: constants.RuntimeSourceConfigAuth,
					Optional:            true,
					Sensitive:           true,
				},
				"git_core_auto_crlf": schema.BoolAttribute{
					MarkdownDescription: constants.RuntimeSourceConfigGitCoreCRLF,
					Optional:            true,
					Computed:            true,
				},
				"git_sparse_checkout_config": schema.StringAttribute{
					MarkdownDescription: constants.RuntimeSourceConfigGitSparse,
					Optional:            true,
				},
				"include_sub_module": schema.BoolAttribute{
					MarkdownDescription: constants.RuntimeSourceConfigIncludeSubmodule,
					Optional:            true,
				},
				"is_private": schema.BoolAttribute{
					MarkdownDescription: constants.RuntimeSourceConfigIsPrivate,
					Optional:            true,
					Computed:            true,
				},
				"ref": schema.StringAttribute{
					MarkdownDescription: constants.RuntimeSourceConfigRef,
					Optional:            true,
					Computed:            true,
				},
				"repo": schema.StringAttribute{
					MarkdownDescription: constants.RuntimeSourceConfigRepo,
					Required:            true,
				},
				"working_dir": schema.StringAttribute{
					MarkdownDescription: constants.RuntimeSourceConfigWorkingDir,
					Optional:            true,
				},
			},
		},
	}
}

// Schema defines the schema for the resource.
func (r *workflowTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a workflow template resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: constants.Id,
				Computed:            true,
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"owner_org": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowTemplateOwnerOrg,
				Computed:            true,
			},
			"template_name": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowTemplateName,
				Required:            true,
			},
			"source_config_kind": schema.StringAttribute{
				MarkdownDescription: constants.SourceConfigKind,
				Required:            true,
			},
			"is_public": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowTemplateIsPublic,
				Optional:            true,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Description, "Description for workflow template"),
				Optional:            true,
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Tags, "Tags for workflow template"),
				ElementType:         types.StringType,
				Optional:            true,
			},
			"context_tags": schema.MapAttribute{
				MarkdownDescription: fmt.Sprintf(constants.ContextTags, "workflow template"),
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"shared_orgs_list": schema.ListAttribute{
				MarkdownDescription: constants.WorkflowTemplateSharedOrgs,
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"runtime_source": schema.SingleNestedAttribute{
				MarkdownDescription: fmt.Sprintf(constants.RuntimeSource, "template"),
				Optional:            true,
				Computed:            true,
				Attributes:          WorkflowTemplateRuntimeSourceConfig(),
			},
			"vcs_triggers": schema.SingleNestedAttribute{
				MarkdownDescription: constants.VCSTriggers,
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: constants.VCSTriggersType,
						Required:            true,
					},
					"create_tag": schema.SingleNestedAttribute{
						MarkdownDescription: constants.VCSTriggersCreateTag,
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"create_revision": schema.SingleNestedAttribute{
								MarkdownDescription: constants.VCSTriggersCreateTagRevision,
								Required:            true,
								Attributes: map[string]schema.Attribute{
									"enabled": schema.BoolAttribute{
										MarkdownDescription: constants.VCSTriggersCreateTagRevisionEnabled,
										Optional:            true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

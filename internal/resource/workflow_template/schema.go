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
			MarkdownDescription: `Destination kind for the source configuration (e.g., GITHUB_COM, GITHUB_APP_CUSTOM, GITLAB_OAUTH_SSH, GITLAB_COM, AZURE_DEVOPS).`,
			Optional:            true,
		},
		"config": schema.SingleNestedAttribute{
			MarkdownDescription: "Configuration for the runtime environment.",
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"auth": schema.StringAttribute{
					MarkdownDescription: "Connector id to access private git repository",
					Optional:            true,
					Sensitive:           true,
				},
				"git_core_auto_crlf": schema.BoolAttribute{
					MarkdownDescription: "Whether to automatically handle CRLF line endings.",
					Optional:            true,
					Computed:            true,
				},
				"git_sparse_checkout_config": schema.StringAttribute{
					MarkdownDescription: "Git sparse checkout command line git cli options.",
					Optional:            true,
				},
				"include_sub_module": schema.BoolAttribute{
					MarkdownDescription: "Whether to include git submodules.",
					Optional:            true,
				},
				"is_private": schema.BoolAttribute{
					MarkdownDescription: "Whether the repository is private. Auth is required if the repository is private",
					Optional:            true,
					Computed:            true,
				},
				"ref": schema.StringAttribute{
					MarkdownDescription: "Git reference (branch, tag, or commit hash).",
					Optional:            true,
					Computed:            true,
				},
				"repo": schema.StringAttribute{
					MarkdownDescription: "Git repository URL.",
					Required:            true,
				},
				"working_dir": schema.StringAttribute{
					MarkdownDescription: "Working directory within the repository.",
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
				},
			},
			"owner_org": schema.StringAttribute{
				MarkdownDescription: "Organization the template belongs to",
				Computed:            true,
			},
			"template_name": schema.StringAttribute{
				MarkdownDescription: "Name of the workflow template.",
				Required:            true,
			},
			"source_config_kind": schema.StringAttribute{
				MarkdownDescription: "Source configuration kind (TERRAFORM, OPENTOFU, ANSIBLE_PLAYBOOK, HELM, KUBECTL, CLOUDFORMATION, CUSTOM).",
				Required:            true,
			},
			"is_public": schema.StringAttribute{
				MarkdownDescription: "Make template available to other organisations. Available values (\"0\" or \"1\")",
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
				MarkdownDescription: "Contextual tags to give context to your tags.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"shared_orgs_list": schema.ListAttribute{
				MarkdownDescription: "List of organizations the template is shared with.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"runtime_source": schema.SingleNestedAttribute{
				MarkdownDescription: "Runtime source configuration for the template.",
				Optional:            true,
				Computed:            true,
				Attributes:          WorkflowTemplateRuntimeSourceConfig(),
			},
			"vcs_triggers": schema.SingleNestedAttribute{
				MarkdownDescription: "VCS trigger configuration for the workflow.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "VCS provider type (GITHUB_COM, GITHUB_APP_CUSTOM, GITLAB_OAUTH_SSH, GITLAB_COM).",
						Required:            true,
					},
					"create_tag": schema.SingleNestedAttribute{
						MarkdownDescription: "Trigger configuration on tag creation in VCS",
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"create_revision": schema.SingleNestedAttribute{
								MarkdownDescription: "Create new revision on tag creation",
								Required:            true,
								Attributes: map[string]schema.Attribute{
									"enabled": schema.BoolAttribute{
										MarkdownDescription: "Whether to create revision when tag is created.",
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

package workflowsteptemplate

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
func (r *workflowStepTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a workflow step template resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: constants.Id,
				Computed:            true,
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"template_name": schema.StringAttribute{
				MarkdownDescription: "Name of the workflow step template.",
				Required:            true,
			},
			"template_type": schema.StringAttribute{
				MarkdownDescription: "Type of the template (WORKFLOW_STEP, IAC, IAC_GROUP, IAC_POLICY).",
				Computed:            true,
			},
			"is_active": schema.StringAttribute{
				MarkdownDescription: "Whether the template is active.",
				Optional:            true,
				Computed:            true,
			},
			"is_public": schema.StringAttribute{
				MarkdownDescription: "Whether the template is public.",
				Optional:            true,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Description, "workflow step template"),
				Optional:            true,
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Tags, "workflow step template"),
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
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
			"source_config_kind": schema.StringAttribute{
				MarkdownDescription: "Source configuration kind (DOCKER_IMAGE).",
				Required:            true,
			},
			"latest_revision": schema.Int32Attribute{
				MarkdownDescription: "Latest revision of the template.",
				Computed:            true,
			},
			"next_revision": schema.Int32Attribute{
				MarkdownDescription: "Next revision of the template.",
				Computed:            true,
			},
			"runtime_source": schema.SingleNestedAttribute{
				MarkdownDescription: "Runtime source configuration for the template.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"source_config_dest_kind": schema.StringAttribute{
						MarkdownDescription: "Destination kind for the source configuration.",
						Optional:            true,
						Computed:            true,
					},
					"additional_config": schema.MapAttribute{
						MarkdownDescription: "Additional configuration for the runtime source.",
						ElementType:         types.StringType,
						Optional:            true,
						Computed:            true,
					},
					"config": schema.SingleNestedAttribute{
						MarkdownDescription: "Configuration for the runtime source.",
						Optional:            true,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"is_private": schema.BoolAttribute{
								MarkdownDescription: "Whether the docker image is private.",
								Optional:            true,
								Computed:            true,
							},
							"auth": schema.StringAttribute{
								MarkdownDescription: "Authentication credentials for the docker image.",
								Optional:            true,
								Computed:            true,
								Sensitive:           true,
							},
							"docker_image": schema.StringAttribute{
								MarkdownDescription: "Docker image URI.",
								Optional:            true,
								Computed:            true,
							},
							"docker_registry_username": schema.StringAttribute{
								MarkdownDescription: "Username for the docker registry.",
								Optional:            true,
								Computed:            true,
							},
						},
					},
				},
			},
		},
	}
}

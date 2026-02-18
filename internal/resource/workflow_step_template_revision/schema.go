package workflowsteptemplaterevision

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
func (r *workflowStepTemplateRevisionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a workflow step template revision resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the revision in the format `templateId:revisionNumber`.",
				Computed:            true,
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"template_id": schema.StringAttribute{
				MarkdownDescription: "ID of the parent workflow step template.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"alias": schema.StringAttribute{
				MarkdownDescription: "Alias for the revision.",
				Optional:            true,
				Computed:            true,
			},
			"notes": schema.StringAttribute{
				MarkdownDescription: "Notes for the revision.",
				Optional:            true,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Description, "workflow step template revision"),
				Optional:            true,
				Computed:            true,
			},
			"template_type": schema.StringAttribute{
				MarkdownDescription: "Type of the template (WORKFLOW_STEP).",
				Computed:            true,
			},
			"source_config_kind": schema.StringAttribute{
				MarkdownDescription: "Source configuration kind (DOCKER_IMAGE).",
				Required:            true,
			},
			"is_active": schema.StringAttribute{
				MarkdownDescription: "Whether the revision is active.",
				Optional:            true,
				Computed:            true,
			},
			"is_public": schema.StringAttribute{
				MarkdownDescription: "Whether the revision is public.",
				Optional:            true,
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Tags, "workflow step template revision"),
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
			"runtime_source": schema.SingleNestedAttribute{
				MarkdownDescription: "Runtime source configuration for the revision.",
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
			"deprecation": schema.SingleNestedAttribute{
				MarkdownDescription: "Deprecation information for the revision.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"effective_date": schema.StringAttribute{
						MarkdownDescription: "Effective date for the deprecation.",
						Optional:            true,
						Computed:            true,
					},
					"message": schema.StringAttribute{
						MarkdownDescription: "Deprecation message.",
						Optional:            true,
						Computed:            true,
					},
				},
			},
		},
	}
}

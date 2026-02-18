package workflowsteptemplaterevision

import (
	"context"
	"fmt"

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
				MarkdownDescription: "ID of the revision in the format `templateId:revisionNumber`.",
				Required:            true,
			},
			"template_id": schema.StringAttribute{
				MarkdownDescription: "ID of the parent workflow step template.",
				Computed:            true,
			},
			"alias": schema.StringAttribute{
				MarkdownDescription: "Alias for the revision.",
				Computed:            true,
			},
			"notes": schema.StringAttribute{
				MarkdownDescription: "Notes for the revision.",
				Computed:            true,
			},
			"long_description": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Description, "workflow step template revision"),
				Computed:            true,
			},
			"template_type": schema.StringAttribute{
				MarkdownDescription: "Type of the template (WORKFLOW_STEP).",
				Computed:            true,
			},
			"source_config_kind": schema.StringAttribute{
				MarkdownDescription: "Source configuration kind (DOCKER_IMAGE).",
				Computed:            true,
			},
			"is_active": schema.StringAttribute{
				MarkdownDescription: "Whether the revision is active.",
				Computed:            true,
			},
			"is_public": schema.StringAttribute{
				MarkdownDescription: "Whether the revision is public.",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Tags, "workflow step template revision"),
				ElementType:         types.StringType,
				Computed:            true,
			},
			"context_tags": schema.MapAttribute{
				MarkdownDescription: "Contextual tags to give context to your tags.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"runtime_source": schema.SingleNestedAttribute{
				MarkdownDescription: "Runtime source configuration for the revision.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"source_config_dest_kind": schema.StringAttribute{
						MarkdownDescription: "Destination kind for the source configuration.",
						Computed:            true,
					},
					"additional_config": schema.MapAttribute{
						MarkdownDescription: "Additional configuration for the runtime source.",
						ElementType:         types.StringType,
						Computed:            true,
					},
					"config": schema.SingleNestedAttribute{
						MarkdownDescription: "Configuration for the runtime source.",
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"is_private": schema.BoolAttribute{
								MarkdownDescription: "Whether the docker image is private.",
								Computed:            true,
							},
							"auth": schema.StringAttribute{
								MarkdownDescription: "Authentication credentials for the docker image.",
								Computed:            true,
								Sensitive:           true,
							},
							"docker_image": schema.StringAttribute{
								MarkdownDescription: "Docker image URI.",
								Computed:            true,
							},
							"docker_registry_username": schema.StringAttribute{
								MarkdownDescription: "Username for the docker registry.",
								Computed:            true,
							},
						},
					},
				},
			},
			"deprecation": schema.SingleNestedAttribute{
				MarkdownDescription: "Deprecation information for the revision.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"effective_date": schema.StringAttribute{
						MarkdownDescription: "Effective date for the deprecation.",
						Computed:            true,
					},
					"message": schema.StringAttribute{
						MarkdownDescription: "Deprecation message.",
						Computed:            true,
					},
				},
			},
		},
	}
}

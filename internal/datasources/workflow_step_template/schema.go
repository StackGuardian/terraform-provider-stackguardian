package workflowsteptemplate

import (
	"context"
	"fmt"

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
				MarkdownDescription: "Name of the workflow step template.",
				Computed:            true,
			},
			"template_type": schema.StringAttribute{
				MarkdownDescription: "Type of the template (WORKFLOW_STEP).",
				Computed:            true,
			},
			"is_active": schema.StringAttribute{
				MarkdownDescription: "Whether the template is active.",
				Computed:            true,
			},
			"is_public": schema.StringAttribute{
				MarkdownDescription: "Whether the template is public.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Description, "workflow step template"),
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Tags, "workflow step template"),
				ElementType:         types.StringType,
				Computed:            true,
			},
			"context_tags": schema.MapAttribute{
				MarkdownDescription: "Contextual tags to give context to your tags.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"shared_orgs_list": schema.ListAttribute{
				MarkdownDescription: "List of organizations the template is shared with.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"source_config_kind": schema.StringAttribute{
				MarkdownDescription: "Source configuration kind (DOCKER_IMAGE).",
				Computed:            true,
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
		},
	}
}

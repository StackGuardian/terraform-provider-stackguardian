package workflowtemplate

import (
	"context"
	"fmt"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *workflowTemplateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to read a workflow template.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: constants.DatasourceId,
				Required:            true,
			},
			"owner_org": schema.StringAttribute{
				MarkdownDescription: "Organization the template belongs to.",
				Computed:            true,
			},
			"template_name": schema.StringAttribute{
				MarkdownDescription: "Name of the workflow template.",
				Computed:            true,
			},
			"template_type": schema.StringAttribute{
				MarkdownDescription: "Type of the template.",
				Computed:            true,
			},
			"source_config_kind": schema.StringAttribute{
				Computed: true,
			},
			"is_public": schema.StringAttribute{
				MarkdownDescription: "Whether the template is public.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Description, "workflow template"),
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Tags, "workflow template"),
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
			"runtime_source": schema.SingleNestedAttribute{
				MarkdownDescription: "Runtime source configuration for the template.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"source_config_dest_kind": schema.StringAttribute{
						Computed: true,
					},
					"config": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"auth": schema.StringAttribute{
								Computed:  true,
								Sensitive: true,
							},
							"git_core_auto_crlf": schema.BoolAttribute{
								Computed: true,
							},
							"git_sparse_checkout_config": schema.StringAttribute{
								Computed: true,
							},
							"include_sub_module": schema.BoolAttribute{
								Computed: true,
							},
							"is_private": schema.BoolAttribute{
								Computed: true,
							},
							"ref": schema.StringAttribute{
								Computed: true,
							},
							"repo": schema.StringAttribute{
								Computed: true,
							},
							"working_dir": schema.StringAttribute{
								Computed: true,
							},
						},
					},
				},
			},
			"vcs_triggers": schema.SingleNestedAttribute{
				MarkdownDescription: "VCS trigger configuration for the template.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Computed: true,
					},
					"create_tag": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"create_revision": schema.SingleNestedAttribute{
								Computed: true,
								Attributes: map[string]schema.Attribute{
									"enabled": schema.BoolAttribute{
										Computed: true,
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

package policy

import (
	"context"
	"fmt"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema defines the schema for the resource.
func (r *policyDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: constants.DatasourceId,
				Optional:            true,
				Computed:            true,
			},
			"resource_name": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.ResourceName, "policy") + constants.DatasourceResourceNameDeprecation,
				Optional:            true,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Description, "policy"),
				Computed:            true,
			},
			"approvers": schema.ListAttribute{
				MarkdownDescription: constants.Approvers,
				ElementType:         types.StringType,
				Computed:            true,
			},
			"number_of_approvals_required": schema.Int32Attribute{
				MarkdownDescription: constants.NumberOfApprovalsRequired,
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Tags, "policy"),
				ElementType:         types.StringType,
				Computed:            true,
			},
			"enforced_on": schema.ListAttribute{
				MarkdownDescription: constants.EnforcedOn,
				ElementType:         types.StringType,
				Computed:            true,
			},
			"policies_config": schema.ListNestedAttribute{
				MarkdownDescription: constants.PolicyConfig,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: fmt.Sprintf(constants.ResourceName, "policy config"),
							Computed:            true,
						},
						"skip": schema.BoolAttribute{
							MarkdownDescription: constants.PolicyConfigSkip,
							Computed:            true,
						},
						"on_fail": schema.StringAttribute{
							MarkdownDescription: constants.PolicyConfigOnFail,
							Computed:            true,
						},
						"on_pass": schema.StringAttribute{
							MarkdownDescription: constants.PolicyConfigOnPass,
							Computed:            true,
						},
						"policy_input_data": schema.SingleNestedAttribute{
							MarkdownDescription: constants.PolicyConfigInputData,
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"schema_type": schema.StringAttribute{
									MarkdownDescription: constants.PolicyConfigInputDataSchemaType,
									Computed:            true,
								},
								"data": schema.StringAttribute{
									MarkdownDescription: constants.PolicyConfigInputDataData,
									Computed:            true,
								},
							},
						},
						"policy_vcs_config": schema.SingleNestedAttribute{
							MarkdownDescription: constants.PolicyVCSConfig,
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"use_marketplace_template": schema.BoolAttribute{
									MarkdownDescription: constants.PolicyVCSConfigMarketplaceTemplate,
									Computed:            true,
								},
								"policy_template_id": schema.StringAttribute{
									MarkdownDescription: constants.PolicyVCSConfigTemplateId,
									Computed:            true,
								},
								"custom_source": schema.SingleNestedAttribute{
									MarkdownDescription: constants.PolicyVCSConfigCustomSource,
									Computed:            true,
									Attributes: map[string]schema.Attribute{
										"source_config_dest_kind": schema.StringAttribute{
											MarkdownDescription: constants.DiscoverySettingsBenchmarksRuntimeSourceSourceConfigDestKind,
											Computed:            true,
										},
										"source_config_kind": schema.StringAttribute{
											MarkdownDescription: "",
											Computed:            true,
										},
										"config": schema.SingleNestedAttribute{
											MarkdownDescription: "",
											Computed:            true,
											Attributes: map[string]schema.Attribute{
												"include_submodule": schema.BoolAttribute{
													MarkdownDescription: "",
													Computed:            true,
												},
												"ref": schema.StringAttribute{
													MarkdownDescription: "",
													Computed:            true,
												},
												"git_core_auto_crlf": schema.BoolAttribute{
													MarkdownDescription: "",
													Computed:            true,
												},
												"git_sparse_checkout_config": schema.StringAttribute{
													MarkdownDescription: "",
													Computed:            true,
												},
												"auth": schema.StringAttribute{
													MarkdownDescription: "",
													Computed:            true,
												},
												"working_dir": schema.StringAttribute{
													MarkdownDescription: "",
													Computed:            true,
												},
												"repo": schema.StringAttribute{
													MarkdownDescription: "",
													Computed:            true,
												},
												"is_private": schema.BoolAttribute{
													MarkdownDescription: "",
													Computed:            true,
												},
											},
										},
										"additional_config": schema.StringAttribute{
											MarkdownDescription: "",
											Computed:            true,
										},
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

package policy

import (
	"context"
	"fmt"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema defines the schema for the resource.
func (r *policyResrouce) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_name": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.ResourceName, "policy"),
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Description, "policy"),
				Optional:            true,
				Computed:            true,
			},
			"policy_type": schema.StringAttribute{
				MarkdownDescription: constants.PolicyType,
				Required:            true,
			},
			"approvers": schema.ListAttribute{
				MarkdownDescription: constants.Approvers,
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"number_of_approvals_required": schema.Int32Attribute{
				MarkdownDescription: constants.NumberOfApprovalsRequired,
				Optional:            true,
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Tags, "policy"),
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"enforced_on": schema.ListAttribute{
				MarkdownDescription: constants.EnforcedOn,
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"policies_config": schema.ListNestedAttribute{
				MarkdownDescription: constants.PolicyConfig,
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: fmt.Sprintf(constants.ResourceName, "policy config"),
							Required:            true,
						},
						"skip": schema.BoolAttribute{
							MarkdownDescription: constants.PolicyConfigSkip,
							Optional:            true,
						},
						"on_fail": schema.StringAttribute{
							MarkdownDescription: constants.PolicyConfigOnFail,
							Optional:            true,
						},
						"on_pass": schema.StringAttribute{
							MarkdownDescription: constants.PolicyConfigOnPass,
							Optional:            true,
						},
						"policy_input_data": schema.SingleNestedAttribute{
							MarkdownDescription: constants.PolicyConfigInputData,
							Optional:            true,
							Attributes: map[string]schema.Attribute{
								"schema_type": schema.StringAttribute{
									MarkdownDescription: constants.PolicyConfigInputDataSchemaType,
									Required:            true,
								},
								"data": schema.StringAttribute{
									MarkdownDescription: constants.PolicyConfigInputDataData,
									Required:            true,
								},
							},
						},
						"policy_vcs_config": schema.SingleNestedAttribute{
							MarkdownDescription: constants.PolicyVCSConfig,
							Optional:            true,
							Attributes: map[string]schema.Attribute{
								"use_marketplace_template": schema.BoolAttribute{
									MarkdownDescription: constants.PolicyVCSConfigMarketplaceTemplate,
									Required:            true,
								},
								"policy_template_id": schema.StringAttribute{
									MarkdownDescription: "Must atmost 100 characters",
									Optional:            true,
								},
								"custom_source": schema.SingleNestedAttribute{
									MarkdownDescription: constants.PolicyVCSConfigCustomSource,
									Optional:            true,
									Attributes: map[string]schema.Attribute{
										"source_config_dest_kind": schema.StringAttribute{
											MarkdownDescription: constants.DiscoverySettingsBenchmarksRuntimeSourceSourceConfigDestKind,
											Optional:            true,
											Computed:            true,
										},
										"source_config_kind": schema.StringAttribute{
											MarkdownDescription: "",
											Required:            true,
										},
										"config": schema.SingleNestedAttribute{
											MarkdownDescription: "",
											Optional:            true,
											Computed:            true,
											Attributes: map[string]schema.Attribute{
												"include_submodule": schema.BoolAttribute{
													MarkdownDescription: "",
													Optional:            true,
												},
												"ref": schema.StringAttribute{
													MarkdownDescription: "",
													Optional:            true,
												},
												"git_core_auto_crlf": schema.BoolAttribute{
													MarkdownDescription: "",
													Optional:            true,
													Computed:            true,
												},
												"git_sparse_checkout_config": schema.StringAttribute{
													MarkdownDescription: "",
													Optional:            true,
												},
												"auth": schema.StringAttribute{
													MarkdownDescription: "",
													Optional:            true,
												},
												"working_dir": schema.StringAttribute{
													MarkdownDescription: "",
													Optional:            true,
												},
												"repo": schema.StringAttribute{
													MarkdownDescription: "",
													Optional:            true,
												},
												"is_private": schema.BoolAttribute{
													MarkdownDescription: "",
													Optional:            true,
												},
											},
										},
										"additional_config": schema.StringAttribute{
											MarkdownDescription: "",
											Optional:            true,
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

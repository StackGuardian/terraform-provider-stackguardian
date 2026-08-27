package policy

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
func (r *policyResrouce) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of the resource. Use this attribute to reference the resource in other resources. Allowed characters are ^[a-zA-Z0-9_]+$",
				Computed:            true,
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"resource_name": schema.StringAttribute{
				MarkdownDescription: "Name of the policy.",
				Required:            true,
			},

			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the policy check. Must be less than 256 characters.",
				Optional:            true,
				Computed:            true,
			},
			"policy_type": schema.StringAttribute{
				MarkdownDescription: "Type of policy created <span style=\"background-color: #eff0f0; color: #e53835;\">GENERAL</span> or <span style=\"background-color: #eff0f0; color: #e53835;\">FILTER.INSIGHT</span>.",
				Required:            true,
			},
			"approvers": schema.ListAttribute{
				MarkdownDescription: "List of stackguardian users found in Access Management who can approve the policy check.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"number_of_approvals_required": schema.Int32Attribute{
				MarkdownDescription: "Number of approvals required for a policy check to run.",
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
				MarkdownDescription: "List of resources on which this policy is to be applied on. Resources supported: <span style=\"background-color: #eff0f0; color: #e53835;\">*</span>, <span style=\"background-color: #eff0f0; color: #e53835;\">/wfgrps/<grp>[/<subgrp>...]</span>, <span style=\"background-color: #eff0f0; color: #e53835;\">/stacks/<stackId></span>, <span style=\"background-color: #eff0f0; color: #e53835;\">/integrations/<integrationId></span>",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"policies_config": schema.ListNestedAttribute{
				MarkdownDescription: "Policy Rules configuration.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the policy rule.",
							Required:            true,
						},
						"skip": schema.BoolAttribute{
							MarkdownDescription: "If true, the policy rule will be skipped.",
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
							MarkdownDescription: "Reference a policy template",
							Optional:            true,
							Attributes: map[string]schema.Attribute{
								"use_marketplace_template": schema.BoolAttribute{
									MarkdownDescription: "If true use a policy template or use a custom source.",
									Required:            true,
								},
								"policy_template_id": schema.StringAttribute{
									MarkdownDescription: "",
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

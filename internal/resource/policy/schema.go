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
		MarkdownDescription: "Manages a policy: a guardrail evaluated during workflow runs, which can block or flag a run that violates it.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: constants.Id,
				Computed:            true,
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
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
									MarkdownDescription: constants.PolicyVCSConfigTemplateId,
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
											MarkdownDescription: constants.PolicyVCSConfigCustomSourceSourceConfigKind,
											Required:            true,
										},
										"config": schema.SingleNestedAttribute{
											MarkdownDescription: constants.PolicyVCSConfigCustomSourceConfig,
											Optional:            true,
											Computed:            true,
											Attributes: map[string]schema.Attribute{
												"include_submodule": schema.BoolAttribute{
													MarkdownDescription: constants.PolicyVCSConfigCustomSourceIncludeSubmodule,
													Optional:            true,
												},
												"ref": schema.StringAttribute{
													MarkdownDescription: constants.PolicyVCSConfigCustomSourceRef,
													Optional:            true,
												},
												"git_core_auto_crlf": schema.BoolAttribute{
													MarkdownDescription: constants.PolicyVCSConfigCustomSourceGitCoreAutoCRLF,
													Optional:            true,
													Computed:            true,
												},
												"git_sparse_checkout_config": schema.StringAttribute{
													MarkdownDescription: constants.PolicyVCSConfigCustomSourceGitSparseCheckoutConfig,
													Optional:            true,
												},
												"auth": schema.StringAttribute{
													MarkdownDescription: constants.PolicyVCSConfigCustomSourceAuth,
													Optional:            true,
												},
												"working_dir": schema.StringAttribute{
													MarkdownDescription: constants.PolicyVCSConfigCustomSourceWorkingDir,
													Optional:            true,
												},
												"repo": schema.StringAttribute{
													MarkdownDescription: constants.PolicyVCSConfigCustomSourceRepo,
													Optional:            true,
												},
												"is_private": schema.BoolAttribute{
													MarkdownDescription: constants.PolicyVCSConfigCustomSourceIsPrivate,
													Optional:            true,
												},
											},
										},
										"additional_config": schema.StringAttribute{
											MarkdownDescription: constants.PolicyVCSConfigAdditionalConfig,
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

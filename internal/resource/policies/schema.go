package policies

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema defines the schema for the resource.
func (r *policyResrouce) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_name": schema.StringAttribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
			},
			"approvers": schema.ListAttribute{
				MarkdownDescription: "",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"number_of_approvals_required": schema.Int32Attribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"enforced_on": schema.ListAttribute{
				MarkdownDescription: "",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"policies_config": schema.ListNestedAttribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Must be atmost 100 characters. Should match ^[-a-zA-Z0-9_]+$",
							Required:            true,
						},
						"skip": schema.BoolAttribute{
							MarkdownDescription: "",
							Optional:            true,
						},
						"on_fail": schema.StringAttribute{
							MarkdownDescription: `<span style="background-color: #eff0f0; color: #e53835;">FAIL</span>,
							<span style="background-color: #eff0f0; color: #e53835;">WARN</span>,
							<span style="background-color: #eff0f0; color: #e53835;">PASS</span>,
							<span style="background-color: #eff0f0; color: #e53835;">APPROVAL_REQUIRED</span>`,
							Required: true,
						},
						"on_pass": schema.StringAttribute{
							MarkdownDescription: `<span style="background-color: #eff0f0; color: #e53835;">FAIL</span>,
							<span style="background-color: #eff0f0; color: #e53835;">WARN</span>,
							<span style="background-color: #eff0f0; color: #e53835;">PASS</span>,
							<span style="background-color: #eff0f0; color: #e53835;">APPROVAL_REQUIRED</span>`,
							Required: true,
						},
						"policy_input_data": schema.SingleNestedAttribute{
							MarkdownDescription: "",
							Optional:            true,
							Attributes: map[string]schema.Attribute{
								"schema_type": schema.StringAttribute{
									MarkdownDescription: `<span style="background-color: #eff0f0; color: #e53835;">FORM_JSONSCHEMA</span>,
									<span style="background-color: #eff0f0; color: #e53835;">RAW_JSON</span>,
									<span style="background-color: #eff0f0; color: #e53835;">TIRITH_JSON</span>,
									<span style="background-color: #eff0f0; color: #e53835;">NONE</span>`,
									Required: true,
								},
								"data": schema.StringAttribute{
									MarkdownDescription: "JSON encoded string",
									Required:            true,
								},
							},
						},
						"policy_vcs_config": schema.SingleNestedAttribute{
							MarkdownDescription: "Version control config",
							Required:            true,
							Attributes: map[string]schema.Attribute{
								"use_marketplace_template": schema.BoolAttribute{
									Required: true,
								},
								"policy_template_id": schema.StringAttribute{
									MarkdownDescription: "Must atmost 100 characters",
									Optional:            true,
								},
								"custom_source": schema.SingleNestedAttribute{
									MarkdownDescription: "",
									Optional:            true,
									Attributes: map[string]schema.Attribute{
										"source_config_dest_kind": schema.StringAttribute{
											MarkdownDescription: "",
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

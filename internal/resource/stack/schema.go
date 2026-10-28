package stack

import (
	"context"
	"fmt"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Shared schema components
var ministepsNotificationRecipients = schema.ListNestedAttribute{
	Optional: true,
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"recipients": schema.ListAttribute{
				MarkdownDescription: constants.MiniStepsNotificationsRecipients,
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	},
}

var ministepsWebhooks = schema.ListNestedAttribute{
	Optional: true,
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"webhook_name": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWebhookName,
				Required:            true,
			},
			"webhook_url": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWebhookURL,
				Required:            true,
			},
			"webhook_secret": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWebhookSecret,
				Optional:            true,
				Sensitive:           true,
			},
		},
	},
}

var ministepsWorkflowChaining = schema.ListNestedAttribute{
	Optional: true,
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"workflow_group_id": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWfChainingWorkflowGroupId,
				Required:            true,
			},
			"stack_id": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWfChainingStackId,
				Optional:            true,
			},
			"stack_run_payload": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWfChainingStackPayload,
				Optional:            true,
				Validators: []validator.String{
					validators.JSONObject(),
				},
			},
			"workflow_id": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWfChainingWorkflowId,
				Optional:            true,
			},
			"workflow_run_payload": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWfChainingWorkflowPayload,
				Optional:            true,
				Validators: []validator.String{
					validators.JSONObject(),
				},
			},
		},
	},
}

var envVarsAttrs = map[string]schema.Attribute{
	"config": schema.SingleNestedAttribute{
		MarkdownDescription: constants.EnvVarConfig,
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"var_name": schema.StringAttribute{
				MarkdownDescription: constants.EnvVarConfigVarName,
				Required:            true,
			},
			"secret_id": schema.StringAttribute{
				MarkdownDescription: constants.EnvVarConfigSecretId,
				Optional:            true,
			},
			"text_value": schema.StringAttribute{
				MarkdownDescription: constants.EnvVarConfigTextValue,
				Optional:            true,
				Sensitive:           true,
			},
		},
	},
	"kind": schema.StringAttribute{
		MarkdownDescription: constants.EnvVarKind,
		Required:            true,
	},
}

var mountPointAttrs = map[string]schema.Attribute{
	"source": schema.StringAttribute{
		MarkdownDescription: constants.MountPointSource,
		Optional:            true,
	},
	"target": schema.StringAttribute{
		MarkdownDescription: constants.MountPointTarget,
		Optional:            true,
	},
	"read_only": schema.BoolAttribute{
		MarkdownDescription: constants.MountPointReadOnly,
		Optional:            true,
	},
}

var wfStepsConfigNestedObj = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		"name": schema.StringAttribute{
			MarkdownDescription: constants.WfStepName,
			Required:            true,
		},
		"environment_variables": schema.ListNestedAttribute{
			MarkdownDescription: constants.WfStepEnvVars,
			Optional:            true,
			NestedObject:        schema.NestedAttributeObject{Attributes: envVarsAttrs},
		},
		"approval": schema.BoolAttribute{
			MarkdownDescription: constants.WfStepApproval,
			Optional:            true,
		},
		"timeout": schema.Int64Attribute{
			MarkdownDescription: constants.WfStepTimeout,
			Optional:            true,
		},
		"cmd_override": schema.StringAttribute{
			MarkdownDescription: constants.WfStepCmdOverride,
			Optional:            true,
		},
		"mount_points": schema.ListNestedAttribute{
			MarkdownDescription: constants.WfStepMountPoints,
			Optional:            true,
			NestedObject:        schema.NestedAttributeObject{Attributes: mountPointAttrs},
		},
		"wf_step_template_id": schema.StringAttribute{
			MarkdownDescription: constants.WfStepTemplateId,
			Required:            true,
		},
		"wf_step_input_data": schema.SingleNestedAttribute{
			MarkdownDescription: constants.WfStepInputData,
			Optional:            true,
			Attributes: map[string]schema.Attribute{
				"schema_type": schema.StringAttribute{
					MarkdownDescription: constants.WfStepInputDataSchemaType,
					Optional:            true,
				},
				"data": schema.StringAttribute{
					MarkdownDescription: constants.WfStepInputDataData,
					Optional:            true,
					Validators: []validator.String{
						validators.JSONObject(),
					},
				},
			},
		},
	},
}

var deploymentPlatformConfigAttrs = map[string]schema.Attribute{
	"kind": schema.StringAttribute{
		MarkdownDescription: constants.DeploymentPlatformKind,
		Required:            true,
	},
	"config": schema.SingleNestedAttribute{
		MarkdownDescription: constants.DeploymentPlatformConfigDetails,
		Required:            true,
		Attributes: map[string]schema.Attribute{
			"integration_id": schema.StringAttribute{
				MarkdownDescription: constants.DeploymentPlatformIntegrationId,
				Required:            true,
			},
			"profile_name": schema.StringAttribute{
				MarkdownDescription: constants.DeploymentPlatformProfileName,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	},
}

// terraformConfigAttrs' fields are all Optional+Computed with UseStateForUnknown:
// mergeTerraformConfig (see model.go) fills any field the user leaves unset from
// the stack template revision, then the workflow template revision — a value that
// can come from a template layer, not just the user, needs Computed or Terraform
// core forces it to null on every plan while apply returns the merged value
// ("Provider produced inconsistent result after apply").
var terraformConfigAttrs = map[string]schema.Attribute{
	"terraform_version": schema.StringAttribute{
		MarkdownDescription: constants.TerraformVersion,
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"drift_check": schema.BoolAttribute{
		MarkdownDescription: constants.TerraformDriftCheck,
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.UseStateForUnknown(),
		},
	},
	"drift_cron": schema.StringAttribute{
		MarkdownDescription: constants.TerraformDriftCron,
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"managed_terraform_state": schema.BoolAttribute{
		MarkdownDescription: constants.TerraformManagedState,
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.UseStateForUnknown(),
		},
	},
	"approval_pre_apply": schema.BoolAttribute{
		MarkdownDescription: constants.TerraformApprovalPreApply,
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.UseStateForUnknown(),
		},
	},
	"terraform_plan_options": schema.StringAttribute{
		MarkdownDescription: constants.TerraformPlanOptions,
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"terraform_init_options": schema.StringAttribute{
		MarkdownDescription: constants.TerraformInitOptions,
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"terraform_bin_path": schema.ListNestedAttribute{
		MarkdownDescription: constants.TerraformBinPath,
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.List{
			listplanmodifier.UseStateForUnknown(),
		},
		NestedObject: schema.NestedAttributeObject{Attributes: mountPointAttrs},
	},
	"timeout": schema.Int64Attribute{
		MarkdownDescription: constants.TerraformTimeout,
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.UseStateForUnknown(),
		},
	},
	"post_apply_wf_steps_config": schema.ListNestedAttribute{
		MarkdownDescription: constants.TerraformPostApplyWfSteps,
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.List{
			listplanmodifier.UseStateForUnknown(),
		},
		NestedObject: wfStepsConfigNestedObj,
	},
	"pre_apply_wf_steps_config": schema.ListNestedAttribute{
		MarkdownDescription: constants.TerraformPreApplyWfSteps,
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.List{
			listplanmodifier.UseStateForUnknown(),
		},
		NestedObject: wfStepsConfigNestedObj,
	},
	"pre_plan_wf_steps_config": schema.ListNestedAttribute{
		MarkdownDescription: constants.TerraformPrePlanWfSteps,
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.List{
			listplanmodifier.UseStateForUnknown(),
		},
		NestedObject: wfStepsConfigNestedObj,
	},
	"post_plan_wf_steps_config": schema.ListNestedAttribute{
		MarkdownDescription: constants.TerraformPostPlanWfSteps,
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.List{
			listplanmodifier.UseStateForUnknown(),
		},
		NestedObject: wfStepsConfigNestedObj,
	},
	"pre_init_hooks": schema.ListAttribute{
		MarkdownDescription: constants.TerraformPreInitHooks,
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.List{
			listplanmodifier.UseStateForUnknown(),
		},
		ElementType: types.StringType,
	},
	"pre_plan_hooks": schema.ListAttribute{
		MarkdownDescription: constants.TerraformPrePlanHooks,
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.List{
			listplanmodifier.UseStateForUnknown(),
		},
		ElementType: types.StringType,
	},
	"post_plan_hooks": schema.ListAttribute{
		MarkdownDescription: constants.TerraformPostPlanHooks,
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.List{
			listplanmodifier.UseStateForUnknown(),
		},
		ElementType: types.StringType,
	},
	"pre_apply_hooks": schema.ListAttribute{
		MarkdownDescription: constants.TerraformPreApplyHooks,
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.List{
			listplanmodifier.UseStateForUnknown(),
		},
		ElementType: types.StringType,
	},
	"post_apply_hooks": schema.ListAttribute{
		MarkdownDescription: constants.TerraformPostApplyHooks,
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.List{
			listplanmodifier.UseStateForUnknown(),
		},
		ElementType: types.StringType,
	},
	"run_pre_init_hooks_on_drift": schema.BoolAttribute{
		MarkdownDescription: constants.TerraformRunPreInitHooksOnDrift,
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.UseStateForUnknown(),
		},
	},
}

// actionsAttrs defines the schema attributes for a single action value inside
// the "actions" top-level attribute.
var actionsAttrs = map[string]schema.Attribute{
	"name": schema.StringAttribute{
		MarkdownDescription: "Name of the action.",
		Required:            true,
	},
	"description": schema.StringAttribute{
		MarkdownDescription: "Description of the action.",
		Optional:            true,
	},
	// Computed only: users don't set it. Whether an action is the user's own is
	// decided by whether actions is declared in config, which only a plan
	// modifier can see; a static schema Default can't express that. See
	// actionDefaultModifier.
	"default": schema.BoolAttribute{
		MarkdownDescription: constants.StackActionDefault,
		Computed:            true,
		PlanModifiers: []planmodifier.Bool{
			actionDefaultModifier{},
		},
	},
	"order": schema.MapNestedAttribute{
		MarkdownDescription: "Execution order for workflows in this action, keyed by each workflow's `id` (its UUID) as defined in the stack template revision's `workflows_config.workflows` — the same `id` declared in this stack's `workflows_config.workflows`.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"parameters": schema.SingleNestedAttribute{
					MarkdownDescription: "Run configuration for the workflow.",
					Optional:            true,
					Attributes: map[string]schema.Attribute{
						"terraform_action": schema.SingleNestedAttribute{
							Optional: true,
							Attributes: map[string]schema.Attribute{
								"action": schema.StringAttribute{
									MarkdownDescription: "Terraform action (apply, destroy, plan).",
									Optional:            true,
								},
							},
						},
						"deployment_platform_config": schema.ListNestedAttribute{
							MarkdownDescription: constants.WfDeploymentPlatformConfig,
							Optional:            true,
							NestedObject:        schema.NestedAttributeObject{Attributes: deploymentPlatformConfigAttrs},
						},
						"wf_steps_config": schema.ListNestedAttribute{
							MarkdownDescription: constants.WfStepsConfig,
							Optional:            true,
							NestedObject:        wfStepsConfigNestedObj,
						},
						"environment_variables": schema.ListNestedAttribute{
							MarkdownDescription: constants.WfEnvironmentVariables,
							Optional:            true,
							NestedObject:        schema.NestedAttributeObject{Attributes: envVarsAttrs},
						},
					},
				},
				"dependencies": schema.ListNestedAttribute{
					MarkdownDescription: "Workflow dependencies defining execution order. Defaults to `[]`.",
					Optional:            true,
					// Computed with a [] default, mirroring the API, which stores [] when no
					// dependencies are sent; without it an omitted value plans null but reads back [].
					Computed: true,
					Default: listdefault.StaticValue(types.ListValueMust(
						types.ObjectType{AttrTypes: ActionDependencyModel{}.AttributeTypes()}, []attr.Value{},
					)),
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"id": schema.StringAttribute{
								MarkdownDescription: "Workflow this depends on, as its `id` (its UUID) as defined in the stack template revision's `workflows_config.workflows` — the same `id` declared in this stack's `workflows_config.workflows`.",
								Required:            true,
							},
							"condition": schema.SingleNestedAttribute{
								Optional: true,
								Attributes: map[string]schema.Attribute{
									"latest_status": schema.StringAttribute{
										MarkdownDescription: "Required latest status of the dependency (e.g. COMPLETED).",
										Required:            true,
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

// Schema defines the schema for the stack resource.
func (r *stackResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a stack: a group of workflows created from a stack template revision and run together through the stack's actions.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: constants.Id,
				Required:            true,
				// The SDK has no way to change a stack's id via update (PatchedStack has
				// no Id field), so a change must recreate the resource.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"workflow_group_id": schema.StringAttribute{
				MarkdownDescription: "ID of the workflow group this stack belongs to.",
				Required:            true,
				// A stack lives inside a workflow group; the platform has no move
				// operation, so changing the group must recreate the stack.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_name": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.ResourceName, "stack"),
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Description, "stack"),
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Tags, "stack"),
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"actions": schema.MapNestedAttribute{
				MarkdownDescription: "Actions define the sequence in which the workflows in the Stack are executed. When left unset, inherits the actions resolved from the stack template revision (its own actions verbatim, or a generated apply/plan/destroy set — see the template revision docs); when set, this value is used as-is instead of the template's. Must have at least one entry if set — an empty map is rejected, matching the API's own requirement that a stack always have at least one action.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Map{
					mapvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: actionsAttrs,
				},
			},
			"template_group_id": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateGroupId,
				Required:            true,
				Validators: []validator.String{
					validators.QualifiedTemplateRevisionId(),
				},
			},
			"workflows_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Workflows configuration for the stack. Every workflow defined on the stack template revision (`template_group_id`) must be declared in `workflows.*.id`, in the order the revision defines them. `terraform plan` fails if one is missing or out of order (or `terraform apply`, when `template_group_id` is only known at apply time).",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"workflows": schema.ListNestedAttribute{
						MarkdownDescription: "List of workflows in the stack. Must include the id of every workflow defined on the referenced stack template revision, in the same order.",
						Required:            true,
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									MarkdownDescription: "UUID of the workflow, as defined on the stack template revision.",
									Required:            true,
								},
								"workflow_id": schema.StringAttribute{
									MarkdownDescription: "Resource id the platform assigns to this workflow, derived from the resolved workflow template name and this workflow's `id`.",
									Computed:            true,
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.UseStateForUnknown(),
									},
								},
								"resource_name": schema.StringAttribute{
									MarkdownDescription: "Name of the workflow resource.",
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.UseStateForUnknown(),
									},
								},
								"description": schema.StringAttribute{
									MarkdownDescription: fmt.Sprintf(constants.Description, "workflow"),
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.UseStateForUnknown(),
									},
								},
								"tags": schema.ListAttribute{
									MarkdownDescription: fmt.Sprintf(constants.Tags, "workflow"),
									ElementType:         types.StringType,
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.List{
										listplanmodifier.UseStateForUnknown(),
									},
								},
								"wf_type": schema.StringAttribute{
									MarkdownDescription: constants.WorkflowType,
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.UseStateForUnknown(),
									},
								},
								"parallel_execution": schema.StringAttribute{
									MarkdownDescription: "Enable or disable parallel execution (enabled/disabled).",
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.UseStateForUnknown(),
									},
								},
								"wf_steps_config": schema.ListNestedAttribute{
									MarkdownDescription: constants.WfStepsConfig,
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.List{
										listplanmodifier.UseStateForUnknown(),
									},
									NestedObject: wfStepsConfigNestedObj,
								},
								"terraform_config": schema.SingleNestedAttribute{
									MarkdownDescription: constants.TerraformConfig,
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.Object{
										objectplanmodifier.UseStateForUnknown(),
									},
									Attributes: terraformConfigAttrs,
								},
								"environment_variables": schema.ListNestedAttribute{
									MarkdownDescription: constants.WfEnvironmentVariables,
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.List{
										listplanmodifier.UseStateForUnknown(),
									},
									NestedObject: schema.NestedAttributeObject{
										Attributes: envVarsAttrs,
									},
								},
								"deployment_platform_config": schema.ListNestedAttribute{
									MarkdownDescription: constants.WfDeploymentPlatformConfig,
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.List{
										listplanmodifier.UseStateForUnknown(),
									},
									NestedObject: schema.NestedAttributeObject{
										Attributes: deploymentPlatformConfigAttrs,
									},
								},
								"vcs_config": schema.SingleNestedAttribute{
									MarkdownDescription: constants.WorkflowVcsConfig,
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.Object{
										objectplanmodifier.UseStateForUnknown(),
									},
									Attributes: map[string]schema.Attribute{
										"iac_vcs_config": schema.SingleNestedAttribute{
											MarkdownDescription: "IaC VCS configuration. Not user-editable — always inherited from the matching workflow on the stack template revision.",
											Computed:            true,
											PlanModifiers: []planmodifier.Object{
												objectplanmodifier.UseStateForUnknown(),
											},
											Attributes: map[string]schema.Attribute{
												"use_marketplace_template": schema.BoolAttribute{
													MarkdownDescription: constants.WorkflowUseMarketplaceTemplate,
													Computed:            true,
												},
												"iac_template_id": schema.StringAttribute{
													MarkdownDescription: "Workflow template revision this workflow is created from, as `/<org>/<workflow-template-id>:<revision>`. Inherited from the stack template revision.",
													Computed:            true,
												},
											},
										},
										// iac_input_data = {} (no fields) is valid: it overrides inputs inherited
										// from the stack template revision with "No Inputs" (an explicitly empty
										// iacInputData, which the API keeps rather than replacing with the template's).
										// Once any field is set, schema_type and data are both required
										// (AlsoRequires on each field); schema_type is Optional only to allow {}.
										"iac_input_data": schema.SingleNestedAttribute{
											MarkdownDescription: constants.WorkflowIacInputData + " Set to `{}` for no inputs. When any field is set, `schema_type` and `data` are both required.",
											Optional:            true,
											Attributes: map[string]schema.Attribute{
												"schema_id": schema.StringAttribute{
													MarkdownDescription: constants.WorkflowIacInputDataSchemaId,
													Optional:            true,
													Validators: []validator.String{
														stringvalidator.AlsoRequires(
															path.MatchRelative().AtParent().AtName("schema_type"),
															path.MatchRelative().AtParent().AtName("data"),
														),
													},
												},
												"schema_type": schema.StringAttribute{
													MarkdownDescription: constants.WorkflowIacInputDataSchemaType,
													Optional:            true,
													Validators: []validator.String{
														stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("data")),
													},
												},
												"data": schema.StringAttribute{
													MarkdownDescription: constants.WorkflowIacInputDataData,
													Optional:            true,
													Validators: []validator.String{
														validators.JSONObject(),
														stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("schema_type")),
													},
												},
											},
										},
									},
								},
								"approvers": schema.ListAttribute{
									MarkdownDescription: constants.WfApprovers,
									ElementType:         types.StringType,
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.List{
										listplanmodifier.UseStateForUnknown(),
									},
								},
								"number_of_approvals_required": schema.Int64Attribute{
									MarkdownDescription: constants.WfNumberOfApprovals,
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.Int64{
										int64planmodifier.UseStateForUnknown(),
									},
								},
								"user_job_cpu": schema.Int64Attribute{
									MarkdownDescription: constants.WfUserJobCPU,
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.Int64{
										int64planmodifier.UseStateForUnknown(),
									},
								},
								"user_job_memory": schema.Int64Attribute{
									MarkdownDescription: constants.WfUserJobMemory,
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.Int64{
										int64planmodifier.UseStateForUnknown(),
									},
								},
								"user_schedules": schema.ListNestedAttribute{
									MarkdownDescription: constants.WfUserSchedules,
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.List{
										listplanmodifier.UseStateForUnknown(),
									},
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"name": schema.StringAttribute{
												MarkdownDescription: constants.UserScheduleName,
												Computed:            true,
												PlanModifiers: []planmodifier.String{
													stringplanmodifier.UseStateForUnknown(),
												},
											},
											"desc": schema.StringAttribute{
												MarkdownDescription: constants.UserScheduleDesc,
												Optional:            true,
											},
											"cron": schema.StringAttribute{
												MarkdownDescription: constants.UserScheduleCron,
												Required:            true,
											},
											"state": schema.StringAttribute{
												MarkdownDescription: constants.UserScheduleState,
												Required:            true,
											},
										},
									},
								},
								"mini_steps": schema.SingleNestedAttribute{
									MarkdownDescription: constants.WfMiniSteps,
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.Object{
										objectplanmodifier.UseStateForUnknown(),
									},
									Attributes: map[string]schema.Attribute{
										"notifications": schema.SingleNestedAttribute{
											MarkdownDescription: constants.MiniStepsNotifications,
											Optional:            true,
											Attributes: map[string]schema.Attribute{
												"email": schema.SingleNestedAttribute{
													MarkdownDescription: constants.MiniStepsNotificationsEmail,
													Optional:            true,
													Attributes: map[string]schema.Attribute{
														"approval_required": ministepsNotificationRecipients,
														"cancelled":         ministepsNotificationRecipients,
														"completed":         ministepsNotificationRecipients,
														"drift_detected":    ministepsNotificationRecipients,
														"errored":           ministepsNotificationRecipients,
													},
												},
											},
										},
										"webhooks": schema.SingleNestedAttribute{
											MarkdownDescription: constants.MiniStepsWebhooks,
											Optional:            true,
											Attributes: map[string]schema.Attribute{
												"approval_required": ministepsWebhooks,
												"cancelled":         ministepsWebhooks,
												"completed":         ministepsWebhooks,
												"drift_detected":    ministepsWebhooks,
												"errored":           ministepsWebhooks,
											},
										},
										"wf_chaining": schema.SingleNestedAttribute{
											MarkdownDescription: constants.MiniStepsWorkflowChaining,
											Optional:            true,
											Attributes: map[string]schema.Attribute{
												"completed": ministepsWorkflowChaining,
												"errored":   ministepsWorkflowChaining,
											},
										},
									},
								},
								"context_tags": schema.MapAttribute{
									MarkdownDescription: fmt.Sprintf(constants.ContextTags, "workflow"),
									ElementType:         types.StringType,
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.Map{
										mapplanmodifier.UseStateForUnknown(),
									},
								},
								"runner_constraints": schema.SingleNestedAttribute{
									MarkdownDescription: constants.WorkflowRunnerConstraints,
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.Object{
										objectplanmodifier.UseStateForUnknown(),
									},
									Attributes: map[string]schema.Attribute{
										"type": schema.StringAttribute{
											MarkdownDescription: constants.RunnerConstraintsType,
											Required:            true,
										},
										"names": schema.ListAttribute{
											MarkdownDescription: constants.RunnerConstraintsNames,
											ElementType:         types.StringType,
											Optional:            true,
										},
									},
								},
							},
						},
					},
				},
			},
			"context_tags": schema.MapAttribute{
				MarkdownDescription: fmt.Sprintf(constants.ContextTags, "stack"),
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

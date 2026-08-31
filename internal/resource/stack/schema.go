package stack

import (
	"context"
	"fmt"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
			},
			"workflow_id": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWfChainingWorkflowId,
				Optional:            true,
			},
			"workflow_run_payload": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWfChainingWorkflowPayload,
				Optional:            true,
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

// actionsAttrs defines the schema attributes for a single action value, shared
// between the "default_actions" (computed, server-reflected) and "custom_actions"
// (user-authored) top-level attributes — both describe the same shape.
var actionsAttrs = map[string]schema.Attribute{
	"name": schema.StringAttribute{
		MarkdownDescription: "Name of the action.",
		Required:            true,
	},
	"description": schema.StringAttribute{
		MarkdownDescription: "Description of the action.",
		Optional:            true,
	},
	"order": schema.MapNestedAttribute{
		MarkdownDescription: "Execution order for workflows in this action. Key is the workflow ID.",
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
							Optional:     true,
							NestedObject: schema.NestedAttributeObject{Attributes: deploymentPlatformConfigAttrs},
						},
						"wf_steps_config": schema.ListNestedAttribute{
							Optional:     true,
							NestedObject: wfStepsConfigNestedObj,
						},
						"environment_variables": schema.ListNestedAttribute{
							Optional:     true,
							NestedObject: schema.NestedAttributeObject{Attributes: envVarsAttrs},
						},
					},
				},
				"dependencies": schema.ListNestedAttribute{
					MarkdownDescription: "Workflow dependencies defining execution order.",
					Optional:            true,
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"id": schema.StringAttribute{
								MarkdownDescription: "ID of the workflow this depends on.",
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
		MarkdownDescription: "Manages a stack resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Id for the resource. Use it to reference the resource in other resources. Allowed characters are ^[a-zA-Z0-9_]+$",
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
				MarkdownDescription: "Name of the stack.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the stack. Must be less than 256 characters.",
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
			"environment_variables": schema.ListNestedAttribute{
				MarkdownDescription: "Environment variables for the stack.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: envVarsAttrs,
				},
			},
			"deployment_platform_config": schema.ListNestedAttribute{
				MarkdownDescription: "Deployment platform configuration.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: deploymentPlatformConfigAttrs,
				},
			},
			"default_actions": schema.MapNestedAttribute{
				MarkdownDescription: "Actions define the sequence in which the workflows in the Stack are executed. This reflects the full set of actions on the stack (built-in and custom); use `custom_actions` to author your own.",
				Computed:            true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: actionsAttrs,
				},
			},
			"custom_actions": schema.MapNestedAttribute{
				MarkdownDescription: "User-authored actions to add to the stack. Same shape as `default_actions`. Computed because the API returns one merged actions map with no \"is custom\" marker other than Default=false, so this is re-derived from the server response on every read.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: actionsAttrs,
				},
			},
			"template_group_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the template group that this Stack is mapped to. The format is `/<org>/<template_group_revision_id>`. Required — a stack cannot be created without a template.",
				Required:            true,
			},
			"workflows_config": schema.SingleNestedAttribute{
				MarkdownDescription: "Workflows configuration for the stack.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"workflows": schema.ListNestedAttribute{
						MarkdownDescription: "List of workflows in the stack.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									MarkdownDescription: "UUID of the workflow slot as defined in the stack template. Required so the provider can match this entry against the stack template revision's workflows_config for merging.",
									Required:            true,
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
									MarkdownDescription: "Description of the workflow.",
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.UseStateForUnknown(),
									},
								},
								"tags": schema.ListAttribute{
									MarkdownDescription: "Tags for the workflow.",
									ElementType:         types.StringType,
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.List{
										listplanmodifier.UseStateForUnknown(),
									},
								},
								"wf_type": schema.StringAttribute{
									MarkdownDescription: "Type of workflow.",
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
									MarkdownDescription: "Workflow steps configuration.",
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
									MarkdownDescription: "Environment variables.",
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
									MarkdownDescription: "Deployment platform configuration.",
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.List{
										listplanmodifier.UseStateForUnknown(),
									},
									NestedObject: schema.NestedAttributeObject{
										Attributes: deploymentPlatformConfigAttrs,
									},
								},
								"template_id": schema.StringAttribute{
									MarkdownDescription: "ID of the workflow template that this workflow is based on.",
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.UseStateForUnknown(),
									},
								},
								"vcs_config": schema.SingleNestedAttribute{
									MarkdownDescription: "VCS (version control) configuration for the workflow.",
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.Object{
										objectplanmodifier.UseStateForUnknown(),
									},
									Attributes: map[string]schema.Attribute{
										"iac_vcs_config": schema.SingleNestedAttribute{
											MarkdownDescription: "IaC VCS configuration. Not user-editable — always inherited from the matched workflow slot on the stack template revision.",
											Computed:            true,
											PlanModifiers: []planmodifier.Object{
												objectplanmodifier.UseStateForUnknown(),
											},
											Attributes: map[string]schema.Attribute{
												"use_marketplace_template": schema.BoolAttribute{
													MarkdownDescription: "Whether to use a marketplace template.",
													Computed:            true,
												},
												"iac_template_id": schema.StringAttribute{
													MarkdownDescription: "ID of the IaC template from the marketplace.",
													Computed:            true,
												},
											},
										},
										"iac_input_data": schema.SingleNestedAttribute{
											MarkdownDescription: "IaC input data for the workflow.",
											Optional:            true,
											Attributes: map[string]schema.Attribute{
												"schema_id": schema.StringAttribute{
													Optional: true,
												},
												"schema_type": schema.StringAttribute{
													Required: true,
												},
												"data": schema.StringAttribute{
													MarkdownDescription: "Input data as a JSON string.",
													Optional:            true,
												},
											},
										},
									},
								},
								"input_schemas": schema.ListNestedAttribute{
									MarkdownDescription: "Input schema definitions for this workflow.",
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.List{
										listplanmodifier.UseStateForUnknown(),
									},
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"name": schema.StringAttribute{
												Optional: true,
											},
											"type": schema.StringAttribute{
												MarkdownDescription: "Schema type (e.g. FORM_JSONSCHEMA).",
												Required:            true,
											},
											"encoded_data": schema.StringAttribute{
												MarkdownDescription: "Base64-encoded schema data.",
												Optional:            true,
											},
											"ui_schema_data": schema.StringAttribute{
												Optional: true,
											},
										},
									},
								},
								"approvers": schema.ListAttribute{
									MarkdownDescription: "List of approvers.",
									ElementType:         types.StringType,
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.List{
										listplanmodifier.UseStateForUnknown(),
									},
								},
								"number_of_approvals_required": schema.Int64Attribute{
									MarkdownDescription: "Number of approvals required.",
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.Int64{
										int64planmodifier.UseStateForUnknown(),
									},
								},
								"user_job_cpu": schema.Int64Attribute{
									MarkdownDescription: "CPU limit for the user job.",
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.Int64{
										int64planmodifier.UseStateForUnknown(),
									},
								},
								"user_job_memory": schema.Int64Attribute{
									MarkdownDescription: "Memory limit for the user job.",
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.Int64{
										int64planmodifier.UseStateForUnknown(),
									},
								},
								"user_schedules": schema.ListNestedAttribute{
									MarkdownDescription: "User-defined schedules.",
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.List{
										listplanmodifier.UseStateForUnknown(),
									},
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"name": schema.StringAttribute{
												Computed: true,
												PlanModifiers: []planmodifier.String{
													stringplanmodifier.UseStateForUnknown(),
												},
											},
											"desc":  schema.StringAttribute{Optional: true},
											"cron":  schema.StringAttribute{Required: true},
											"state": schema.StringAttribute{Required: true},
										},
									},
								},
								"mini_steps": schema.SingleNestedAttribute{
									MarkdownDescription: "Mini steps configuration.",
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.Object{
										objectplanmodifier.UseStateForUnknown(),
									},
									Attributes: map[string]schema.Attribute{
										"notifications": schema.SingleNestedAttribute{
											Optional: true,
											Attributes: map[string]schema.Attribute{
												"email": schema.SingleNestedAttribute{
													Optional: true,
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
											Optional: true,
											Attributes: map[string]schema.Attribute{
												"approval_required": ministepsWebhooks,
												"cancelled":         ministepsWebhooks,
												"completed":         ministepsWebhooks,
												"drift_detected":    ministepsWebhooks,
												"errored":           ministepsWebhooks,
											},
										},
										"wf_chaining": schema.SingleNestedAttribute{
											Optional: true,
											Attributes: map[string]schema.Attribute{
												"completed": ministepsWorkflowChaining,
												"errored":   ministepsWorkflowChaining,
											},
										},
									},
								},
								"context_tags": schema.MapAttribute{
									MarkdownDescription: "Contextual tags.",
									ElementType:         types.StringType,
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.Map{
										mapplanmodifier.UseStateForUnknown(),
									},
								},
								"runner_constraints": schema.SingleNestedAttribute{
									MarkdownDescription: "Runner constraints.",
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.Object{
										objectplanmodifier.UseStateForUnknown(),
									},
									Attributes: map[string]schema.Attribute{
										"type": schema.StringAttribute{Required: true},
										"names": schema.ListAttribute{
											ElementType: types.StringType,
											Optional:    true,
										},
									},
								},
							},
						},
					},
				},
			},
			"user_schedules": schema.ListNestedAttribute{
				MarkdownDescription: "User-defined schedules for the stack.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"desc":  schema.StringAttribute{Optional: true},
						"cron":  schema.StringAttribute{Required: true},
						"state": schema.StringAttribute{Required: true},
						"inputs": schema.SingleNestedAttribute{
							MarkdownDescription: "Action to run on this schedule.",
							Required:            true,
							Attributes: map[string]schema.Attribute{
								"action_type": schema.StringAttribute{
									MarkdownDescription: "The action to trigger for this schedule.",
									Required:            true,
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
			"mini_steps": schema.SingleNestedAttribute{
				MarkdownDescription: "Mini steps configuration for the stack.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"notifications": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"email": schema.SingleNestedAttribute{
								Optional: true,
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
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"approval_required": ministepsWebhooks,
							"cancelled":         ministepsWebhooks,
							"completed":         ministepsWebhooks,
							"drift_detected":    ministepsWebhooks,
							"errored":           ministepsWebhooks,
						},
					},
					"wf_chaining": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"completed": ministepsWorkflowChaining,
							"errored":   ministepsWorkflowChaining,
						},
					},
				},
			},
		},
	}
}

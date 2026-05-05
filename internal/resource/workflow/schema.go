package workflow

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

var ministepsNotificationRecepients = schema.ListNestedAttribute{
	Optional: true,
	Computed: true,
	PlanModifiers: []planmodifier.List{
		listplanmodifier.UseStateForUnknown(),
	},
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
	Computed: true,
	PlanModifiers: []planmodifier.List{
		listplanmodifier.UseStateForUnknown(),
	},
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
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	},
}

var ministepsWorkflowChaining = schema.ListNestedAttribute{
	Optional: true,
	Computed: true,
	PlanModifiers: []planmodifier.List{
		listplanmodifier.UseStateForUnknown(),
	},
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"workflow_group_id": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWfChainingWorkflowGroupId,
				Required:            true,
			},
			"stack_id": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWfChainingStackId,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"stack_run_payload": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWfChainingStackPayload,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"workflow_id": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWfChainingWorkflowId,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"workflow_run_payload": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWfChainingWorkflowPayload,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	},
}

var terraformConfigSchema = schema.SingleNestedAttribute{
	MarkdownDescription: constants.TerraformConfig,
	Optional:            true,
	Computed:            true,
	PlanModifiers: []planmodifier.Object{
		objectplanmodifier.UseStateForUnknown(),
	},
	Attributes: map[string]schema.Attribute{
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
			NestedObject: schema.NestedAttributeObject{
				Attributes: mount_point,
			},
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
			NestedObject: wfStepsConfig,
		},
		"pre_apply_wf_steps_config": schema.ListNestedAttribute{
			MarkdownDescription: constants.TerraformPreApplyWfSteps,
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
			NestedObject: wfStepsConfig,
		},
		"pre_plan_wf_steps_config": schema.ListNestedAttribute{
			MarkdownDescription: constants.TerraformPrePlanWfSteps,
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
			NestedObject: wfStepsConfig,
		},
		"post_plan_wf_steps_config": schema.ListNestedAttribute{
			MarkdownDescription: constants.TerraformPostPlanWfSteps,
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
			NestedObject: wfStepsConfig,
		},
		"pre_init_hooks": schema.ListAttribute{
			MarkdownDescription: constants.TerraformPreInitHooks,
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
		"pre_plan_hooks": schema.ListAttribute{
			MarkdownDescription: constants.TerraformPrePlanHooks,
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
		"post_plan_hooks": schema.ListAttribute{
			MarkdownDescription: constants.TerraformPostPlanHooks,
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
		"pre_apply_hooks": schema.ListAttribute{
			MarkdownDescription: constants.TerraformPreApplyHooks,
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
		"post_apply_hooks": schema.ListAttribute{
			MarkdownDescription: constants.TerraformPostApplyHooks,
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
			PlanModifiers: []planmodifier.List{
				listplanmodifier.UseStateForUnknown(),
			},
		},
		"run_pre_init_hooks_on_drift": schema.BoolAttribute{
			MarkdownDescription: constants.TerraformRunPreInitHooksOnDrift,
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
	},
}

var environmentVariables = map[string]schema.Attribute{
	"config": schema.SingleNestedAttribute{
		MarkdownDescription: constants.EnvVarConfig,
		Required:            true,
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
var mount_point = map[string]schema.Attribute{
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
var wfStepsConfig = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		"name": schema.StringAttribute{
			MarkdownDescription: constants.WfStepName,
			Required:            true,
		},
		"environment_variables": schema.ListNestedAttribute{
			MarkdownDescription: constants.WfStepEnvVars,
			Optional:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: environmentVariables,
			},
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
			NestedObject: schema.NestedAttributeObject{
				Attributes: mount_point,
			},
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

// Schema defines the schema for the workflow resource.
func (r *workflowResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a workflow resource in a workflow group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: constants.Id,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"workflow_group_id": schema.StringAttribute{
				MarkdownDescription: "ID of the parent workflow group.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_name": schema.StringAttribute{
				MarkdownDescription: "Name of the workflow.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Description, "workflow"),
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"wf_type": schema.StringAttribute{
				MarkdownDescription: "Type of workflow (e.g., Terraform, Ansible, etc.).",
				Required:            true,
			},
			"environment_variables": schema.ListNestedAttribute{
				MarkdownDescription: "Environment variables for the workflow.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: environmentVariables,
				},
			},
			"mini_steps": schema.SingleNestedAttribute{
				MarkdownDescription: "Mini steps configuration for the workflow.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"notifications": schema.SingleNestedAttribute{
						MarkdownDescription: constants.MiniStepsNotifications,
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"email": schema.SingleNestedAttribute{
								MarkdownDescription: constants.MiniStepsNotificationsEmail,
								Optional:            true,
								Computed:            true,
								PlanModifiers: []planmodifier.Object{
									objectplanmodifier.UseStateForUnknown(),
								},
								Attributes: map[string]schema.Attribute{
									"approval_required": ministepsNotificationRecepients,
									"cancelled":         ministepsNotificationRecepients,
									"completed":         ministepsNotificationRecepients,
									"drift_detected":    ministepsNotificationRecepients,
									"errored":           ministepsNotificationRecepients,
								},
							},
						},
					},
					"webhooks": schema.SingleNestedAttribute{
						MarkdownDescription: constants.MiniStepsWebhooks,
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"approval_required": ministepsWebhooks,
							"cancelled":         ministepsWebhooks,
							"completed":         ministepsWebhooks,
							"drift_detected":    ministepsWebhooks,
							"errored":           ministepsWebhooks,
						},
					},
					"wf_chaining": schema.SingleNestedAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: constants.MiniStepsWorkflowChaining,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"completed": ministepsWorkflowChaining,
							"errored":   ministepsWorkflowChaining,
						},
					},
				},
			},
			"runner_constraints": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
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
						Optional:            true,
						ElementType:         types.StringType,
					},
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
			"user_schedules": schema.ListNestedAttribute{
				MarkdownDescription: "User-defined schedules for the workflow.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cron": schema.StringAttribute{
							MarkdownDescription: constants.UserScheduleCron,
							Required:            true,
						},
						"state": schema.StringAttribute{
							MarkdownDescription: constants.UserScheduleState,
							Optional:            true,
						},
						"desc": schema.StringAttribute{
							MarkdownDescription: constants.UserScheduleDesc,
							Optional:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: constants.UserScheduleName,
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
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
			"approvers": schema.ListAttribute{
				MarkdownDescription: "List of approvers for the workflow.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"number_of_approvals_required": schema.Int64Attribute{
				MarkdownDescription: "Number of approvals required before workflow execution.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"user_job_cpu": schema.Int64Attribute{
				MarkdownDescription: "CPU allocation for workflow execution.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"user_job_memory": schema.Int64Attribute{
				MarkdownDescription: "Memory allocation for workflow execution.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
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
						MarkdownDescription: "IaC VCS configuration.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"use_marketplace_template": schema.BoolAttribute{
								MarkdownDescription: "Whether to use a marketplace template.",
								Required:            true,
								PlanModifiers: []planmodifier.Bool{
									boolplanmodifier.RequiresReplace(),
								},
							},
							"iac_template_id": schema.StringAttribute{
								MarkdownDescription: "ID of the IaC template from the marketplace.",
								Optional:            true,
							},
							"custom_source": schema.SingleNestedAttribute{
								MarkdownDescription: "Custom source configuration.",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"source_config_dest_kind": schema.StringAttribute{
										MarkdownDescription: constants.RuntimeSourceDestKind,
										Required:            true,
									},
									"config": schema.SingleNestedAttribute{
										MarkdownDescription: "Source configuration details.",
										Required:            true,
										Attributes: map[string]schema.Attribute{
											"is_private": schema.BoolAttribute{
												Optional: true,
											},
											"auth": schema.StringAttribute{
												Optional:  true,
												Sensitive: true,
											},
											"working_dir": schema.StringAttribute{
												Optional: true,
											},
											"git_sparse_checkout_config": schema.StringAttribute{
												Optional: true,
											},
											"git_core_auto_crlf": schema.BoolAttribute{
												Optional: true,
												Computed: true,
												PlanModifiers: []planmodifier.Bool{
													boolplanmodifier.UseStateForUnknown(),
												},
											},
											"ref": schema.StringAttribute{
												Optional: true,
											},
											"repo": schema.StringAttribute{
												Optional: true,
											},
											"include_sub_module": schema.BoolAttribute{
												Optional: true,
											},
										},
									},
								},
							},
						},
					},
					"iac_input_data": schema.SingleNestedAttribute{
						MarkdownDescription: "IaC input data for the workflow.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"schema_id": schema.StringAttribute{
								Optional: true,
								Computed: true,
							},
							"schema_type": schema.StringAttribute{
								Required: true,
							},
							"data": schema.StringAttribute{
								MarkdownDescription: "Input data as a JSON string.",
								Required:            true,
							},
						},
					},
				},
			},
			"terraform_config": terraformConfigSchema,
			"deployment_platform_config": schema.ListNestedAttribute{
				MarkdownDescription: "Deployment platform configuration.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
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
								},
							},
						},
					},
				},
			},
			"wf_steps_config": schema.ListNestedAttribute{
				MarkdownDescription: "Workflow steps configuration.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: wfStepsConfig,
			},
		},
	}
}

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
				Required:            true,
			},
			"workflow_group_id": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowWorkflowGroupId,
				Required:            true,
			},
			"resource_name": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.ResourceName, "workflow"),
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
			"wf_type": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowType,
				Required:            true,
			},
			"environment_variables": schema.ListNestedAttribute{
				MarkdownDescription: constants.WfEnvironmentVariables,
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
				MarkdownDescription: constants.WfUserSchedules,
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
			"vcs_config": schema.SingleNestedAttribute{
				MarkdownDescription: constants.WorkflowVcsConfig,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"iac_vcs_config": schema.SingleNestedAttribute{
						MarkdownDescription: constants.WorkflowIacVcsConfig,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"use_marketplace_template": schema.BoolAttribute{
								MarkdownDescription: constants.WorkflowUseMarketplaceTemplate,
								Required:            true,
								PlanModifiers: []planmodifier.Bool{
									boolplanmodifier.RequiresReplace(),
								},
							},
							"iac_template_id": schema.StringAttribute{
								MarkdownDescription: constants.WorkflowIacTemplateId,
								Optional:            true,
							},
							"custom_source": schema.SingleNestedAttribute{
								MarkdownDescription: constants.WorkflowCustomSource,
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"source_config_dest_kind": schema.StringAttribute{
										MarkdownDescription: constants.RuntimeSourceDestKind,
										Required:            true,
									},
									"config": schema.SingleNestedAttribute{
										MarkdownDescription: constants.RuntimeSourceConfig,
										Required:            true,
										Attributes: map[string]schema.Attribute{
											"is_private": schema.BoolAttribute{
												MarkdownDescription: constants.RuntimeSourceConfigIsPrivate,
												Optional:            true,
											},
											"auth": schema.StringAttribute{
												MarkdownDescription: constants.RuntimeSourceConfigAuth,
												Optional:            true,
												Sensitive:           true,
											},
											"working_dir": schema.StringAttribute{
												MarkdownDescription: constants.RuntimeSourceConfigWorkingDir,
												Optional:            true,
											},
											"git_sparse_checkout_config": schema.StringAttribute{
												MarkdownDescription: constants.RuntimeSourceConfigGitSparse,
												Optional:            true,
											},
											"git_core_auto_crlf": schema.BoolAttribute{
												MarkdownDescription: constants.RuntimeSourceConfigGitCoreCRLF,
												Optional:            true,
												Computed:            true,
												PlanModifiers: []planmodifier.Bool{
													boolplanmodifier.UseStateForUnknown(),
												},
											},
											"ref": schema.StringAttribute{
												MarkdownDescription: constants.RuntimeSourceConfigRef,
												Optional:            true,
											},
											"repo": schema.StringAttribute{
												MarkdownDescription: constants.RuntimeSourceConfigRepo,
												Optional:            true,
											},
											"include_sub_module": schema.BoolAttribute{
												MarkdownDescription: constants.RuntimeSourceConfigIncludeSubmodule,
												Optional:            true,
											},
										},
									},
								},
							},
						},
					},
					"iac_input_data": schema.SingleNestedAttribute{
						MarkdownDescription: constants.WorkflowIacInputData,
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"schema_id": schema.StringAttribute{
								MarkdownDescription: constants.WorkflowIacInputDataSchemaId,
								Optional:            true,
								Computed:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"schema_type": schema.StringAttribute{
								MarkdownDescription: "Schema type for the input data. Allowed values are `FORM_JSONSCHEMA`, `RAW_HCL`, `RAW_JSON`, `NO_CODE_JSON`, `NONE`. Required when creating workflow using template and not defined in template.",
								Optional:            true,
								Computed:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"data": schema.StringAttribute{
								MarkdownDescription: "Input data as a JSON string. Required if creating workflow using template not defined in template.",
								Optional:            true,
								Computed:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
						},
					},
				},
			},
			"terraform_config": terraformConfigSchema,
			"deployment_platform_config": schema.ListNestedAttribute{
				MarkdownDescription: constants.WfDeploymentPlatformConfig,
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
				MarkdownDescription: constants.WfStepsConfig,
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

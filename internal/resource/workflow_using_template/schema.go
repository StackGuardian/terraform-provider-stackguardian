package workflowusingtemplate

import (
	"context"
	"fmt"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var nonEmptyString = []validator.String{stringvalidator.LengthAtLeast(1)}

func (r *workflowUsingTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a workflow resource that is created from a workflow template.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: constants.Id,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_name": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.ResourceName, "workflow_using_template"),
				Optional:            true,
			},
			"workflow_group_id": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowWorkflowGroupId,
				Required:            true,
				Validators:          nonEmptyString,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Description, "workflow_using_template"),
				Optional:            true,
				Validators:          nonEmptyString,
			},
			"wf_type": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowType,
				Required:            true,
				Validators:          nonEmptyString,
			},
			"environment_variables": schema.ListNestedAttribute{
				MarkdownDescription: constants.WfEnvironmentVariables,
				Optional:            true,
				NestedObject:        environmentVariable(),
			},
			"mini_steps": schema.SingleNestedAttribute{
				MarkdownDescription: constants.WfMiniSteps,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"notifications": schema.SingleNestedAttribute{
						MarkdownDescription: constants.MiniStepsNotifications,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"email": schema.SingleNestedAttribute{
								MarkdownDescription: constants.MiniStepsNotificationsEmail,
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"approval_required": schema.ListNestedAttribute{Optional: true, NestedObject: miniStepsNotificationRecipients()},
									"cancelled":         schema.ListNestedAttribute{Optional: true, NestedObject: miniStepsNotificationRecipients()},
									"completed":         schema.ListNestedAttribute{Optional: true, NestedObject: miniStepsNotificationRecipients()},
									"drift_detected":    schema.ListNestedAttribute{Optional: true, NestedObject: miniStepsNotificationRecipients()},
									"errored":           schema.ListNestedAttribute{Optional: true, NestedObject: miniStepsNotificationRecipients()},
								},
							},
						},
					},
					"webhooks": schema.SingleNestedAttribute{
						MarkdownDescription: constants.MiniStepsWebhooks,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"approval_required": schema.ListNestedAttribute{Optional: true, NestedObject: miniStepsWebhook()},
							"cancelled":         schema.ListNestedAttribute{Optional: true, NestedObject: miniStepsWebhook()},
							"completed":         schema.ListNestedAttribute{Optional: true, NestedObject: miniStepsWebhook()},
							"drift_detected":    schema.ListNestedAttribute{Optional: true, NestedObject: miniStepsWebhook()},
							"errored":           schema.ListNestedAttribute{Optional: true, NestedObject: miniStepsWebhook()},
						},
					},
					"wf_chaining": schema.SingleNestedAttribute{
						MarkdownDescription: constants.MiniStepsWorkflowChaining,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"completed": schema.ListNestedAttribute{Optional: true, NestedObject: miniStepsWfChaining()},
							"errored":   schema.ListNestedAttribute{Optional: true, NestedObject: miniStepsWfChaining()},
						},
					},
				},
			},
			"runner_constraints": schema.SingleNestedAttribute{
				MarkdownDescription: constants.WorkflowRunnerConstraints,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: constants.RunnerConstraintsType,
						Required:            true,
						Validators:          nonEmptyString,
					},
					"names": schema.ListAttribute{
						MarkdownDescription: constants.RunnerConstraintsNames,
						Optional:            true,
						ElementType:         types.StringType,
					},
				},
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Tags, "workflow_using_template"),
				ElementType:         types.StringType,
				Optional:            true,
			},
			"user_schedules": schema.ListNestedAttribute{
				MarkdownDescription: constants.WfUserSchedules,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cron": schema.StringAttribute{
							MarkdownDescription: constants.UserScheduleCron,
							Required:            true,
							Validators:          nonEmptyString,
						},
						"state": schema.StringAttribute{
							MarkdownDescription: constants.UserScheduleState,
							Required:            true,
							Validators:          nonEmptyString,
						},
						"desc": schema.StringAttribute{
							MarkdownDescription: constants.UserScheduleDesc,
							Optional:            true,
							Validators:          nonEmptyString,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: constants.UserScheduleName,
							Required:            true,
							Validators:          nonEmptyString,
						},
					},
				},
			},
			"context_tags": schema.MapAttribute{
				MarkdownDescription: fmt.Sprintf(constants.ContextTags, "workflow_using_template"),
				ElementType:         types.StringType,
				Optional:            true,
			},
			"approvers": schema.ListAttribute{
				MarkdownDescription: constants.WfApprovers,
				ElementType:         types.StringType,
				Optional:            true,
			},
			"number_of_approvals_required": schema.Int64Attribute{
				MarkdownDescription: constants.WfNumberOfApprovals,
				Optional:            true,
			},
			"user_job_cpu": schema.Int64Attribute{
				MarkdownDescription: constants.WfUserJobCPU,
				Optional:            true,
			},
			"user_job_memory": schema.Int64Attribute{
				MarkdownDescription: constants.WfUserJobMemory,
				Optional:            true,
			},
			"vcs_config": schema.SingleNestedAttribute{
				MarkdownDescription: constants.WorkflowVcsConfig,
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"iac_vcs_config": schema.SingleNestedAttribute{
						MarkdownDescription: constants.WorkflowIacVcsConfig,
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"iac_template_id": schema.StringAttribute{
								MarkdownDescription: "The ID of the workflow template to use for this workflow.",
								Required:            true,
								Validators:          nonEmptyString,
							},
						},
					},
					"iac_input_data": schema.SingleNestedAttribute{
						MarkdownDescription: constants.WorkflowIacInputData,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"schema_id": schema.StringAttribute{
								MarkdownDescription: constants.WorkflowIacInputDataSchemaId,
								Optional:            true,
								Validators:          nonEmptyString,
							},
							"schema_type": schema.StringAttribute{
								MarkdownDescription: constants.WorkflowIacInputDataSchemaType,
								Optional:            true,
								Validators:          nonEmptyString,
							},
							"data": schema.StringAttribute{
								MarkdownDescription: constants.WorkflowIacInputDataData,
								Optional:            true,
								Validators:          nonEmptyString,
							},
						},
					},
				},
			},
			"terraform_config": schema.SingleNestedAttribute{
				MarkdownDescription: constants.TerraformConfig,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"terraform_version": schema.StringAttribute{
						MarkdownDescription: constants.TerraformVersion,
						Optional:            true,
						Validators:          nonEmptyString,
					},
					"drift_check": schema.BoolAttribute{
						MarkdownDescription: constants.TerraformDriftCheck,
						Optional:            true,
					},
					"drift_cron": schema.StringAttribute{
						MarkdownDescription: constants.TerraformDriftCron,
						Optional:            true,
						Validators:          nonEmptyString,
					},
					"managed_terraform_state": schema.BoolAttribute{
						MarkdownDescription: constants.TerraformManagedState,
						Optional:            true,
					},
					"approval_pre_apply": schema.BoolAttribute{
						MarkdownDescription: constants.TerraformApprovalPreApply,
						Optional:            true,
					},
					"terraform_plan_options": schema.StringAttribute{
						MarkdownDescription: constants.TerraformPlanOptions,
						Optional:            true,
						Validators:          nonEmptyString,
					},
					"terraform_init_options": schema.StringAttribute{
						MarkdownDescription: constants.TerraformInitOptions,
						Optional:            true,
						Validators:          nonEmptyString,
					},
					"terraform_bin_path": schema.ListNestedAttribute{
						MarkdownDescription: constants.TerraformBinPath,
						Optional:            true,
						NestedObject:        mountPoint(),
					},
					"timeout": schema.Int64Attribute{
						MarkdownDescription: constants.TerraformTimeout,
						Optional:            true,
					},
					"post_apply_wf_steps_config": schema.ListNestedAttribute{
						MarkdownDescription: constants.TerraformPostApplyWfSteps,
						Optional:            true,
						NestedObject:        wfStepsConfig(),
					},
					"pre_apply_wf_steps_config": schema.ListNestedAttribute{
						MarkdownDescription: constants.TerraformPreApplyWfSteps,
						Optional:            true,
						NestedObject:        wfStepsConfig(),
					},
					"pre_plan_wf_steps_config": schema.ListNestedAttribute{
						MarkdownDescription: constants.TerraformPrePlanWfSteps,
						Optional:            true,
						NestedObject:        wfStepsConfig(),
					},
					"post_plan_wf_steps_config": schema.ListNestedAttribute{
						MarkdownDescription: constants.TerraformPostPlanWfSteps,
						Optional:            true,
						NestedObject:        wfStepsConfig(),
					},
					"pre_init_hooks": schema.ListAttribute{
						MarkdownDescription: constants.TerraformPreInitHooks,
						Optional:            true,
						ElementType:         types.StringType,
					},
					"pre_plan_hooks": schema.ListAttribute{
						MarkdownDescription: constants.TerraformPrePlanHooks,
						Optional:            true,
						ElementType:         types.StringType,
					},
					"post_plan_hooks": schema.ListAttribute{
						MarkdownDescription: constants.TerraformPostPlanHooks,
						Optional:            true,
						ElementType:         types.StringType,
					},
					"pre_apply_hooks": schema.ListAttribute{
						MarkdownDescription: constants.TerraformPreApplyHooks,
						Optional:            true,
						ElementType:         types.StringType,
					},
					"post_apply_hooks": schema.ListAttribute{
						MarkdownDescription: constants.TerraformPostApplyHooks,
						Optional:            true,
						ElementType:         types.StringType,
					},
					"run_pre_init_hooks_on_drift": schema.BoolAttribute{
						MarkdownDescription: constants.TerraformRunPreInitHooksOnDrift,
						Optional:            true,
					},
					"run_pre_plan_hooks_on_drift": schema.BoolAttribute{
						MarkdownDescription: constants.TerraformRunPrePlanHooksOnDrift,
						Optional:            true,
					},
					"run_post_plan_hooks_on_drift": schema.BoolAttribute{
						MarkdownDescription: constants.TerraformRunPostPlanHooksOnDrift,
						Optional:            true,
					},
				},
			},
			"deployment_platform_config": schema.ListNestedAttribute{
				MarkdownDescription: constants.WfDeploymentPlatformConfig,
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"kind": schema.StringAttribute{
							MarkdownDescription: constants.DeploymentPlatformKind,
							Required:            true,
							Validators:          nonEmptyString,
						},
						"config": schema.SingleNestedAttribute{
							MarkdownDescription: constants.DeploymentPlatformConfigDetails,
							Required:            true,
							Attributes: map[string]schema.Attribute{
								"integration_id": schema.StringAttribute{
									MarkdownDescription: constants.DeploymentPlatformIntegrationId,
									Required:            true,
									Validators:          nonEmptyString,
								},
								"profile_name": schema.StringAttribute{
									MarkdownDescription: constants.DeploymentPlatformProfileName,
									Optional:            true,
									Validators:          nonEmptyString,
								},
							},
						},
					},
				},
			},
			"wf_steps_config": schema.ListNestedAttribute{
				MarkdownDescription: constants.WfStepsConfig,
				Optional:            true,
				NestedObject:        wfStepsConfig(),
			},
			"resolved_schema": schema.SingleNestedAttribute{
				MarkdownDescription: "Read-only view of the workflow attributes as returned by the API.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"resource_name": schema.StringAttribute{Computed: true},
					"description":   schema.StringAttribute{Computed: true},
					"environment_variables": schema.ListNestedAttribute{
						Computed:     true,
						NestedObject: computedEnvironmentVariable(),
					},
					"mini_steps": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"notifications": schema.SingleNestedAttribute{
								Computed: true,
								Attributes: map[string]schema.Attribute{
									"email": schema.SingleNestedAttribute{
										Computed: true,
										Attributes: map[string]schema.Attribute{
											"approval_required": schema.ListNestedAttribute{Computed: true, NestedObject: computedMiniStepsNotificationRecipients()},
											"cancelled":         schema.ListNestedAttribute{Computed: true, NestedObject: computedMiniStepsNotificationRecipients()},
											"completed":         schema.ListNestedAttribute{Computed: true, NestedObject: computedMiniStepsNotificationRecipients()},
											"drift_detected":    schema.ListNestedAttribute{Computed: true, NestedObject: computedMiniStepsNotificationRecipients()},
											"errored":           schema.ListNestedAttribute{Computed: true, NestedObject: computedMiniStepsNotificationRecipients()},
										},
									},
								},
							},
							"webhooks": schema.SingleNestedAttribute{
								Computed: true,
								Attributes: map[string]schema.Attribute{
									"approval_required": schema.ListNestedAttribute{Computed: true, NestedObject: computedMiniStepsWebhook()},
									"cancelled":         schema.ListNestedAttribute{Computed: true, NestedObject: computedMiniStepsWebhook()},
									"completed":         schema.ListNestedAttribute{Computed: true, NestedObject: computedMiniStepsWebhook()},
									"drift_detected":    schema.ListNestedAttribute{Computed: true, NestedObject: computedMiniStepsWebhook()},
									"errored":           schema.ListNestedAttribute{Computed: true, NestedObject: computedMiniStepsWebhook()},
								},
							},
							"wf_chaining": schema.SingleNestedAttribute{
								Computed: true,
								Attributes: map[string]schema.Attribute{
									"completed": schema.ListNestedAttribute{Computed: true, NestedObject: computedMiniStepsWfChaining()},
									"errored":   schema.ListNestedAttribute{Computed: true, NestedObject: computedMiniStepsWfChaining()},
								},
							},
						},
					},
					"runner_constraints": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"type":  schema.StringAttribute{Computed: true},
							"names": schema.ListAttribute{Computed: true, ElementType: types.StringType},
						},
					},
					"tags":         schema.ListAttribute{Computed: true, ElementType: types.StringType},
					"approvers":    schema.ListAttribute{Computed: true, ElementType: types.StringType},
					"context_tags": schema.MapAttribute{Computed: true, ElementType: types.StringType},
					"user_schedules": schema.ListNestedAttribute{
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"cron":  schema.StringAttribute{Computed: true},
								"state": schema.StringAttribute{Computed: true},
								"desc":  schema.StringAttribute{Computed: true},
								"name":  schema.StringAttribute{Computed: true},
							},
						},
					},
					"number_of_approvals_required": schema.Int64Attribute{Computed: true},
					"user_job_cpu":                 schema.Int64Attribute{Computed: true},
					"user_job_memory":              schema.Int64Attribute{Computed: true},
					"terraform_config": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"terraform_version":            schema.StringAttribute{Computed: true},
							"drift_check":                  schema.BoolAttribute{Computed: true},
							"drift_cron":                   schema.StringAttribute{Computed: true},
							"managed_terraform_state":      schema.BoolAttribute{Computed: true},
							"approval_pre_apply":           schema.BoolAttribute{Computed: true},
							"terraform_plan_options":       schema.StringAttribute{Computed: true},
							"terraform_init_options":       schema.StringAttribute{Computed: true},
							"terraform_bin_path":           schema.ListNestedAttribute{Computed: true, NestedObject: computedMountPoint()},
							"timeout":                      schema.Int64Attribute{Computed: true},
							"post_apply_wf_steps_config":   schema.ListNestedAttribute{Computed: true, NestedObject: computedWfStepsConfig()},
							"pre_apply_wf_steps_config":    schema.ListNestedAttribute{Computed: true, NestedObject: computedWfStepsConfig()},
							"pre_plan_wf_steps_config":     schema.ListNestedAttribute{Computed: true, NestedObject: computedWfStepsConfig()},
							"post_plan_wf_steps_config":    schema.ListNestedAttribute{Computed: true, NestedObject: computedWfStepsConfig()},
							"pre_init_hooks":               schema.ListAttribute{Computed: true, ElementType: types.StringType},
							"pre_plan_hooks":               schema.ListAttribute{Computed: true, ElementType: types.StringType},
							"post_plan_hooks":              schema.ListAttribute{Computed: true, ElementType: types.StringType},
							"pre_apply_hooks":              schema.ListAttribute{Computed: true, ElementType: types.StringType},
							"post_apply_hooks":             schema.ListAttribute{Computed: true, ElementType: types.StringType},
							"run_pre_init_hooks_on_drift":  schema.BoolAttribute{Computed: true},
							"run_pre_plan_hooks_on_drift":  schema.BoolAttribute{Computed: true},
							"run_post_plan_hooks_on_drift": schema.BoolAttribute{Computed: true},
						},
					},
					"deployment_platform_config": schema.ListNestedAttribute{
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"kind": schema.StringAttribute{Computed: true},
								"config": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"integration_id": schema.StringAttribute{Computed: true},
										"profile_name":   schema.StringAttribute{Computed: true},
									},
								},
							},
						},
					},
					"wf_steps_config": schema.ListNestedAttribute{
						Computed:     true,
						NestedObject: computedWfStepsConfig(),
					},
				},
			},
		},
	}
}

func wfStepsConfig() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: constants.WfStepName,
				Required:            true,
				Validators:          nonEmptyString,
			},
			"environment_variables": schema.ListNestedAttribute{
				MarkdownDescription: constants.WfStepEnvVars,
				Optional:            true,
				NestedObject:        environmentVariable(),
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
				Validators:          nonEmptyString,
			},
			"mount_points": schema.ListNestedAttribute{
				MarkdownDescription: constants.WfStepMountPoints,
				Optional:            true,
				NestedObject:        mountPoint(),
			},
			"wf_step_template_id": schema.StringAttribute{
				MarkdownDescription: constants.WfStepTemplateId,
				Required:            true,
				Validators:          nonEmptyString,
			},
			"wf_step_input_data": schema.SingleNestedAttribute{
				MarkdownDescription: constants.WfStepInputData,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"schema_type": schema.StringAttribute{
						MarkdownDescription: constants.WfStepInputDataSchemaType,
						Optional:            true,
						Validators:          nonEmptyString,
					},
					"data": schema.StringAttribute{
						MarkdownDescription: constants.WfStepInputDataData,
						Optional:            true,
						Validators:          nonEmptyString,
					},
				},
			},
		},
	}
}

func environmentVariable() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"config": schema.SingleNestedAttribute{
				MarkdownDescription: constants.EnvVarConfig,
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"var_name": schema.StringAttribute{
						MarkdownDescription: constants.EnvVarConfigVarName,
						Required:            true,
						Validators:          nonEmptyString,
					},
					"secret_id": schema.StringAttribute{
						MarkdownDescription: constants.EnvVarConfigSecretId,
						Optional:            true,
						Validators:          nonEmptyString,
					},
					"text_value": schema.StringAttribute{
						MarkdownDescription: constants.EnvVarConfigTextValue,
						Optional:            true,
						Validators:          nonEmptyString,
					},
				},
			},
			"kind": schema.StringAttribute{
				MarkdownDescription: constants.EnvVarKind,
				Required:            true,
				Validators:          nonEmptyString,
			},
		},
	}
}

func mountPoint() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"source": schema.StringAttribute{
				MarkdownDescription: constants.MountPointSource,
				Required:            true,
				Validators:          nonEmptyString,
			},
			"target": schema.StringAttribute{
				MarkdownDescription: constants.MountPointTarget,
				Required:            true,
				Validators:          nonEmptyString,
			},
			"read_only": schema.BoolAttribute{
				MarkdownDescription: constants.MountPointReadOnly,
				Optional:            true,
			},
		},
	}
}

func miniStepsNotificationRecipients() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"recipients": schema.ListAttribute{
				MarkdownDescription: constants.MiniStepsNotificationsRecipients,
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func miniStepsWebhook() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"webhook_name": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWebhookName,
				Required:            true,
				Validators:          nonEmptyString,
			},
			"webhook_url": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWebhookURL,
				Required:            true,
				Validators:          nonEmptyString,
			},
			"webhook_secret": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWebhookSecret,
				Optional:            true,
				Validators:          nonEmptyString,
			},
		},
	}
}

func computedWfStepsConfig() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name":                  schema.StringAttribute{Computed: true},
			"environment_variables": schema.ListNestedAttribute{Computed: true, NestedObject: computedEnvironmentVariable()},
			"approval":              schema.BoolAttribute{Computed: true},
			"timeout":               schema.Int64Attribute{Computed: true},
			"cmd_override":          schema.StringAttribute{Computed: true},
			"mount_points":          schema.ListNestedAttribute{Computed: true, NestedObject: computedMountPoint()},
			"wf_step_template_id":   schema.StringAttribute{Computed: true},
			"wf_step_input_data": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"schema_type": schema.StringAttribute{Computed: true},
					"data":        schema.StringAttribute{Computed: true},
				},
			},
		},
	}
}

func computedEnvironmentVariable() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"config": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"var_name":   schema.StringAttribute{Computed: true},
					"secret_id":  schema.StringAttribute{Computed: true},
					"text_value": schema.StringAttribute{Computed: true},
				},
			},
			"kind": schema.StringAttribute{Computed: true},
		},
	}
}

func computedMountPoint() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"source":    schema.StringAttribute{Computed: true},
			"target":    schema.StringAttribute{Computed: true},
			"read_only": schema.BoolAttribute{Computed: true},
		},
	}
}

func computedMiniStepsNotificationRecipients() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"recipients": schema.ListAttribute{Computed: true, ElementType: types.StringType},
		},
	}
}

func computedMiniStepsWebhook() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"webhook_name":   schema.StringAttribute{Computed: true},
			"webhook_url":    schema.StringAttribute{Computed: true},
			"webhook_secret": schema.StringAttribute{Computed: true},
		},
	}
}

func computedMiniStepsWfChaining() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"workflow_group_id":    schema.StringAttribute{Computed: true},
			"stack_id":             schema.StringAttribute{Computed: true},
			"stack_run_payload":    schema.StringAttribute{Computed: true},
			"workflow_id":          schema.StringAttribute{Computed: true},
			"workflow_run_payload": schema.StringAttribute{Computed: true},
		},
	}
}

func miniStepsWfChaining() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"workflow_group_id": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWfChainingWorkflowGroupId,
				Required:            true,
				Validators:          nonEmptyString,
			},
			"stack_id": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWfChainingStackId,
				Optional:            true,
				Validators:          nonEmptyString,
			},
			"stack_run_payload": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWfChainingStackPayload,
				Optional:            true,
				Validators:          nonEmptyString,
			},
			"workflow_id": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWfChainingWorkflowId,
				Optional:            true,
				Validators:          nonEmptyString,
			},
			"workflow_run_payload": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWfChainingWorkflowPayload,
				Optional:            true,
				Validators:          nonEmptyString,
			},
		},
	}
}

package workflowtemplaterevision

import (
	"context"
	"fmt"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	workflowtemplate "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/workflow_template"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var ministepsNotificationRecepients = schema.ListNestedAttribute{
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

var terraformConfigSchema = schema.SingleNestedAttribute{
	MarkdownDescription: constants.TerraformConfig,
	Optional:            true,
	Computed:            true,
	Attributes: map[string]schema.Attribute{
		"terraform_version": schema.StringAttribute{
			MarkdownDescription: constants.TerraformVersion,
			Optional:            true,
			Computed:            true,
		},
		"drift_check": schema.BoolAttribute{
			MarkdownDescription: constants.TerraformDriftCheck,
			Optional:            true,
			Computed:            true,
		},
		"drift_cron": schema.StringAttribute{
			MarkdownDescription: constants.TerraformDriftCron,
			Optional:            true,
			Computed:            true,
		},
		"managed_terraform_state": schema.BoolAttribute{
			MarkdownDescription: constants.TerraformManagedState,
			Optional:            true,
			Computed:            true,
		},
		"approval_pre_apply": schema.BoolAttribute{
			MarkdownDescription: constants.TerraformApprovalPreApply,
			Optional:            true,
			Computed:            true,
		},
		"terraform_plan_options": schema.StringAttribute{
			MarkdownDescription: constants.TerraformPlanOptions,
			Optional:            true,
			Computed:            true,
		},
		"terraform_init_options": schema.StringAttribute{
			MarkdownDescription: constants.TerraformInitOptions,
			Optional:            true,
			Computed:            true,
		},
		"terraform_bin_path": schema.ListNestedAttribute{
			MarkdownDescription: constants.TerraformBinPath,
			Optional:            true,
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: mount_point,
			},
		},
		// TODO: confirm the description
		"timeout": schema.Int64Attribute{
			MarkdownDescription: constants.TerraformTimeout,
			Optional:            true,
			Computed:            true,
		},
		"post_apply_wf_steps_config": schema.ListNestedAttribute{
			MarkdownDescription: constants.TerraformPostApplyWfSteps,
			Optional:            true,
			Computed:            true,
			NestedObject:        wfStepsConfig,
		},
		"pre_apply_wf_steps_config": schema.ListNestedAttribute{
			MarkdownDescription: constants.TerraformPreApplyWfSteps,
			Optional:            true,
			Computed:            true,
			NestedObject:        wfStepsConfig,
		},
		"pre_plan_wf_steps_config": schema.ListNestedAttribute{
			MarkdownDescription: constants.TerraformPrePlanWfSteps,
			Optional:            true,
			Computed:            true,
			NestedObject:        wfStepsConfig,
		},
		"post_plan_wf_steps_config": schema.ListNestedAttribute{
			MarkdownDescription: constants.TerraformPostPlanWfSteps,
			Optional:            true,
			Computed:            true,
			NestedObject:        wfStepsConfig,
		},
		"pre_init_hooks": schema.ListAttribute{
			MarkdownDescription: constants.TerraformPreInitHooks,
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"pre_plan_hooks": schema.ListAttribute{
			MarkdownDescription: constants.TerraformPrePlanHooks,
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"post_plan_hooks": schema.ListAttribute{
			MarkdownDescription: constants.TerraformPostPlanHooks,
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"pre_apply_hooks": schema.ListAttribute{
			MarkdownDescription: constants.TerraformPreApplyHooks,
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"post_apply_hooks": schema.ListAttribute{
			MarkdownDescription: constants.TerraformPostApplyHooks,
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"run_pre_init_hooks_on_drift": schema.BoolAttribute{
			MarkdownDescription: constants.TerraformRunPreInitHooksOnDrift,
			Optional:            true,
			Computed:            true,
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
	// TODO: confirm the description
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

// Schema defines the schema for the resource.
func (r *workflowTemplateRevisionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a workflow template revision resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: constants.Id,
				Computed:            true,
			},
			"template_id": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowTemplateRevisionTemplateId,
				Required:            true,
			},
			"revision_id": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Description, "workflow template revision"),
				Optional:            true,
				Computed:            true,
			},
			"alias": schema.StringAttribute{
				MarkdownDescription: constants.TemplateRevisionAlias,
				Optional:            true,
			},
			"notes": schema.StringAttribute{
				MarkdownDescription: constants.TemplateRevisionNotes,
				Optional:            true,
				Computed:            true,
			},
			"source_config_kind": schema.StringAttribute{
				MarkdownDescription: constants.SourceConfigKind,
				Required:            true,
			},
			"is_public": schema.StringAttribute{
				MarkdownDescription: constants.TemplateRevisionIsPublic,
				Optional:            true,
				Computed:            true,
			},
			"deprecation": schema.SingleNestedAttribute{
				MarkdownDescription: constants.TemplateRevisionDeprecation,
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"effective_date": schema.StringAttribute{
						MarkdownDescription: constants.TemplateRevisionDeprecationEffectiveDate,
						Optional:            true,
					},
					"message": schema.StringAttribute{
						MarkdownDescription: constants.DeprecationMessage,
						Optional:            true,
					},
				},
			},
			"environment_variables": schema.ListNestedAttribute{
				MarkdownDescription: constants.WorkflowTemplateRevisionEnvironmentVariables,
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: environmentVariables,
				},
			},
			// TODO: Update descriptions for encoded_data and ui_schema_data
			"input_schemas": schema.ListNestedAttribute{
				MarkdownDescription: constants.WorkflowTemplateRevisionInputSchemas,
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: constants.InputSchemaType,
							Optional:            true,
						},
						"encoded_data": schema.StringAttribute{
							MarkdownDescription: constants.InputSchemaEncodedData,
							Optional:            true,
						},
						"ui_schema_data": schema.StringAttribute{
							MarkdownDescription: constants.InputSchemaUISchemaData,
							Optional:            true,
						},
					},
				},
			},
			"mini_steps": schema.SingleNestedAttribute{
				MarkdownDescription: constants.WorkflowTemplateRevisionMiniSteps,
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
						MarkdownDescription: constants.MiniStepsWorkflowChaining,
						Attributes: map[string]schema.Attribute{
							"completed": ministepsWorkflowChaining,
							"errored":   ministepsWorkflowChaining,
						},
					},
				},
			},
			"runner_constraints": schema.SingleNestedAttribute{
				Optional: true,
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
				MarkdownDescription: fmt.Sprintf(constants.Tags, "workflow template revision"),
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"user_schedules": schema.ListNestedAttribute{
				MarkdownDescription: constants.WorkflowTemplateRevisionUserSchedules,
				Optional:            true,
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
						},
					},
				},
			},
			"context_tags": schema.MapAttribute{
				MarkdownDescription: fmt.Sprintf(constants.ContextTags, "workflow template revision"),
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"approvers": schema.ListAttribute{
				MarkdownDescription: constants.WorkflowTemplateRevisionApprovers,
				ElementType:         types.StringType,
				Optional:            true,
			},
			"number_of_approvals_required": schema.Int64Attribute{
				MarkdownDescription: constants.WorkflowTemplateRevisionNumberOfApprovals,
				Optional:            true,
				Computed:            true,
			},
			"user_job_cpu": schema.Int64Attribute{
				MarkdownDescription: constants.WorkflowTemplateRevisionUserJobCPU,
				Optional:            true,
				Computed:            true,
			},
			"user_job_memory": schema.Int64Attribute{
				MarkdownDescription: constants.WorkflowTemplateRevisionUserJobMemory,
				Optional:            true,
				Computed:            true,
			},
			"runtime_source": schema.SingleNestedAttribute{
				MarkdownDescription: fmt.Sprintf(constants.RuntimeSource, "revision"),
				Optional:            true,
				Computed:            true,
				Attributes:          workflowtemplate.WorkflowTemplateRuntimeSourceConfig(),
			},
			"terraform_config": terraformConfigSchema,
			"deployment_platform_config": schema.SingleNestedAttribute{
				MarkdownDescription: constants.WorkflowTemplateRevisionDeploymentPlatformConfig,
				Optional:            true,
				Computed:            true,
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
								Computed:            true,
							},
						},
					},
				},
			},
			"wf_steps_config": schema.ListNestedAttribute{
				MarkdownDescription: constants.WorkflowTemplateRevisionWfStepsConfig,
				Optional:            true,
				Computed:            true,
				NestedObject:        wfStepsConfig,
			},
		},
	}
}

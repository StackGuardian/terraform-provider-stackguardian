package workflowtemplaterevision

import (
	"context"
	"fmt"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var ministepsNotificationRecepients = schema.ListNestedAttribute{
	Computed: true,
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"recipients": schema.ListAttribute{
				MarkdownDescription: constants.MiniStepsNotificationsRecipients,
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	},
}

var ministepsWebhooks = schema.ListNestedAttribute{
	Computed: true,
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"webhook_name": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWebhookName,
				Computed:            true,
			},
			"webhook_url": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWebhookURL,
				Computed:            true,
			},
			"webhook_secret": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWebhookSecret,
				Computed:            true,
			},
		},
	},
}

var ministepsWorkflowChaining = schema.ListNestedAttribute{
	Computed: true,
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"workflow_group_id": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWfChainingWorkflowGroupId,
				Computed:            true,
			},
			"stack_id": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWfChainingStackId,
				Computed:            true,
			},
			"stack_run_payload": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWfChainingStackPayload,
				Computed:            true,
			},
			"workflow_id": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWfChainingWorkflowId,
				Computed:            true,
			},
			"workflow_run_payload": schema.StringAttribute{
				MarkdownDescription: constants.MiniStepsWfChainingWorkflowPayload,
				Computed:            true,
			},
		},
	},
}

var miniStepsSchema = schema.SingleNestedAttribute{
	Computed: true,
	Attributes: map[string]schema.Attribute{
		"notifications": schema.SingleNestedAttribute{
			MarkdownDescription: constants.MiniStepsNotifications,
			Computed:            true,
			Attributes: map[string]schema.Attribute{
				"email": schema.SingleNestedAttribute{
					MarkdownDescription: constants.MiniStepsNotificationsEmail,
					Computed:            true,
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
			Computed:            true,
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
			Computed:            true,
			Attributes: map[string]schema.Attribute{
				"completed": ministepsWorkflowChaining,
				"errored":   ministepsWorkflowChaining,
			},
		},
	},
}

var environmentVariablesSchema = schema.ListNestedAttribute{
	Computed: true,
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"config": schema.SingleNestedAttribute{
				MarkdownDescription: constants.EnvVarConfig,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"var_name": schema.StringAttribute{
						MarkdownDescription: constants.EnvVarConfigVarName,
						Computed:            true,
					},
					"secret_id": schema.StringAttribute{
						MarkdownDescription: constants.EnvVarConfigSecretId,
						Computed:            true,
					},
					"text_value": schema.StringAttribute{
						MarkdownDescription: constants.EnvVarConfigTextValue,
						Computed:            true,
					},
				},
			},
			"kind": schema.StringAttribute{
				MarkdownDescription: constants.EnvVarKind,
				Computed:            true,
			},
		},
	},
}

var mountPointsSchema = schema.ListNestedAttribute{
	Computed: true,
	NestedObject: schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"source": schema.StringAttribute{
				MarkdownDescription: constants.MountPointSource,
				Computed:            true,
			},
			"target": schema.StringAttribute{
				MarkdownDescription: constants.MountPointTarget,
				Computed:            true,
			},
			"read_only": schema.BoolAttribute{
				MarkdownDescription: constants.MountPointReadOnly,
				Computed:            true,
			},
		},
	},
}

var wfStepsConfigSchema = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		"name": schema.StringAttribute{
			MarkdownDescription: constants.WfStepName,
			Computed:            true,
		},
		"wf_step_template_id": schema.StringAttribute{
			MarkdownDescription: constants.WfStepTemplateId,
			Computed:            true,
		},
		"timeout": schema.Int64Attribute{
			MarkdownDescription: constants.WfStepTimeout,
			Computed:            true,
		},
		"approval": schema.BoolAttribute{
			MarkdownDescription: constants.WfStepApproval,
			Computed:            true,
		},
		"cmd_override": schema.StringAttribute{
			MarkdownDescription: constants.WfStepCmdOverride,
			Computed:            true,
		},
		"environment_variables": environmentVariablesSchema,
		"mount_points":          mountPointsSchema,
		"wf_step_input_data": schema.SingleNestedAttribute{
			MarkdownDescription: constants.WfStepInputData,
			Computed:            true,
			Attributes: map[string]schema.Attribute{
				"schema_type": schema.StringAttribute{
					MarkdownDescription: constants.WfStepInputDataSchemaType,
					Computed:            true,
				},
				"data": schema.StringAttribute{
					MarkdownDescription: constants.WfStepInputDataData,
					Computed:            true,
				},
			},
		},
	},
}

var terraformConfigSchema = schema.SingleNestedAttribute{
	Computed: true,
	Attributes: map[string]schema.Attribute{
		"terraform_version": schema.StringAttribute{
			MarkdownDescription: "Terraform/OpenTofu version, returned in the bare form (e.g. `1.5.7`). The engine prefix the API stores (`TERRAFORM-` / `OPENTOFU-`) is stripped so the value can be referenced directly into a `stackguardian_workflow_from_template` resource without producing a perpetual diff.",
			Computed:            true,
		},
		"terraform_plan_options": schema.StringAttribute{
			MarkdownDescription: constants.TerraformPlanOptions,
			Computed:            true,
		},
		"terraform_init_options": schema.StringAttribute{
			MarkdownDescription: constants.TerraformInitOptions,
			Computed:            true,
		},
		"terraform_bin_path": schema.ListNestedAttribute{
			MarkdownDescription: constants.TerraformBinPath,
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"source": schema.StringAttribute{
						MarkdownDescription: constants.MountPointSource,
						Computed:            true,
					},
					"target": schema.StringAttribute{
						MarkdownDescription: constants.MountPointTarget,
						Computed:            true,
					},
					"read_only": schema.BoolAttribute{
						MarkdownDescription: constants.MountPointReadOnly,
						Computed:            true,
					},
				},
			},
		},
		"timeout": schema.Int64Attribute{
			MarkdownDescription: constants.TerraformTimeout,
			Computed:            true,
		},
		"managed_terraform_state": schema.BoolAttribute{
			MarkdownDescription: constants.TerraformManagedState,
			Computed:            true,
		},
		"drift_check": schema.BoolAttribute{
			MarkdownDescription: constants.TerraformDriftCheck,
			Computed:            true,
		},
		"drift_cron": schema.StringAttribute{
			MarkdownDescription: constants.TerraformDriftCron,
			Computed:            true,
		},
		"approval_pre_apply": schema.BoolAttribute{
			MarkdownDescription: constants.TerraformApprovalPreApply,
			Computed:            true,
		},
		"run_pre_init_hooks_on_drift": schema.BoolAttribute{
			MarkdownDescription: constants.TerraformRunPreInitHooksOnDrift,
			Computed:            true,
		},
		"pre_init_hooks": schema.ListAttribute{
			MarkdownDescription: constants.TerraformPreInitHooks,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"pre_plan_hooks": schema.ListAttribute{
			MarkdownDescription: constants.TerraformPrePlanHooks,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"post_plan_hooks": schema.ListAttribute{
			MarkdownDescription: constants.TerraformPostPlanHooks,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"pre_apply_hooks": schema.ListAttribute{
			MarkdownDescription: constants.TerraformPreApplyHooks,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"post_apply_hooks": schema.ListAttribute{
			MarkdownDescription: constants.TerraformPostApplyHooks,
			Computed:            true,
			ElementType:         types.StringType,
		},
		"post_apply_wf_steps_config": schema.ListNestedAttribute{
			MarkdownDescription: constants.TerraformPostApplyWfSteps,
			Computed:            true,
			NestedObject:        wfStepsConfigSchema,
		},
		"pre_apply_wf_steps_config": schema.ListNestedAttribute{
			MarkdownDescription: constants.TerraformPreApplyWfSteps,
			Computed:            true,
			NestedObject:        wfStepsConfigSchema,
		},
		"pre_plan_wf_steps_config": schema.ListNestedAttribute{
			MarkdownDescription: constants.TerraformPrePlanWfSteps,
			Computed:            true,
			NestedObject:        wfStepsConfigSchema,
		},
		"post_plan_wf_steps_config": schema.ListNestedAttribute{
			MarkdownDescription: constants.TerraformPostPlanWfSteps,
			Computed:            true,
			NestedObject:        wfStepsConfigSchema,
		},
	},
}

var runtimeSourceSchema = schema.SingleNestedAttribute{
	Computed: true,
	Attributes: map[string]schema.Attribute{
		"source_config_dest_kind": schema.StringAttribute{
			MarkdownDescription: constants.RuntimeSourceDestKind,
			Computed:            true,
		},
		"config": schema.SingleNestedAttribute{
			MarkdownDescription: constants.RuntimeSourceConfig,
			Computed:            true,
			Attributes: map[string]schema.Attribute{
				"is_private": schema.BoolAttribute{
					MarkdownDescription: constants.RuntimeSourceConfigIsPrivate,
					Computed:            true,
				},
				"auth": schema.StringAttribute{
					MarkdownDescription: constants.RuntimeSourceConfigAuth,
					Computed:            true,
				},
				"git_core_auto_crlf": schema.BoolAttribute{
					MarkdownDescription: constants.RuntimeSourceConfigGitCoreCRLF,
					Computed:            true,
				},
				"git_sparse_checkout_config": schema.StringAttribute{
					MarkdownDescription: constants.RuntimeSourceConfigGitSparse,
					Computed:            true,
				},
				"include_sub_module": schema.BoolAttribute{
					MarkdownDescription: constants.RuntimeSourceConfigIncludeSubmodule,
					Computed:            true,
				},
				"ref": schema.StringAttribute{
					MarkdownDescription: constants.RuntimeSourceConfigRef,
					Computed:            true,
				},
				"repo": schema.StringAttribute{
					MarkdownDescription: constants.RuntimeSourceConfigRepo,
					Computed:            true,
				},
				"working_dir": schema.StringAttribute{
					MarkdownDescription: constants.RuntimeSourceConfigWorkingDir,
					Computed:            true,
				},
			},
		},
	},
}

func (d *workflowTemplateRevisionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "> **Note:** This data source is currently in **BETA**. Features and behavior may change.\n\nReads a workflow template revision — for example to fetch a revision's default values and reference them from a `stackguardian_workflow_from_template` resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Revision ID in the bare `<template-name>:<revision>` form (e.g. `my-template:1`). Do not use the full `/<org>/<template-name>:<revision>` path — it returns an `Unauthorized` error.",
				Required:            true,
			},
			"template_id": schema.StringAttribute{
				MarkdownDescription: "Resource ID of the parent workflow template.",
				Optional:            true,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Description, "workflow template revision"),
				Computed:            true,
			},
			"alias": schema.StringAttribute{
				MarkdownDescription: constants.TemplateRevisionAlias,
				Computed:            true,
			},
			"notes": schema.StringAttribute{
				MarkdownDescription: constants.TemplateRevisionNotes,
				Computed:            true,
			},
			"source_config_kind": schema.StringAttribute{
				MarkdownDescription: constants.SourceConfigKind,
				Computed:            true,
			},
			"is_public": schema.StringAttribute{
				MarkdownDescription: constants.TemplateRevisionIsPublic,
				Computed:            true,
			},
			"deprecation": schema.SingleNestedAttribute{
				MarkdownDescription: constants.TemplateRevisionDeprecation,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"effective_date": schema.StringAttribute{
						MarkdownDescription: constants.TemplateRevisionDeprecationEffectiveDate,
						Computed:            true,
					},
					"message": schema.StringAttribute{
						MarkdownDescription: constants.DeprecationMessage,
						Computed:            true,
					},
				},
			},
			"environment_variables": environmentVariablesSchema,
			"input_schemas": schema.ListNestedAttribute{
				MarkdownDescription: constants.WorkflowTemplateRevisionInputSchemas,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						// name is present on the shared InputSchemaModel (and the resource
						// schema); omitting it here caused a schema/model type mismatch.
						"name": schema.StringAttribute{
							MarkdownDescription: constants.InputSchemaName,
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: constants.RunnerConstraintsType,
							Computed:            true,
						},
						"encoded_data": schema.StringAttribute{
							MarkdownDescription: constants.InputSchemaEncodedData,
							Computed:            true,
						},
						"ui_schema_data": schema.StringAttribute{
							MarkdownDescription: constants.InputSchemaUISchemaData,
							Computed:            true,
						},
					},
				},
			},
			"mini_steps": miniStepsSchema,
			"runner_constraints": schema.SingleNestedAttribute{
				MarkdownDescription: constants.WorkflowRunnerConstraints,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: constants.RunnerConstraintsType,
						Computed:            true,
					},
					"names": schema.ListAttribute{
						MarkdownDescription: constants.RunnerConstraintsNames,
						Computed:            true,
						ElementType:         types.StringType,
					},
				},
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Tags, "workflow template revision"),
				ElementType:         types.StringType,
				Computed:            true,
			},
			"user_schedules": schema.ListNestedAttribute{
				MarkdownDescription: constants.WfUserSchedules,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cron": schema.StringAttribute{
							MarkdownDescription: constants.UserScheduleCron,
							Computed:            true,
						},
						"state": schema.StringAttribute{
							MarkdownDescription: constants.UserScheduleState,
							Computed:            true,
						},
						"desc": schema.StringAttribute{
							MarkdownDescription: constants.UserScheduleDesc,
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: constants.UserScheduleName,
							Computed:            true,
						},
					},
				},
			},
			"context_tags": schema.MapAttribute{
				MarkdownDescription: "Context tags for the revision.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"approvers": schema.ListAttribute{
				MarkdownDescription: constants.WfApprovers,
				ElementType:         types.StringType,
				Computed:            true,
			},
			"number_of_approvals_required": schema.Int64Attribute{
				MarkdownDescription: constants.WfNumberOfApprovals,
				Computed:            true,
			},
			"user_job_cpu": schema.Int64Attribute{
				MarkdownDescription: constants.WfUserJobCPU,
				Computed:            true,
			},
			"user_job_memory": schema.Int64Attribute{
				MarkdownDescription: constants.WfUserJobMemory,
				Computed:            true,
			},
			"runtime_source":   runtimeSourceSchema,
			"terraform_config": terraformConfigSchema,
			// deployment_platform_config is a LIST in the API (and in the shared revision
			// model, which is types.List). Declaring it as a single nested object here caused
			// a schema/model type mismatch that errored at runtime whenever a revision carried
			// a populated deployment_platform_config.
			"deployment_platform_config": schema.ListNestedAttribute{
				MarkdownDescription: constants.WfDeploymentPlatformConfig,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"kind": schema.StringAttribute{
							MarkdownDescription: constants.DeploymentPlatformKind,
							Computed:            true,
						},
						"config": schema.SingleNestedAttribute{
							MarkdownDescription: constants.DeploymentPlatformConfigDetails,
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"integration_id": schema.StringAttribute{
									MarkdownDescription: constants.DeploymentPlatformIntegrationId,
									Computed:            true,
								},
								"profile_name": schema.StringAttribute{
									MarkdownDescription: constants.DeploymentPlatformProfileName,
									Computed:            true,
								},
							},
						},
					},
				},
			},
			"wf_steps_config": schema.ListNestedAttribute{
				MarkdownDescription: constants.WfStepsConfig,
				Computed:            true,
				NestedObject:        wfStepsConfigSchema,
			},
		},
	}
}

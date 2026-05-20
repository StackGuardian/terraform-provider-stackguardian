package workflowgit

import (
	"context"
	"fmt"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *workflowGitDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to read a git-based workflow.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: constants.DatasourceId,
				Required:            true,
			},
			"workflow_group_id": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowWorkflowGroupId,
				Required:            true,
			},
			"resource_name": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.ResourceName, "workflow_git"),
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Description, "workflow_git"),
				Computed:            true,
			},
			"wf_type": schema.StringAttribute{
				MarkdownDescription: constants.WorkflowType,
				Computed:            true,
			},
			"environment_variables": schema.ListNestedAttribute{
				MarkdownDescription: constants.WfEnvironmentVariables,
				Computed:            true,
				NestedObject:        dsEnvironmentVariable(),
			},
			"mini_steps": schema.SingleNestedAttribute{
				MarkdownDescription: constants.WfMiniSteps,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"notifications": schema.SingleNestedAttribute{
						MarkdownDescription: constants.MiniStepsNotifications,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"email": schema.SingleNestedAttribute{
								MarkdownDescription: constants.MiniStepsNotificationsEmail,
								Computed:            true,
								Attributes: map[string]schema.Attribute{
									"approval_required": schema.ListNestedAttribute{Computed: true, NestedObject: dsMiniStepsNotificationRecipients()},
									"cancelled":         schema.ListNestedAttribute{Computed: true, NestedObject: dsMiniStepsNotificationRecipients()},
									"completed":         schema.ListNestedAttribute{Computed: true, NestedObject: dsMiniStepsNotificationRecipients()},
									"drift_detected":    schema.ListNestedAttribute{Computed: true, NestedObject: dsMiniStepsNotificationRecipients()},
									"errored":           schema.ListNestedAttribute{Computed: true, NestedObject: dsMiniStepsNotificationRecipients()},
								},
							},
						},
					},
					"webhooks": schema.SingleNestedAttribute{
						MarkdownDescription: constants.MiniStepsWebhooks,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"approval_required": schema.ListNestedAttribute{Computed: true, NestedObject: dsMiniStepsWebhook()},
							"cancelled":         schema.ListNestedAttribute{Computed: true, NestedObject: dsMiniStepsWebhook()},
							"completed":         schema.ListNestedAttribute{Computed: true, NestedObject: dsMiniStepsWebhook()},
							"drift_detected":    schema.ListNestedAttribute{Computed: true, NestedObject: dsMiniStepsWebhook()},
							"errored":           schema.ListNestedAttribute{Computed: true, NestedObject: dsMiniStepsWebhook()},
						},
					},
					"wf_chaining": schema.SingleNestedAttribute{
						MarkdownDescription: constants.MiniStepsWorkflowChaining,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"completed": schema.ListNestedAttribute{Computed: true, NestedObject: dsMiniStepsWfChaining()},
							"errored":   schema.ListNestedAttribute{Computed: true, NestedObject: dsMiniStepsWfChaining()},
						},
					},
				},
			},
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
				MarkdownDescription: fmt.Sprintf(constants.Tags, "workflow_git"),
				Computed:            true,
				ElementType:         types.StringType,
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
				MarkdownDescription: fmt.Sprintf(constants.ContextTags, "workflow_git"),
				Computed:            true,
				ElementType:         types.StringType,
			},
			"approvers": schema.ListAttribute{
				MarkdownDescription: constants.WfApprovers,
				Computed:            true,
				ElementType:         types.StringType,
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
			"vcs_config": schema.SingleNestedAttribute{
				MarkdownDescription: constants.WorkflowVcsConfig,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"iac_vcs_config": schema.SingleNestedAttribute{
						MarkdownDescription: constants.WorkflowIacVcsConfig,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"custom_source": schema.SingleNestedAttribute{
								MarkdownDescription: constants.WorkflowCustomSource,
								Computed:            true,
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
												Sensitive:           true,
											},
											"working_dir": schema.StringAttribute{
												MarkdownDescription: constants.RuntimeSourceConfigWorkingDir,
												Computed:            true,
											},
											"git_sparse_checkout_config": schema.StringAttribute{
												MarkdownDescription: constants.RuntimeSourceConfigGitSparse,
												Computed:            true,
											},
											"git_core_auto_crlf": schema.BoolAttribute{
												MarkdownDescription: constants.RuntimeSourceConfigGitCoreCRLF,
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
											"include_sub_module": schema.BoolAttribute{
												MarkdownDescription: constants.RuntimeSourceConfigIncludeSubmodule,
												Computed:            true,
											},
										},
									},
								},
							},
						},
					},
					"iac_input_data": schema.SingleNestedAttribute{
						MarkdownDescription: constants.WorkflowIacInputData,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"schema_id": schema.StringAttribute{
								MarkdownDescription: constants.WorkflowIacInputDataSchemaId,
								Computed:            true,
							},
							"schema_type": schema.StringAttribute{
								MarkdownDescription: constants.WorkflowIacInputDataSchemaType,
								Computed:            true,
							},
							"data": schema.StringAttribute{
								MarkdownDescription: constants.WorkflowIacInputDataData,
								Computed:            true,
							},
						},
					},
				},
			},
			"terraform_config": schema.SingleNestedAttribute{
				MarkdownDescription: constants.TerraformConfig,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"terraform_version": schema.StringAttribute{
						MarkdownDescription: constants.TerraformVersion,
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
					"managed_terraform_state": schema.BoolAttribute{
						MarkdownDescription: constants.TerraformManagedState,
						Computed:            true,
					},
					"approval_pre_apply": schema.BoolAttribute{
						MarkdownDescription: constants.TerraformApprovalPreApply,
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
						NestedObject:        dsMountPoint(),
					},
					"timeout": schema.Int64Attribute{
						MarkdownDescription: constants.TerraformTimeout,
						Computed:            true,
					},
					"post_apply_wf_steps_config": schema.ListNestedAttribute{
						MarkdownDescription: constants.TerraformPostApplyWfSteps,
						Computed:            true,
						NestedObject:        dsWfStepsConfig(),
					},
					"pre_apply_wf_steps_config": schema.ListNestedAttribute{
						MarkdownDescription: constants.TerraformPreApplyWfSteps,
						Computed:            true,
						NestedObject:        dsWfStepsConfig(),
					},
					"pre_plan_wf_steps_config": schema.ListNestedAttribute{
						MarkdownDescription: constants.TerraformPrePlanWfSteps,
						Computed:            true,
						NestedObject:        dsWfStepsConfig(),
					},
					"post_plan_wf_steps_config": schema.ListNestedAttribute{
						MarkdownDescription: constants.TerraformPostPlanWfSteps,
						Computed:            true,
						NestedObject:        dsWfStepsConfig(),
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
					"run_pre_init_hooks_on_drift": schema.BoolAttribute{
						MarkdownDescription: constants.TerraformRunPreInitHooksOnDrift,
						Computed:            true,
					},
					"run_pre_plan_hooks_on_drift": schema.BoolAttribute{
						MarkdownDescription: constants.TerraformRunPrePlanHooksOnDrift,
						Computed:            true,
					},
					"run_post_plan_hooks_on_drift": schema.BoolAttribute{
						MarkdownDescription: constants.TerraformRunPostPlanHooksOnDrift,
						Computed:            true,
					},
				},
			},
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
				NestedObject:        dsWfStepsConfig(),
			},
			"vcs_triggers": schema.SingleNestedAttribute{
				MarkdownDescription: constants.VCSTriggers,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"tracked_branch": schema.StringAttribute{
						MarkdownDescription: constants.VCSTriggersTrackedBranch,
						Computed:            true,
					},
					"approval_pre_apply": schema.BoolAttribute{
						MarkdownDescription: constants.VCSTriggersApprovalPreApply,
						Computed:            true,
					},
					"plan_only": schema.BoolAttribute{
						MarkdownDescription: constants.VCSTriggersPlanOnly,
						Computed:            true,
					},
					"file_triggers_enabled": schema.BoolAttribute{
						MarkdownDescription: constants.VCSTriggersFileTriggersEnabled,
						Computed:            true,
					},
					"file_trigger_patterns": schema.ListAttribute{
						MarkdownDescription: constants.VCSTriggersFileTriggerPatterns,
						Computed:            true,
						ElementType:         types.StringType,
					},
					"gh_webhook_url": schema.StringAttribute{
						MarkdownDescription: constants.VCSTriggersGhWebhookUrl,
						Computed:            true,
					},
					"all_pull_requests": schema.MapNestedAttribute{
						MarkdownDescription: constants.VCSTriggersAllPullRequests,
						Computed:            true,
						NestedObject:        dsVcsTriggerActionNestedObject(),
					},
					"pull_request_opened": schema.MapNestedAttribute{
						MarkdownDescription: constants.VCSTriggersPullRequestOpened,
						Computed:            true,
						NestedObject:        dsVcsTriggerActionNestedObject(),
					},
					"pull_request_modified": schema.MapNestedAttribute{
						MarkdownDescription: constants.VCSTriggersPullRequestModified,
						Computed:            true,
						NestedObject:        dsVcsTriggerActionNestedObject(),
					},
					"create_tag": schema.MapNestedAttribute{
						MarkdownDescription: constants.VCSTriggersCreateTagAction,
						Computed:            true,
						NestedObject:        dsVcsTriggerActionNestedObject(),
					},
					"push": schema.MapNestedAttribute{
						MarkdownDescription: constants.VCSTriggersPush,
						Computed:            true,
						NestedObject:        dsVcsTriggerActionNestedObject(),
					},
				},
			},
		},
	}
}

func dsWfStepsConfig() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: constants.WfStepName,
				Computed:            true,
			},
			"environment_variables": schema.ListNestedAttribute{
				MarkdownDescription: constants.WfStepEnvVars,
				Computed:            true,
				NestedObject:        dsEnvironmentVariable(),
			},
			"approval": schema.BoolAttribute{
				MarkdownDescription: constants.WfStepApproval,
				Computed:            true,
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: constants.WfStepTimeout,
				Computed:            true,
			},
			"cmd_override": schema.StringAttribute{
				MarkdownDescription: constants.WfStepCmdOverride,
				Computed:            true,
			},
			"mount_points": schema.ListNestedAttribute{
				MarkdownDescription: constants.WfStepMountPoints,
				Computed:            true,
				NestedObject:        dsMountPoint(),
			},
			"wf_step_template_id": schema.StringAttribute{
				MarkdownDescription: constants.WfStepTemplateId,
				Computed:            true,
			},
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
}

func dsEnvironmentVariable() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
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
	}
}

func dsMountPoint() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
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
	}
}

func dsMiniStepsNotificationRecipients() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"recipients": schema.ListAttribute{
				MarkdownDescription: constants.MiniStepsNotificationsRecipients,
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func dsMiniStepsWebhook() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
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
	}
}

func dsVcsTriggerActionNestedObject() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func dsMiniStepsWfChaining() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
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
	}
}

package stacktemplaterevision

import (
	"context"
	"fmt"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Shared schema attribute maps reused across workflows_config and actions.

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
	"config": schema.StringAttribute{
		MarkdownDescription: constants.DeploymentPlatformConfigDetails + " (JSON string)",
		Optional:            true,
	},
}

// workflowInStackAttrs defines the schema attributes for a single workflow inside workflows_config.
// Fields correspond to the SDK's StackTemplateRevisionWorkflow struct.
var workflowInStackAttrs = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		MarkdownDescription: "UUID identifying the workflow within the stack template.",
		Required:            true,
	},
	"template_id": schema.StringAttribute{
		MarkdownDescription: "ID of the workflow template that this workflow is based on.",
		Required:            true,
	},
	"resource_name": schema.StringAttribute{
		MarkdownDescription: "Name of the workflow resource within the stack.",
		Optional:            true,
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
		Attributes: map[string]schema.Attribute{
			"terraform_version": schema.StringAttribute{
				MarkdownDescription: constants.TerraformVersion,
				Optional:            true,
			},
			"drift_check": schema.BoolAttribute{
				MarkdownDescription: constants.TerraformDriftCheck,
				Optional:            true,
			},
			"drift_cron": schema.StringAttribute{
				MarkdownDescription: constants.TerraformDriftCron,
				Optional:            true,
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
			},
			"terraform_init_options": schema.StringAttribute{
				MarkdownDescription: constants.TerraformInitOptions,
				Optional:            true,
			},
			"terraform_bin_path": schema.ListNestedAttribute{
				MarkdownDescription: constants.TerraformBinPath,
				Optional:            true,
				NestedObject:        schema.NestedAttributeObject{Attributes: mountPointAttrs},
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: constants.TerraformTimeout,
				Optional:            true,
			},
			"post_apply_wf_steps_config": schema.ListNestedAttribute{
				MarkdownDescription: constants.TerraformPostApplyWfSteps,
				Optional:            true,
				NestedObject:        wfStepsConfigNestedObj,
			},
			"pre_apply_wf_steps_config": schema.ListNestedAttribute{
				MarkdownDescription: constants.TerraformPreApplyWfSteps,
				Optional:            true,
				NestedObject:        wfStepsConfigNestedObj,
			},
			"pre_plan_wf_steps_config": schema.ListNestedAttribute{
				MarkdownDescription: constants.TerraformPrePlanWfSteps,
				Optional:            true,
				NestedObject:        wfStepsConfigNestedObj,
			},
			"post_plan_wf_steps_config": schema.ListNestedAttribute{
				MarkdownDescription: constants.TerraformPostPlanWfSteps,
				Optional:            true,
				NestedObject:        wfStepsConfigNestedObj,
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
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	},
	"environment_variables": schema.ListNestedAttribute{
		MarkdownDescription: "Environment variables for the workflow.",
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.List{
			listplanmodifier.UseStateForUnknown(),
		},
		NestedObject: schema.NestedAttributeObject{Attributes: envVarsAttrs},
	},
	"deployment_platform_config": schema.ListNestedAttribute{
		MarkdownDescription: "Deployment platform configuration.",
		Optional:            true,
		NestedObject:        schema.NestedAttributeObject{Attributes: deploymentPlatformConfigAttrs},
	},
	"vcs_config": schema.SingleNestedAttribute{
		MarkdownDescription: "VCS (version control) configuration for the workflow.",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"iac_vcs_config": schema.SingleNestedAttribute{
				MarkdownDescription: "IaC VCS configuration.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"use_marketplace_template": schema.BoolAttribute{
						MarkdownDescription: "Whether to use a marketplace template.",
						Optional:            true,
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
								Optional:            true,
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
	// iac_input_data at the workflow level corresponds to TemplatesIacInputData in the SDK.
	// Used when the workflow is instantiated from a workflow template (template_id).
	"iac_input_data": schema.SingleNestedAttribute{
		MarkdownDescription: "Top-level IaC input data for this workflow, used when the workflow is instantiated from a workflow template (`template_id`).",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"schema_type": schema.StringAttribute{
				MarkdownDescription: "Schema type for the input data (e.g. RAW_JSON).",
				Required:            true,
			},
			"data": schema.StringAttribute{
				MarkdownDescription: "Input data as a JSON string.",
				Optional:            true,
			},
		},
	},
	"user_schedules": schema.ListNestedAttribute{
		MarkdownDescription: "Scheduled run configuration.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Optional: true,
				},
				"desc": schema.StringAttribute{
					Optional: true,
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
	"approvers": schema.ListAttribute{
		MarkdownDescription: "List of approvers.",
		ElementType:         types.StringType,
		Optional:            true,
	},
	"number_of_approvals_required": schema.Int64Attribute{
		MarkdownDescription: "Number of approvals required.",
		Optional:            true,
	},
	"runner_constraints": schema.SingleNestedAttribute{
		MarkdownDescription: "Runner constraints for the workflow.",
		Optional:            true,
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
	"user_job_cpu": schema.Int64Attribute{
		MarkdownDescription: "CPU limit for the user job.",
		Optional:            true,
	},
	"user_job_memory": schema.Int64Attribute{
		MarkdownDescription: "Memory limit for the user job.",
		Optional:            true,
	},
	"input_schemas": schema.ListNestedAttribute{
		MarkdownDescription: "Input schema definitions for this workflow.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					Computed: true,
					Optional: true,
				},
				"name": schema.StringAttribute{
					Optional: true,
				},
				"description": schema.StringAttribute{
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
				"is_committed": schema.BoolAttribute{
					Optional: true,
					Computed: true,
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
				Optional:            true,
				MarkdownDescription: constants.MiniStepsWorkflowChaining,
				Attributes: map[string]schema.Attribute{
					"completed": ministepsWorkflowChaining,
					"errored":   ministepsWorkflowChaining,
				},
			},
		},
	},
}

// actionOrderAttrs defines the schema for a single ActionOrder (value in the order map).
var actionOrderAttrs = map[string]schema.Attribute{
	"parameters": schema.SingleNestedAttribute{
		MarkdownDescription: "Run configuration parameters for the action step.",
		Optional:            true,
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"terraform_action": schema.SingleNestedAttribute{
				MarkdownDescription: "Terraform-specific action parameters.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"action": schema.StringAttribute{
						MarkdownDescription: `Terraform action to execute. E.g., "apply", "plan", "destroy".`,
						Computed:            true,
					},
				},
			},
			"deployment_platform_config": schema.ListNestedAttribute{
				MarkdownDescription: "Deployment platform configuration.",
				Optional:            true,
				NestedObject:        schema.NestedAttributeObject{Attributes: deploymentPlatformConfigAttrs},
			},
			"wf_steps_config": schema.ListNestedAttribute{
				MarkdownDescription: "Workflow steps configuration.",
				Optional:            true,
				Computed:            true,
				NestedObject:        wfStepsConfigNestedObj,
			},
			"environment_variables": schema.ListNestedAttribute{
				MarkdownDescription: "Environment variables.",
				Optional:            true,
				Computed:            true,
				NestedObject:        schema.NestedAttributeObject{Attributes: envVarsAttrs},
			},
		},
	},
	"dependencies": schema.ListNestedAttribute{
		MarkdownDescription: "List of workflow dependencies that must complete before this step runs.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					MarkdownDescription: "ID of the dependent workflow.",
					Required:            true,
				},
				"condition": schema.SingleNestedAttribute{
					MarkdownDescription: "Condition that must be met by the dependency.",
					Optional:            true,
					Attributes: map[string]schema.Attribute{
						"latest_status": schema.StringAttribute{
							MarkdownDescription: "Required latest status of the dependency (e.g., COMPLETED).",
							Required:            true,
						},
					},
				},
			},
		},
	},
}

// Schema defines the schema for the resource.
func (r *stackTemplateRevisionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateRevisionId,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"parent_template_id": schema.StringAttribute{
				MarkdownDescription: "ID of the parent stack template to create the revision under.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"template_id": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateRevisionTemplateId,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"alias": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateRevisionAlias,
				Optional:            true,
				Computed:            true,
			},
			"notes": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateRevisionNotes,
				Optional:            true,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateRevisionDescription,
				Optional:            true,
				Computed:            true,
			},
			"source_config_kind": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateSourceConfigKindCommon,
				Required:            true,
			},
			"is_active": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateIsActiveCommon,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_public": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateIsPublicCommon,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Tags, "stack template revision"),
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"context_tags": schema.MapAttribute{
				MarkdownDescription: fmt.Sprintf(constants.ContextTags, "stack template revision"),
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			"deprecation": schema.SingleNestedAttribute{
				MarkdownDescription: constants.Deprecation,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"effective_date": schema.StringAttribute{
						MarkdownDescription: constants.TemplateRevisionDeprecationEffectiveDate,
						Optional:            true,
					},
					"message": schema.StringAttribute{
						MarkdownDescription: constants.TemplateRevisionDeprecation,
						Optional:            true,
					},
				},
			},
			"workflows_config": schema.SingleNestedAttribute{
				MarkdownDescription: constants.StackTemplateRevisionWorkflowsConfig,
				Optional:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"workflows": schema.ListNestedAttribute{
						MarkdownDescription: "List of workflows that make up the stack.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: workflowInStackAttrs,
						},
					},
				},
			},
			"actions": schema.MapNestedAttribute{
				MarkdownDescription: constants.StackTemplateRevisionActions,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the action.",
							Optional:            true,
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description of the action.",
							Optional:            true,
							Computed:            true,
						},
						"default": schema.BoolAttribute{
							MarkdownDescription: "Whether this is the default action.",
							Optional:            true,
							Computed:            true,
						},
						"order": schema.MapNestedAttribute{
							MarkdownDescription: "Ordered map of workflow IDs to their action configurations.",
							Optional:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: actionOrderAttrs,
							},
						},
					},
				},
			},
		},
	}
}

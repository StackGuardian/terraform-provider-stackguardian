package stacktemplaterevisiondatasource

import (
	"context"
	"fmt"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	stacktemplaterevision "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/stack_template_revision"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &stackTemplateRevisionDataSource{}
	_ datasource.DataSourceWithConfigure = &stackTemplateRevisionDataSource{}
)

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &stackTemplateRevisionDataSource{}
}

type stackTemplateRevisionDataSource struct {
	client  *sgclient.Client
	orgName string
}

func (d *stackTemplateRevisionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_template_revision"
}

func (d *stackTemplateRevisionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	provInfo, ok := req.ProviderData.(*customTypes.ProviderInfo)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *customTypes.ProviderInfo, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = provInfo.Client
	d.orgName = provInfo.Org_name
}

// Datasource-scoped schema helpers (mirrors resource/stack_template_revision/schema.go using datasource/schema types).

var dsEnvVarsAttrs = map[string]schema.Attribute{
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
}

var dsMountPointAttrs = map[string]schema.Attribute{
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
}

var dsWfStepsConfigNestedObj = schema.NestedAttributeObject{
	Attributes: map[string]schema.Attribute{
		"name": schema.StringAttribute{
			MarkdownDescription: constants.WfStepName,
			Computed:            true,
		},
		"environment_variables": schema.ListNestedAttribute{
			MarkdownDescription: constants.WfStepEnvVars,
			Computed:            true,
			NestedObject:        schema.NestedAttributeObject{Attributes: dsEnvVarsAttrs},
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
			NestedObject:        schema.NestedAttributeObject{Attributes: dsMountPointAttrs},
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

var dsDeploymentPlatformConfigAttrs = map[string]schema.Attribute{
	"kind": schema.StringAttribute{
		MarkdownDescription: constants.DeploymentPlatformKind,
		Computed:            true,
	},
	"config": schema.StringAttribute{
		MarkdownDescription: constants.DeploymentPlatformConfigDetails + " (JSON string)",
		Computed:            true,
	},
}

var dsWorkflowInStackAttrs = map[string]schema.Attribute{
	"resource_name": schema.StringAttribute{
		MarkdownDescription: "Name of the workflow resource within the stack.",
		Computed:            true,
	},
	"description": schema.StringAttribute{
		MarkdownDescription: "Description of this workflow.",
		Computed:            true,
	},
	"tags": schema.ListAttribute{
		MarkdownDescription: "Tags for the workflow.",
		ElementType:         types.StringType,
		Computed:            true,
	},
	"is_active": schema.StringAttribute{
		MarkdownDescription: constants.StackTemplateIsActiveCommon,
		Computed:            true,
	},
	"wf_type": schema.StringAttribute{
		MarkdownDescription: `Workflow type. Valid values: TERRAFORM, OPENTOFU, CUSTOM`,
		Computed:            true,
	},
	"wf_steps_config": schema.ListNestedAttribute{
		MarkdownDescription: "Workflow steps configuration.",
		Computed:            true,
		NestedObject:        dsWfStepsConfigNestedObj,
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
				NestedObject:        schema.NestedAttributeObject{Attributes: dsMountPointAttrs},
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: constants.TerraformTimeout,
				Computed:            true,
			},
			"post_apply_wf_steps_config": schema.ListNestedAttribute{
				MarkdownDescription: constants.TerraformPostApplyWfSteps,
				Computed:            true,
				NestedObject:        dsWfStepsConfigNestedObj,
			},
			"pre_apply_wf_steps_config": schema.ListNestedAttribute{
				MarkdownDescription: constants.TerraformPreApplyWfSteps,
				Computed:            true,
				NestedObject:        dsWfStepsConfigNestedObj,
			},
			"pre_plan_wf_steps_config": schema.ListNestedAttribute{
				MarkdownDescription: constants.TerraformPrePlanWfSteps,
				Computed:            true,
				NestedObject:        dsWfStepsConfigNestedObj,
			},
			"post_plan_wf_steps_config": schema.ListNestedAttribute{
				MarkdownDescription: constants.TerraformPostPlanWfSteps,
				Computed:            true,
				NestedObject:        dsWfStepsConfigNestedObj,
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
		},
	},
	"environment_variables": schema.ListNestedAttribute{
		MarkdownDescription: "Environment variables for the workflow.",
		Computed:            true,
		NestedObject:        schema.NestedAttributeObject{Attributes: dsEnvVarsAttrs},
	},
	"deployment_platform_config": schema.ListNestedAttribute{
		MarkdownDescription: "Deployment platform configuration.",
		Computed:            true,
		NestedObject:        schema.NestedAttributeObject{Attributes: dsDeploymentPlatformConfigAttrs},
	},
	"vcs_config": schema.SingleNestedAttribute{
		MarkdownDescription: "VCS (version control) configuration for the workflow.",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"iac_vcs_config": schema.SingleNestedAttribute{
				MarkdownDescription: "IaC VCS configuration.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"use_marketplace_template": schema.BoolAttribute{
						MarkdownDescription: "Whether to use a marketplace template.",
						Computed:            true,
					},
					"iac_template_id": schema.StringAttribute{
						MarkdownDescription: "ID of the IaC template from the marketplace.",
						Computed:            true,
					},
					"custom_source": schema.SingleNestedAttribute{
						MarkdownDescription: "Custom source configuration.",
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"source_config_dest_kind": schema.StringAttribute{
								MarkdownDescription: constants.RuntimeSourceDestKind,
								Computed:            true,
							},
							"config": schema.SingleNestedAttribute{
								MarkdownDescription: "Source configuration details.",
								Computed:            true,
								Attributes: map[string]schema.Attribute{
									"is_private": schema.BoolAttribute{
										Computed: true,
									},
									"auth": schema.StringAttribute{
										Computed:  true,
										Sensitive: true,
									},
									"working_dir": schema.StringAttribute{
										Computed: true,
									},
									"git_sparse_checkout_config": schema.StringAttribute{
										Computed: true,
									},
									"git_core_auto_crlf": schema.BoolAttribute{
										Computed: true,
									},
									"ref": schema.StringAttribute{
										Computed: true,
									},
									"repo": schema.StringAttribute{
										Computed: true,
									},
									"include_sub_module": schema.BoolAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"iac_input_data": schema.SingleNestedAttribute{
				MarkdownDescription: "IaC input data for the workflow.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"schema_id": schema.StringAttribute{
						Computed: true,
					},
					"schema_type": schema.StringAttribute{
						Computed: true,
					},
					"data": schema.StringAttribute{
						MarkdownDescription: "Input data as a JSON string.",
						Computed:            true,
					},
				},
			},
		},
	},
	"user_schedules": schema.ListNestedAttribute{
		MarkdownDescription: "Scheduled run configuration.",
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Computed: true,
				},
				"desc": schema.StringAttribute{
					Computed: true,
				},
				"cron": schema.StringAttribute{
					MarkdownDescription: constants.UserScheduleCron,
					Computed:            true,
				},
				"state": schema.StringAttribute{
					MarkdownDescription: constants.UserScheduleState,
					Computed:            true,
				},
			},
		},
	},
	"approvers": schema.ListAttribute{
		MarkdownDescription: "List of approvers.",
		ElementType:         types.StringType,
		Computed:            true,
	},
	"number_of_approvals_required": schema.Int64Attribute{
		MarkdownDescription: "Number of approvals required.",
		Computed:            true,
	},
	"runner_constraints": schema.SingleNestedAttribute{
		MarkdownDescription: "Runner constraints for the workflow.",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				MarkdownDescription: constants.RunnerConstraintsType,
				Computed:            true,
			},
			"names": schema.ListAttribute{
				MarkdownDescription: constants.RunnerConstraintsNames,
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	},
	"user_job_cpu": schema.Int64Attribute{
		MarkdownDescription: "CPU limit for the user job.",
		Computed:            true,
	},
	"user_job_memory": schema.Int64Attribute{
		MarkdownDescription: "Memory limit for the user job.",
		Computed:            true,
	},
}

var dsActionOrderAttrs = map[string]schema.Attribute{
	"parameters": schema.SingleNestedAttribute{
		MarkdownDescription: "Run configuration parameters for the action step.",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"terraform_action": schema.SingleNestedAttribute{
				MarkdownDescription: "Terraform-specific action parameters.",
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
				Computed:            true,
				NestedObject:        schema.NestedAttributeObject{Attributes: dsDeploymentPlatformConfigAttrs},
			},
			"wf_steps_config": schema.ListNestedAttribute{
				MarkdownDescription: "Workflow steps configuration.",
				Computed:            true,
				NestedObject:        dsWfStepsConfigNestedObj,
			},
			"environment_variables": schema.ListNestedAttribute{
				MarkdownDescription: "Environment variables.",
				Computed:            true,
				NestedObject:        schema.NestedAttributeObject{Attributes: dsEnvVarsAttrs},
			},
		},
	},
	"dependencies": schema.ListNestedAttribute{
		MarkdownDescription: "List of workflow dependencies that must complete before this step runs.",
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					MarkdownDescription: "ID of the dependent workflow.",
					Computed:            true,
				},
				"condition": schema.SingleNestedAttribute{
					MarkdownDescription: "Condition that must be met by the dependency.",
					Computed:            true,
					Attributes: map[string]schema.Attribute{
						"latest_status": schema.StringAttribute{
							MarkdownDescription: "Required latest status of the dependency (e.g., COMPLETED).",
							Computed:            true,
						},
					},
				},
			},
		},
	},
}

func (d *stackTemplateRevisionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "> **Note:** This data source is currently in **BETA**. Features and behavior may change.\n\nUse this data source to read a stack template revision.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: constants.DatasourceId,
				Required:            true,
			},
			"template_id": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateRevisionTemplateId,
				Computed:            true,
			},
			"alias": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateRevisionAlias,
				Computed:            true,
			},
			"notes": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateRevisionNotes,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateRevisionDescription,
				Computed:            true,
			},
			"source_config_kind": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateSourceConfigKindCommon,
				Computed:            true,
			},
			"is_active": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateIsActiveCommon,
				Computed:            true,
			},
			"is_public": schema.StringAttribute{
				MarkdownDescription: constants.StackTemplateIsPublicCommon,
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: fmt.Sprintf(constants.Tags, "stack template revision"),
				ElementType:         types.StringType,
				Computed:            true,
			},
			"context_tags": schema.MapAttribute{
				MarkdownDescription: fmt.Sprintf(constants.ContextTags, "stack template revision"),
				ElementType:         types.StringType,
				Computed:            true,
			},
			"deprecation": schema.SingleNestedAttribute{
				MarkdownDescription: constants.Deprecation,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"effective_date": schema.StringAttribute{
						MarkdownDescription: constants.TemplateRevisionDeprecationEffectiveDate,
						Computed:            true,
					},
					"message": schema.StringAttribute{
						MarkdownDescription: constants.TemplateRevisionDeprecation,
						Computed:            true,
					},
				},
			},
			"workflows_config": schema.SingleNestedAttribute{
				MarkdownDescription: constants.StackTemplateRevisionWorkflowsConfig,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"workflows": schema.ListNestedAttribute{
						MarkdownDescription: "List of workflows that make up the stack.",
						Computed:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: dsWorkflowInStackAttrs,
						},
					},
				},
			},
			"actions": schema.MapNestedAttribute{
				MarkdownDescription: constants.StackTemplateRevisionActions,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the action.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description of the action.",
							Computed:            true,
						},
						"default": schema.BoolAttribute{
							MarkdownDescription: "Whether this is the default action.",
							Computed:            true,
						},
						"order": schema.MapNestedAttribute{
							MarkdownDescription: "Ordered map of workflow IDs to their action configurations.",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: dsActionOrderAttrs,
							},
						},
					},
				},
			},
		},
	}
}

func (d *stackTemplateRevisionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config stacktemplaterevision.StackTemplateRevisionResourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	revisionID := config.Id.ValueString()
	if revisionID == "" {
		resp.Diagnostics.AddError("id must be provided", "")
		return
	}

	readResp, err := d.client.StackTemplateRevisions.ReadStackTemplateRevision(ctx, d.orgName, revisionID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read stack template revision.", err.Error())
		return
	}

	if readResp == nil {
		resp.Diagnostics.AddError("Error reading stack template revision", "API response is empty")
		return
	}

	model, diags := stacktemplaterevision.BuildAPIModelToStackTemplateRevisionModel(ctx, &readResp.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

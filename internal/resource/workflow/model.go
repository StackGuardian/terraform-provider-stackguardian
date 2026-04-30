package workflow

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type workflowResourceModel struct {
	Id                         types.String `tfsdk:"id"`
	WorkflowGroupId            types.String `tfsdk:"workflow_group_id"`
	ResourceName               types.String `tfsdk:"resource_name"`
	Description                types.String `tfsdk:"description"`
	WfType                     types.String `tfsdk:"wf_type"`
	EnvironmentVariables       types.List   `tfsdk:"environment_variables"`
	InputSchemas               types.List   `tfsdk:"input_schemas"`
	MiniSteps                  types.Object `tfsdk:"mini_steps"`
	RunnerConstraints          types.Object `tfsdk:"runner_constraints"`
	Tags                       types.List   `tfsdk:"tags"`
	UserSchedules              types.List   `tfsdk:"user_schedules"`
	ContextTags                types.Map    `tfsdk:"context_tags"`
	Approvers                  types.List   `tfsdk:"approvers"`
	NumberOfApprovalsRequired  types.Int64  `tfsdk:"number_of_approvals_required"`
	UserJobCpu                 types.Int64  `tfsdk:"user_job_cpu"`
	UserJobMemory              types.Int64  `tfsdk:"user_job_memory"`
	VcsConfig                  types.Object `tfsdk:"vcs_config"`
	TerraformConfig            types.Object `tfsdk:"terraform_config"`
	DeploymentPlatformConfig   types.Object `tfsdk:"deployment_platform_config"`
	WfStepsConfig              types.List   `tfsdk:"wf_steps_config"`
}

type EnvVarModel struct {
	Kind   types.String `tfsdk:"kind"`
	Config types.Object `tfsdk:"config"`
}

type EnvVarConfigModel struct {
	VarName   types.String `tfsdk:"var_name"`
	SecretId  types.String `tfsdk:"secret_id"`
	TextValue types.String `tfsdk:"text_value"`
}

type InputSchemaModel struct {
	Name         types.String `tfsdk:"name"`
	Type         types.String `tfsdk:"type"`
	EncodedData  types.String `tfsdk:"encoded_data"`
	UiSchemaData types.String `tfsdk:"ui_schema_data"`
}

type UserScheduleModel struct {
	Cron  types.String `tfsdk:"cron"`
	State types.String `tfsdk:"state"`
	Desc  types.String `tfsdk:"desc"`
	Name  types.String `tfsdk:"name"`
}

type VcsConfigModel struct {
	RepoUrl                        types.String `tfsdk:"repo_url"`
	RepoBranch                     types.String `tfsdk:"repo_branch"`
	TerraformBackendConfigFilePath types.String `tfsdk:"terraform_backend_config_file_path"`
}

type TerraformConfigModel struct {
	TerraformVersion        types.String `tfsdk:"terraform_version"`
	DriftCheck              types.Bool   `tfsdk:"drift_check"`
	DriftCron               types.String `tfsdk:"drift_cron"`
	ManagedTerraformState   types.Bool   `tfsdk:"managed_terraform_state"`
	ApprovalPreApply        types.Bool   `tfsdk:"approval_pre_apply"`
	TerraformPlanOptions    types.String `tfsdk:"terraform_plan_options"`
	TerraformInitOptions    types.String `tfsdk:"terraform_init_options"`
	TerraformBinPath        types.List   `tfsdk:"terraform_bin_path"`
	Timeout                 types.Int64  `tfsdk:"timeout"`
	PostApplyWfStepsConfig  types.List   `tfsdk:"post_apply_wf_steps_config"`
	PreApplyWfStepsConfig   types.List   `tfsdk:"pre_apply_wf_steps_config"`
	PrePlanWfStepsConfig    types.List   `tfsdk:"pre_plan_wf_steps_config"`
	PostPlanWfStepsConfig   types.List   `tfsdk:"post_plan_wf_steps_config"`
	PreInitHooks            types.List   `tfsdk:"pre_init_hooks"`
	PrePlanHooks            types.List   `tfsdk:"pre_plan_hooks"`
	PostPlanHooks           types.List   `tfsdk:"post_plan_hooks"`
	PreApplyHooks           types.List   `tfsdk:"pre_apply_hooks"`
	PostApplyHooks          types.List   `tfsdk:"post_apply_hooks"`
	RunPreInitHooksOnDrift  types.Bool   `tfsdk:"run_pre_init_hooks_on_drift"`
}

type DeploymentPlatformConfigModel struct {
	Kind   types.String `tfsdk:"kind"`
	Config types.Object `tfsdk:"config"`
}

type DeploymentPlatformConfigDetailModel struct {
	IntegrationId types.String `tfsdk:"integration_id"`
	ProfileName   types.String `tfsdk:"profile_name"`
}

type RunnerConstraintsModel struct {
	Type  types.String `tfsdk:"type"`
	Names types.List   `tfsdk:"names"`
}

type MiniStepsModel struct {
	Notifications types.Object `tfsdk:"notifications"`
	Webhooks      types.Object `tfsdk:"webhooks"`
	WfChaining    types.Object `tfsdk:"wf_chaining"`
}

type WfStepsConfigModel struct {
	Name                  types.String `tfsdk:"name"`
	EnvironmentVariables  types.List   `tfsdk:"environment_variables"`
	Approval              types.Bool   `tfsdk:"approval"`
	Timeout               types.Int64  `tfsdk:"timeout"`
	CmdOverride           types.String `tfsdk:"cmd_override"`
	MountPoints           types.List   `tfsdk:"mount_points"`
	WfStepTemplateId      types.String `tfsdk:"wf_step_template_id"`
	WfStepInputData       types.Object `tfsdk:"wf_step_input_data"`
}

type WfStepInputDataModel struct {
	SchemaType types.String `tfsdk:"schema_type"`
	Data       types.String `tfsdk:"data"`
}

type MountPointModel struct {
	Source   types.String `tfsdk:"source"`
	Target   types.String `tfsdk:"target"`
	ReadOnly types.Bool   `tfsdk:"read_only"`
}

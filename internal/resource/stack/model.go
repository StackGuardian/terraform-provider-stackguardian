package stack

import (
	"context"
	"fmt"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/sg-sdk-go/stacktemplaterevisions"
	"github.com/StackGuardian/sg-sdk-go/workflowtemplaterevisions"
	"github.com/StackGuardian/sg-sdk-go/workflowtemplates"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/expanders"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/flatteners"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ---------------------------------------------------------------------------
// Root resource model
// ---------------------------------------------------------------------------

type StackResourceModel struct {
	Id                       types.String `tfsdk:"id"`
	WorkflowGroupId          types.String `tfsdk:"workflow_group_id"`
	ResourceName             types.String `tfsdk:"resource_name"`
	Description              types.String `tfsdk:"description"`
	Tags                     types.List   `tfsdk:"tags"`
	EnvironmentVariables     types.List   `tfsdk:"environment_variables"`
	DeploymentPlatformConfig types.List   `tfsdk:"deployment_platform_config"`
	DefaultActions           types.Map    `tfsdk:"default_actions"`
	CustomActions            types.Map    `tfsdk:"custom_actions"`
	TemplateGroupId          types.String `tfsdk:"template_group_id"`
	WorkflowsConfig          types.Object `tfsdk:"workflows_config"`
	UserSchedules            types.List   `tfsdk:"user_schedules"`
	ContextTags              types.Map    `tfsdk:"context_tags"`
	MiniSteps                types.Object `tfsdk:"mini_steps"`
}

// ---------------------------------------------------------------------------
// Environment variables
// ---------------------------------------------------------------------------

type EnvVarConfigModel struct {
	VarName   types.String `tfsdk:"var_name"`
	SecretId  types.String `tfsdk:"secret_id"`
	TextValue types.String `tfsdk:"text_value"`
}

func (EnvVarConfigModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"var_name":   types.StringType,
		"secret_id":  types.StringType,
		"text_value": types.StringType,
	}
}

// EnvironmentVariableModel represents a single environment variable configuration.
type EnvironmentVariableModel struct {
	Config types.Object `tfsdk:"config"`
	Kind   types.String `tfsdk:"kind"`
}

func (EnvironmentVariableModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"config": types.ObjectType{AttrTypes: EnvVarConfigModel{}.AttributeTypes()},
		"kind":   types.StringType,
	}
}

// ---------------------------------------------------------------------------
// Mount points
// ---------------------------------------------------------------------------

type MountPointModel struct {
	Source   types.String `tfsdk:"source"`
	Target   types.String `tfsdk:"target"`
	ReadOnly types.Bool   `tfsdk:"read_only"`
}

func (MountPointModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"source":    types.StringType,
		"target":    types.StringType,
		"read_only": types.BoolType,
	}
}

// ---------------------------------------------------------------------------
// Wf step input data
// ---------------------------------------------------------------------------

type WfStepInputDataModel struct {
	SchemaType types.String `tfsdk:"schema_type"`
	Data       types.String `tfsdk:"data"`
}

func (WfStepInputDataModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"schema_type": types.StringType,
		"data":        types.StringType,
	}
}

// ---------------------------------------------------------------------------
// WfStepsConfig
// ---------------------------------------------------------------------------

type WfStepsConfigModel struct {
	Name                 types.String `tfsdk:"name"`
	EnvironmentVariables types.List   `tfsdk:"environment_variables"`
	Approval             types.Bool   `tfsdk:"approval"`
	Timeout              types.Int64  `tfsdk:"timeout"`
	CmdOverride          types.String `tfsdk:"cmd_override"`
	MountPoints          types.List   `tfsdk:"mount_points"`
	WfStepTemplateId     types.String `tfsdk:"wf_step_template_id"`
	WfStepInputData      types.Object `tfsdk:"wf_step_input_data"`
}

func (WfStepsConfigModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                  types.StringType,
		"environment_variables": types.ListType{ElemType: types.ObjectType{AttrTypes: EnvironmentVariableModel{}.AttributeTypes()}},
		"approval":              types.BoolType,
		"timeout":               types.Int64Type,
		"cmd_override":          types.StringType,
		"mount_points":          types.ListType{ElemType: types.ObjectType{AttrTypes: MountPointModel{}.AttributeTypes()}},
		"wf_step_template_id":   types.StringType,
		"wf_step_input_data":    types.ObjectType{AttrTypes: WfStepInputDataModel{}.AttributeTypes()},
	}
}

// ---------------------------------------------------------------------------
// Terraform config
// ---------------------------------------------------------------------------

type TerraformConfigModel struct {
	TerraformVersion       types.String `tfsdk:"terraform_version"`
	DriftCheck             types.Bool   `tfsdk:"drift_check"`
	DriftCron              types.String `tfsdk:"drift_cron"`
	ManagedTerraformState  types.Bool   `tfsdk:"managed_terraform_state"`
	ApprovalPreApply       types.Bool   `tfsdk:"approval_pre_apply"`
	TerraformPlanOptions   types.String `tfsdk:"terraform_plan_options"`
	TerraformInitOptions   types.String `tfsdk:"terraform_init_options"`
	TerraformBinPath       types.List   `tfsdk:"terraform_bin_path"`
	Timeout                types.Int64  `tfsdk:"timeout"`
	PostApplyWfStepsConfig types.List   `tfsdk:"post_apply_wf_steps_config"`
	PreApplyWfStepsConfig  types.List   `tfsdk:"pre_apply_wf_steps_config"`
	PrePlanWfStepsConfig   types.List   `tfsdk:"pre_plan_wf_steps_config"`
	PostPlanWfStepsConfig  types.List   `tfsdk:"post_plan_wf_steps_config"`
	PreInitHooks           types.List   `tfsdk:"pre_init_hooks"`
	PrePlanHooks           types.List   `tfsdk:"pre_plan_hooks"`
	PostPlanHooks          types.List   `tfsdk:"post_plan_hooks"`
	PreApplyHooks          types.List   `tfsdk:"pre_apply_hooks"`
	PostApplyHooks         types.List   `tfsdk:"post_apply_hooks"`
	RunPreInitHooksOnDrift types.Bool   `tfsdk:"run_pre_init_hooks_on_drift"`
}

func (TerraformConfigModel) AttributeTypes() map[string]attr.Type {
	wfStepsListType := types.ListType{ElemType: types.ObjectType{AttrTypes: WfStepsConfigModel{}.AttributeTypes()}}
	return map[string]attr.Type{
		"terraform_version":           types.StringType,
		"drift_check":                 types.BoolType,
		"drift_cron":                  types.StringType,
		"managed_terraform_state":     types.BoolType,
		"approval_pre_apply":          types.BoolType,
		"terraform_plan_options":      types.StringType,
		"terraform_init_options":      types.StringType,
		"terraform_bin_path":          types.ListType{ElemType: types.ObjectType{AttrTypes: MountPointModel{}.AttributeTypes()}},
		"timeout":                     types.Int64Type,
		"post_apply_wf_steps_config":  wfStepsListType,
		"pre_apply_wf_steps_config":   wfStepsListType,
		"pre_plan_wf_steps_config":    wfStepsListType,
		"post_plan_wf_steps_config":   wfStepsListType,
		"pre_init_hooks":              types.ListType{ElemType: types.StringType},
		"pre_plan_hooks":              types.ListType{ElemType: types.StringType},
		"post_plan_hooks":             types.ListType{ElemType: types.StringType},
		"pre_apply_hooks":             types.ListType{ElemType: types.StringType},
		"post_apply_hooks":            types.ListType{ElemType: types.StringType},
		"run_pre_init_hooks_on_drift": types.BoolType,
	}
}

// ---------------------------------------------------------------------------
// Deployment platform config
// ---------------------------------------------------------------------------

type DeploymentPlatformConfigConfigModel struct {
	IntegrationId types.String `tfsdk:"integration_id"`
	ProfileName   types.String `tfsdk:"profile_name"`
}

func (DeploymentPlatformConfigConfigModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"integration_id": types.StringType,
		"profile_name":   types.StringType,
	}
}

// DeploymentPlatformConfigModel represents deployment platform configuration.
type DeploymentPlatformConfigModel struct {
	Kind   types.String `tfsdk:"kind"`
	Config types.Object `tfsdk:"config"`
}

func (DeploymentPlatformConfigModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"kind":   types.StringType,
		"config": types.ObjectType{AttrTypes: DeploymentPlatformConfigConfigModel{}.AttributeTypes()},
	}
}

// ---------------------------------------------------------------------------
// Runner constraints
// ---------------------------------------------------------------------------

type RunnerConstraintsModel struct {
	Type  types.String `tfsdk:"type"`
	Names types.List   `tfsdk:"names"`
}

func (RunnerConstraintsModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":  types.StringType,
		"names": types.ListType{ElemType: types.StringType},
	}
}

// ---------------------------------------------------------------------------
// User schedules
// ---------------------------------------------------------------------------

// StackActionInputsModel is the "inputs" payload for a top-level stack user
// schedule. It corresponds to the SDK's StackAction, which is just an action
// type identifier (not the richer {action, resource_name} shape).
type StackActionInputsModel struct {
	ActionType types.String `tfsdk:"action_type"`
}

func (StackActionInputsModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"action_type": types.StringType,
	}
}

// UserSchedulesModel represents a top-level stack user schedule
// (sgsdkgo.StackUserSchedules), which carries a required Inputs payload.
type UserSchedulesModel struct {
	Name   types.String `tfsdk:"name"`
	Desc   types.String `tfsdk:"desc"`
	Cron   types.String `tfsdk:"cron"`
	State  types.String `tfsdk:"state"`
	Inputs types.Object `tfsdk:"inputs"`
}

func (UserSchedulesModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":   types.StringType,
		"desc":   types.StringType,
		"cron":   types.StringType,
		"state":  types.StringType,
		"inputs": types.ObjectType{AttrTypes: StackActionInputsModel{}.AttributeTypes()},
	}
}

// WfUserSchedulesModel represents a per-workflow user schedule
// (sgsdkgo.UserSchedules), which has no Inputs field.
type WfUserSchedulesModel struct {
	Name  types.String `tfsdk:"name"`
	Desc  types.String `tfsdk:"desc"`
	Cron  types.String `tfsdk:"cron"`
	State types.String `tfsdk:"state"`
}

func (WfUserSchedulesModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":  types.StringType,
		"desc":  types.StringType,
		"cron":  types.StringType,
		"state": types.StringType,
	}
}

// ---------------------------------------------------------------------------
// VCS config (for a workflow in the stack)
// ---------------------------------------------------------------------------

// IacVcsConfigModel has no custom_source: stacks never use it (unlike
// stack_template_revision/workflow_from_template's copies of this shape).
type IacVcsConfigModel struct {
	UseMarketplaceTemplate types.Bool   `tfsdk:"use_marketplace_template"`
	IacTemplateId          types.String `tfsdk:"iac_template_id"`
}

func (IacVcsConfigModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"use_marketplace_template": types.BoolType,
		"iac_template_id":          types.StringType,
	}
}

// VcsIacInputDataModel is vcs_config.iac_input_data, corresponding to the
// SDK's sgsdkgo.IacInputData (schema_id + schema_type + data).
type VcsIacInputDataModel struct {
	SchemaId   types.String `tfsdk:"schema_id"`
	SchemaType types.String `tfsdk:"schema_type"`
	Data       types.String `tfsdk:"data"`
}

func (VcsIacInputDataModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"schema_id":   types.StringType,
		"schema_type": types.StringType,
		"data":        types.StringType,
	}
}

type VcsConfigModel struct {
	IacVcsConfig types.Object `tfsdk:"iac_vcs_config"`
	IacInputData types.Object `tfsdk:"iac_input_data"`
}

func (VcsConfigModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"iac_vcs_config": types.ObjectType{AttrTypes: IacVcsConfigModel{}.AttributeTypes()},
		"iac_input_data": types.ObjectType{AttrTypes: VcsIacInputDataModel{}.AttributeTypes()},
	}
}

// ---------------------------------------------------------------------------
// Input schemas (for workflows_config workflows)
// ---------------------------------------------------------------------------

type StackInputSchemaModel struct {
	Name         types.String `tfsdk:"name"`
	Type         types.String `tfsdk:"type"`
	EncodedData  types.String `tfsdk:"encoded_data"`
	UiSchemaData types.String `tfsdk:"ui_schema_data"`
}

func (StackInputSchemaModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":           types.StringType,
		"type":           types.StringType,
		"encoded_data":   types.StringType,
		"ui_schema_data": types.StringType,
	}
}

// ---------------------------------------------------------------------------
// MiniSteps (shared shape for the stack-level and per-workflow mini_steps)
// ---------------------------------------------------------------------------

type MinistepsNotificationRecipientsModel struct {
	Recipients types.List `tfsdk:"recipients"`
}

func (MinistepsNotificationRecipientsModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"recipients": types.ListType{ElemType: types.StringType},
	}
}

type MinistepsWebhooksModel struct {
	WebhookName   types.String `tfsdk:"webhook_name"`
	WebhookUrl    types.String `tfsdk:"webhook_url"`
	WebhookSecret types.String `tfsdk:"webhook_secret"`
}

func (MinistepsWebhooksModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"webhook_name":   types.StringType,
		"webhook_url":    types.StringType,
		"webhook_secret": types.StringType,
	}
}

type MinistepsWorkflowChainingModel struct {
	WorkflowGroupId    types.String `tfsdk:"workflow_group_id"`
	StackId            types.String `tfsdk:"stack_id"`
	StackRunPayload    types.String `tfsdk:"stack_run_payload"`
	WorkflowId         types.String `tfsdk:"workflow_id"`
	WorkflowRunPayload types.String `tfsdk:"workflow_run_payload"`
}

func (MinistepsWorkflowChainingModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"workflow_group_id":    types.StringType,
		"stack_id":             types.StringType,
		"stack_run_payload":    types.StringType,
		"workflow_id":          types.StringType,
		"workflow_run_payload": types.StringType,
	}
}

type MinistepsEmailModel struct {
	ApprovalRequired types.List `tfsdk:"approval_required"`
	Cancelled        types.List `tfsdk:"cancelled"`
	Completed        types.List `tfsdk:"completed"`
	DriftDetected    types.List `tfsdk:"drift_detected"`
	Errored          types.List `tfsdk:"errored"`
}

func (MinistepsEmailModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"approval_required": types.ListType{ElemType: types.ObjectType{AttrTypes: MinistepsNotificationRecipientsModel{}.AttributeTypes()}},
		"cancelled":         types.ListType{ElemType: types.ObjectType{AttrTypes: MinistepsNotificationRecipientsModel{}.AttributeTypes()}},
		"completed":         types.ListType{ElemType: types.ObjectType{AttrTypes: MinistepsNotificationRecipientsModel{}.AttributeTypes()}},
		"drift_detected":    types.ListType{ElemType: types.ObjectType{AttrTypes: MinistepsNotificationRecipientsModel{}.AttributeTypes()}},
		"errored":           types.ListType{ElemType: types.ObjectType{AttrTypes: MinistepsNotificationRecipientsModel{}.AttributeTypes()}},
	}
}

type MinistepsNotificationsModel struct {
	Email types.Object `tfsdk:"email"`
}

func (MinistepsNotificationsModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"email": types.ObjectType{AttrTypes: MinistepsEmailModel{}.AttributeTypes()},
	}
}

type MinistepsWebhooksContainerModel struct {
	ApprovalRequired types.List `tfsdk:"approval_required"`
	Cancelled        types.List `tfsdk:"cancelled"`
	Completed        types.List `tfsdk:"completed"`
	DriftDetected    types.List `tfsdk:"drift_detected"`
	Errored          types.List `tfsdk:"errored"`
}

func (MinistepsWebhooksContainerModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"approval_required": types.ListType{ElemType: types.ObjectType{AttrTypes: MinistepsWebhooksModel{}.AttributeTypes()}},
		"cancelled":         types.ListType{ElemType: types.ObjectType{AttrTypes: MinistepsWebhooksModel{}.AttributeTypes()}},
		"completed":         types.ListType{ElemType: types.ObjectType{AttrTypes: MinistepsWebhooksModel{}.AttributeTypes()}},
		"drift_detected":    types.ListType{ElemType: types.ObjectType{AttrTypes: MinistepsWebhooksModel{}.AttributeTypes()}},
		"errored":           types.ListType{ElemType: types.ObjectType{AttrTypes: MinistepsWebhooksModel{}.AttributeTypes()}},
	}
}

type MinistepsWfChainingContainerModel struct {
	Completed types.List `tfsdk:"completed"`
	Errored   types.List `tfsdk:"errored"`
}

func (MinistepsWfChainingContainerModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"completed": types.ListType{ElemType: types.ObjectType{AttrTypes: MinistepsWorkflowChainingModel{}.AttributeTypes()}},
		"errored":   types.ListType{ElemType: types.ObjectType{AttrTypes: MinistepsWorkflowChainingModel{}.AttributeTypes()}},
	}
}

type MinistepsModel struct {
	Notifications types.Object `tfsdk:"notifications"`
	Webhooks      types.Object `tfsdk:"webhooks"`
	WfChaining    types.Object `tfsdk:"wf_chaining"`
}

func (MinistepsModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"notifications": types.ObjectType{AttrTypes: MinistepsNotificationsModel{}.AttributeTypes()},
		"webhooks":      types.ObjectType{AttrTypes: MinistepsWebhooksContainerModel{}.AttributeTypes()},
		"wf_chaining":   types.ObjectType{AttrTypes: MinistepsWfChainingContainerModel{}.AttributeTypes()},
	}
}

// ---------------------------------------------------------------------------
// WorkflowsConfig
// ---------------------------------------------------------------------------

// WorkflowInStackModel corresponds to sgsdkgo.StackWorkflowsConfigWorkflow.
type WorkflowInStackModel struct {
	Id                        types.String `tfsdk:"id"`
	ResourceName              types.String `tfsdk:"resource_name"`
	Description               types.String `tfsdk:"description"`
	Tags                      types.List   `tfsdk:"tags"`
	WfType                    types.String `tfsdk:"wf_type"`
	ParallelExecution         types.String `tfsdk:"parallel_execution"`
	WfStepsConfig             types.List   `tfsdk:"wf_steps_config"`
	TerraformConfig           types.Object `tfsdk:"terraform_config"`
	EnvironmentVariables      types.List   `tfsdk:"environment_variables"`
	DeploymentPlatformConfig  types.List   `tfsdk:"deployment_platform_config"`
	TemplateId                types.String `tfsdk:"template_id"`
	IsActive                  types.String `tfsdk:"is_active"`
	VcsConfig                 types.Object `tfsdk:"vcs_config"`
	InputSchemas              types.List   `tfsdk:"input_schemas"`
	Approvers                 types.List   `tfsdk:"approvers"`
	NumberOfApprovalsRequired types.Int64  `tfsdk:"number_of_approvals_required"`
	UserJobCpu                types.Int64  `tfsdk:"user_job_cpu"`
	UserJobMemory             types.Int64  `tfsdk:"user_job_memory"`
	UserSchedules             types.List   `tfsdk:"user_schedules"`
	MiniSteps                 types.Object `tfsdk:"mini_steps"`
	ContextTags               types.Map    `tfsdk:"context_tags"`
	RunnerConstraints         types.Object `tfsdk:"runner_constraints"`
}

func (WorkflowInStackModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                           types.StringType,
		"resource_name":                types.StringType,
		"description":                  types.StringType,
		"tags":                         types.ListType{ElemType: types.StringType},
		"wf_type":                      types.StringType,
		"parallel_execution":           types.StringType,
		"wf_steps_config":              types.ListType{ElemType: types.ObjectType{AttrTypes: WfStepsConfigModel{}.AttributeTypes()}},
		"terraform_config":             types.ObjectType{AttrTypes: TerraformConfigModel{}.AttributeTypes()},
		"environment_variables":        types.ListType{ElemType: types.ObjectType{AttrTypes: EnvironmentVariableModel{}.AttributeTypes()}},
		"deployment_platform_config":   types.ListType{ElemType: types.ObjectType{AttrTypes: DeploymentPlatformConfigModel{}.AttributeTypes()}},
		"template_id":                  types.StringType,
		"is_active":                    types.StringType,
		"vcs_config":                   types.ObjectType{AttrTypes: VcsConfigModel{}.AttributeTypes()},
		"input_schemas":                types.ListType{ElemType: types.ObjectType{AttrTypes: StackInputSchemaModel{}.AttributeTypes()}},
		"approvers":                    types.ListType{ElemType: types.StringType},
		"number_of_approvals_required": types.Int64Type,
		"user_job_cpu":                 types.Int64Type,
		"user_job_memory":              types.Int64Type,
		"user_schedules":               types.ListType{ElemType: types.ObjectType{AttrTypes: WfUserSchedulesModel{}.AttributeTypes()}},
		"mini_steps":                   types.ObjectType{AttrTypes: MinistepsModel{}.AttributeTypes()},
		"context_tags":                 types.MapType{ElemType: types.StringType},
		"runner_constraints":           types.ObjectType{AttrTypes: RunnerConstraintsModel{}.AttributeTypes()},
	}
}

type WorkflowsConfigModel struct {
	Workflows types.List `tfsdk:"workflows"`
}

func (WorkflowsConfigModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"workflows": types.ListType{ElemType: types.ObjectType{AttrTypes: WorkflowInStackModel{}.AttributeTypes()}},
	}
}

// ---------------------------------------------------------------------------
// Actions
// ---------------------------------------------------------------------------

type TerraformActionModel struct {
	Action types.String `tfsdk:"action"`
}

func (TerraformActionModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"action": types.StringType,
	}
}

type StackActionParametersModel struct {
	TerraformAction          types.Object `tfsdk:"terraform_action"`
	DeploymentPlatformConfig types.List   `tfsdk:"deployment_platform_config"`
	WfStepsConfig            types.List   `tfsdk:"wf_steps_config"`
	EnvironmentVariables     types.List   `tfsdk:"environment_variables"`
}

func (StackActionParametersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"terraform_action":           types.ObjectType{AttrTypes: TerraformActionModel{}.AttributeTypes()},
		"deployment_platform_config": types.ListType{ElemType: types.ObjectType{AttrTypes: DeploymentPlatformConfigModel{}.AttributeTypes()}},
		"wf_steps_config":            types.ListType{ElemType: types.ObjectType{AttrTypes: WfStepsConfigModel{}.AttributeTypes()}},
		"environment_variables":      types.ListType{ElemType: types.ObjectType{AttrTypes: EnvironmentVariableModel{}.AttributeTypes()}},
	}
}

type ActionDependencyConditionModel struct {
	LatestStatus types.String `tfsdk:"latest_status"`
}

func (ActionDependencyConditionModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"latest_status": types.StringType,
	}
}

type ActionDependencyModel struct {
	Id        types.String `tfsdk:"id"`
	Condition types.Object `tfsdk:"condition"`
}

func (ActionDependencyModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":        types.StringType,
		"condition": types.ObjectType{AttrTypes: ActionDependencyConditionModel{}.AttributeTypes()},
	}
}

type ActionOrderModel struct {
	Parameters   types.Object `tfsdk:"parameters"`
	Dependencies types.List   `tfsdk:"dependencies"`
}

func (ActionOrderModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"parameters":   types.ObjectType{AttrTypes: StackActionParametersModel{}.AttributeTypes()},
		"dependencies": types.ListType{ElemType: types.ObjectType{AttrTypes: ActionDependencyModel{}.AttributeTypes()}},
	}
}

// ActionsModel represents a single action value in the actions map.
type ActionsModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Order       types.Map    `tfsdk:"order"`
}

func (ActionsModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":        types.StringType,
		"description": types.StringType,
		"order":       types.MapType{ElemType: types.ObjectType{AttrTypes: ActionOrderModel{}.AttributeTypes()}},
	}
}

// ---------------------------------------------------------------------------
// Converters: Terraform model -> SDK API types
// ---------------------------------------------------------------------------

func expandEnvironmentVariables(ctx context.Context, list types.List) ([]*sgsdkgo.EnvVars, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var models []EnvironmentVariableModel
	if diags := list.ElementsAs(ctx, &models, false); diags.HasError() {
		return nil, diags
	}
	result := make([]*sgsdkgo.EnvVars, len(models))
	for i, m := range models {
		ev := &sgsdkgo.EnvVars{
			Kind: sgsdkgo.EnvVarsKindEnum(m.Kind.ValueString()),
		}
		if !m.Config.IsNull() && !m.Config.IsUnknown() {
			var cfgModel EnvVarConfigModel
			if diags := m.Config.As(ctx, &cfgModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); diags.HasError() {
				return nil, diags
			}
			ev.Config = &sgsdkgo.EnvVarConfig{
				VarName:   cfgModel.VarName.ValueString(),
				SecretId:  cfgModel.SecretId.ValueStringPointer(),
				TextValue: cfgModel.TextValue.ValueStringPointer(),
			}
		}
		result[i] = ev
	}
	return result, nil
}

func flattenEnvironmentVariables(ctx context.Context, envVars []*sgsdkgo.EnvVars) (types.List, diag.Diagnostics) {
	nullList := types.ListNull(types.ObjectType{AttrTypes: EnvironmentVariableModel{}.AttributeTypes()})
	if len(envVars) == 0 {
		return nullList, nil
	}
	models := make([]EnvironmentVariableModel, 0, len(envVars))
	for _, ev := range envVars {
		if ev == nil {
			continue
		}
		configObj := types.ObjectNull(EnvVarConfigModel{}.AttributeTypes())
		if ev.Config != nil {
			cfgModel := EnvVarConfigModel{
				VarName:   flatteners.String(ev.Config.VarName),
				SecretId:  flatteners.StringPtr(ev.Config.SecretId),
				TextValue: flatteners.StringPtr(ev.Config.TextValue),
			}
			obj, diags := types.ObjectValueFrom(ctx, EnvVarConfigModel{}.AttributeTypes(), cfgModel)
			if diags.HasError() {
				return nullList, diags
			}
			configObj = obj
		}
		models = append(models, EnvironmentVariableModel{
			Config: configObj,
			Kind:   flatteners.String(string(ev.Kind)),
		})
	}
	if len(models) == 0 {
		return nullList, nil
	}
	list, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: EnvironmentVariableModel{}.AttributeTypes()}, models)
	if diags.HasError() {
		return nullList, diags
	}
	return list, nil
}

func expandMountPoints(ctx context.Context, list types.List) ([]sgsdkgo.MountPoint, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var models []MountPointModel
	if diags := list.ElementsAs(ctx, &models, false); diags.HasError() {
		return nil, diags
	}
	result := make([]sgsdkgo.MountPoint, len(models))
	for i, m := range models {
		result[i] = sgsdkgo.MountPoint{
			Source:   m.Source.ValueString(),
			Target:   m.Target.ValueString(),
			ReadOnly: m.ReadOnly.ValueBoolPointer(),
		}
	}
	return result, nil
}

func flattenMountPoints(ctx context.Context, mps []sgsdkgo.MountPoint) (types.List, diag.Diagnostics) {
	nullList := types.ListNull(types.ObjectType{AttrTypes: MountPointModel{}.AttributeTypes()})
	if len(mps) == 0 {
		return nullList, nil
	}
	models := make([]MountPointModel, len(mps))
	for i, mp := range mps {
		models[i] = MountPointModel{
			Source:   flatteners.String(mp.Source),
			Target:   flatteners.String(mp.Target),
			ReadOnly: flatteners.BoolPtr(mp.ReadOnly),
		}
	}
	list, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: MountPointModel{}.AttributeTypes()}, models)
	if diags.HasError() {
		return nullList, diags
	}
	return list, nil
}

func expandWfStepsConfig(ctx context.Context, list types.List) ([]*sgsdkgo.WfStepsConfig, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var models []WfStepsConfigModel
	if diags := list.ElementsAs(ctx, &models, false); diags.HasError() {
		return nil, diags
	}
	result := make([]*sgsdkgo.WfStepsConfig, len(models))
	for i, m := range models {
		step := &sgsdkgo.WfStepsConfig{
			Name:             m.Name.ValueStringPointer(),
			Approval:         m.Approval.ValueBoolPointer(),
			Timeout:          expanders.IntPtr(m.Timeout.ValueInt64Pointer()),
			WfStepTemplateId: m.WfStepTemplateId.ValueStringPointer(),
			CmdOverride:      m.CmdOverride.ValueStringPointer(),
		}
		if !m.EnvironmentVariables.IsNull() && !m.EnvironmentVariables.IsUnknown() {
			envVars, diags := expandEnvironmentVariables(ctx, m.EnvironmentVariables)
			if diags.HasError() {
				return nil, diags
			}
			step.EnvironmentVariables = envVarSlice(envVars)
		}
		if !m.MountPoints.IsNull() && !m.MountPoints.IsUnknown() {
			mps, diags := expandMountPoints(ctx, m.MountPoints)
			if diags.HasError() {
				return nil, diags
			}
			step.MountPoints = mps
		}
		if !m.WfStepInputData.IsNull() && !m.WfStepInputData.IsUnknown() {
			var idm WfStepInputDataModel
			if diags := m.WfStepInputData.As(ctx, &idm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); diags.HasError() {
				return nil, diags
			}
			schemaType, err := sgsdkgo.NewWfStepInputDataSchemaTypeEnumFromString(idm.SchemaType.ValueString())
			if err != nil {
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Invalid schema type", "Value '"+idm.SchemaType.ValueString()+"' is not a valid schema type")}
			}
			step.WfStepInputData = &sgsdkgo.WfStepInputData{
				SchemaType: &schemaType,
				Data:       expanders.JSONStringToMap(idm.Data.ValueString()),
			}
		}
		result[i] = step
	}
	return result, nil
}

// envVarSlice converts a []*sgsdkgo.EnvVars (as produced by expandEnvironmentVariables)
// into the value slice []sgsdkgo.EnvVars that WfStepsConfig expects.
func envVarSlice(ptrs []*sgsdkgo.EnvVars) []sgsdkgo.EnvVars {
	if ptrs == nil {
		return nil
	}
	result := make([]sgsdkgo.EnvVars, 0, len(ptrs))
	for _, p := range ptrs {
		if p != nil {
			result = append(result, *p)
		}
	}
	return result
}

func flattenWfStepsConfig(ctx context.Context, steps []*sgsdkgo.WfStepsConfig) (types.List, diag.Diagnostics) {
	nullList := types.ListNull(types.ObjectType{AttrTypes: WfStepsConfigModel{}.AttributeTypes()})
	if len(steps) == 0 {
		return nullList, nil
	}
	models := make([]WfStepsConfigModel, 0, len(steps))
	for _, s := range steps {
		if s == nil {
			continue
		}
		envList, diags := flattenEnvironmentVariables(ctx, envVarPointerSlice(s.EnvironmentVariables))
		if diags.HasError() {
			return nullList, diags
		}
		mpList, diags := flattenMountPoints(ctx, s.MountPoints)
		if diags.HasError() {
			return nullList, diags
		}
		inputDataObj := types.ObjectNull(WfStepInputDataModel{}.AttributeTypes())
		if s.WfStepInputData != nil {
			idm := WfStepInputDataModel{
				SchemaType: flatteners.String(string(*s.WfStepInputData.SchemaType)),
				Data:       flatteners.JSONInterfaceToString(s.WfStepInputData.Data),
			}
			obj, diags := types.ObjectValueFrom(ctx, WfStepInputDataModel{}.AttributeTypes(), idm)
			if diags.HasError() {
				return nullList, diags
			}
			inputDataObj = obj
		}
		models = append(models, WfStepsConfigModel{
			Name:                 flatteners.StringPtr(s.Name),
			EnvironmentVariables: envList,
			Approval:             flatteners.BoolPtr(s.Approval),
			Timeout:              flatteners.Int64Ptr(s.Timeout),
			CmdOverride:          flatteners.StringPtr(s.CmdOverride),
			MountPoints:          mpList,
			WfStepTemplateId:     flatteners.StringPtr(s.WfStepTemplateId),
			WfStepInputData:      inputDataObj,
		})
	}
	if len(models) == 0 {
		return nullList, nil
	}
	list, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: WfStepsConfigModel{}.AttributeTypes()}, models)
	if diags.HasError() {
		return nullList, diags
	}
	return list, nil
}

func envVarPointerSlice(vals []sgsdkgo.EnvVars) []*sgsdkgo.EnvVars {
	if vals == nil {
		return nil
	}
	result := make([]*sgsdkgo.EnvVars, len(vals))
	for i := range vals {
		result[i] = &vals[i]
	}
	return result
}

func expandTerraformConfig(ctx context.Context, obj types.Object) (*sgsdkgo.TerraformConfig, diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil
	}
	var m TerraformConfigModel
	if diags := obj.As(ctx, &m, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); diags.HasError() {
		return nil, diags
	}
	tc := &sgsdkgo.TerraformConfig{
		TerraformVersion:       m.TerraformVersion.ValueStringPointer(),
		DriftCheck:             m.DriftCheck.ValueBoolPointer(),
		DriftCron:              m.DriftCron.ValueStringPointer(),
		ManagedTerraformState:  m.ManagedTerraformState.ValueBoolPointer(),
		ApprovalPreApply:       m.ApprovalPreApply.ValueBoolPointer(),
		TerraformPlanOptions:   m.TerraformPlanOptions.ValueStringPointer(),
		TerraformInitOptions:   m.TerraformInitOptions.ValueStringPointer(),
		Timeout:                expanders.IntPtr(m.Timeout.ValueInt64Pointer()),
		RunPreInitHooksOnDrift: m.RunPreInitHooksOnDrift.ValueBoolPointer(),
	}
	if !m.TerraformBinPath.IsNull() && !m.TerraformBinPath.IsUnknown() {
		mps, diags := expandMountPoints(ctx, m.TerraformBinPath)
		if diags.HasError() {
			return nil, diags
		}
		tc.TerraformBinPath = mps
	}
	for _, pair := range []struct {
		list *types.List
		dest *[]sgsdkgo.WfStepsConfig
	}{
		{&m.PostApplyWfStepsConfig, &tc.PostApplyWfStepsConfig},
		{&m.PreApplyWfStepsConfig, &tc.PreApplyWfStepsConfig},
		{&m.PrePlanWfStepsConfig, &tc.PrePlanWfStepsConfig},
		{&m.PostPlanWfStepsConfig, &tc.PostPlanWfStepsConfig},
	} {
		if !pair.list.IsNull() && !pair.list.IsUnknown() {
			steps, diags := expandWfStepsConfig(ctx, *pair.list)
			if diags.HasError() {
				return nil, diags
			}
			*pair.dest = wfStepValueSlice(steps)
		}
	}
	for _, pair := range []struct {
		list *types.List
		dest *[]string
	}{
		{&m.PreInitHooks, &tc.PreInitHooks},
		{&m.PrePlanHooks, &tc.PrePlanHooks},
		{&m.PostPlanHooks, &tc.PostPlanHooks},
		{&m.PreApplyHooks, &tc.PreApplyHooks},
		{&m.PostApplyHooks, &tc.PostApplyHooks},
	} {
		if !pair.list.IsNull() && !pair.list.IsUnknown() {
			hooks, diags := expanders.StringList(ctx, *pair.list)
			if diags.HasError() {
				return nil, diags
			}
			*pair.dest = hooks
		}
	}
	return tc, nil
}

func wfStepValueSlice(ptrs []*sgsdkgo.WfStepsConfig) []sgsdkgo.WfStepsConfig {
	if ptrs == nil {
		return nil
	}
	result := make([]sgsdkgo.WfStepsConfig, 0, len(ptrs))
	for _, p := range ptrs {
		if p != nil {
			result = append(result, *p)
		}
	}
	return result
}

func flattenTerraformConfig(ctx context.Context, tc *sgsdkgo.TerraformConfig) (types.Object, diag.Diagnostics) {
	nullObj := types.ObjectNull(TerraformConfigModel{}.AttributeTypes())
	if tc == nil {
		return nullObj, nil
	}
	binPath, diags := flattenMountPoints(ctx, tc.TerraformBinPath)
	if diags.HasError() {
		return nullObj, diags
	}
	postApply, diags := flattenWfStepsConfig(ctx, wfStepPointerSlice(tc.PostApplyWfStepsConfig))
	if diags.HasError() {
		return nullObj, diags
	}
	preApply, diags := flattenWfStepsConfig(ctx, wfStepPointerSlice(tc.PreApplyWfStepsConfig))
	if diags.HasError() {
		return nullObj, diags
	}
	prePlan, diags := flattenWfStepsConfig(ctx, wfStepPointerSlice(tc.PrePlanWfStepsConfig))
	if diags.HasError() {
		return nullObj, diags
	}
	postPlan, diags := flattenWfStepsConfig(ctx, wfStepPointerSlice(tc.PostPlanWfStepsConfig))
	if diags.HasError() {
		return nullObj, diags
	}

	strListNull := types.ListNull(types.StringType)
	makeStringList := func(hooks []string) (types.List, diag.Diagnostics) {
		if len(hooks) == 0 {
			return strListNull, nil
		}
		return types.ListValueFrom(ctx, types.StringType, hooks)
	}
	preInit, diags := makeStringList(tc.PreInitHooks)
	if diags.HasError() {
		return nullObj, diags
	}
	prePlanHooks, diags := makeStringList(tc.PrePlanHooks)
	if diags.HasError() {
		return nullObj, diags
	}
	postPlanHooks, diags := makeStringList(tc.PostPlanHooks)
	if diags.HasError() {
		return nullObj, diags
	}
	preApplyHooks, diags := makeStringList(tc.PreApplyHooks)
	if diags.HasError() {
		return nullObj, diags
	}
	postApplyHooks, diags := makeStringList(tc.PostApplyHooks)
	if diags.HasError() {
		return nullObj, diags
	}

	m := TerraformConfigModel{
		TerraformVersion:       flatteners.StringPtr(tc.TerraformVersion),
		DriftCheck:             flatteners.BoolPtr(tc.DriftCheck),
		DriftCron:              flatteners.StringPtr(tc.DriftCron),
		ManagedTerraformState:  flatteners.BoolPtr(tc.ManagedTerraformState),
		ApprovalPreApply:       flatteners.BoolPtr(tc.ApprovalPreApply),
		TerraformPlanOptions:   flatteners.StringPtr(tc.TerraformPlanOptions),
		TerraformInitOptions:   flatteners.StringPtr(tc.TerraformInitOptions),
		TerraformBinPath:       binPath,
		Timeout:                flatteners.Int64Ptr(tc.Timeout),
		PostApplyWfStepsConfig: postApply,
		PreApplyWfStepsConfig:  preApply,
		PrePlanWfStepsConfig:   prePlan,
		PostPlanWfStepsConfig:  postPlan,
		PreInitHooks:           preInit,
		PrePlanHooks:           prePlanHooks,
		PostPlanHooks:          postPlanHooks,
		PreApplyHooks:          preApplyHooks,
		PostApplyHooks:         postApplyHooks,
		RunPreInitHooksOnDrift: flatteners.BoolPtr(tc.RunPreInitHooksOnDrift),
	}
	obj, diags := types.ObjectValueFrom(ctx, TerraformConfigModel{}.AttributeTypes(), m)
	if diags.HasError() {
		return nullObj, diags
	}
	return obj, nil
}

func wfStepPointerSlice(vals []sgsdkgo.WfStepsConfig) []*sgsdkgo.WfStepsConfig {
	if vals == nil {
		return nil
	}
	result := make([]*sgsdkgo.WfStepsConfig, len(vals))
	for i := range vals {
		result[i] = &vals[i]
	}
	return result
}

func expandDeploymentPlatformConfig(ctx context.Context, list types.List) ([]*sgsdkgo.DeploymentPlatformConfig, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var models []DeploymentPlatformConfigModel
	if diags := list.ElementsAs(ctx, &models, false); diags.HasError() {
		return nil, diags
	}
	result := make([]*sgsdkgo.DeploymentPlatformConfig, len(models))
	for i, m := range models {
		kind, err := sgsdkgo.NewDeploymentPlatformConfigKindEnumFromString(m.Kind.ValueString())
		if err != nil {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Invalid deployment platform config kind", "Value '"+m.Kind.ValueString()+"' is not a valid kind")}
		}
		dpc := &sgsdkgo.DeploymentPlatformConfig{Kind: &kind}
		if !m.Config.IsNull() && !m.Config.IsUnknown() {
			var cfgModel DeploymentPlatformConfigConfigModel
			if diags := m.Config.As(ctx, &cfgModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); diags.HasError() {
				return nil, diags
			}
			dpc.Config = &sgsdkgo.DeploymentPlatformConfigConfig{
				IntegrationId: cfgModel.IntegrationId.ValueStringPointer(),
			}
			// profile_name is Optional+Computed: ValueStringPointer() returns &"" for
			// unknown, which would clear it on create/update whenever the user hasn't set
			// it. Only set when known.
			if !cfgModel.ProfileName.IsNull() && !cfgModel.ProfileName.IsUnknown() {
				dpc.Config.ProfileName = cfgModel.ProfileName.ValueStringPointer()
			}
		}
		result[i] = dpc
	}
	return result, nil
}

func flattenDeploymentPlatformConfig(ctx context.Context, dpcs []*sgsdkgo.DeploymentPlatformConfig) (types.List, diag.Diagnostics) {
	listNull := types.ListNull(types.ObjectType{AttrTypes: DeploymentPlatformConfigModel{}.AttributeTypes()})
	if len(dpcs) == 0 {
		return listNull, nil
	}
	models := make([]DeploymentPlatformConfigModel, 0, len(dpcs))
	for _, dpc := range dpcs {
		if dpc == nil {
			continue
		}
		configObj := types.ObjectNull(DeploymentPlatformConfigConfigModel{}.AttributeTypes())
		if dpc.Config != nil {
			cfgModel := DeploymentPlatformConfigConfigModel{
				IntegrationId: flatteners.StringPtr(dpc.Config.IntegrationId),
				ProfileName:   flatteners.StringPtr(dpc.Config.ProfileName),
			}
			obj, diags := types.ObjectValueFrom(ctx, DeploymentPlatformConfigConfigModel{}.AttributeTypes(), cfgModel)
			if diags.HasError() {
				return listNull, diags
			}
			configObj = obj
		}
		models = append(models, DeploymentPlatformConfigModel{
			Kind:   flatteners.StringPtr((*string)(dpc.Kind)),
			Config: configObj,
		})
	}
	if len(models) == 0 {
		return listNull, nil
	}
	list, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: DeploymentPlatformConfigModel{}.AttributeTypes()}, models)
	if diags.HasError() {
		return listNull, diags
	}
	return list, nil
}

func expandRunnerConstraints(ctx context.Context, obj types.Object) (*sgsdkgo.RunnerConstraints, diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil
	}
	var m RunnerConstraintsModel
	if diags := obj.As(ctx, &m, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); diags.HasError() {
		return nil, diags
	}
	names, diags := expanders.StringList(ctx, m.Names)
	if diags.HasError() {
		return nil, diags
	}
	return &sgsdkgo.RunnerConstraints{
		Type:  (*sgsdkgo.RunnerConstraintsTypeEnum)(m.Type.ValueStringPointer()),
		Names: names,
	}, nil
}

func flattenRunnerConstraints(ctx context.Context, rc *sgsdkgo.RunnerConstraints) (types.Object, diag.Diagnostics) {
	nullObj := types.ObjectNull(RunnerConstraintsModel{}.AttributeTypes())
	if rc == nil {
		return nullObj, nil
	}
	names, diags := types.ListValueFrom(ctx, types.StringType, rc.Names)
	if diags.HasError() {
		return nullObj, diags
	}
	m := RunnerConstraintsModel{
		Type:  flatteners.StringPtr((*string)(rc.Type)),
		Names: names,
	}
	obj, diags := types.ObjectValueFrom(ctx, RunnerConstraintsModel{}.AttributeTypes(), m)
	if diags.HasError() {
		return nullObj, diags
	}
	return obj, nil
}

func expandVcsConfig(ctx context.Context, obj types.Object) (*sgsdkgo.VcsConfig, diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil
	}
	var m VcsConfigModel
	if diags := obj.As(ctx, &m, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); diags.HasError() {
		return nil, diags
	}
	vcsConfig := &sgsdkgo.VcsConfig{}

	if !m.IacVcsConfig.IsNull() && !m.IacVcsConfig.IsUnknown() {
		var iacModel IacVcsConfigModel
		if diags := m.IacVcsConfig.As(ctx, &iacModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); diags.HasError() {
			return nil, diags
		}
		vcsConfig.IacVcsConfig = &sgsdkgo.IacvcsConfig{
			UseMarketplaceTemplate: iacModel.UseMarketplaceTemplate.ValueBoolPointer(),
			IacTemplateId:          iacModel.IacTemplateId.ValueStringPointer(),
		}
	}

	if !m.IacInputData.IsNull() && !m.IacInputData.IsUnknown() {
		var idModel VcsIacInputDataModel
		if diags := m.IacInputData.As(ctx, &idModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); diags.HasError() {
			return nil, diags
		}
		schemaType, err := sgsdkgo.NewIacInputDataSchemaTypeEnumFromString(idModel.SchemaType.ValueString())
		if err != nil {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Invalid schema type", "Value '"+idModel.SchemaType.ValueString()+"' is not a valid schema type")}
		}
		vcsConfig.IacInputData = &sgsdkgo.IacInputData{
			SchemaId:   idModel.SchemaId.ValueStringPointer(),
			SchemaType: &schemaType,
		}
		if dataMap := expanders.JSONStringToMap(idModel.Data.ValueString()); dataMap != nil {
			vcsConfig.IacInputData.Data = &dataMap
		}
	}

	return vcsConfig, nil
}

func flattenVcsConfig(ctx context.Context, vc *sgsdkgo.VcsConfig) (types.Object, diag.Diagnostics) {
	nullObj := types.ObjectNull(VcsConfigModel{}.AttributeTypes())
	if vc == nil {
		return nullObj, nil
	}
	m := VcsConfigModel{}

	iacVcsObj := types.ObjectNull(IacVcsConfigModel{}.AttributeTypes())
	if vc.IacVcsConfig != nil {
		iacM := IacVcsConfigModel{
			UseMarketplaceTemplate: flatteners.BoolPtr(vc.IacVcsConfig.UseMarketplaceTemplate),
			IacTemplateId:          flatteners.StringPtr(vc.IacVcsConfig.IacTemplateId),
		}
		obj, diags := types.ObjectValueFrom(ctx, IacVcsConfigModel{}.AttributeTypes(), iacM)
		if diags.HasError() {
			return nullObj, diags
		}
		iacVcsObj = obj
	}
	m.IacVcsConfig = iacVcsObj

	iacInputObj := types.ObjectNull(VcsIacInputDataModel{}.AttributeTypes())
	if vc.IacInputData != nil {
		dataStr := types.StringNull()
		if vc.IacInputData.Data != nil {
			dataStr = flatteners.JSONInterfaceToString(*vc.IacInputData.Data)
		}
		idM := VcsIacInputDataModel{
			SchemaId:   flatteners.StringPtr(vc.IacInputData.SchemaId),
			SchemaType: flatteners.String(string(*vc.IacInputData.SchemaType)),
			Data:       dataStr,
		}
		obj, diags := types.ObjectValueFrom(ctx, VcsIacInputDataModel{}.AttributeTypes(), idM)
		if diags.HasError() {
			return nullObj, diags
		}
		iacInputObj = obj
	}
	m.IacInputData = iacInputObj

	obj, diags := types.ObjectValueFrom(ctx, VcsConfigModel{}.AttributeTypes(), m)
	if diags.HasError() {
		return nullObj, diags
	}
	return obj, nil
}

func expandInputSchemas(ctx context.Context, list types.List) ([]*sgsdkgo.InputSchemas, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var models []StackInputSchemaModel
	if diags := list.ElementsAs(ctx, &models, false); diags.HasError() {
		return nil, diags
	}
	result := make([]*sgsdkgo.InputSchemas, len(models))
	for i, ism := range models {
		result[i] = &sgsdkgo.InputSchemas{
			Name:         ism.Name.ValueStringPointer(),
			Type:         sgsdkgo.InputSchemasTypeEnum(ism.Type.ValueString()),
			EncodedData:  ism.EncodedData.ValueStringPointer(),
			UiSchemaData: ism.UiSchemaData.ValueStringPointer(),
		}
	}
	return result, nil
}

func flattenInputSchemas(ctx context.Context, items []*sgsdkgo.InputSchemas) (types.List, diag.Diagnostics) {
	nullList := types.ListNull(types.ObjectType{AttrTypes: StackInputSchemaModel{}.AttributeTypes()})
	if len(items) == 0 {
		return nullList, nil
	}
	models := make([]StackInputSchemaModel, 0, len(items))
	for _, is := range items {
		if is == nil {
			continue
		}
		models = append(models, StackInputSchemaModel{
			Name:         flatteners.StringPtr(is.Name),
			Type:         flatteners.String(string(is.Type)),
			EncodedData:  flatteners.StringPtr(is.EncodedData),
			UiSchemaData: flatteners.StringPtr(is.UiSchemaData),
		})
	}
	if len(models) == 0 {
		return nullList, nil
	}
	list, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: StackInputSchemaModel{}.AttributeTypes()}, models)
	if diags.HasError() {
		return nullList, diags
	}
	return list, nil
}

// expandWfUserSchedules converts a per-workflow user_schedules list to
// []*sgsdkgo.UserSchedules (no Inputs field, unlike the stack-level schedules).
func expandWfUserSchedules(ctx context.Context, list types.List) ([]*sgsdkgo.UserSchedules, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var models []WfUserSchedulesModel
	if diags := list.ElementsAs(ctx, &models, false); diags.HasError() {
		return nil, diags
	}
	result := make([]*sgsdkgo.UserSchedules, len(models))
	for i, m := range models {
		state := sgsdkgo.StateEnum(m.State.ValueString())
		result[i] = &sgsdkgo.UserSchedules{
			Name:  m.Name.ValueStringPointer(),
			Desc:  m.Desc.ValueStringPointer(),
			Cron:  m.Cron.ValueStringPointer(),
			State: &state,
		}
	}
	return result, nil
}

func flattenWfUserSchedules(ctx context.Context, us []*sgsdkgo.UserSchedules) (types.List, diag.Diagnostics) {
	nullList := types.ListNull(types.ObjectType{AttrTypes: WfUserSchedulesModel{}.AttributeTypes()})
	if len(us) == 0 {
		return nullList, nil
	}
	models := make([]WfUserSchedulesModel, 0, len(us))
	for _, s := range us {
		if s == nil {
			continue
		}
		state := ""
		if s.State != nil {
			state = string(*s.State)
		}
		models = append(models, WfUserSchedulesModel{
			Name:  flatteners.StringPtr(s.Name),
			Desc:  flatteners.StringPtr(s.Desc),
			Cron:  flatteners.StringPtr(s.Cron),
			State: flatteners.String(state),
		})
	}
	if len(models) == 0 {
		return nullList, nil
	}
	list, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: WfUserSchedulesModel{}.AttributeTypes()}, models)
	if diags.HasError() {
		return nullList, diags
	}
	return list, nil
}

// expandUserSchedules converts the stack-level user_schedules list to
// []*sgsdkgo.StackUserSchedules, including the Inputs (StackAction) payload.
func expandUserSchedules(ctx context.Context, list types.List) ([]*sgsdkgo.StackUserSchedules, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var models []UserSchedulesModel
	if diags := list.ElementsAs(ctx, &models, false); diags.HasError() {
		return nil, diags
	}
	result := make([]*sgsdkgo.StackUserSchedules, len(models))
	for i, m := range models {
		var inputsModel StackActionInputsModel
		if diags := m.Inputs.As(ctx, &inputsModel, basetypes.ObjectAsOptions{}); diags.HasError() {
			return nil, diags
		}
		// name is Computed-only — server-assigned, never sent.
		result[i] = &sgsdkgo.StackUserSchedules{
			Desc:   m.Desc.ValueStringPointer(),
			Cron:   m.Cron.ValueString(),
			State:  sgsdkgo.StateEnum(m.State.ValueString()),
			Inputs: &sgsdkgo.StackAction{ActionType: inputsModel.ActionType.ValueString()},
		}
	}
	return result, nil
}

func flattenUserSchedules(ctx context.Context, us []*sgsdkgo.StackUserSchedules) (types.List, diag.Diagnostics) {
	nullList := types.ListNull(types.ObjectType{AttrTypes: UserSchedulesModel{}.AttributeTypes()})
	if len(us) == 0 {
		return nullList, nil
	}
	models := make([]UserSchedulesModel, 0, len(us))
	for _, s := range us {
		if s == nil {
			continue
		}
		inputsObj := types.ObjectNull(StackActionInputsModel{}.AttributeTypes())
		if s.Inputs != nil {
			obj, diags := types.ObjectValueFrom(ctx, StackActionInputsModel{}.AttributeTypes(), StackActionInputsModel{
				ActionType: flatteners.String(s.Inputs.ActionType),
			})
			if diags.HasError() {
				return nullList, diags
			}
			inputsObj = obj
		}
		models = append(models, UserSchedulesModel{
			Name:   flatteners.StringPtr(s.Name),
			Desc:   flatteners.StringPtr(s.Desc),
			Cron:   flatteners.String(s.Cron),
			State:  flatteners.String(string(s.State)),
			Inputs: inputsObj,
		})
	}
	if len(models) == 0 {
		return nullList, nil
	}
	list, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: UserSchedulesModel{}.AttributeTypes()}, models)
	if diags.HasError() {
		return nullList, diags
	}
	return list, nil
}

func expandContextTags(ctx context.Context, ct types.Map) (map[string]*string, diag.Diagnostics) {
	if ct.IsNull() || ct.IsUnknown() {
		return nil, nil
	}
	var strs map[string]string
	if diags := ct.ElementsAs(ctx, &strs, false); diags.HasError() {
		return nil, diags
	}
	result := make(map[string]*string, len(strs))
	for k, v := range strs {
		val := v
		result[k] = &val
	}
	return result, nil
}

func flattenContextTags(ctx context.Context, ct map[string]*string) (types.Map, diag.Diagnostics) {
	nullMap := types.MapNull(types.StringType)
	if len(ct) == 0 {
		return nullMap, nil
	}
	elements := make(map[string]attr.Value, len(ct))
	for k, v := range ct {
		elements[k] = flatteners.StringPtr(v)
	}
	m, diags := types.MapValue(types.StringType, elements)
	if diags.HasError() {
		return nullMap, diags
	}
	return m, nil
}

// ---------------------------------------------------------------------------
// MiniSteps converters (Terraform -> SDK), shared by stack-level and
// per-workflow mini_steps.
// ---------------------------------------------------------------------------

func expandNotificationRecipients(ctx context.Context, list types.List) ([]*sgsdkgo.Notifications, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var models []MinistepsNotificationRecipientsModel
	if diags := list.ElementsAs(ctx, &models, false); diags.HasError() {
		return nil, diags
	}
	result := make([]*sgsdkgo.Notifications, len(models))
	for i, m := range models {
		recipients, diags := expanders.StringList(ctx, m.Recipients)
		if diags.HasError() {
			return nil, diags
		}
		result[i] = &sgsdkgo.Notifications{Recipients: recipients}
	}
	return result, nil
}

func expandWebhooks(ctx context.Context, list types.List) ([]*sgsdkgo.Webhook, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var models []MinistepsWebhooksModel
	if diags := list.ElementsAs(ctx, &models, false); diags.HasError() {
		return nil, diags
	}
	result := make([]*sgsdkgo.Webhook, len(models))
	for i, m := range models {
		wh := &sgsdkgo.Webhook{
			WebhookName: m.WebhookName.ValueString(),
			WebhookUrl:  m.WebhookUrl.ValueString(),
		}
		if !m.WebhookSecret.IsNull() && !m.WebhookSecret.IsUnknown() {
			wh.WebhookSecret = m.WebhookSecret.ValueStringPointer()
		}
		result[i] = wh
	}
	return result, nil
}

func expandWfChaining(ctx context.Context, list types.List) ([]*sgsdkgo.MiniSteps, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var models []MinistepsWorkflowChainingModel
	if diags := list.ElementsAs(ctx, &models, false); diags.HasError() {
		return nil, diags
	}
	result := make([]*sgsdkgo.MiniSteps, len(models))
	for i, m := range models {
		ms := &sgsdkgo.MiniSteps{
			WorkflowGroupId: m.WorkflowGroupId.ValueString(),
			WorkflowId:      m.WorkflowId.ValueStringPointer(),
			StackId:         m.StackId.ValueStringPointer(),
		}
		if !m.WorkflowRunPayload.IsNull() && !m.WorkflowRunPayload.IsUnknown() {
			ms.WorkflowRunPayload = expanders.JSONStringToMap(m.WorkflowRunPayload.ValueString())
		}
		if !m.StackRunPayload.IsNull() && !m.StackRunPayload.IsUnknown() {
			ms.StackRunPayload = expanders.JSONStringToMap(m.StackRunPayload.ValueString())
		}
		result[i] = ms
	}
	return result, nil
}

func expandMiniSteps(ctx context.Context, obj types.Object) (*sgsdkgo.MiniStepsSchema, diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil
	}
	var m MinistepsModel
	if diags := obj.As(ctx, &m, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); diags.HasError() {
		return nil, diags
	}
	result := &sgsdkgo.MiniStepsSchema{}

	if !m.Notifications.IsNull() && !m.Notifications.IsUnknown() {
		var notifModel MinistepsNotificationsModel
		if diags := m.Notifications.As(ctx, &notifModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); diags.HasError() {
			return nil, diags
		}
		notif := &sgsdkgo.NotificationTypes{}
		if !notifModel.Email.IsNull() && !notifModel.Email.IsUnknown() {
			var emailModel MinistepsEmailModel
			if diags := notifModel.Email.As(ctx, &emailModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); diags.HasError() {
				return nil, diags
			}
			email := &sgsdkgo.NotificationEmailType{}
			for _, pair := range []struct {
				list *types.List
				dest *[]*sgsdkgo.Notifications
			}{
				{&emailModel.ApprovalRequired, &email.ApprovalRequired},
				{&emailModel.Cancelled, &email.Cancelled},
				{&emailModel.Completed, &email.Completed},
				{&emailModel.DriftDetected, &email.DriftDetected},
				{&emailModel.Errored, &email.Errored},
			} {
				if !pair.list.IsNull() && !pair.list.IsUnknown() {
					vals, diags := expandNotificationRecipients(ctx, *pair.list)
					if diags.HasError() {
						return nil, diags
					}
					*pair.dest = vals
				}
			}
			notif.Email = email
		}
		result.Notifications = notif
	}

	if !m.Webhooks.IsNull() && !m.Webhooks.IsUnknown() {
		var whModel MinistepsWebhooksContainerModel
		if diags := m.Webhooks.As(ctx, &whModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); diags.HasError() {
			return nil, diags
		}
		wh := &sgsdkgo.WebhookTypes{}
		for _, pair := range []struct {
			list *types.List
			dest *[]*sgsdkgo.Webhook
		}{
			{&whModel.ApprovalRequired, &wh.ApprovalRequired},
			{&whModel.Cancelled, &wh.Cancelled},
			{&whModel.Completed, &wh.Completed},
			{&whModel.DriftDetected, &wh.DriftDetected},
			{&whModel.Errored, &wh.Errored},
		} {
			if !pair.list.IsNull() && !pair.list.IsUnknown() {
				vals, diags := expandWebhooks(ctx, *pair.list)
				if diags.HasError() {
					return nil, diags
				}
				*pair.dest = vals
			}
		}
		result.Webhooks = wh
	}

	if !m.WfChaining.IsNull() && !m.WfChaining.IsUnknown() {
		var wcModel MinistepsWfChainingContainerModel
		if diags := m.WfChaining.As(ctx, &wcModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); diags.HasError() {
			return nil, diags
		}
		wc := &sgsdkgo.WfChainingPayloadPayload{}
		for _, pair := range []struct {
			list *types.List
			dest *[]*sgsdkgo.MiniSteps
		}{
			{&wcModel.Completed, &wc.Completed},
			{&wcModel.Errored, &wc.Errored},
		} {
			if !pair.list.IsNull() && !pair.list.IsUnknown() {
				vals, diags := expandWfChaining(ctx, *pair.list)
				if diags.HasError() {
					return nil, diags
				}
				*pair.dest = vals
			}
		}
		result.WfChaining = wc
	}

	return result, nil
}

func flattenNotificationRecipients(ctx context.Context, recipients []*sgsdkgo.Notifications) (types.List, diag.Diagnostics) {
	nullList := types.ListNull(types.ObjectType{AttrTypes: MinistepsNotificationRecipientsModel{}.AttributeTypes()})
	if len(recipients) == 0 {
		return nullList, nil
	}
	models := make([]MinistepsNotificationRecipientsModel, 0, len(recipients))
	for _, r := range recipients {
		if r == nil {
			continue
		}
		recList, diags := types.ListValueFrom(ctx, types.StringType, r.Recipients)
		if diags.HasError() {
			return nullList, diags
		}
		models = append(models, MinistepsNotificationRecipientsModel{Recipients: recList})
	}
	if len(models) == 0 {
		return nullList, nil
	}
	list, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: MinistepsNotificationRecipientsModel{}.AttributeTypes()}, models)
	if diags.HasError() {
		return nullList, diags
	}
	return list, nil
}

func flattenWebhooks(ctx context.Context, webhooks []*sgsdkgo.Webhook) (types.List, diag.Diagnostics) {
	nullList := types.ListNull(types.ObjectType{AttrTypes: MinistepsWebhooksModel{}.AttributeTypes()})
	if len(webhooks) == 0 {
		return nullList, nil
	}
	models := make([]MinistepsWebhooksModel, 0, len(webhooks))
	for _, wh := range webhooks {
		if wh == nil {
			continue
		}
		models = append(models, MinistepsWebhooksModel{
			WebhookName:   flatteners.String(wh.WebhookName),
			WebhookUrl:    flatteners.String(wh.WebhookUrl),
			WebhookSecret: flatteners.StringPtr(wh.WebhookSecret),
		})
	}
	if len(models) == 0 {
		return nullList, nil
	}
	list, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: MinistepsWebhooksModel{}.AttributeTypes()}, models)
	if diags.HasError() {
		return nullList, diags
	}
	return list, nil
}

func flattenWfChaining(ctx context.Context, items []*sgsdkgo.MiniSteps) (types.List, diag.Diagnostics) {
	nullList := types.ListNull(types.ObjectType{AttrTypes: MinistepsWorkflowChainingModel{}.AttributeTypes()})
	if len(items) == 0 {
		return nullList, nil
	}
	models := make([]MinistepsWorkflowChainingModel, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		models = append(models, MinistepsWorkflowChainingModel{
			WorkflowGroupId:    flatteners.String(item.WorkflowGroupId),
			StackId:            flatteners.StringPtr(item.StackId),
			StackRunPayload:    flatteners.JSONInterfaceToString(item.StackRunPayload),
			WorkflowId:         flatteners.StringPtr(item.WorkflowId),
			WorkflowRunPayload: flatteners.JSONInterfaceToString(item.WorkflowRunPayload),
		})
	}
	list, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: MinistepsWorkflowChainingModel{}.AttributeTypes()}, models)
	if diags.HasError() {
		return nullList, diags
	}
	return list, nil
}

func flattenMiniSteps(ctx context.Context, miniSteps *sgsdkgo.MiniStepsSchema) (types.Object, diag.Diagnostics) {
	nullObj := types.ObjectNull(MinistepsModel{}.AttributeTypes())
	if miniSteps == nil {
		return nullObj, nil
	}

	var diags diag.Diagnostics
	msModel := MinistepsModel{}

	if miniSteps.Notifications != nil {
		notifModel := MinistepsNotificationsModel{}
		if miniSteps.Notifications.Email != nil {
			emailModel := MinistepsEmailModel{}
			emailModel.ApprovalRequired, diags = flattenNotificationRecipients(ctx, miniSteps.Notifications.Email.ApprovalRequired)
			if diags.HasError() {
				return nullObj, diags
			}
			emailModel.Cancelled, diags = flattenNotificationRecipients(ctx, miniSteps.Notifications.Email.Cancelled)
			if diags.HasError() {
				return nullObj, diags
			}
			emailModel.Completed, diags = flattenNotificationRecipients(ctx, miniSteps.Notifications.Email.Completed)
			if diags.HasError() {
				return nullObj, diags
			}
			emailModel.DriftDetected, diags = flattenNotificationRecipients(ctx, miniSteps.Notifications.Email.DriftDetected)
			if diags.HasError() {
				return nullObj, diags
			}
			emailModel.Errored, diags = flattenNotificationRecipients(ctx, miniSteps.Notifications.Email.Errored)
			if diags.HasError() {
				return nullObj, diags
			}
			emailObj, d := types.ObjectValueFrom(ctx, MinistepsEmailModel{}.AttributeTypes(), emailModel)
			if d.HasError() {
				return nullObj, d
			}
			notifModel.Email = emailObj
		} else {
			notifModel.Email = types.ObjectNull(MinistepsEmailModel{}.AttributeTypes())
		}
		notifObj, d := types.ObjectValueFrom(ctx, MinistepsNotificationsModel{}.AttributeTypes(), notifModel)
		if d.HasError() {
			return nullObj, d
		}
		msModel.Notifications = notifObj
	} else {
		msModel.Notifications = types.ObjectNull(MinistepsNotificationsModel{}.AttributeTypes())
	}

	if miniSteps.Webhooks != nil {
		whModel := MinistepsWebhooksContainerModel{}
		whModel.ApprovalRequired, diags = flattenWebhooks(ctx, miniSteps.Webhooks.ApprovalRequired)
		if diags.HasError() {
			return nullObj, diags
		}
		whModel.Cancelled, diags = flattenWebhooks(ctx, miniSteps.Webhooks.Cancelled)
		if diags.HasError() {
			return nullObj, diags
		}
		whModel.Completed, diags = flattenWebhooks(ctx, miniSteps.Webhooks.Completed)
		if diags.HasError() {
			return nullObj, diags
		}
		whModel.DriftDetected, diags = flattenWebhooks(ctx, miniSteps.Webhooks.DriftDetected)
		if diags.HasError() {
			return nullObj, diags
		}
		whModel.Errored, diags = flattenWebhooks(ctx, miniSteps.Webhooks.Errored)
		if diags.HasError() {
			return nullObj, diags
		}
		whObj, d := types.ObjectValueFrom(ctx, MinistepsWebhooksContainerModel{}.AttributeTypes(), whModel)
		if d.HasError() {
			return nullObj, d
		}
		msModel.Webhooks = whObj
	} else {
		msModel.Webhooks = types.ObjectNull(MinistepsWebhooksContainerModel{}.AttributeTypes())
	}

	if miniSteps.WfChaining != nil {
		wcModel := MinistepsWfChainingContainerModel{}
		wcModel.Completed, diags = flattenWfChaining(ctx, miniSteps.WfChaining.Completed)
		if diags.HasError() {
			return nullObj, diags
		}
		wcModel.Errored, diags = flattenWfChaining(ctx, miniSteps.WfChaining.Errored)
		if diags.HasError() {
			return nullObj, diags
		}
		wcObj, d := types.ObjectValueFrom(ctx, MinistepsWfChainingContainerModel{}.AttributeTypes(), wcModel)
		if d.HasError() {
			return nullObj, d
		}
		msModel.WfChaining = wcObj
	} else {
		msModel.WfChaining = types.ObjectNull(MinistepsWfChainingContainerModel{}.AttributeTypes())
	}

	obj, diags := types.ObjectValueFrom(ctx, MinistepsModel{}.AttributeTypes(), msModel)
	if diags.HasError() {
		return nullObj, diags
	}
	return obj, nil
}

// ---------------------------------------------------------------------------
// Workflow template merge (provider-side resolution)
//
// A workflow slot in a stack's workflows_config resolves from up to three
// layers, lowest to highest precedence:
//  1. the workflow template revision the slot's stack-template entry points
//     to (vcs_config.iac_vcs_config.iac_template_id)
//  2. the matching workflow entry on the stack template revision (an
//     override layer)
//  3. what the user declared on the stack resource's own workflows_config
//
// Mirrors workflow_from_template's mergeTemplateDefaults/mergeTerraformConfig
// pattern: layer 3 is built first (respecting Optional+Computed guards, so a
// field is nil exactly when the user left it unset), then layers 2 and 1 each
// fill in whatever is STILL nil, in precedence order. TerraformConfig is
// deep-merged field-by-field rather than replaced wholesale, chaining the
// pairwise merge twice to fold in all three layers.
// ---------------------------------------------------------------------------

// mergeTerraformConfig deep-merges a TerraformConfig field-by-field: the higher-
// precedence value (user) is kept when set, otherwise the lower-precedence
// value (tpl) fills it. Returns the user config unchanged if tpl has none, and
// tpl's config if user has none. Chain two calls to fold in a third layer:
// mergeTerraformConfig(mergeTerraformConfig(ownCfg, stackTplCfg), workflowTplCfg).
// drift_cron is cleared whenever the resolved drift_check is false.
func mergeTerraformConfig(user, tpl *sgsdkgo.TerraformConfig) *sgsdkgo.TerraformConfig {
	if tpl == nil {
		return coupleDriftFields(user)
	}
	if user == nil {
		return coupleDriftFields(tpl)
	}

	if user.TerraformVersion == nil {
		user.TerraformVersion = tpl.TerraformVersion
	}
	if user.DriftCheck == nil {
		user.DriftCheck = tpl.DriftCheck
	}
	if user.DriftCron == nil {
		user.DriftCron = tpl.DriftCron
	}
	if user.ManagedTerraformState == nil {
		user.ManagedTerraformState = tpl.ManagedTerraformState
	}
	if user.ApprovalPreApply == nil {
		user.ApprovalPreApply = tpl.ApprovalPreApply
	}
	if user.TerraformPlanOptions == nil {
		user.TerraformPlanOptions = tpl.TerraformPlanOptions
	}
	if user.TerraformInitOptions == nil {
		user.TerraformInitOptions = tpl.TerraformInitOptions
	}
	if user.WfStepTemplateRevisionId == nil {
		user.WfStepTemplateRevisionId = tpl.WfStepTemplateRevisionId
	}
	if user.Timeout == nil {
		user.Timeout = tpl.Timeout
	}
	if user.RunPreInitHooksOnDrift == nil {
		user.RunPreInitHooksOnDrift = tpl.RunPreInitHooksOnDrift
	}
	if user.RunPrePlanHooksOnDrift == nil {
		user.RunPrePlanHooksOnDrift = tpl.RunPrePlanHooksOnDrift
	}
	if user.RunPostPlanHooksOnDrift == nil {
		user.RunPostPlanHooksOnDrift = tpl.RunPostPlanHooksOnDrift
	}
	if len(user.TerraformBinPath) == 0 {
		user.TerraformBinPath = tpl.TerraformBinPath
	}
	if len(user.PostApplyWfStepsConfig) == 0 {
		user.PostApplyWfStepsConfig = tpl.PostApplyWfStepsConfig
	}
	if len(user.PreApplyWfStepsConfig) == 0 {
		user.PreApplyWfStepsConfig = tpl.PreApplyWfStepsConfig
	}
	if len(user.PrePlanWfStepsConfig) == 0 {
		user.PrePlanWfStepsConfig = tpl.PrePlanWfStepsConfig
	}
	if len(user.PostPlanWfStepsConfig) == 0 {
		user.PostPlanWfStepsConfig = tpl.PostPlanWfStepsConfig
	}
	if len(user.PreInitHooks) == 0 {
		user.PreInitHooks = tpl.PreInitHooks
	}
	if len(user.PrePlanHooks) == 0 {
		user.PrePlanHooks = tpl.PrePlanHooks
	}
	if len(user.PostPlanHooks) == 0 {
		user.PostPlanHooks = tpl.PostPlanHooks
	}
	if len(user.PreApplyHooks) == 0 {
		user.PreApplyHooks = tpl.PreApplyHooks
	}
	if len(user.PostApplyHooks) == 0 {
		user.PostApplyHooks = tpl.PostApplyHooks
	}

	return coupleDriftFields(user)
}

// coupleDriftFields enforces the drift_check/drift_cron coupling: a cron is only
// meaningful when drift checking is enabled, so whenever the resolved drift_check
// is absent or false the cron is cleared, regardless of which layer it came from.
func coupleDriftFields(cfg *sgsdkgo.TerraformConfig) *sgsdkgo.TerraformConfig {
	if cfg == nil {
		return nil
	}
	if cfg.DriftCheck == nil || !*cfg.DriftCheck {
		cfg.DriftCron = nil
	}
	return cfg
}

// wfStepsConfigPtrSlice converts a value slice (as carried by both stack
// template revisions and workflow template revisions) to the pointer slice
// StackWorkflowsConfigWorkflow expects.
func wfStepsConfigPtrSlice(vals []sgsdkgo.WfStepsConfig) []*sgsdkgo.WfStepsConfig {
	if len(vals) == 0 {
		return nil
	}
	result := make([]*sgsdkgo.WfStepsConfig, len(vals))
	for i := range vals {
		result[i] = &vals[i]
	}
	return result
}

// envVarsPtrSliceFromValues converts a value slice to the pointer slice
// StackWorkflowsConfigWorkflow expects.
func envVarsPtrSliceFromValues(vals []sgsdkgo.EnvVars) []*sgsdkgo.EnvVars {
	if len(vals) == 0 {
		return nil
	}
	result := make([]*sgsdkgo.EnvVars, len(vals))
	for i := range vals {
		result[i] = &vals[i]
	}
	return result
}

// inputSchemasPtrSliceFromValues converts a value slice to the pointer slice
// StackWorkflowsConfigWorkflow expects.
func inputSchemasPtrSliceFromValues(vals []sgsdkgo.InputSchemas) []*sgsdkgo.InputSchemas {
	if len(vals) == 0 {
		return nil
	}
	result := make([]*sgsdkgo.InputSchemas, len(vals))
	for i := range vals {
		result[i] = &vals[i]
	}
	return result
}

// deploymentPlatformConfigFromWorkflowTemplate adapts the workflow template
// revision's own DeploymentPlatformConfig type (workflowtemplaterevisions
// package) to the root SDK type StackWorkflowsConfigWorkflow expects — these
// are distinct Go types with the same shape, not aliases of each other.
func deploymentPlatformConfigFromWorkflowTemplate(items []*workflowtemplaterevisions.DeploymentPlatformConfig) []*sgsdkgo.DeploymentPlatformConfig {
	if len(items) == 0 {
		return nil
	}
	result := make([]*sgsdkgo.DeploymentPlatformConfig, 0, len(items))
	for _, it := range items {
		if it == nil {
			continue
		}
		kind := sgsdkgo.DeploymentPlatformConfigKindEnum(it.Kind)
		dpc := &sgsdkgo.DeploymentPlatformConfig{
			Kind: &kind,
			Config: &sgsdkgo.DeploymentPlatformConfigConfig{
				ProfileName: it.Config.ProfileName,
			},
		}
		if it.Config.IntegrationId != "" {
			id := it.Config.IntegrationId
			dpc.Config.IntegrationId = &id
		}
		result = append(result, dpc)
	}
	return result
}

// userSchedulesFromWorkflowTemplate adapts the workflow template revision's own
// UserSchedules type (workflowtemplaterevisions package, which also carries a
// per-schedule Inputs payload with no counterpart at this level) to the generic
// sgsdkgo.UserSchedules shape StackWorkflowsConfigWorkflow expects.
func userSchedulesFromWorkflowTemplate(items []workflowtemplaterevisions.UserSchedules) []*sgsdkgo.UserSchedules {
	if len(items) == 0 {
		return nil
	}
	result := make([]*sgsdkgo.UserSchedules, len(items))
	for i := range items {
		it := items[i]
		cron := it.Cron
		state := sgsdkgo.StateEnum(it.State)
		result[i] = &sgsdkgo.UserSchedules{
			Name:  it.Name,
			Desc:  it.Desc,
			Cron:  &cron,
			State: &state,
		}
	}
	return result
}

// mergeWorkflowWithStackTemplateOverride fills wf's still-nil/empty fields from
// stackTplWf, the matching workflow slot on the stack template revision — the
// middle precedence layer. vcs_config.iac_vcs_config is a special case: it is
// unconditionally overwritten (never merely filled), because it is Computed-only
// on the stack resource and the stack template is its ONLY source of truth,
// never something to preserve from a stale prior value.
func mergeWorkflowWithStackTemplateOverride(wf *sgsdkgo.StackWorkflowsConfigWorkflow, stackTplWf *stacktemplaterevisions.StackTemplateRevisionWorkflow) {
	if wf == nil || stackTplWf == nil {
		return
	}

	if wf.ResourceName == nil {
		wf.ResourceName = stackTplWf.ResourceName
	}
	if wf.TemplateId == nil {
		wf.TemplateId = stackTplWf.TemplateId
	}
	if len(wf.WfStepsConfig) == 0 {
		wf.WfStepsConfig = stackTplWf.WfStepsConfig
	}
	wf.TerraformConfig = mergeTerraformConfig(wf.TerraformConfig, stackTplWf.TerraformConfig)
	if len(wf.EnvironmentVariables) == 0 {
		wf.EnvironmentVariables = stackTplWf.EnvironmentVariables
	}
	if len(wf.DeploymentPlatformConfig) == 0 {
		wf.DeploymentPlatformConfig = stackTplWf.DeploymentPlatformConfig
	}
	if len(wf.UserSchedules) == 0 {
		wf.UserSchedules = stackTplWf.UserSchedules
	}
	if wf.MiniSteps == nil {
		wf.MiniSteps = stackTplWf.MiniSteps
	}
	if len(wf.Approvers) == 0 {
		wf.Approvers = stackTplWf.Approvers
	}
	if wf.NumberOfApprovalsRequired == nil {
		wf.NumberOfApprovalsRequired = stackTplWf.NumberOfApprovalsRequired
	}
	if wf.RunnerConstraints == nil {
		wf.RunnerConstraints = stackTplWf.RunnerConstraints
	}
	if wf.UserJobCpu == nil {
		wf.UserJobCpu = stackTplWf.UserJobCpu
	}
	if wf.UserJobMemory == nil {
		wf.UserJobMemory = stackTplWf.UserJobMemory
	}
	if len(wf.InputSchemas) == 0 {
		wf.InputSchemas = stackTplWf.InputSchemas
	}

	// iac_vcs_config is never user-settable on the stack resource (Computed-only) —
	// always take the stack template's value, whatever it is (including clearing it
	// if the stack template doesn't have one either).
	if stackTplWf.VcsConfig != nil {
		if wf.VcsConfig == nil {
			wf.VcsConfig = &sgsdkgo.VcsConfig{}
		}
		wf.VcsConfig.IacVcsConfig = stackTplWf.VcsConfig.IacVcsConfig
	} else if wf.VcsConfig != nil {
		wf.VcsConfig.IacVcsConfig = nil
	}
}

// mergeWorkflowWithWorkflowTemplateDefaults fills wf's still-nil/empty fields
// from workflowTpl — the lowest precedence layer. mini_steps has no fallback
// here: the workflow template revision's Ministeps type is structurally
// distinct from sgsdkgo.MiniStepsSchema (separate type trees for
// notifications/webhooks/wf_chaining), so it is intentionally not bridged;
// mini_steps still resolves fully from the stack's own config and the stack
// template override layer, which share the same type.
func mergeWorkflowWithWorkflowTemplateDefaults(wf *sgsdkgo.StackWorkflowsConfigWorkflow, workflowTpl *workflowtemplaterevisions.ReadWorkflowTemplateRevisionModel) {
	if wf == nil || workflowTpl == nil {
		return
	}

	if wf.Description == nil {
		wf.Description = workflowTpl.LongDescription
	}
	if wf.NumberOfApprovalsRequired == nil {
		wf.NumberOfApprovalsRequired = workflowTpl.NumberOfApprovalsRequired
	}
	if wf.UserJobCpu == nil {
		wf.UserJobCpu = workflowTpl.UserJobCPU
	}
	if wf.UserJobMemory == nil {
		wf.UserJobMemory = workflowTpl.UserJobMemory
	}
	wf.TerraformConfig = mergeTerraformConfig(wf.TerraformConfig, workflowTpl.TerraformConfig)
	if wf.RunnerConstraints == nil {
		wf.RunnerConstraints = workflowTpl.RunnerConstraints
	}
	if len(wf.Tags) == 0 {
		wf.Tags = workflowTpl.Tags
	}
	if len(wf.Approvers) == 0 {
		wf.Approvers = workflowTpl.Approvers
	}
	if len(wf.ContextTags) == 0 && len(workflowTpl.ContextTags) > 0 {
		wf.ContextTags = contextTagsFromTemplate(workflowTpl.ContextTags)
	}
	if wf.IsActive == nil {
		wf.IsActive = workflowTpl.IsActive
	}
	if len(wf.WfStepsConfig) == 0 {
		wf.WfStepsConfig = wfStepsConfigPtrSlice(workflowTpl.WfStepsConfig)
	}
	if len(wf.EnvironmentVariables) == 0 {
		wf.EnvironmentVariables = envVarsPtrSliceFromValues(workflowTpl.EnvironmentVariables)
	}
	if len(wf.InputSchemas) == 0 {
		wf.InputSchemas = inputSchemasPtrSliceFromValues(workflowTpl.InputSchemas)
	}
	if len(wf.DeploymentPlatformConfig) == 0 {
		wf.DeploymentPlatformConfig = deploymentPlatformConfigFromWorkflowTemplate(workflowTpl.DeploymentPlatformConfig)
	}
	if len(wf.UserSchedules) == 0 {
		wf.UserSchedules = userSchedulesFromWorkflowTemplate(workflowTpl.UserSchedules)
	}
}

// ---------------------------------------------------------------------------
// WorkflowsConfig converters
// ---------------------------------------------------------------------------

func expandWorkflowsConfig(ctx context.Context, wfc types.Object, stackTpl *stacktemplaterevisions.ReadStackTemplateRevisionModel, workflowTemplates map[string]*workflowtemplaterevisions.ReadWorkflowTemplateRevisionModel) (*sgsdkgo.StackWorkflowsConfig, diag.Diagnostics) {
	if wfc.IsNull() || wfc.IsUnknown() {
		return nil, nil
	}
	var m WorkflowsConfigModel
	if diags := wfc.As(ctx, &m, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); diags.HasError() {
		return nil, diags
	}
	if m.Workflows.IsNull() || m.Workflows.IsUnknown() {
		return &sgsdkgo.StackWorkflowsConfig{}, nil
	}
	var wfModels []WorkflowInStackModel
	if diags := m.Workflows.ElementsAs(ctx, &wfModels, false); diags.HasError() {
		return nil, diags
	}

	// stackTplWorkflowsById indexes the stack template revision's own workflow
	// entries by slot id, for matching against this stack's own entries below.
	stackTplWorkflowsById := make(map[string]*stacktemplaterevisions.StackTemplateRevisionWorkflow)
	if stackTpl != nil && stackTpl.WorkflowsConfig != nil {
		for _, w := range stackTpl.WorkflowsConfig.Workflows {
			if w != nil && w.Id != nil {
				stackTplWorkflowsById[*w.Id] = w
			}
		}
	}

	workflows := make([]*sgsdkgo.StackWorkflowsConfigWorkflow, len(wfModels))
	for i, wm := range wfModels {
		wf := &sgsdkgo.StackWorkflowsConfigWorkflow{
			// id is Required, always known here — it's also the key used to match this
			// entry against the stack template revision's workflows_config below.
			Id: wm.Id.ValueStringPointer(),
		}
		// resource_name/description/template_id/number_of_approvals_required/
		// user_job_cpu/user_job_memory are all Optional+Computed: ValueStringPointer()/
		// ValueInt64Pointer() return &""/&0 for unknown, which would send spurious empty/zero
		// values on create whenever the user hasn't set them. Only set when known.
		if !wm.ResourceName.IsNull() && !wm.ResourceName.IsUnknown() {
			wf.ResourceName = wm.ResourceName.ValueStringPointer()
		}
		if !wm.Description.IsNull() && !wm.Description.IsUnknown() {
			wf.Description = wm.Description.ValueStringPointer()
		}
		if !wm.TemplateId.IsNull() && !wm.TemplateId.IsUnknown() {
			wf.TemplateId = wm.TemplateId.ValueStringPointer()
		}
		if !wm.NumberOfApprovalsRequired.IsNull() && !wm.NumberOfApprovalsRequired.IsUnknown() {
			wf.NumberOfApprovalsRequired = expanders.IntPtr(wm.NumberOfApprovalsRequired.ValueInt64Pointer())
		}
		if !wm.UserJobCpu.IsNull() && !wm.UserJobCpu.IsUnknown() {
			wf.UserJobCpu = expanders.IntPtr(wm.UserJobCpu.ValueInt64Pointer())
		}
		if !wm.UserJobMemory.IsNull() && !wm.UserJobMemory.IsUnknown() {
			wf.UserJobMemory = expanders.IntPtr(wm.UserJobMemory.ValueInt64Pointer())
		}

		if !wm.Tags.IsNull() && !wm.Tags.IsUnknown() {
			tags, diags := expanders.StringList(ctx, wm.Tags)
			if diags.HasError() {
				return nil, diags
			}
			wf.Tags = tags
		}
		if !wm.WfType.IsNull() && !wm.WfType.IsUnknown() {
			wfType, err := sgsdkgo.NewWfTypeEnumFromString(wm.WfType.ValueString())
			if err != nil {
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Invalid wf_type", "Value '"+wm.WfType.ValueString()+"' is not a valid workflow type")}
			}
			wf.WfType = &wfType
		}
		if !wm.ParallelExecution.IsNull() && !wm.ParallelExecution.IsUnknown() {
			pe, err := sgsdkgo.NewParallelExecutionEnumFromString(wm.ParallelExecution.ValueString())
			if err != nil {
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Invalid parallel_execution", "Value '"+wm.ParallelExecution.ValueString()+"' is not a valid value")}
			}
			wf.ParallelExecution = &pe
		}
		if !wm.IsActive.IsNull() && !wm.IsActive.IsUnknown() {
			isActive, err := sgsdkgo.NewIsPublicEnumFromString(wm.IsActive.ValueString())
			if err != nil {
				return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Invalid is_active", "Value '"+wm.IsActive.ValueString()+"' is not a valid value")}
			}
			wf.IsActive = &isActive
		}
		if !wm.WfStepsConfig.IsNull() && !wm.WfStepsConfig.IsUnknown() {
			steps, diags := expandWfStepsConfig(ctx, wm.WfStepsConfig)
			if diags.HasError() {
				return nil, diags
			}
			wf.WfStepsConfig = steps
		}
		if !wm.TerraformConfig.IsNull() && !wm.TerraformConfig.IsUnknown() {
			tc, diags := expandTerraformConfig(ctx, wm.TerraformConfig)
			if diags.HasError() {
				return nil, diags
			}
			wf.TerraformConfig = tc
		}
		if !wm.EnvironmentVariables.IsNull() && !wm.EnvironmentVariables.IsUnknown() {
			envVars, diags := expandEnvironmentVariables(ctx, wm.EnvironmentVariables)
			if diags.HasError() {
				return nil, diags
			}
			wf.EnvironmentVariables = envVars
		}
		if !wm.DeploymentPlatformConfig.IsNull() && !wm.DeploymentPlatformConfig.IsUnknown() {
			dpcs, diags := expandDeploymentPlatformConfig(ctx, wm.DeploymentPlatformConfig)
			if diags.HasError() {
				return nil, diags
			}
			wf.DeploymentPlatformConfig = dpcs
		}
		if !wm.VcsConfig.IsNull() && !wm.VcsConfig.IsUnknown() {
			vcs, diags := expandVcsConfig(ctx, wm.VcsConfig)
			if diags.HasError() {
				return nil, diags
			}
			wf.VcsConfig = vcs
		}
		if !wm.InputSchemas.IsNull() && !wm.InputSchemas.IsUnknown() {
			schemas, diags := expandInputSchemas(ctx, wm.InputSchemas)
			if diags.HasError() {
				return nil, diags
			}
			wf.InputSchemas = schemas
		}
		if !wm.Approvers.IsNull() && !wm.Approvers.IsUnknown() {
			approvers, diags := expanders.StringList(ctx, wm.Approvers)
			if diags.HasError() {
				return nil, diags
			}
			wf.Approvers = approvers
		}
		if !wm.RunnerConstraints.IsNull() && !wm.RunnerConstraints.IsUnknown() {
			rc, diags := expandRunnerConstraints(ctx, wm.RunnerConstraints)
			if diags.HasError() {
				return nil, diags
			}
			wf.RunnerConstraints = rc
		}
		if !wm.UserSchedules.IsNull() && !wm.UserSchedules.IsUnknown() {
			us, diags := expandWfUserSchedules(ctx, wm.UserSchedules)
			if diags.HasError() {
				return nil, diags
			}
			wf.UserSchedules = us
		}
		if !wm.MiniSteps.IsNull() && !wm.MiniSteps.IsUnknown() {
			ms, diags := expandMiniSteps(ctx, wm.MiniSteps)
			if diags.HasError() {
				return nil, diags
			}
			wf.MiniSteps = ms
		}
		if !wm.ContextTags.IsNull() && !wm.ContextTags.IsUnknown() {
			ct, diags := expandContextTags(ctx, wm.ContextTags)
			if diags.HasError() {
				return nil, diags
			}
			wf.ContextTags = ct
		}

		// Layer 2: fill whatever the user left unset from the matching stack
		// template workflow slot (also the sole source for iac_vcs_config).
		stackTplWf := stackTplWorkflowsById[wm.Id.ValueString()]
		mergeWorkflowWithStackTemplateOverride(wf, stackTplWf)

		// Layer 1: fill whatever is STILL unset from the workflow template
		// revision the matched slot points to.
		mergeWorkflowWithWorkflowTemplateDefaults(wf, workflowTemplates[wm.Id.ValueString()])

		workflows[i] = wf
	}
	return &sgsdkgo.StackWorkflowsConfig{Workflows: workflows}, nil
}

func flattenWorkflowsConfig(ctx context.Context, wfc *sgsdkgo.StackWorkflowsConfig) (types.Object, diag.Diagnostics) {
	nullObj := types.ObjectNull(WorkflowsConfigModel{}.AttributeTypes())
	if wfc == nil {
		return nullObj, nil
	}
	wfListNull := types.ListNull(types.ObjectType{AttrTypes: WorkflowInStackModel{}.AttributeTypes()})
	if len(wfc.Workflows) == 0 {
		m := WorkflowsConfigModel{Workflows: wfListNull}
		return types.ObjectValueFrom(ctx, WorkflowsConfigModel{}.AttributeTypes(), m)
	}

	elements := make([]attr.Value, 0, len(wfc.Workflows))
	for _, wf := range wfc.Workflows {
		if wf == nil {
			continue
		}
		wfSteps, diags := flattenWfStepsConfig(ctx, wf.WfStepsConfig)
		if diags.HasError() {
			return nullObj, diags
		}
		tcObj, diags := flattenTerraformConfig(ctx, wf.TerraformConfig)
		if diags.HasError() {
			return nullObj, diags
		}
		envVars, diags := flattenEnvironmentVariables(ctx, wf.EnvironmentVariables)
		if diags.HasError() {
			return nullObj, diags
		}
		dpcs, diags := flattenDeploymentPlatformConfig(ctx, wf.DeploymentPlatformConfig)
		if diags.HasError() {
			return nullObj, diags
		}
		vcs, diags := flattenVcsConfig(ctx, wf.VcsConfig)
		if diags.HasError() {
			return nullObj, diags
		}
		inputSchemas, diags := flattenInputSchemas(ctx, wf.InputSchemas)
		if diags.HasError() {
			return nullObj, diags
		}
		us, diags := flattenWfUserSchedules(ctx, wf.UserSchedules)
		if diags.HasError() {
			return nullObj, diags
		}
		rc, diags := flattenRunnerConstraints(ctx, wf.RunnerConstraints)
		if diags.HasError() {
			return nullObj, diags
		}
		msObj, diags := flattenMiniSteps(ctx, wf.MiniSteps)
		if diags.HasError() {
			return nullObj, diags
		}
		ctMap, diags := flattenContextTags(ctx, wf.ContextTags)
		if diags.HasError() {
			return nullObj, diags
		}

		tagsList := types.ListNull(types.StringType)
		if wf.Tags != nil {
			l, diags := types.ListValueFrom(ctx, types.StringType, wf.Tags)
			if diags.HasError() {
				return nullObj, diags
			}
			tagsList = l
		}
		approversList := types.ListNull(types.StringType)
		if wf.Approvers != nil {
			l, diags := types.ListValueFrom(ctx, types.StringType, wf.Approvers)
			if diags.HasError() {
				return nullObj, diags
			}
			approversList = l
		}

		wfType := ""
		if wf.WfType != nil {
			wfType = string(*wf.WfType)
		}
		parallelExecution := ""
		if wf.ParallelExecution != nil {
			parallelExecution = string(*wf.ParallelExecution)
		}
		isActive := ""
		if wf.IsActive != nil {
			isActive = string(*wf.IsActive)
		}

		wm := WorkflowInStackModel{
			Id:                        flatteners.StringPtr(wf.Id),
			ResourceName:              flatteners.StringPtr(wf.ResourceName),
			Description:               flatteners.StringPtr(wf.Description),
			Tags:                      tagsList,
			WfType:                    flatteners.String(wfType),
			ParallelExecution:         flatteners.String(parallelExecution),
			WfStepsConfig:             wfSteps,
			TerraformConfig:           tcObj,
			EnvironmentVariables:      envVars,
			DeploymentPlatformConfig:  dpcs,
			TemplateId:                flatteners.StringPtr(wf.TemplateId),
			IsActive:                  flatteners.String(isActive),
			VcsConfig:                 vcs,
			InputSchemas:              inputSchemas,
			Approvers:                 approversList,
			NumberOfApprovalsRequired: flatteners.Int64Ptr(wf.NumberOfApprovalsRequired),
			UserJobCpu:                flatteners.Int64Ptr(wf.UserJobCpu),
			UserJobMemory:             flatteners.Int64Ptr(wf.UserJobMemory),
			UserSchedules:             us,
			MiniSteps:                 msObj,
			ContextTags:               ctMap,
			RunnerConstraints:         rc,
		}
		obj, diags := types.ObjectValueFrom(ctx, WorkflowInStackModel{}.AttributeTypes(), wm)
		if diags.HasError() {
			return nullObj, diags
		}
		elements = append(elements, obj)
	}

	wfList, diags := types.ListValue(types.ObjectType{AttrTypes: WorkflowInStackModel{}.AttributeTypes()}, elements)
	if diags.HasError() {
		return nullObj, diags
	}
	m := WorkflowsConfigModel{Workflows: wfList}
	return types.ObjectValueFrom(ctx, WorkflowsConfigModel{}.AttributeTypes(), m)
}

// ---------------------------------------------------------------------------
// Actions converters
// ---------------------------------------------------------------------------

// expandSingleActionsMap converts one Terraform actions map (either
// default_actions or custom_actions) to API format, stamping isDefault onto
// every action's Default field.
func expandSingleActionsMap(ctx context.Context, actions types.Map, isDefault bool) (map[string]*sgsdkgo.Actions, diag.Diagnostics) {
	if actions.IsNull() || actions.IsUnknown() {
		return nil, nil
	}

	var models map[string]ActionsModel
	if diags := actions.ElementsAs(ctx, &models, false); diags.HasError() {
		return nil, diags
	}

	result := make(map[string]*sgsdkgo.Actions, len(models))
	for k, am := range models {
		def := isDefault
		action := &sgsdkgo.Actions{
			Name:        am.Name.ValueString(),
			Description: am.Description.ValueStringPointer(),
			Default:     &def,
		}

		if !am.Order.IsNull() && !am.Order.IsUnknown() {
			var orderModels map[string]ActionOrderModel
			if diags := am.Order.ElementsAs(ctx, &orderModels, false); diags.HasError() {
				return nil, diags
			}

			order := make(map[string]*sgsdkgo.ActionOrder, len(orderModels))
			for wfId, om := range orderModels {
				ao := &sgsdkgo.ActionOrder{}

				if !om.Parameters.IsNull() && !om.Parameters.IsUnknown() {
					var pm StackActionParametersModel
					if diags := om.Parameters.As(ctx, &pm, basetypes.ObjectAsOptions{
						UnhandledNullAsEmpty:    true,
						UnhandledUnknownAsEmpty: true,
					}); diags.HasError() {
						return nil, diags
					}

					params := &sgsdkgo.StackActionParameters{}

					if !pm.TerraformAction.IsNull() && !pm.TerraformAction.IsUnknown() {
						var tam TerraformActionModel
						if diags := pm.TerraformAction.As(ctx, &tam, basetypes.ObjectAsOptions{
							UnhandledNullAsEmpty:    true,
							UnhandledUnknownAsEmpty: true,
						}); diags.HasError() {
							return nil, diags
						}
						if !tam.Action.IsNull() && !tam.Action.IsUnknown() {
							actionEnum := sgsdkgo.ActionEnum(tam.Action.ValueString())
							params.TerraformAction = &sgsdkgo.TerraformAction{Action: &actionEnum}
						}
					}

					if !pm.DeploymentPlatformConfig.IsNull() && !pm.DeploymentPlatformConfig.IsUnknown() {
						dpcs, diags := expandDeploymentPlatformConfig(ctx, pm.DeploymentPlatformConfig)
						if diags.HasError() {
							return nil, diags
						}
						params.DeploymentPlatformConfig = dpcs
					}

					if !pm.WfStepsConfig.IsNull() && !pm.WfStepsConfig.IsUnknown() {
						steps, diags := expandWfStepsConfig(ctx, pm.WfStepsConfig)
						if diags.HasError() {
							return nil, diags
						}
						params.WfStepsConfig = steps
					}

					if !pm.EnvironmentVariables.IsNull() && !pm.EnvironmentVariables.IsUnknown() {
						envVars, diags := expandEnvironmentVariables(ctx, pm.EnvironmentVariables)
						if diags.HasError() {
							return nil, diags
						}
						params.EnvironmentVariables = envVars
					}

					ao.Parameters = params
				}

				if !om.Dependencies.IsNull() && !om.Dependencies.IsUnknown() {
					var depModels []ActionDependencyModel
					if diags := om.Dependencies.ElementsAs(ctx, &depModels, false); diags.HasError() {
						return nil, diags
					}

					deps := make([]*sgsdkgo.ActionDependency, len(depModels))
					for j, dm := range depModels {
						dep := &sgsdkgo.ActionDependency{Id: dm.Id.ValueString()}

						if !dm.Condition.IsNull() && !dm.Condition.IsUnknown() {
							var cond ActionDependencyConditionModel
							if diags := dm.Condition.As(ctx, &cond, basetypes.ObjectAsOptions{
								UnhandledNullAsEmpty:    true,
								UnhandledUnknownAsEmpty: true,
							}); diags.HasError() {
								return nil, diags
							}
							dep.Condition = &sgsdkgo.ActionDependencyCondition{
								LatestStatus: cond.LatestStatus.ValueString(),
							}
						}

						deps[j] = dep
					}

					ao.Dependencies = deps
				}

				order[wfId] = ao
			}

			action.Order = order
		}

		result[k] = action
	}

	return result, nil
}

// expandActionsMap merges the Terraform default_actions and custom_actions maps
// into the single map the SDK's Actions field expects, stamping Default=true on
// entries from defaultActions and Default=false on entries from customActions.
// An action key present in both is a config error, since it's ambiguous which
// Default value should win. Any action the stack template revision (tpl)
// resolves to — its own Actions verbatim, or a freshly generated set — that
// isn't already covered by defaultActions/customActions is filled in as a
// lowest-precedence default; see generateStackActions.
func expandActionsMap(ctx context.Context, defaultActions, customActions types.Map, tpl *stacktemplaterevisions.ReadStackTemplateRevisionModel, workflowTemplates map[string]*workflowtemplaterevisions.ReadWorkflowTemplateRevisionModel) (map[string]*sgsdkgo.Actions, diag.Diagnostics) {
	var diags diag.Diagnostics

	defaults, d := expandSingleActionsMap(ctx, defaultActions, true)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	customs, d := expandSingleActionsMap(ctx, customActions, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	result := make(map[string]*sgsdkgo.Actions, len(defaults)+len(customs))
	for k, v := range defaults {
		result[k] = v
	}
	for k, v := range customs {
		if _, exists := result[k]; exists {
			diags.Append(diag.NewErrorDiagnostic(
				"Duplicate action key",
				"Action \""+k+"\" is defined in both default_actions and custom_actions; it must only appear in one.",
			))
			continue
		}
		result[k] = v
	}
	if diags.HasError() {
		return nil, diags
	}

	isDefault := true
	for k, v := range generateStackActions(tpl, workflowTemplates) {
		if _, exists := result[k]; exists || v == nil {
			continue
		}
		// Force Default=true on a shallow copy: an action resolved from the
		// template is always a default action for the stack, regardless of
		// what Default flag it carries. custom_actions is exclusively
		// user-authored, so these entries never count as custom.
		action := *v
		action.Default = &isDefault
		result[k] = &action
	}
	if len(result) == 0 {
		return nil, diags
	}

	return result, diags
}

// generateStackActions resolves the stack template revision's Actions,
// following the platform's own create-flow algorithm exactly (see
// default_actions_generation_doc.txt):
//
//  1. If the revision already defines its own Actions, they're used verbatim
//     — nothing is generated.
//  2. Otherwise, generate exactly three actions (apply/plan/destroy) from the
//     revision's own WorkflowsConfig.workflows list, in declaration order —
//     never the stack's own workflows_config, which this algorithm doesn't
//     consult at all. apply/plan chain forward through that list; destroy
//     chains backward. Each entry depends on the previous one visited in its
//     own chain's direction (a COMPLETED-status dependency); the first
//     workflow visited in a given direction has none. Order map keys are the
//     workflow slot id (StackTemplateRevisionWorkflow.Id) — bare, no prefix,
//     since at create time the workflow doesn't have its own resource id yet.
//  3. Post-process: for any workflow slot whose referenced workflow template
//     resolves to a non-Terraform source_config_kind (HELM, KUBECTL,
//     ANSIBLE_PLAYBOOK, CUSTOM, or CLOUDFORMATION), its parameters are
//     cleared to empty in all three actions — dependencies stay untouched.
//     terraform apply/plan/destroy has no meaning for those step types.
//
// workflowTemplates is keyed by workflow slot id (see resolveWorkflowTemplates
// in resource.go) — nil is safe (every slot is then treated as Terraform-typed
// for step 3, since there's nothing to classify it by) but only accurate when
// tpl.Actions is already populated (step 1 doesn't need workflowTemplates at
// all); see reResolveOnRevisionChange's use of that fact.
func generateStackActions(tpl *stacktemplaterevisions.ReadStackTemplateRevisionModel, workflowTemplates map[string]*workflowtemplaterevisions.ReadWorkflowTemplateRevisionModel) map[string]*sgsdkgo.Actions {
	if tpl == nil {
		return nil
	}
	if len(tpl.Actions) > 0 {
		return tpl.Actions
	}
	if tpl.WorkflowsConfig == nil || len(tpl.WorkflowsConfig.Workflows) == 0 {
		return nil
	}

	ids := make([]string, 0, len(tpl.WorkflowsConfig.Workflows))
	nonTerraformIds := make(map[string]bool)
	for _, w := range tpl.WorkflowsConfig.Workflows {
		if w == nil || w.Id == nil {
			continue
		}
		ids = append(ids, *w.Id)
		if wt := workflowTemplates[*w.Id]; wt != nil && wt.SourceConfigKind != nil {
			switch *wt.SourceConfigKind {
			case workflowtemplates.WorkflowTemplateSourceConfigKindHelm,
				workflowtemplates.WorkflowTemplateSourceConfigKindKubectl,
				workflowtemplates.WorkflowTemplateSourceConfigKindAnsiblePlaybook,
				workflowtemplates.WorkflowTemplateSourceConfigKindCustom,
				workflowtemplates.WorkflowTemplateSourceConfigKindCloudformation:
				nonTerraformIds[*w.Id] = true
			}
		}
	}
	if len(ids) == 0 {
		return nil
	}
	reversedIds := make([]string, len(ids))
	for i, wfId := range ids {
		reversedIds[len(ids)-1-i] = wfId
	}

	chain := func(action sgsdkgo.ActionEnum, order []string) map[string]*sgsdkgo.ActionOrder {
		result := make(map[string]*sgsdkgo.ActionOrder, len(order))
		for i, wfId := range order {
			var deps []*sgsdkgo.ActionDependency
			if i > 0 {
				deps = []*sgsdkgo.ActionDependency{
					{
						Id:        order[i-1],
						Condition: &sgsdkgo.ActionDependencyCondition{LatestStatus: "COMPLETED"},
					},
				}
			}
			a := action
			result[wfId] = &sgsdkgo.ActionOrder{
				Parameters:   &sgsdkgo.StackActionParameters{TerraformAction: &sgsdkgo.TerraformAction{Action: &a}},
				Dependencies: deps,
			}
		}
		return result
	}

	isDefault := true
	applyDesc := "use this action to create resources in the stack"
	planDesc := "use this action to plan resources in the stack"
	destroyDesc := "use this action to destroy resources in the stack"
	actions := map[string]*sgsdkgo.Actions{
		"apply":   {Name: "Create", Description: &applyDesc, Default: &isDefault, Order: chain(sgsdkgo.ActionEnumApply, ids)},
		"plan":    {Name: "Plan", Description: &planDesc, Default: &isDefault, Order: chain(sgsdkgo.ActionEnumPlan, ids)},
		"destroy": {Name: "Destroy", Description: &destroyDesc, Default: &isDefault, Order: chain(sgsdkgo.ActionEnumDestroy, reversedIds)},
	}

	// Post-process: blank parameters for non-Terraform workflow types across
	// all three actions; dependencies stay as chained above.
	for wfId := range nonTerraformIds {
		for _, action := range actions {
			if entry := action.Order[wfId]; entry != nil {
				entry.Parameters = &sgsdkgo.StackActionParameters{}
			}
		}
	}

	return actions
}

// translateActionsOrderKeys rewrites action order map keys AND dependency ids
// back to slot uuid using relations (WorkflowRelationsMap: {slot uuid:
// workflow's own resource id, e.g. "/wfs/<name>"}). Once a workflow exists,
// the platform substitutes its real id for the slot uuid everywhere in
// returned Actions — both the order key and dependencies[].id (see
// default_actions.json) — flattening that directly caused "inconsistent
// result after apply" (state must stay keyed by the uuid the user
// configured). Ids with no match in relations are left as-is.
func translateActionsOrderKeys(actions map[string]*sgsdkgo.Actions, relations map[string]interface{}) map[string]*sgsdkgo.Actions {
	if len(actions) == 0 || len(relations) == 0 {
		return actions
	}
	reverse := make(map[string]string, len(relations))
	for slotId, v := range relations {
		if workflowId, ok := v.(string); ok && workflowId != "" {
			reverse[workflowId] = slotId
		}
	}
	if len(reverse) == 0 {
		return actions
	}

	result := make(map[string]*sgsdkgo.Actions, len(actions))
	for k, a := range actions {
		if a == nil || len(a.Order) == 0 {
			result[k] = a
			continue
		}
		translatedOrder := make(map[string]*sgsdkgo.ActionOrder, len(a.Order))
		for orderKey, ao := range a.Order {
			newKey := orderKey
			if slotId, ok := reverse[orderKey]; ok {
				newKey = slotId
			}
			translatedOrder[newKey] = translateDependencyIds(ao, reverse)
		}
		action := *a
		action.Order = translatedOrder
		result[k] = &action
	}
	return result
}

// translateDependencyIds applies the same slot-uuid translation to an order
// entry's dependency ids — they reference other workflows, so the API
// substitutes real workflow ids there too, not just in the order key.
func translateDependencyIds(ao *sgsdkgo.ActionOrder, reverse map[string]string) *sgsdkgo.ActionOrder {
	if ao == nil || len(ao.Dependencies) == 0 {
		return ao
	}
	deps := make([]*sgsdkgo.ActionDependency, len(ao.Dependencies))
	for i, dep := range ao.Dependencies {
		if dep == nil {
			continue
		}
		d := *dep
		if slotId, ok := reverse[d.Id]; ok {
			d.Id = slotId
		}
		deps[i] = &d
	}
	result := *ao
	result.Dependencies = deps
	return &result
}

// flattenActionsMap converts the API actions map to Terraform format,
// including the order/parameters/dependencies structure. Actions with
// Default == true are split into defaultActions; everything else goes into
// customActions — unless forceDefault is set, in which case every action goes
// into defaultActions regardless of its own Default flag. forceDefault is for
// flattening a stack TEMPLATE's actions: custom_actions is exclusively
// user-authored on the stack itself, so nothing from a template — even an
// action the template itself marked non-default — may ever land there.
func flattenActionsMap(ctx context.Context, actions map[string]*sgsdkgo.Actions, forceDefault bool) (defaultActions types.Map, customActions types.Map, diags diag.Diagnostics) {
	mapNull := types.MapNull(types.ObjectType{AttrTypes: ActionsModel{}.AttributeTypes()})
	if actions == nil {
		return mapNull, mapNull, nil
	}
	defaultElements := make(map[string]attr.Value)
	customElements := make(map[string]attr.Value)
	for k, a := range actions {
		if a == nil {
			continue
		}
		orderObj := types.MapNull(types.ObjectType{AttrTypes: ActionOrderModel{}.AttributeTypes()})
		if a.Order != nil {
			orderElements := make(map[string]attr.Value, len(a.Order))
			for wfId, ao := range a.Order {
				if ao == nil {
					continue
				}
				paramsObj := types.ObjectNull(StackActionParametersModel{}.AttributeTypes())
				if ao.Parameters != nil {
					p := ao.Parameters
					taObj := types.ObjectNull(TerraformActionModel{}.AttributeTypes())
					if p.TerraformAction != nil && p.TerraformAction.Action != nil {
						tam := TerraformActionModel{Action: flatteners.String(string(*p.TerraformAction.Action))}
						obj, d := types.ObjectValueFrom(ctx, TerraformActionModel{}.AttributeTypes(), tam)
						if d.HasError() {
							return mapNull, mapNull, d
						}
						taObj = obj
					}
					dpcList, d := flattenDeploymentPlatformConfig(ctx, p.DeploymentPlatformConfig)
					if d.HasError() {
						return mapNull, mapNull, d
					}
					wfStepsList, d := flattenWfStepsConfig(ctx, p.WfStepsConfig)
					if d.HasError() {
						return mapNull, mapNull, d
					}
					envList, d := flattenEnvironmentVariables(ctx, p.EnvironmentVariables)
					if d.HasError() {
						return mapNull, mapNull, d
					}
					pm := StackActionParametersModel{
						TerraformAction:          taObj,
						DeploymentPlatformConfig: dpcList,
						WfStepsConfig:            wfStepsList,
						EnvironmentVariables:     envList,
					}
					obj, d := types.ObjectValueFrom(ctx, StackActionParametersModel{}.AttributeTypes(), pm)
					if d.HasError() {
						return mapNull, mapNull, d
					}
					paramsObj = obj
				}

				// len == 0, not != nil: config had no dependencies (null), but the API
				// echoes an explicit [] instead of omitting the field — flattening that
				// as a non-null empty list caused "inconsistent result after apply".
				depListObj := types.ListNull(types.ObjectType{AttrTypes: ActionDependencyModel{}.AttributeTypes()})
				if len(ao.Dependencies) > 0 {
					depElems := make([]attr.Value, 0, len(ao.Dependencies))
					for _, dep := range ao.Dependencies {
						if dep == nil {
							continue
						}
						condObj := types.ObjectNull(ActionDependencyConditionModel{}.AttributeTypes())
						if dep.Condition != nil {
							condM := ActionDependencyConditionModel{LatestStatus: flatteners.String(dep.Condition.LatestStatus)}
							obj, d := types.ObjectValueFrom(ctx, ActionDependencyConditionModel{}.AttributeTypes(), condM)
							if d.HasError() {
								return mapNull, mapNull, d
							}
							condObj = obj
						}
						dm := ActionDependencyModel{Id: flatteners.String(dep.Id), Condition: condObj}
						depObj, d := types.ObjectValueFrom(ctx, ActionDependencyModel{}.AttributeTypes(), dm)
						if d.HasError() {
							return mapNull, mapNull, d
						}
						depElems = append(depElems, depObj)
					}
					l, d := types.ListValue(types.ObjectType{AttrTypes: ActionDependencyModel{}.AttributeTypes()}, depElems)
					if d.HasError() {
						return mapNull, mapNull, d
					}
					depListObj = l
				}
				aoM := ActionOrderModel{Parameters: paramsObj, Dependencies: depListObj}
				aoObj, d := types.ObjectValueFrom(ctx, ActionOrderModel{}.AttributeTypes(), aoM)
				if d.HasError() {
					return mapNull, mapNull, d
				}
				orderElements[wfId] = aoObj
			}
			m, d := types.MapValue(types.ObjectType{AttrTypes: ActionOrderModel{}.AttributeTypes()}, orderElements)
			if d.HasError() {
				return mapNull, mapNull, d
			}
			orderObj = m
		}
		am := ActionsModel{
			Name:        flatteners.String(a.Name),
			Description: flatteners.StringPtr(a.Description),
			Order:       orderObj,
		}
		obj, d := types.ObjectValueFrom(ctx, ActionsModel{}.AttributeTypes(), am)
		if d.HasError() {
			return mapNull, mapNull, d
		}
		if forceDefault || (a.Default != nil && *a.Default) {
			defaultElements[k] = obj
		} else {
			customElements[k] = obj
		}
	}
	defaultMap, d := types.MapValue(types.ObjectType{AttrTypes: ActionsModel{}.AttributeTypes()}, defaultElements)
	if d.HasError() {
		return mapNull, mapNull, d
	}
	customMap, d := types.MapValue(types.ObjectType{AttrTypes: ActionsModel{}.AttributeTypes()}, customElements)
	if d.HasError() {
		return mapNull, mapNull, d
	}
	return defaultMap, customMap, nil
}

// ---------------------------------------------------------------------------
// ToAPIModel / ToUpdateAPIModel / BuildAPIModelToStackModel
// ---------------------------------------------------------------------------

// contextTagsFromTemplate converts a stack template revision's ContextTags
// (map[string]string) to the map[string]*string shape sgsdkgo.Stack expects.
func contextTagsFromTemplate(ct map[string]string) map[string]*string {
	if ct == nil {
		return nil
	}
	result := make(map[string]*string, len(ct))
	for k, v := range ct {
		val := v
		result[k] = &val
	}
	return result
}

// ToAPIModel converts the plan into a Create payload. orgName prefixes
// template_group_id to the wire format the API expects ("/<org>/<name>:<rev>"
// — see the template_group_id schema description); the Terraform-facing
// value stays bare ("<name>:<rev>"), matching what fetchTemplateRevision
// reads against the stack's own org. tpl is the stack template revision
// resolved from template_group_id (nil if unset) —
// description/tags/context_tags/actions fall back to the template's value
// when the user left the corresponding field unset. workflows_config has no
// template-based resolution yet.
func (m *StackResourceModel) ToAPIModel(ctx context.Context, orgName string, tpl *stacktemplaterevisions.ReadStackTemplateRevisionModel, workflowTemplates map[string]*workflowtemplaterevisions.ReadWorkflowTemplateRevisionModel) (*sgsdkgo.Stack, diag.Diagnostics) {
	var diags diag.Diagnostics
	apiModel := &sgsdkgo.Stack{
		// id is Required, always known here.
		Id: m.Id.ValueStringPointer(),
	}
	// resource_name is Optional+Computed: ValueStringPointer() returns &"" for unknown,
	// which would send an explicit empty name on create whenever the user hasn't set it.
	// Only set when known. No template counterpart exists for resource_name.
	if !m.ResourceName.IsUnknown() && !m.ResourceName.IsNull() {
		apiModel.ResourceName = m.ResourceName.ValueStringPointer()
	}

	if !m.Description.IsUnknown() && !m.Description.IsNull() {
		apiModel.Description = m.Description.ValueStringPointer()
	} else if tpl != nil {
		apiModel.Description = tpl.LongDescription
	}

	if !m.Tags.IsUnknown() && !m.Tags.IsNull() {
		tags, tagDiags := expanders.StringList(ctx, m.Tags)
		diags.Append(tagDiags...)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.Tags = tags
	} else if tpl != nil {
		apiModel.Tags = tpl.Tags
	}

	// No template counterpart exists for environment_variables.
	if !m.EnvironmentVariables.IsUnknown() && !m.EnvironmentVariables.IsNull() {
		envVars, envDiags := expandEnvironmentVariables(ctx, m.EnvironmentVariables)
		diags.Append(envDiags...)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.EnvironmentVariables = envVars
	}

	// No template counterpart exists for deployment_platform_config.
	if !m.DeploymentPlatformConfig.IsUnknown() && !m.DeploymentPlatformConfig.IsNull() {
		dpc, dpcDiags := expandDeploymentPlatformConfig(ctx, m.DeploymentPlatformConfig)
		diags.Append(dpcDiags...)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.DeploymentPlatformConfig = dpc
	}

	if !m.TemplateGroupId.IsUnknown() && !m.TemplateGroupId.IsNull() {
		prefixed := fmt.Sprintf("/%s/%s", orgName, m.TemplateGroupId.ValueString())
		apiModel.TemplateGroupId = &prefixed
	}

	// default_actions and custom_actions are merged into the SDK's single Actions
	// map, with Default set true/false based on which Terraform attribute each
	// action came from; any template action not already covered fills in as a
	// lowest-precedence default.
	actions, actionDiags := expandActionsMap(ctx, m.DefaultActions, m.CustomActions, tpl, workflowTemplates)
	diags.Append(actionDiags...)
	if diags.HasError() {
		return nil, diags
	}
	if actions != nil {
		apiModel.Actions = actions
	}

	// workflows_config resolves per-slot from up to three layers — see
	// expandWorkflowsConfig.
	if !m.WorkflowsConfig.IsUnknown() && !m.WorkflowsConfig.IsNull() {
		wfc, wfcDiags := expandWorkflowsConfig(ctx, m.WorkflowsConfig, tpl, workflowTemplates)
		diags.Append(wfcDiags...)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.WorkflowsConfig = wfc
	}

	// No template counterpart exists for user_schedules.
	if !m.UserSchedules.IsUnknown() && !m.UserSchedules.IsNull() {
		userSchedules, usDiags := expandUserSchedules(ctx, m.UserSchedules)
		diags.Append(usDiags...)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.UserSchedules = userSchedules
	}

	if !m.ContextTags.IsUnknown() && !m.ContextTags.IsNull() {
		contextTags, ctDiags := expandContextTags(ctx, m.ContextTags)
		diags.Append(ctDiags...)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.ContextTags = contextTags
	} else if tpl != nil {
		apiModel.ContextTags = contextTagsFromTemplate(tpl.ContextTags)
	}

	// No template counterpart exists for mini_steps.
	if !m.MiniSteps.IsUnknown() && !m.MiniSteps.IsNull() {
		miniSteps, msDiags := expandMiniSteps(ctx, m.MiniSteps)
		diags.Append(msDiags...)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.MiniSteps = miniSteps
	}

	return apiModel, diags
}

// ToUpdateAPIModel converts the plan into an Update payload. orgName prefixes
// template_group_id to the wire format the API expects — see ToAPIModel. tpl
// is the stack template revision resolved from template_group_id (nil if
// unset) — description/tags/context_tags/actions fall back to the template's
// value when the field is unknown (no prior state and no config value). A
// field the user explicitly nulled out still clears via sgsdkgo.Null, even
// with a template present — explicit null always means "clear it", not
// "inherit".
func (m *StackResourceModel) ToUpdateAPIModel(ctx context.Context, orgName string, tpl *stacktemplaterevisions.ReadStackTemplateRevisionModel, workflowTemplates map[string]*workflowtemplaterevisions.ReadWorkflowTemplateRevisionModel) (*sgsdkgo.PatchedStack, diag.Diagnostics) {
	var diags diag.Diagnostics
	apiModel := &sgsdkgo.PatchedStack{}

	// No template counterpart exists for resource_name.
	if !m.ResourceName.IsUnknown() && !m.ResourceName.IsNull() {
		apiModel.ResourceName = sgsdkgo.Optional(m.ResourceName.ValueString())
	} else if m.ResourceName.IsNull() {
		apiModel.ResourceName = sgsdkgo.Null[string]()
	}

	if !m.Description.IsUnknown() && !m.Description.IsNull() {
		apiModel.Description = sgsdkgo.Optional(m.Description.ValueString())
	} else if m.Description.IsNull() {
		apiModel.Description = sgsdkgo.Null[string]()
	} else if tpl != nil && tpl.LongDescription != nil {
		apiModel.Description = sgsdkgo.Optional(*tpl.LongDescription)
	}

	if !m.Tags.IsUnknown() && !m.Tags.IsNull() {
		tags, tagDiags := expanders.StringList(ctx, m.Tags)
		diags.Append(tagDiags...)
		if diags.HasError() {
			return nil, diags
		}
		if tags != nil {
			apiModel.Tags = sgsdkgo.Optional(tags)
		} else {
			apiModel.Tags = sgsdkgo.Null[[]string]()
		}
	} else if m.Tags.IsNull() {
		apiModel.Tags = sgsdkgo.Null[[]string]()
	} else if tpl != nil && tpl.Tags != nil {
		apiModel.Tags = sgsdkgo.Optional(tpl.Tags)
	}

	// No template counterpart exists for environment_variables.
	if !m.EnvironmentVariables.IsUnknown() && !m.EnvironmentVariables.IsNull() {
		envVars, envDiags := expandEnvironmentVariables(ctx, m.EnvironmentVariables)
		diags.Append(envDiags...)
		if diags.HasError() {
			return nil, diags
		}
		if envVars != nil {
			apiModel.EnvironmentVariables = sgsdkgo.Optional(envVars)
		} else {
			apiModel.EnvironmentVariables = sgsdkgo.Null[[]*sgsdkgo.EnvVars]()
		}
	} else if m.EnvironmentVariables.IsNull() {
		apiModel.EnvironmentVariables = sgsdkgo.Null[[]*sgsdkgo.EnvVars]()
	}

	// No template counterpart exists for deployment_platform_config.
	if !m.DeploymentPlatformConfig.IsUnknown() && !m.DeploymentPlatformConfig.IsNull() {
		dpc, dpcDiags := expandDeploymentPlatformConfig(ctx, m.DeploymentPlatformConfig)
		diags.Append(dpcDiags...)
		if diags.HasError() {
			return nil, diags
		}
		if dpc != nil {
			apiModel.DeploymentPlatformConfig = sgsdkgo.Optional(dpc)
		} else {
			apiModel.DeploymentPlatformConfig = sgsdkgo.Null[[]*sgsdkgo.DeploymentPlatformConfig]()
		}
	} else if m.DeploymentPlatformConfig.IsNull() {
		apiModel.DeploymentPlatformConfig = sgsdkgo.Null[[]*sgsdkgo.DeploymentPlatformConfig]()
	}

	if !m.TemplateGroupId.IsUnknown() && !m.TemplateGroupId.IsNull() {
		apiModel.TemplateGroupId = sgsdkgo.Optional(fmt.Sprintf("/%s/%s", orgName, m.TemplateGroupId.ValueString()))
	} else if m.TemplateGroupId.IsNull() {
		apiModel.TemplateGroupId = sgsdkgo.Null[string]()
	}

	// default_actions and custom_actions are merged into the SDK's single Actions
	// map, with Default set true/false based on which Terraform attribute each
	// action came from; any template action not already covered fills in as a
	// lowest-precedence default.
	actions, actionDiags := expandActionsMap(ctx, m.DefaultActions, m.CustomActions, tpl, workflowTemplates)
	diags.Append(actionDiags...)
	if diags.HasError() {
		return nil, diags
	}
	if actions != nil {
		apiModel.Actions = sgsdkgo.Optional(actions)
	} else if m.DefaultActions.IsNull() || m.CustomActions.IsNull() {
		apiModel.Actions = sgsdkgo.Null[map[string]*sgsdkgo.Actions]()
	}

	// workflows_config resolves per-slot from up to three layers — see
	// expandWorkflowsConfig. updateWorkflowsFromConfig is a query param (not
	// part of the request body) telling the API to propagate this resolved
	// WorkflowsConfig down to the actual live workflow resources; it's only
	// set when we're actually sending workflow config data, not on an
	// explicit clear.
	if !m.WorkflowsConfig.IsUnknown() && !m.WorkflowsConfig.IsNull() {
		wfc, wfcDiags := expandWorkflowsConfig(ctx, m.WorkflowsConfig, tpl, workflowTemplates)
		diags.Append(wfcDiags...)
		if diags.HasError() {
			return nil, diags
		}
		if wfc != nil {
			apiModel.WorkflowsConfig = sgsdkgo.Optional(*wfc)
			sync := true
			apiModel.UpdateWorkflowsFromConfig = &sync
		} else {
			apiModel.WorkflowsConfig = sgsdkgo.Null[sgsdkgo.StackWorkflowsConfig]()
		}
	} else if m.WorkflowsConfig.IsNull() {
		apiModel.WorkflowsConfig = sgsdkgo.Null[sgsdkgo.StackWorkflowsConfig]()
	}

	// No template counterpart exists for user_schedules.
	if !m.UserSchedules.IsUnknown() && !m.UserSchedules.IsNull() {
		userSchedules, usDiags := expandUserSchedules(ctx, m.UserSchedules)
		diags.Append(usDiags...)
		if diags.HasError() {
			return nil, diags
		}
		if userSchedules != nil {
			apiModel.UserSchedules = sgsdkgo.Optional(userSchedules)
		} else {
			apiModel.UserSchedules = sgsdkgo.Null[[]*sgsdkgo.StackUserSchedules]()
		}
	} else if m.UserSchedules.IsNull() {
		apiModel.UserSchedules = sgsdkgo.Null[[]*sgsdkgo.StackUserSchedules]()
	}

	if !m.ContextTags.IsUnknown() && !m.ContextTags.IsNull() {
		contextTags, ctDiags := expandContextTags(ctx, m.ContextTags)
		diags.Append(ctDiags...)
		if diags.HasError() {
			return nil, diags
		}
		if contextTags != nil {
			apiModel.ContextTags = sgsdkgo.Optional(contextTags)
		} else {
			apiModel.ContextTags = sgsdkgo.Null[map[string]*string]()
		}
	} else if m.ContextTags.IsNull() {
		apiModel.ContextTags = sgsdkgo.Null[map[string]*string]()
	} else if tpl != nil && tpl.ContextTags != nil {
		apiModel.ContextTags = sgsdkgo.Optional(contextTagsFromTemplate(tpl.ContextTags))
	}

	// No template counterpart exists for mini_steps.
	if !m.MiniSteps.IsUnknown() && !m.MiniSteps.IsNull() {
		miniSteps, msDiags := expandMiniSteps(ctx, m.MiniSteps)
		diags.Append(msDiags...)
		if diags.HasError() {
			return nil, diags
		}
		if miniSteps != nil {
			apiModel.MiniSteps = sgsdkgo.Optional(*miniSteps)
		} else {
			apiModel.MiniSteps = sgsdkgo.Null[sgsdkgo.MiniStepsSchema]()
		}
	} else if m.MiniSteps.IsNull() {
		apiModel.MiniSteps = sgsdkgo.Null[sgsdkgo.MiniStepsSchema]()
	}

	return apiModel, diags
}

// validateCustomActionsAgainstRevision errors when a custom_actions entry's
// order key or dependency id references a workflow slot that no longer
// exists on the new stack template revision tpl. Once template_group_id
// changes, generateStackActions rebuilds default_actions from tpl's own
// workflow list automatically, but custom_actions is user-authored and can't
// be silently rewritten — a dangling reference must be caught here rather
// than surfacing as an opaque API error. Ids also present in the stack's own
// workflows_config.workflows are exempt: those are user-declared slots that
// may have no template backing at all.
func validateCustomActionsAgainstRevision(ctx context.Context, customActions types.Map, workflowsConfig types.Object, tpl *stacktemplaterevisions.ReadStackTemplateRevisionModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if customActions.IsNull() || customActions.IsUnknown() {
		return diags
	}

	valid := make(map[string]bool)
	if tpl != nil && tpl.WorkflowsConfig != nil {
		for _, w := range tpl.WorkflowsConfig.Workflows {
			if w != nil && w.Id != nil {
				valid[*w.Id] = true
			}
		}
	}
	if !workflowsConfig.IsNull() && !workflowsConfig.IsUnknown() {
		var wfc WorkflowsConfigModel
		if d := workflowsConfig.As(ctx, &wfc, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); !d.HasError() && !wfc.Workflows.IsNull() && !wfc.Workflows.IsUnknown() {
			var wfModels []WorkflowInStackModel
			if d2 := wfc.Workflows.ElementsAs(ctx, &wfModels, false); !d2.HasError() {
				for _, wm := range wfModels {
					if !wm.Id.IsNull() && !wm.Id.IsUnknown() {
						valid[wm.Id.ValueString()] = true
					}
				}
			}
		}
	}

	customs, d := expandSingleActionsMap(ctx, customActions, false)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	for actionKey, a := range customs {
		if a == nil {
			continue
		}
		for orderKey, ao := range a.Order {
			if !valid[orderKey] {
				diags.AddError(
					"custom_actions references a removed workflow",
					fmt.Sprintf("custom_actions[%q].order references workflow %q, which no longer exists on the new stack template revision. Update custom_actions to remove it before changing template_group_id.", actionKey, orderKey),
				)
			}
			if ao == nil {
				continue
			}
			for _, dep := range ao.Dependencies {
				if dep != nil && !valid[dep.Id] {
					diags.AddError(
						"custom_actions references a removed workflow",
						fmt.Sprintf("custom_actions[%q].order[%q].dependencies references workflow %q, which no longer exists on the new stack template revision. Update custom_actions to remove it before changing template_group_id.", actionKey, orderKey, dep.Id),
					)
				}
			}
		}
	}
	return diags
}

// reResolveOnRevisionChange re-resolves the template-derived fields the user
// left unset against the NEW stack template revision tpl, used by ModifyPlan
// on a template_group_id change. Values are computed via the same flatteners
// the Read path (BuildAPIModelToStackModel) uses, so plan == apply.
// default_actions is always re-resolved — it's Computed-only, so it's always
// "user-unset" by definition. custom_actions and the scalar fields are only
// re-resolved when the user left them unset in config. Fields with no
// template counterpart (resource_name, environment_variables,
// deployment_platform_config, user_schedules, mini_steps) and
// workflows_config are untouched.
func reResolveOnRevisionChange(ctx context.Context, plan *StackResourceModel, config StackResourceModel, tpl *stacktemplaterevisions.ReadStackTemplateRevisionModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if config.Description.IsNull() {
		plan.Description = flatteners.StringPtr(tpl.LongDescription)
	}

	if config.Tags.IsNull() {
		if tpl.Tags != nil {
			tagsList, d := flatteners.ListOfStringToTerraformList(tpl.Tags)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			plan.Tags = tagsList
		} else {
			plan.Tags = types.ListNull(types.StringType)
		}
	}

	if config.ContextTags.IsNull() {
		ctMap, d := flattenContextTags(ctx, contextTagsFromTemplate(tpl.ContextTags))
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		plan.ContextTags = ctMap
	}

	// forceDefault=true: custom_actions is exclusively user-authored on the
	// stack, so it's never touched here — everything the template resolves to
	// counts as a default action for the stack, regardless of what Default
	// flag it carries. generateStackActions(tpl, nil) is passed a nil
	// workflowTemplates since ModifyPlan never calls resolveWorkflowTemplates
	// — safe (and exactly accurate) whenever tpl.Actions is already populated,
	// since that path (step 1: verbatim copy) never consults
	// workflowTemplates at all; see actionsNeedGeneration below for the case
	// where it doesn't.
	defaultActions, _, d := flattenActionsMap(ctx, generateStackActions(tpl, nil), true)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	plan.DefaultActions = defaultActions

	// When tpl.Actions is empty, ToUpdateAPIModel's actual expandActionsMap
	// call generates a fresh apply/plan/destroy set AND blanks parameters for
	// non-Terraform workflow types (generateStackActions step 3) — a
	// classification that needs workflowTemplates, which this function has no
	// way to resolve. A known plan.DefaultActions value that then doesn't
	// match what apply actually returns is exactly what Terraform's "Provider
	// produced inconsistent result after apply" guards against, so mark it
	// unknown instead whenever that's a possibility — deferring to apply-time
	// truth is always safe, since Terraform's consistency check only applies
	// to values that were known in the plan.
	if actionsNeedGeneration(tpl) {
		plan.DefaultActions = types.MapUnknown(types.ObjectType{AttrTypes: ActionsModel{}.AttributeTypes()})
	}

	return diags
}

// actionsNeedGeneration reports whether generateStackActions would need to
// synthesize a fresh apply/plan/destroy set (tpl.Actions is empty) rather
// than just copying tpl.Actions verbatim.
func actionsNeedGeneration(tpl *stacktemplaterevisions.ReadStackTemplateRevisionModel) bool {
	return tpl == nil || len(tpl.Actions) == 0
}

// reResolveWorkflowsConfigOnRevisionChange re-derives workflows_config's
// template-derived per-workflow fields (see mergeWorkflowWithStackTemplateOverride
// / mergeWorkflowWithWorkflowTemplateDefaults) against the new revision tpl.
// That merge only runs inside expandWorkflowsConfig, which apply calls with
// the PLAN value — so if plan still carries the OLD revision's merged values
// forward (UseStateForUnknown, since config left them unset), apply would
// silently send stale data instead of re-deriving from the new revision.
// wf_type/parallel_execution have no template counterpart, so they're
// preserved from prior state rather than wiped to null; a slot newly added in
// this same apply has no prior value to preserve, so those are left as
// expandWorkflowsConfig(config, ...) computed them. state, not plan,
// is the source for prior values — plan can be Unknown here for reasons
// unrelated to the revision change (e.g. a slot added/removed in the same
// apply), while state is always fully known.
func reResolveWorkflowsConfigOnRevisionChange(ctx context.Context, plan *StackResourceModel, config, state StackResourceModel, tpl *stacktemplaterevisions.ReadStackTemplateRevisionModel, workflowTemplates map[string]*workflowtemplaterevisions.ReadWorkflowTemplateRevisionModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if config.WorkflowsConfig.IsNull() || config.WorkflowsConfig.IsUnknown() {
		return diags
	}

	prior, d := expandWorkflowsConfig(ctx, state.WorkflowsConfig, tpl, workflowTemplates)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	priorById := make(map[string]*sgsdkgo.StackWorkflowsConfigWorkflow)
	if prior != nil {
		for _, w := range prior.Workflows {
			if w != nil && w.Id != nil {
				priorById[*w.Id] = w
			}
		}
	}

	fresh, d := expandWorkflowsConfig(ctx, config.WorkflowsConfig, tpl, workflowTemplates)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	if fresh == nil {
		return diags
	}
	for _, w := range fresh.Workflows {
		if w == nil || w.Id == nil {
			continue
		}
		if p := priorById[*w.Id]; p != nil {
			w.WfType = p.WfType
			w.ParallelExecution = p.ParallelExecution
		}
	}

	wfcObj, d := flattenWorkflowsConfig(ctx, fresh)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	plan.WorkflowsConfig = wfcObj

	return diags
}

// BuildAPIModelToStackModel converts the API response into a StackResourceModel.
// workflowGroupId isn't part of the API response — the platform has no concept
// of "workflow group" on a stack response — so it must be threaded through from
// the caller's own plan/state rather than derived here. orgName strips the
// "/<org>/" prefix the API returns on template_group_id back off — see
// ToAPIModel — so state stays bare, matching what the user types in config.
func BuildAPIModelToStackModel(ctx context.Context, orgName string, apiResponse *sgsdkgo.StackData, workflowGroupId types.String) (*StackResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Strip the owning-org prefix the API returns ("/<org>/<name>:<rev>") back
	// to bare ("<name>:<rev>") — mirrors stack_template_revision's own
	// template_id/iac_template_id stripping. Falls back to the raw value if it
	// has no "/<org>/" prefix (e.g. server omitted it).
	strippedTemplateGroupId := apiResponse.TemplateGroupId
	if strippedTemplateGroupId != nil {
		_, bare := splitTemplateOrg(*strippedTemplateGroupId, orgName)
		strippedTemplateGroupId = &bare
	}

	stackModel := &StackResourceModel{
		Id:              flatteners.String(apiResponse.Id),
		WorkflowGroupId: workflowGroupId,
		ResourceName:    flatteners.StringPtr(apiResponse.ResourceName),
		Description:     flatteners.StringPtr(apiResponse.Description),
		TemplateGroupId: flatteners.StringPtr(strippedTemplateGroupId),
	}

	if apiResponse.Tags != nil {
		tagsList, tagDiags := flatteners.ListOfStringToTerraformList(apiResponse.Tags)
		diags.Append(tagDiags...)
		if diags.HasError() {
			return nil, diags
		}
		stackModel.Tags = tagsList
	} else {
		stackModel.Tags = types.ListNull(types.StringType)
	}

	envVarsList, diagsEnv := flattenEnvironmentVariables(ctx, apiResponse.EnvironmentVariables)
	diags.Append(diagsEnv...)
	if diags.HasError() {
		return nil, diags
	}
	stackModel.EnvironmentVariables = envVarsList

	dpcList, diagsDpc := flattenDeploymentPlatformConfig(ctx, apiResponse.DeploymentPlatformConfig)
	diags.Append(diagsDpc...)
	if diags.HasError() {
		return nil, diags
	}
	stackModel.DeploymentPlatformConfig = dpcList

	defaultActions, customActions, diagsAct := flattenActionsMap(ctx, translateActionsOrderKeys(apiResponse.Actions, apiResponse.WorkflowRelationsMap), false)
	diags.Append(diagsAct...)
	if diags.HasError() {
		return nil, diags
	}
	stackModel.DefaultActions = defaultActions
	stackModel.CustomActions = customActions

	wfcObj, diagsWfc := flattenWorkflowsConfig(ctx, apiResponse.WorkflowsConfig)
	diags.Append(diagsWfc...)
	if diags.HasError() {
		return nil, diags
	}
	stackModel.WorkflowsConfig = wfcObj

	usList, diagsUs := flattenUserSchedules(ctx, apiResponse.UserSchedules)
	diags.Append(diagsUs...)
	if diags.HasError() {
		return nil, diags
	}
	stackModel.UserSchedules = usList

	ctMap, diagsCt := flattenContextTags(ctx, apiResponse.ContextTags)
	diags.Append(diagsCt...)
	if diags.HasError() {
		return nil, diags
	}
	stackModel.ContextTags = ctMap

	msObj, diagsMs := flattenMiniSteps(ctx, apiResponse.MiniSteps)
	diags.Append(diagsMs...)
	if diags.HasError() {
		return nil, diags
	}
	stackModel.MiniSteps = msObj

	return stackModel, diags
}

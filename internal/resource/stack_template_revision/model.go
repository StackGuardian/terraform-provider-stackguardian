package stacktemplaterevision

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/sg-sdk-go/stacktemplaterevisions"
	"github.com/StackGuardian/sg-sdk-go/stacktemplates"
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

type StackTemplateRevisionResourceModel struct {
	Id               types.String `tfsdk:"id"`
	ParentTemplateId types.String `tfsdk:"parent_template_id"`
	TemplateId       types.String `tfsdk:"template_id"`
	Alias            types.String `tfsdk:"alias"`
	Notes            types.String `tfsdk:"notes"`
	LongDescription  types.String `tfsdk:"description"`
	SourceConfigKind types.String `tfsdk:"source_config_kind"`
	IsActive         types.String `tfsdk:"is_active"`
	IsPublic         types.String `tfsdk:"is_public"`
	Tags             types.List   `tfsdk:"tags"`
	ContextTags      types.Map    `tfsdk:"context_tags"`
	Deprecation      types.Object `tfsdk:"deprecation"`
	WorkflowsConfig  types.Object `tfsdk:"workflows_config"`
	Actions          types.Map    `tfsdk:"actions"`
}

// ---------------------------------------------------------------------------
// Deprecation
// ---------------------------------------------------------------------------

type DeprecationModel struct {
	EffectiveDate types.String `tfsdk:"effective_date"`
	Message       types.String `tfsdk:"message"`
}

func (DeprecationModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"effective_date": types.StringType,
		"message":        types.StringType,
	}
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

type EnvVarModel struct {
	Config types.Object `tfsdk:"config"`
	Kind   types.String `tfsdk:"kind"`
}

func (EnvVarModel) AttributeTypes() map[string]attr.Type {
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
// Input schemas (for WorkflowsConfig workflows)
// ---------------------------------------------------------------------------

type StackInputSchemaModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Type         types.String `tfsdk:"type"`
	EncodedData  types.String `tfsdk:"encoded_data"`
	UiSchemaData types.String `tfsdk:"ui_schema_data"`
	IsCommitted  types.Bool   `tfsdk:"is_committed"`
}

func (StackInputSchemaModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":             types.StringType,
		"name":           types.StringType,
		"description":    types.StringType,
		"type":           types.StringType,
		"encoded_data":   types.StringType,
		"ui_schema_data": types.StringType,
		"is_committed":   types.BoolType,
	}
}

// ---------------------------------------------------------------------------
// MiniSteps
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
		"environment_variables": types.ListType{ElemType: types.ObjectType{AttrTypes: EnvVarModel{}.AttributeTypes()}},
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

type DeploymentPlatformConfigModel struct {
	Kind   types.String `tfsdk:"kind"`
	Config types.String `tfsdk:"config"` // JSON string (map[string]interface{})
}

func (DeploymentPlatformConfigModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"kind":   types.StringType,
		"config": types.StringType,
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

type UserSchedulesModel struct {
	Name  types.String `tfsdk:"name"`
	Desc  types.String `tfsdk:"desc"`
	Cron  types.String `tfsdk:"cron"`
	State types.String `tfsdk:"state"`
}

func (UserSchedulesModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":  types.StringType,
		"desc":  types.StringType,
		"cron":  types.StringType,
		"state": types.StringType,
	}
}

// ---------------------------------------------------------------------------
// VCS config (for WorkflowsConfigWorkflow)
// ---------------------------------------------------------------------------

type CustomSourceConfigModel struct {
	IsPrivate               types.Bool   `tfsdk:"is_private"`
	Auth                    types.String `tfsdk:"auth"`
	WorkingDir              types.String `tfsdk:"working_dir"`
	GitSparseCheckoutConfig types.String `tfsdk:"git_sparse_checkout_config"`
	GitCoreAutoCrlf         types.Bool   `tfsdk:"git_core_auto_crlf"`
	Ref                     types.String `tfsdk:"ref"`
	Repo                    types.String `tfsdk:"repo"`
	IncludeSubModule        types.Bool   `tfsdk:"include_sub_module"`
}

func (CustomSourceConfigModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"is_private":                 types.BoolType,
		"auth":                       types.StringType,
		"working_dir":                types.StringType,
		"git_sparse_checkout_config": types.StringType,
		"git_core_auto_crlf":         types.BoolType,
		"ref":                        types.StringType,
		"repo":                       types.StringType,
		"include_sub_module":         types.BoolType,
	}
}

type CustomSourceModel struct {
	SourceConfigDestKind types.String `tfsdk:"source_config_dest_kind"`
	Config               types.Object `tfsdk:"config"`
}

func (CustomSourceModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"source_config_dest_kind": types.StringType,
		"config":                  types.ObjectType{AttrTypes: CustomSourceConfigModel{}.AttributeTypes()},
	}
}

type IacVcsConfigModel struct {
	UseMarketplaceTemplate types.Bool   `tfsdk:"use_marketplace_template"`
	IacTemplateId          types.String `tfsdk:"iac_template_id"`
	CustomSource           types.Object `tfsdk:"custom_source"`
}

func (IacVcsConfigModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"use_marketplace_template": types.BoolType,
		"iac_template_id":          types.StringType,
		"custom_source":            types.ObjectType{AttrTypes: CustomSourceModel{}.AttributeTypes()},
	}
}

type IacInputDataModel struct {
	SchemaId   types.String `tfsdk:"schema_id"`
	SchemaType types.String `tfsdk:"schema_type"`
	Data       types.String `tfsdk:"data"` // JSON string
}

func (IacInputDataModel) AttributeTypes() map[string]attr.Type {
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
		"iac_input_data": types.ObjectType{AttrTypes: IacInputDataModel{}.AttributeTypes()},
	}
}

// ---------------------------------------------------------------------------
// WorkflowsConfig
// ---------------------------------------------------------------------------

type WorkflowInStackModel struct {
	Id                        types.String `tfsdk:"id"`
	TemplateId                types.String `tfsdk:"template_id"`
	ResourceName              types.String `tfsdk:"resource_name"`
	WfStepsConfig             types.List   `tfsdk:"wf_steps_config"`
	TerraformConfig           types.Object `tfsdk:"terraform_config"`
	EnvironmentVariables      types.List   `tfsdk:"environment_variables"`
	DeploymentPlatformConfig  types.List   `tfsdk:"deployment_platform_config"`
	VcsConfig                 types.Object `tfsdk:"vcs_config"`
	IacInputData              types.Object `tfsdk:"iac_input_data"`
	UserSchedules             types.List   `tfsdk:"user_schedules"`
	Approvers                 types.List   `tfsdk:"approvers"`
	NumberOfApprovalsRequired types.Int64  `tfsdk:"number_of_approvals_required"`
	RunnerConstraints         types.Object `tfsdk:"runner_constraints"`
	UserJobCpu                types.Int64  `tfsdk:"user_job_cpu"`
	UserJobMemory             types.Int64  `tfsdk:"user_job_memory"`
	InputSchemas              types.List   `tfsdk:"input_schemas"`
	MiniSteps                 types.Object `tfsdk:"mini_steps"`
}

func (WorkflowInStackModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                           types.StringType,
		"template_id":                  types.StringType,
		"resource_name":                types.StringType,
		"wf_steps_config":              types.ListType{ElemType: types.ObjectType{AttrTypes: WfStepsConfigModel{}.AttributeTypes()}},
		"terraform_config":             types.ObjectType{AttrTypes: TerraformConfigModel{}.AttributeTypes()},
		"environment_variables":        types.ListType{ElemType: types.ObjectType{AttrTypes: EnvVarModel{}.AttributeTypes()}},
		"deployment_platform_config":   types.ListType{ElemType: types.ObjectType{AttrTypes: DeploymentPlatformConfigModel{}.AttributeTypes()}},
		"vcs_config":                   types.ObjectType{AttrTypes: VcsConfigModel{}.AttributeTypes()},
		"iac_input_data":               types.ObjectType{AttrTypes: WfStepInputDataModel{}.AttributeTypes()},
		"user_schedules":               types.ListType{ElemType: types.ObjectType{AttrTypes: UserSchedulesModel{}.AttributeTypes()}},
		"approvers":                    types.ListType{ElemType: types.StringType},
		"number_of_approvals_required": types.Int64Type,
		"runner_constraints":           types.ObjectType{AttrTypes: RunnerConstraintsModel{}.AttributeTypes()},
		"user_job_cpu":                 types.Int64Type,
		"user_job_memory":              types.Int64Type,
		"input_schemas":                types.ListType{ElemType: types.ObjectType{AttrTypes: StackInputSchemaModel{}.AttributeTypes()}},
		"mini_steps":                   types.ObjectType{AttrTypes: MinistepsModel{}.AttributeTypes()},
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
		"environment_variables":      types.ListType{ElemType: types.ObjectType{AttrTypes: EnvVarModel{}.AttributeTypes()}},
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
	Default     types.Bool   `tfsdk:"default"`
	Order       types.Map    `tfsdk:"order"` // map[string]ActionOrderModel
}

func (ActionsModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":        types.StringType,
		"description": types.StringType,
		"default":     types.BoolType,
		"order":       types.MapType{ElemType: types.ObjectType{AttrTypes: ActionOrderModel{}.AttributeTypes()}},
	}
}

// ---------------------------------------------------------------------------
// Helper: parse JSON string to map[string]interface{}
// ---------------------------------------------------------------------------

func parseJSONToMap(s string) map[string]interface{} {
	if s == "" {
		return nil
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return nil
	}
	return result
}

func marshalToJSONString(v interface{}) types.String {
	if v == nil {
		return types.StringNull()
	}
	b, err := json.Marshal(v)
	if err != nil {
		return types.StringNull()
	}
	return flatteners.String(string(b))
}

// ---------------------------------------------------------------------------
// Converters: Terraform model → SDK API types
// ---------------------------------------------------------------------------

func convertEnvVarsToAPI(ctx context.Context, list types.List) ([]sgsdkgo.EnvVars, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var models []EnvVarModel
	if diags := list.ElementsAs(ctx, &models, false); diags.HasError() {
		return nil, diags
	}
	result := make([]sgsdkgo.EnvVars, len(models))
	for i, m := range models {
		var cfgModel EnvVarConfigModel
		if diags := m.Config.As(ctx, &cfgModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); diags.HasError() {
			return nil, diags
		}
		result[i] = sgsdkgo.EnvVars{
			Kind: sgsdkgo.EnvVarsKindEnum(m.Kind.ValueString()),
			Config: &sgsdkgo.EnvVarConfig{
				VarName:   cfgModel.VarName.ValueString(),
				SecretId:  cfgModel.SecretId.ValueStringPointer(),
				TextValue: cfgModel.TextValue.ValueStringPointer(),
			},
		}
	}
	return result, nil
}

func convertEnvVarPointersToAPI(ctx context.Context, list types.List) ([]*sgsdkgo.EnvVars, diag.Diagnostics) {
	vals, diags := convertEnvVarsToAPI(ctx, list)
	if diags.HasError() {
		return nil, diags
	}
	if vals == nil {
		return nil, nil
	}
	result := make([]*sgsdkgo.EnvVars, len(vals))
	for i := range vals {
		result[i] = &vals[i]
	}
	return result, nil
}

func convertMountPointsToAPI(ctx context.Context, list types.List) ([]sgsdkgo.MountPoint, diag.Diagnostics) {
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

func convertWfStepsConfigToAPI(ctx context.Context, list types.List) ([]sgsdkgo.WfStepsConfig, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var models []WfStepsConfigModel
	if diags := list.ElementsAs(ctx, &models, false); diags.HasError() {
		return nil, diags
	}
	result := make([]sgsdkgo.WfStepsConfig, len(models))
	for i, m := range models {
		step := sgsdkgo.WfStepsConfig{
			Name:             m.Name.ValueString(),
			Approval:         m.Approval.ValueBoolPointer(),
			Timeout:          expanders.IntPtr(m.Timeout.ValueInt64Pointer()),
			WfStepTemplateId: m.WfStepTemplateId.ValueStringPointer(),
			CmdOverride:      m.CmdOverride.ValueStringPointer(),
		}
		if !m.EnvironmentVariables.IsNull() && !m.EnvironmentVariables.IsUnknown() {
			envVars, diags := convertEnvVarsToAPI(ctx, m.EnvironmentVariables)
			if diags.HasError() {
				return nil, diags
			}
			step.EnvironmentVariables = envVars
		}
		if !m.MountPoints.IsNull() && !m.MountPoints.IsUnknown() {
			mps, diags := convertMountPointsToAPI(ctx, m.MountPoints)
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
			step.WfStepInputData = &sgsdkgo.WfStepInputData{
				SchemaType: sgsdkgo.WfStepInputDataSchemaTypeEnum(idm.SchemaType.ValueString()),
				Data:       parseJSONToMap(idm.Data.ValueString()),
			}
		}
		result[i] = step
	}
	return result, nil
}

func convertWfStepsConfigPointersToAPI(ctx context.Context, list types.List) ([]*sgsdkgo.WfStepsConfig, diag.Diagnostics) {
	vals, diags := convertWfStepsConfigToAPI(ctx, list)
	if diags.HasError() {
		return nil, diags
	}
	if vals == nil {
		return nil, nil
	}
	result := make([]*sgsdkgo.WfStepsConfig, len(vals))
	for i := range vals {
		result[i] = &vals[i]
	}
	return result, nil
}

func convertTerraformConfigToAPI(ctx context.Context, obj types.Object) (*sgsdkgo.TerraformConfig, diag.Diagnostics) {
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
		mps, diags := convertMountPointsToAPI(ctx, m.TerraformBinPath)
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
			steps, diags := convertWfStepsConfigToAPI(ctx, *pair.list)
			if diags.HasError() {
				return nil, diags
			}
			*pair.dest = steps
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

func convertDeploymentPlatformConfigToAPI(ctx context.Context, list types.List) ([]*sgsdkgo.DeploymentPlatformConfig, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var models []DeploymentPlatformConfigModel
	if diags := list.ElementsAs(ctx, &models, false); diags.HasError() {
		return nil, diags
	}
	result := make([]*sgsdkgo.DeploymentPlatformConfig, len(models))
	for i, m := range models {
		dpc := &sgsdkgo.DeploymentPlatformConfig{
			Kind:   sgsdkgo.DeploymentPlatformConfigKindEnum(m.Kind.ValueString()),
			Config: parseJSONToMap(m.Config.ValueString()),
		}
		result[i] = dpc
	}
	return result, nil
}

func convertRunnerConstraintsToAPI(ctx context.Context, obj types.Object) (*sgsdkgo.RunnerConstraints, diag.Diagnostics) {
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
		Type:  sgsdkgo.RunnerConstraintsTypeEnum(m.Type.ValueString()),
		Names: names,
	}, nil
}

func convertUserSchedulesToAPI(ctx context.Context, list types.List) ([]*sgsdkgo.UserSchedules, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var models []UserSchedulesModel
	if diags := list.ElementsAs(ctx, &models, false); diags.HasError() {
		return nil, diags
	}
	result := make([]*sgsdkgo.UserSchedules, len(models))
	for i, m := range models {
		us := &sgsdkgo.UserSchedules{
			Name:  m.Name.ValueStringPointer(),
			Desc:  m.Desc.ValueStringPointer(),
			Cron:  m.Cron.ValueString(),
			State: sgsdkgo.StateEnum(m.State.ValueString()),
		}
		result[i] = us
	}
	return result, nil
}

func convertVcsConfigToAPI(ctx context.Context, obj types.Object, orgName string) (*sgsdkgo.VcsConfig, diag.Diagnostics) {
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
		var prefixedIacTemplateId *string
		if tid := iacModel.IacTemplateId.ValueStringPointer(); tid != nil {
			v := fmt.Sprintf("/%s/%s", orgName, *tid)
			prefixedIacTemplateId = &v
		}
		iacVcs := &sgsdkgo.IacvcsConfig{
			UseMarketplaceTemplate: iacModel.UseMarketplaceTemplate.ValueBool(),
			IacTemplateId:          prefixedIacTemplateId,
		}
		if !iacModel.CustomSource.IsNull() && !iacModel.CustomSource.IsUnknown() {
			var csModel CustomSourceModel
			if diags := iacModel.CustomSource.As(ctx, &csModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); diags.HasError() {
				return nil, diags
			}
			cs := &sgsdkgo.CustomSource{
				SourceConfigDestKind: sgsdkgo.CustomSourceSourceConfigDestKindEnum(csModel.SourceConfigDestKind.ValueString()),
			}
			if !csModel.Config.IsNull() && !csModel.Config.IsUnknown() {
				var csCfgModel CustomSourceConfigModel
				if diags := csModel.Config.As(ctx, &csCfgModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); diags.HasError() {
					return nil, diags
				}
				cs.Config = &sgsdkgo.CustomSourceConfig{
					IsPrivate:               csCfgModel.IsPrivate.ValueBoolPointer(),
					Auth:                    csCfgModel.Auth.ValueStringPointer(),
					WorkingDir:              csCfgModel.WorkingDir.ValueStringPointer(),
					GitSparseCheckoutConfig: csCfgModel.GitSparseCheckoutConfig.ValueStringPointer(),
					GitCoreAutoCrlf:         csCfgModel.GitCoreAutoCrlf.ValueBoolPointer(),
					Ref:                     csCfgModel.Ref.ValueStringPointer(),
					Repo:                    csCfgModel.Repo.ValueStringPointer(),
					IncludeSubModule:        csCfgModel.IncludeSubModule.ValueBoolPointer(),
				}
			}
			iacVcs.CustomSource = cs
		}
		vcsConfig.IacVcsConfig = iacVcs
	}

	if !m.IacInputData.IsNull() && !m.IacInputData.IsUnknown() {
		var idModel IacInputDataModel
		if diags := m.IacInputData.As(ctx, &idModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); diags.HasError() {
			return nil, diags
		}
		vcsConfig.IacInputData = &sgsdkgo.IacInputData{
			SchemaId:   idModel.SchemaId.ValueStringPointer(),
			SchemaType: sgsdkgo.IacInputDataSchemaTypeEnum(idModel.SchemaType.ValueString()),
			Data:       parseJSONToMap(idModel.Data.ValueString()),
		}
	}

	return vcsConfig, nil
}

// ---------------------------------------------------------------------------
// MiniSteps converters (Terraform → SDK)
// ---------------------------------------------------------------------------

func convertNotificationRecipientsToAPI(ctx context.Context, list types.List) ([]*sgsdkgo.Notifications, diag.Diagnostics) {
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

func convertWebhooksToAPI(ctx context.Context, list types.List) ([]map[string]interface{}, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var models []MinistepsWebhooksModel
	if diags := list.ElementsAs(ctx, &models, false); diags.HasError() {
		return nil, diags
	}
	result := make([]map[string]interface{}, len(models))
	for i, m := range models {
		wh := map[string]interface{}{
			"webhookName": m.WebhookName.ValueString(),
			"webhookUrl":  m.WebhookUrl.ValueString(),
		}
		if !m.WebhookSecret.IsNull() && !m.WebhookSecret.IsUnknown() {
			wh["webhookSecret"] = m.WebhookSecret.ValueString()
		}
		result[i] = wh
	}
	return result, nil
}

func convertWfChainingToAPI(ctx context.Context, list types.List) ([]*sgsdkgo.MiniSteps, diag.Diagnostics) {
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
			ms.WorkflowRunPayload = parseJSONToMap(m.WorkflowRunPayload.ValueString())
		}
		if !m.StackRunPayload.IsNull() && !m.StackRunPayload.IsUnknown() {
			ms.StackRunPayload = parseJSONToMap(m.StackRunPayload.ValueString())
		}
		result[i] = ms
	}
	return result, nil
}

func convertMiniStepsToAPI(ctx context.Context, obj types.Object) (*sgsdkgo.MiniStepsSchema, diag.Diagnostics) {
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
					vals, diags := convertNotificationRecipientsToAPI(ctx, *pair.list)
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
			dest *[]map[string]interface{}
		}{
			{&whModel.ApprovalRequired, &wh.ApprovalRequired},
			{&whModel.Cancelled, &wh.Cancelled},
			{&whModel.Completed, &wh.Completed},
			{&whModel.DriftDetected, &wh.DriftDetected},
			{&whModel.Errored, &wh.Errored},
		} {
			if !pair.list.IsNull() && !pair.list.IsUnknown() {
				vals, diags := convertWebhooksToAPI(ctx, *pair.list)
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
				vals, diags := convertWfChainingToAPI(ctx, *pair.list)
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

func convertNotificationRecipientsFromAPI(ctx context.Context, recipients []*sgsdkgo.Notifications) (types.List, diag.Diagnostics) {
	nullObj := types.ListNull(types.ObjectType{AttrTypes: MinistepsNotificationRecipientsModel{}.AttributeTypes()})
	if recipients == nil {
		return nullObj, nil
	}
	elems := make([]MinistepsNotificationRecipientsModel, 0, len(recipients))
	for _, r := range recipients {
		if r == nil {
			continue
		}
		recList, diags := types.ListValueFrom(ctx, types.StringType, r.Recipients)
		if diags.HasError() {
			return nullObj, diags
		}
		elems = append(elems, MinistepsNotificationRecipientsModel{Recipients: recList})
	}
	obj, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: MinistepsNotificationRecipientsModel{}.AttributeTypes()}, elems)
	if diags.HasError() {
		return nullObj, diags
	}
	return obj, nil
}

func convertWebhooksFromAPI(ctx context.Context, webhooks []map[string]interface{}) (types.List, diag.Diagnostics) {
	nullObj := types.ListNull(types.ObjectType{AttrTypes: MinistepsWebhooksModel{}.AttributeTypes()})
	if webhooks == nil {
		return nullObj, nil
	}
	elems := make([]MinistepsWebhooksModel, 0, len(webhooks))
	for _, wh := range webhooks {
		name, _ := wh["webhookName"].(string)
		url, _ := wh["webhookUrl"].(string)
		secret, _ := wh["webhookSecret"].(string)
		m := MinistepsWebhooksModel{
			WebhookName:   flatteners.String(name),
			WebhookUrl:    flatteners.String(url),
			WebhookSecret: flatteners.String(secret),
		}
		elems = append(elems, m)
	}
	obj, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: MinistepsWebhooksModel{}.AttributeTypes()}, elems)
	if diags.HasError() {
		return nullObj, diags
	}
	return obj, nil
}

func convertWfChainingFromAPI(ctx context.Context, items []*sgsdkgo.MiniSteps) (types.List, diag.Diagnostics) {
	nullObj := types.ListNull(types.ObjectType{AttrTypes: MinistepsWorkflowChainingModel{}.AttributeTypes()})
	if items == nil {
		return nullObj, nil
	}
	elems := make([]MinistepsWorkflowChainingModel, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		m := MinistepsWorkflowChainingModel{
			WorkflowGroupId:    flatteners.String(item.WorkflowGroupId),
			StackId:            flatteners.StringPtr(item.StackId),
			StackRunPayload:    marshalToJSONString(item.StackRunPayload),
			WorkflowId:         flatteners.StringPtr(item.WorkflowId),
			WorkflowRunPayload: marshalToJSONString(item.WorkflowRunPayload),
		}
		elems = append(elems, m)
	}
	obj, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: MinistepsWorkflowChainingModel{}.AttributeTypes()}, elems)
	if diags.HasError() {
		return nullObj, diags
	}
	return obj, nil
}

func miniStepsFromAPI(ctx context.Context, miniSteps *sgsdkgo.MiniStepsSchema) (types.Object, diag.Diagnostics) {
	nullObj := types.ObjectNull(MinistepsModel{}.AttributeTypes())
	if miniSteps == nil {
		return nullObj, nil
	}

	var diags diag.Diagnostics
	msModel := MinistepsModel{}

	// Notifications
	if miniSteps.Notifications != nil {
		notifModel := MinistepsNotificationsModel{}
		if miniSteps.Notifications.Email != nil {
			emailModel := MinistepsEmailModel{}
			emailModel.ApprovalRequired, diags = convertNotificationRecipientsFromAPI(ctx, miniSteps.Notifications.Email.ApprovalRequired)
			if diags.HasError() {
				return nullObj, diags
			}
			emailModel.Cancelled, diags = convertNotificationRecipientsFromAPI(ctx, miniSteps.Notifications.Email.Cancelled)
			if diags.HasError() {
				return nullObj, diags
			}
			emailModel.Completed, diags = convertNotificationRecipientsFromAPI(ctx, miniSteps.Notifications.Email.Completed)
			if diags.HasError() {
				return nullObj, diags
			}
			emailModel.DriftDetected, diags = convertNotificationRecipientsFromAPI(ctx, miniSteps.Notifications.Email.DriftDetected)
			if diags.HasError() {
				return nullObj, diags
			}
			emailModel.Errored, diags = convertNotificationRecipientsFromAPI(ctx, miniSteps.Notifications.Email.Errored)
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

	// Webhooks
	if miniSteps.Webhooks != nil {
		whModel := MinistepsWebhooksContainerModel{}
		whModel.ApprovalRequired, diags = convertWebhooksFromAPI(ctx, miniSteps.Webhooks.ApprovalRequired)
		if diags.HasError() {
			return nullObj, diags
		}
		whModel.Cancelled, diags = convertWebhooksFromAPI(ctx, miniSteps.Webhooks.Cancelled)
		if diags.HasError() {
			return nullObj, diags
		}
		whModel.Completed, diags = convertWebhooksFromAPI(ctx, miniSteps.Webhooks.Completed)
		if diags.HasError() {
			return nullObj, diags
		}
		whModel.DriftDetected, diags = convertWebhooksFromAPI(ctx, miniSteps.Webhooks.DriftDetected)
		if diags.HasError() {
			return nullObj, diags
		}
		whModel.Errored, diags = convertWebhooksFromAPI(ctx, miniSteps.Webhooks.Errored)
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

	// WfChaining
	if miniSteps.WfChaining != nil {
		wcModel := MinistepsWfChainingContainerModel{}
		wcModel.Completed, diags = convertWfChainingFromAPI(ctx, miniSteps.WfChaining.Completed)
		if diags.HasError() {
			return nullObj, diags
		}
		wcModel.Errored, diags = convertWfChainingFromAPI(ctx, miniSteps.WfChaining.Errored)
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

	return types.ObjectValueFrom(ctx, MinistepsModel{}.AttributeTypes(), msModel)
}

func convertWorkflowsConfigToAPI(ctx context.Context, obj types.Object, orgName string) (*stacktemplaterevisions.StackTemplateRevisionWorkflowsConfig, diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil
	}
	var m WorkflowsConfigModel
	if diags := obj.As(ctx, &m, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); diags.HasError() {
		return nil, diags
	}
	if m.Workflows.IsNull() || m.Workflows.IsUnknown() {
		return &stacktemplaterevisions.StackTemplateRevisionWorkflowsConfig{}, nil
	}
	var wfModels []WorkflowInStackModel
	if diags := m.Workflows.ElementsAs(ctx, &wfModels, false); diags.HasError() {
		return nil, diags
	}

	workflows := make([]*stacktemplaterevisions.StackTemplateRevisionWorkflow, len(wfModels))
	for i, wm := range wfModels {
		// Prefix template_id with /<orgName>/
		var prefixedTemplateId *string
		if tid := wm.TemplateId.ValueStringPointer(); tid != nil {
			v := fmt.Sprintf("/%s/%s", orgName, *tid)
			prefixedTemplateId = &v
		}

		wf := &stacktemplaterevisions.StackTemplateRevisionWorkflow{
			Id:                        wm.Id.ValueStringPointer(),
			TemplateId:                prefixedTemplateId,
			ResourceName:              wm.ResourceName.ValueStringPointer(),
			NumberOfApprovalsRequired: expanders.IntPtr(wm.NumberOfApprovalsRequired.ValueInt64Pointer()),
			UserJobCpu:                expanders.IntPtr(wm.UserJobCpu.ValueInt64Pointer()),
			UserJobMemory:             expanders.IntPtr(wm.UserJobMemory.ValueInt64Pointer()),
		}
		if !wm.WfStepsConfig.IsNull() && !wm.WfStepsConfig.IsUnknown() {
			steps, diags := convertWfStepsConfigPointersToAPI(ctx, wm.WfStepsConfig)
			if diags.HasError() {
				return nil, diags
			}
			wf.WfStepsConfig = steps
		}
		if !wm.TerraformConfig.IsNull() && !wm.TerraformConfig.IsUnknown() {
			tc, diags := convertTerraformConfigToAPI(ctx, wm.TerraformConfig)
			if diags.HasError() {
				return nil, diags
			}
			wf.TerraformConfig = tc
		}
		if !wm.EnvironmentVariables.IsNull() && !wm.EnvironmentVariables.IsUnknown() {
			envVars, diags := convertEnvVarPointersToAPI(ctx, wm.EnvironmentVariables)
			if diags.HasError() {
				return nil, diags
			}
			wf.EnvironmentVariables = envVars
		}
		if !wm.DeploymentPlatformConfig.IsNull() && !wm.DeploymentPlatformConfig.IsUnknown() {
			dpcs, diags := convertDeploymentPlatformConfigToAPI(ctx, wm.DeploymentPlatformConfig)
			if diags.HasError() {
				return nil, diags
			}
			wf.DeploymentPlatformConfig = dpcs
		}
		if !wm.VcsConfig.IsNull() && !wm.VcsConfig.IsUnknown() {
			vcs, diags := convertVcsConfigToAPI(ctx, wm.VcsConfig, orgName)
			if diags.HasError() {
				return nil, diags
			}
			wf.VcsConfig = vcs
		}
		if !wm.IacInputData.IsNull() && !wm.IacInputData.IsUnknown() {
			var idm WfStepInputDataModel
			if diags := wm.IacInputData.As(ctx, &idm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); diags.HasError() {
				return nil, diags
			}
			wf.IacInputData = &sgsdkgo.TemplatesIacInputData{
				SchemaType: idm.SchemaType.ValueString(),
				Data:       parseJSONToMap(idm.Data.ValueString()),
			}
		}
		if !wm.UserSchedules.IsNull() && !wm.UserSchedules.IsUnknown() {
			us, diags := convertUserSchedulesToAPI(ctx, wm.UserSchedules)
			if diags.HasError() {
				return nil, diags
			}
			wf.UserSchedules = us
		}
		if !wm.Approvers.IsNull() && !wm.Approvers.IsUnknown() {
			approvers, diags := expanders.StringList(ctx, wm.Approvers)
			if diags.HasError() {
				return nil, diags
			}
			wf.Approvers = approvers
		}
		if !wm.RunnerConstraints.IsNull() && !wm.RunnerConstraints.IsUnknown() {
			rc, diags := convertRunnerConstraintsToAPI(ctx, wm.RunnerConstraints)
			if diags.HasError() {
				return nil, diags
			}
			wf.RunnerConstraints = rc
		}
		if !wm.InputSchemas.IsNull() && !wm.InputSchemas.IsUnknown() {
			var isModels []StackInputSchemaModel
			if diags := wm.InputSchemas.ElementsAs(ctx, &isModels, false); diags.HasError() {
				return nil, diags
			}
			schemas := make([]*sgsdkgo.InputSchemas, len(isModels))
			for j, ism := range isModels {
				schemas[j] = &sgsdkgo.InputSchemas{
					Id:           ism.Id.ValueStringPointer(),
					Name:         ism.Name.ValueStringPointer(),
					Description:  ism.Description.ValueStringPointer(),
					Type:         sgsdkgo.InputSchemasTypeEnum(ism.Type.ValueString()),
					EncodedData:  ism.EncodedData.ValueStringPointer(),
					UiSchemaData: ism.UiSchemaData.ValueStringPointer(),
					IsCommitted:  ism.IsCommitted.ValueBoolPointer(),
				}
			}
			wf.InputSchemas = schemas
		}
		if !wm.MiniSteps.IsNull() && !wm.MiniSteps.IsUnknown() {
			ms, diags := convertMiniStepsToAPI(ctx, wm.MiniSteps)
			if diags.HasError() {
				return nil, diags
			}
			wf.MiniSteps = ms
		}
		workflows[i] = wf
	}
	return &stacktemplaterevisions.StackTemplateRevisionWorkflowsConfig{Workflows: workflows}, nil
}

func convertActionsToAPI(ctx context.Context, actionsMap types.Map) (map[string]*sgsdkgo.Actions, diag.Diagnostics) {
	if actionsMap.IsNull() || actionsMap.IsUnknown() {
		return nil, nil
	}
	var models map[string]ActionsModel
	if diags := actionsMap.ElementsAs(ctx, &models, false); diags.HasError() {
		return nil, diags
	}
	result := make(map[string]*sgsdkgo.Actions, len(models))
	for k, am := range models {
		action := &sgsdkgo.Actions{
			Name:        am.Name.ValueString(),
			Description: am.Description.ValueStringPointer(),
			Default:     am.Default.ValueBoolPointer(),
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
					if diags := om.Parameters.As(ctx, &pm, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); diags.HasError() {
						return nil, diags
					}
					params := &sgsdkgo.StackActionParameters{}
					if !pm.TerraformAction.IsNull() && !pm.TerraformAction.IsUnknown() {
						var tam TerraformActionModel
						if diags := pm.TerraformAction.As(ctx, &tam, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); diags.HasError() {
							return nil, diags
						}
						actionEnum := sgsdkgo.ActionEnum(tam.Action.ValueString())
						params.TerraformAction = &sgsdkgo.TerraformAction{Action: &actionEnum}
					}
					if !pm.DeploymentPlatformConfig.IsNull() && !pm.DeploymentPlatformConfig.IsUnknown() {
						dpcs, diags := convertDeploymentPlatformConfigToAPI(ctx, pm.DeploymentPlatformConfig)
						if diags.HasError() {
							return nil, diags
						}
						params.DeploymentPlatformConfig = dpcs
					}
					if !pm.WfStepsConfig.IsNull() && !pm.WfStepsConfig.IsUnknown() {
						steps, diags := convertWfStepsConfigPointersToAPI(ctx, pm.WfStepsConfig)
						if diags.HasError() {
							return nil, diags
						}
						params.WfStepsConfig = steps
					}
					if !pm.EnvironmentVariables.IsNull() && !pm.EnvironmentVariables.IsUnknown() {
						envVars, diags := convertEnvVarPointersToAPI(ctx, pm.EnvironmentVariables)
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
							if diags := dm.Condition.As(ctx, &cond, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); diags.HasError() {
								return nil, diags
							}
							dep.Condition = &sgsdkgo.ActionDependencyCondition{LatestStatus: cond.LatestStatus.ValueString()}
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

// ---------------------------------------------------------------------------
// Converters: SDK API types → Terraform model
// ---------------------------------------------------------------------------

func envVarsFromAPI(envVars []sgsdkgo.EnvVars) (types.List, diag.Diagnostics) {
	if envVars == nil {
		return types.ListNull(types.ObjectType{AttrTypes: EnvVarModel{}.AttributeTypes()}), nil
	}
	elements := make([]attr.Value, len(envVars))
	for i, ev := range envVars {
		cfgModel := EnvVarConfigModel{}
		if ev.Config != nil {
			cfgModel.VarName = flatteners.String(ev.Config.VarName)
			cfgModel.SecretId = flatteners.StringPtr(ev.Config.SecretId)
			cfgModel.TextValue = flatteners.StringPtr(ev.Config.TextValue)
		}
		cfgObj, diags := types.ObjectValueFrom(context.Background(), EnvVarConfigModel{}.AttributeTypes(), cfgModel)
		if diags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: EnvVarModel{}.AttributeTypes()}), diags
		}
		evModel := EnvVarModel{
			Config: cfgObj,
			Kind:   flatteners.String(string(ev.Kind)),
		}
		obj, diags := types.ObjectValueFrom(context.Background(), EnvVarModel{}.AttributeTypes(), evModel)
		if diags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: EnvVarModel{}.AttributeTypes()}), diags
		}
		elements[i] = obj
	}
	return types.ListValue(types.ObjectType{AttrTypes: EnvVarModel{}.AttributeTypes()}, elements)
}

func envVarPointersFromAPI(envVars []*sgsdkgo.EnvVars) (types.List, diag.Diagnostics) {
	vals := make([]sgsdkgo.EnvVars, 0, len(envVars))
	for _, p := range envVars {
		if p != nil {
			vals = append(vals, *p)
		}
	}
	return envVarsFromAPI(vals)
}

func mountPointsFromAPI(mps []sgsdkgo.MountPoint) (types.List, diag.Diagnostics) {
	if mps == nil {
		return types.ListNull(types.ObjectType{AttrTypes: MountPointModel{}.AttributeTypes()}), nil
	}
	elements := make([]attr.Value, len(mps))
	for i, mp := range mps {
		m := MountPointModel{
			Source:   flatteners.String(mp.Source),
			Target:   flatteners.String(mp.Target),
			ReadOnly: flatteners.BoolPtr(mp.ReadOnly),
		}
		obj, diags := types.ObjectValueFrom(context.Background(), MountPointModel{}.AttributeTypes(), m)
		if diags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: MountPointModel{}.AttributeTypes()}), diags
		}
		elements[i] = obj
	}
	return types.ListValue(types.ObjectType{AttrTypes: MountPointModel{}.AttributeTypes()}, elements)
}

func wfStepsConfigFromAPI(steps []sgsdkgo.WfStepsConfig) (types.List, diag.Diagnostics) {
	listNull := types.ListNull(types.ObjectType{AttrTypes: WfStepsConfigModel{}.AttributeTypes()})
	if steps == nil {
		return listNull, nil
	}
	elements := make([]attr.Value, len(steps))
	for i, s := range steps {
		envList, diags := envVarsFromAPI(s.EnvironmentVariables)
		if diags.HasError() {
			return listNull, diags
		}
		mpList, diags := mountPointsFromAPI(s.MountPoints)
		if diags.HasError() {
			return listNull, diags
		}
		inputDataObj := types.ObjectNull(WfStepInputDataModel{}.AttributeTypes())
		if s.WfStepInputData != nil {
			dataStr := marshalToJSONString(s.WfStepInputData.Data)
			idm := WfStepInputDataModel{
				SchemaType: flatteners.String(string(s.WfStepInputData.SchemaType)),
				Data:       dataStr,
			}
			var diags2 diag.Diagnostics
			inputDataObj, diags2 = types.ObjectValueFrom(context.Background(), WfStepInputDataModel{}.AttributeTypes(), idm)
			if diags2.HasError() {
				return listNull, diags2
			}
		}
		m := WfStepsConfigModel{
			Name:                 flatteners.String(s.Name),
			EnvironmentVariables: envList,
			Approval:             flatteners.BoolPtr(s.Approval),
			Timeout:              flatteners.Int64Ptr(s.Timeout),
			CmdOverride:          flatteners.StringPtr(s.CmdOverride),
			MountPoints:          mpList,
			WfStepTemplateId:     flatteners.StringPtr(s.WfStepTemplateId),
			WfStepInputData:      inputDataObj,
		}
		obj, diags2 := types.ObjectValueFrom(context.Background(), WfStepsConfigModel{}.AttributeTypes(), m)
		if diags2.HasError() {
			return listNull, diags2
		}
		elements[i] = obj
	}
	return types.ListValue(types.ObjectType{AttrTypes: WfStepsConfigModel{}.AttributeTypes()}, elements)
}

func wfStepsConfigPointersFromAPI(steps []*sgsdkgo.WfStepsConfig) (types.List, diag.Diagnostics) {
	vals := make([]sgsdkgo.WfStepsConfig, 0, len(steps))
	for _, p := range steps {
		if p != nil {
			vals = append(vals, *p)
		}
	}
	return wfStepsConfigFromAPI(vals)
}

func terraformConfigFromAPI(tc *sgsdkgo.TerraformConfig) (types.Object, diag.Diagnostics) {
	nullObj := types.ObjectNull(TerraformConfigModel{}.AttributeTypes())
	if tc == nil {
		return nullObj, nil
	}
	strListNull := types.ListNull(types.StringType)

	binPath, diags := mountPointsFromAPI(tc.TerraformBinPath)
	if diags.HasError() {
		return nullObj, diags
	}

	makeWfStepsList := func(steps []sgsdkgo.WfStepsConfig) (types.List, diag.Diagnostics) {
		return wfStepsConfigFromAPI(steps)
	}

	postApply, diags := makeWfStepsList(tc.PostApplyWfStepsConfig)
	if diags.HasError() {
		return nullObj, diags
	}
	preApply, diags := makeWfStepsList(tc.PreApplyWfStepsConfig)
	if diags.HasError() {
		return nullObj, diags
	}
	prePlan, diags := makeWfStepsList(tc.PrePlanWfStepsConfig)
	if diags.HasError() {
		return nullObj, diags
	}
	postPlan, diags := makeWfStepsList(tc.PostPlanWfStepsConfig)
	if diags.HasError() {
		return nullObj, diags
	}

	makeStringList := func(hooks []string) (types.List, diag.Diagnostics) {
		if hooks == nil {
			return strListNull, nil
		}
		elems := make([]attr.Value, len(hooks))
		for i, h := range hooks {
			elems[i] = flatteners.String(h)
		}
		return types.ListValue(types.StringType, elems)
	}

	preInit, diags := makeStringList(tc.PreInitHooks)
	if diags.HasError() {
		return nullObj, diags
	}
	prePlan2, diags := makeStringList(tc.PrePlanHooks)
	if diags.HasError() {
		return nullObj, diags
	}
	postPlan2, diags := makeStringList(tc.PostPlanHooks)
	if diags.HasError() {
		return nullObj, diags
	}
	preApply2, diags := makeStringList(tc.PreApplyHooks)
	if diags.HasError() {
		return nullObj, diags
	}
	postApply2, diags := makeStringList(tc.PostApplyHooks)
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
		PrePlanHooks:           prePlan2,
		PostPlanHooks:          postPlan2,
		PreApplyHooks:          preApply2,
		PostApplyHooks:         postApply2,
		RunPreInitHooksOnDrift: flatteners.BoolPtr(tc.RunPreInitHooksOnDrift),
	}
	return types.ObjectValueFrom(context.Background(), TerraformConfigModel{}.AttributeTypes(), m)
}

func deploymentPlatformConfigFromAPI(dpcs []*sgsdkgo.DeploymentPlatformConfig) (types.List, diag.Diagnostics) {
	listNull := types.ListNull(types.ObjectType{AttrTypes: DeploymentPlatformConfigModel{}.AttributeTypes()})
	if dpcs == nil {
		return listNull, nil
	}
	elements := make([]attr.Value, 0, len(dpcs))
	for _, dpc := range dpcs {
		if dpc == nil {
			continue
		}
		m := DeploymentPlatformConfigModel{
			Kind:   flatteners.String(string(dpc.Kind)),
			Config: marshalToJSONString(dpc.Config),
		}
		obj, diags := types.ObjectValueFrom(context.Background(), DeploymentPlatformConfigModel{}.AttributeTypes(), m)
		if diags.HasError() {
			return listNull, diags
		}
		elements = append(elements, obj)
	}
	return types.ListValue(types.ObjectType{AttrTypes: DeploymentPlatformConfigModel{}.AttributeTypes()}, elements)
}

func runnerConstraintsFromAPI(rc *sgsdkgo.RunnerConstraints) (types.Object, diag.Diagnostics) {
	nullObj := types.ObjectNull(RunnerConstraintsModel{}.AttributeTypes())
	if rc == nil {
		return nullObj, nil
	}
	namesList, diags := types.ListValueFrom(context.Background(), types.StringType, rc.Names)
	if diags.HasError() {
		return nullObj, diags
	}
	m := RunnerConstraintsModel{
		Type:  flatteners.String(string(rc.Type)),
		Names: namesList,
	}
	return types.ObjectValueFrom(context.Background(), RunnerConstraintsModel{}.AttributeTypes(), m)
}

func userSchedulesFromAPI(uss []*sgsdkgo.UserSchedules) (types.List, diag.Diagnostics) {
	listNull := types.ListNull(types.ObjectType{AttrTypes: UserSchedulesModel{}.AttributeTypes()})
	if uss == nil {
		return listNull, nil
	}
	elements := make([]attr.Value, 0, len(uss))
	for _, us := range uss {
		if us == nil {
			continue
		}
		m := UserSchedulesModel{
			Name:  flatteners.StringPtr(us.Name),
			Desc:  flatteners.StringPtr(us.Desc),
			Cron:  flatteners.String(us.Cron),
			State: flatteners.String(string(us.State)),
		}
		obj, diags := types.ObjectValueFrom(context.Background(), UserSchedulesModel{}.AttributeTypes(), m)
		if diags.HasError() {
			return listNull, diags
		}
		elements = append(elements, obj)
	}
	return types.ListValue(types.ObjectType{AttrTypes: UserSchedulesModel{}.AttributeTypes()}, elements)
}

func vcsConfigFromAPI(vc *sgsdkgo.VcsConfig) (types.Object, diag.Diagnostics) {
	nullObj := types.ObjectNull(VcsConfigModel{}.AttributeTypes())
	if vc == nil {
		return nullObj, nil
	}
	m := VcsConfigModel{}

	iacVcsNull := types.ObjectNull(IacVcsConfigModel{}.AttributeTypes())
	if vc.IacVcsConfig != nil {
		var strippedIacTemplateId *string
		if vc.IacVcsConfig.IacTemplateId != nil {
			parts := strings.Split(*vc.IacVcsConfig.IacTemplateId, "/")
			base := parts[len(parts)-1]
			strippedIacTemplateId = &base
		}
		iacM := IacVcsConfigModel{
			UseMarketplaceTemplate: types.BoolValue(vc.IacVcsConfig.UseMarketplaceTemplate),
			IacTemplateId:          flatteners.StringPtr(strippedIacTemplateId),
			CustomSource:           types.ObjectNull(CustomSourceModel{}.AttributeTypes()),
		}
		if vc.IacVcsConfig.CustomSource != nil {
			cs := vc.IacVcsConfig.CustomSource
			csM := CustomSourceModel{
				SourceConfigDestKind: flatteners.String(string(cs.SourceConfigDestKind)),
				Config:               types.ObjectNull(CustomSourceConfigModel{}.AttributeTypes()),
			}
			if cs.Config != nil {
				csCfgM := CustomSourceConfigModel{
					IsPrivate:               flatteners.BoolPtr(cs.Config.IsPrivate),
					Auth:                    flatteners.StringPtr(cs.Config.Auth),
					WorkingDir:              flatteners.StringPtr(cs.Config.WorkingDir),
					GitSparseCheckoutConfig: flatteners.StringPtr(cs.Config.GitSparseCheckoutConfig),
					GitCoreAutoCrlf:         flatteners.BoolPtr(cs.Config.GitCoreAutoCrlf),
					Ref:                     flatteners.StringPtr(cs.Config.Ref),
					Repo:                    flatteners.StringPtr(cs.Config.Repo),
					IncludeSubModule:        flatteners.BoolPtr(cs.Config.IncludeSubModule),
				}
				cfgObj, diags := types.ObjectValueFrom(context.Background(), CustomSourceConfigModel{}.AttributeTypes(), csCfgM)
				if diags.HasError() {
					return nullObj, diags
				}
				csM.Config = cfgObj
			}
			csObj, diags := types.ObjectValueFrom(context.Background(), CustomSourceModel{}.AttributeTypes(), csM)
			if diags.HasError() {
				return nullObj, diags
			}
			iacM.CustomSource = csObj
		}
		var diags diag.Diagnostics
		iacVcsNull, diags = types.ObjectValueFrom(context.Background(), IacVcsConfigModel{}.AttributeTypes(), iacM)
		if diags.HasError() {
			return nullObj, diags
		}
	}
	m.IacVcsConfig = iacVcsNull

	iacInputNull := types.ObjectNull(IacInputDataModel{}.AttributeTypes())
	if vc.IacInputData != nil {
		idM := IacInputDataModel{
			SchemaId:   flatteners.StringPtr(vc.IacInputData.SchemaId),
			SchemaType: flatteners.String(string(vc.IacInputData.SchemaType)),
			Data:       marshalToJSONString(vc.IacInputData.Data),
		}
		var diags diag.Diagnostics
		iacInputNull, diags = types.ObjectValueFrom(context.Background(), IacInputDataModel{}.AttributeTypes(), idM)
		if diags.HasError() {
			return nullObj, diags
		}
	}
	m.IacInputData = iacInputNull

	return types.ObjectValueFrom(context.Background(), VcsConfigModel{}.AttributeTypes(), m)
}

func workflowsConfigFromAPI(wc *stacktemplaterevisions.StackTemplateRevisionWorkflowsConfig) (types.Object, diag.Diagnostics) {
	nullObj := types.ObjectNull(WorkflowsConfigModel{}.AttributeTypes())
	if wc == nil {
		return nullObj, nil
	}
	wfListNull := types.ListNull(types.ObjectType{AttrTypes: WorkflowInStackModel{}.AttributeTypes()})
	if wc.Workflows == nil {
		m := WorkflowsConfigModel{Workflows: wfListNull}
		return types.ObjectValueFrom(context.Background(), WorkflowsConfigModel{}.AttributeTypes(), m)
	}

	elements := make([]attr.Value, 0, len(wc.Workflows))
	for _, wf := range wc.Workflows {
		if wf == nil {
			continue
		}
		wfSteps, diags := wfStepsConfigPointersFromAPI(wf.WfStepsConfig)
		if diags.HasError() {
			return nullObj, diags
		}
		tcObj, diags := terraformConfigFromAPI(wf.TerraformConfig)
		if diags.HasError() {
			return nullObj, diags
		}
		envVars, diags := envVarPointersFromAPI(wf.EnvironmentVariables)
		if diags.HasError() {
			return nullObj, diags
		}
		dpcs, diags := deploymentPlatformConfigFromAPI(wf.DeploymentPlatformConfig)
		if diags.HasError() {
			return nullObj, diags
		}
		vcs, diags := vcsConfigFromAPI(wf.VcsConfig)
		if diags.HasError() {
			return nullObj, diags
		}
		// IacInputData (TemplatesIacInputData)
		iacInputDataObj := types.ObjectNull(WfStepInputDataModel{}.AttributeTypes())
		if wf.IacInputData != nil {
			idm := WfStepInputDataModel{
				SchemaType: flatteners.String(wf.IacInputData.SchemaType),
				Data:       marshalToJSONString(wf.IacInputData.Data),
			}
			iacInputDataObj, diags = types.ObjectValueFrom(context.Background(), WfStepInputDataModel{}.AttributeTypes(), idm)
			if diags.HasError() {
				return nullObj, diags
			}
		}
		us, diags := userSchedulesFromAPI(wf.UserSchedules)
		if diags.HasError() {
			return nullObj, diags
		}
		approvers, diags := types.ListValueFrom(context.Background(), types.StringType, wf.Approvers)
		if diags.HasError() {
			return nullObj, diags
		}
		rc, diags := runnerConstraintsFromAPI(wf.RunnerConstraints)
		if diags.HasError() {
			return nullObj, diags
		}
		// MiniSteps
		msObj, diags := miniStepsFromAPI(context.Background(), wf.MiniSteps)
		if diags.HasError() {
			return nullObj, diags
		}

		// InputSchemas
		inputSchemasNull := types.ListNull(types.ObjectType{AttrTypes: StackInputSchemaModel{}.AttributeTypes()})
		if wf.InputSchemas != nil {
			isElems := make([]attr.Value, 0, len(wf.InputSchemas))
			for _, is := range wf.InputSchemas {
				if is == nil {
					continue
				}
				ism := StackInputSchemaModel{
					Id:           flatteners.StringPtr(is.Id),
					Name:         flatteners.StringPtr(is.Name),
					Description:  flatteners.StringPtr(is.Description),
					Type:         flatteners.String(string(is.Type)),
					EncodedData:  flatteners.StringPtr(is.EncodedData),
					UiSchemaData: flatteners.StringPtr(is.UiSchemaData),
					IsCommitted:  flatteners.BoolPtr(is.IsCommitted),
				}
				isObj, diags2 := types.ObjectValueFrom(context.Background(), StackInputSchemaModel{}.AttributeTypes(), ism)
				if diags2.HasError() {
					return nullObj, diags2
				}
				isElems = append(isElems, isObj)
			}
			inputSchemasNull, diags = types.ListValue(types.ObjectType{AttrTypes: StackInputSchemaModel{}.AttributeTypes()}, isElems)
			if diags.HasError() {
				return nullObj, diags
			}
		}

		// Strip /<ownerOrg>/ prefix from template_id returned by the API
		var strippedTemplateId *string
		if wf.TemplateId != nil {
			parts := strings.Split(*wf.TemplateId, "/")
			base := parts[len(parts)-1]
			strippedTemplateId = &base
		}

		wm := WorkflowInStackModel{
			Id:                        flatteners.StringPtr(wf.Id),
			TemplateId:                flatteners.StringPtr(strippedTemplateId),
			ResourceName:              flatteners.StringPtr(wf.ResourceName),
			WfStepsConfig:             wfSteps,
			TerraformConfig:           tcObj,
			EnvironmentVariables:      envVars,
			DeploymentPlatformConfig:  dpcs,
			VcsConfig:                 vcs,
			IacInputData:              iacInputDataObj,
			UserSchedules:             us,
			Approvers:                 approvers,
			NumberOfApprovalsRequired: flatteners.Int64Ptr(wf.NumberOfApprovalsRequired),
			RunnerConstraints:         rc,
			UserJobCpu:                flatteners.Int64Ptr(wf.UserJobCpu),
			UserJobMemory:             flatteners.Int64Ptr(wf.UserJobMemory),
			InputSchemas:              inputSchemasNull,
			MiniSteps:                 msObj,
		}
		obj, diags := types.ObjectValueFrom(context.Background(), WorkflowInStackModel{}.AttributeTypes(), wm)
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
	return types.ObjectValueFrom(context.Background(), WorkflowsConfigModel{}.AttributeTypes(), m)
}

func actionsFromAPI(actions map[string]*sgsdkgo.Actions) (types.Map, diag.Diagnostics) {
	mapNull := types.MapNull(types.ObjectType{AttrTypes: ActionsModel{}.AttributeTypes()})
	if actions == nil {
		return mapNull, nil
	}
	elements := make(map[string]attr.Value, len(actions))
	for k, a := range actions {
		if a == nil {
			continue
		}
		orderNull := types.MapNull(types.ObjectType{AttrTypes: ActionOrderModel{}.AttributeTypes()})
		if a.Order != nil {
			orderElements := make(map[string]attr.Value, len(a.Order))
			for wfId, ao := range a.Order {
				if ao == nil {
					continue
				}
				// Parameters
				paramsNull := types.ObjectNull(StackActionParametersModel{}.AttributeTypes())
				if ao.Parameters != nil {
					p := ao.Parameters
					// TerraformAction
					taNull := types.ObjectNull(TerraformActionModel{}.AttributeTypes())
					if p.TerraformAction != nil && p.TerraformAction.Action != nil {
						tam := TerraformActionModel{Action: flatteners.String(string(*p.TerraformAction.Action))}
						var diags2 diag.Diagnostics
						taNull, diags2 = types.ObjectValueFrom(context.Background(), TerraformActionModel{}.AttributeTypes(), tam)
						if diags2.HasError() {
							return mapNull, diags2
						}
					}
					// DPC
					dpcList, diags2 := deploymentPlatformConfigFromAPI(p.DeploymentPlatformConfig)
					if diags2.HasError() {
						return mapNull, diags2
					}
					// WfStepsConfig
					wfStepsList, diags2 := wfStepsConfigPointersFromAPI(p.WfStepsConfig)
					if diags2.HasError() {
						return mapNull, diags2
					}
					// EnvVars
					envList, diags2 := envVarPointersFromAPI(p.EnvironmentVariables)
					if diags2.HasError() {
						return mapNull, diags2
					}
					pm := StackActionParametersModel{
						TerraformAction:          taNull,
						DeploymentPlatformConfig: dpcList,
						WfStepsConfig:            wfStepsList,
						EnvironmentVariables:     envList,
					}
					var diags3 diag.Diagnostics
					paramsNull, diags3 = types.ObjectValueFrom(context.Background(), StackActionParametersModel{}.AttributeTypes(), pm)
					if diags3.HasError() {
						return mapNull, diags3
					}
				}
				// Dependencies
				depListNull := types.ListNull(types.ObjectType{AttrTypes: ActionDependencyModel{}.AttributeTypes()})
				if ao.Dependencies != nil {
					depElems := make([]attr.Value, 0, len(ao.Dependencies))
					for _, dep := range ao.Dependencies {
						if dep == nil {
							continue
						}
						condNull := types.ObjectNull(ActionDependencyConditionModel{}.AttributeTypes())
						if dep.Condition != nil {
							condM := ActionDependencyConditionModel{LatestStatus: flatteners.String(dep.Condition.LatestStatus)}
							var diags2 diag.Diagnostics
							condNull, diags2 = types.ObjectValueFrom(context.Background(), ActionDependencyConditionModel{}.AttributeTypes(), condM)
							if diags2.HasError() {
								return mapNull, diags2
							}
						}
						dm := ActionDependencyModel{Id: flatteners.String(dep.Id), Condition: condNull}
						depObj, diags2 := types.ObjectValueFrom(context.Background(), ActionDependencyModel{}.AttributeTypes(), dm)
						if diags2.HasError() {
							return mapNull, diags2
						}
						depElems = append(depElems, depObj)
					}
					var diags2 diag.Diagnostics
					depListNull, diags2 = types.ListValue(types.ObjectType{AttrTypes: ActionDependencyModel{}.AttributeTypes()}, depElems)
					if diags2.HasError() {
						return mapNull, diags2
					}
				}
				aoM := ActionOrderModel{Parameters: paramsNull, Dependencies: depListNull}
				aoObj, diags2 := types.ObjectValueFrom(context.Background(), ActionOrderModel{}.AttributeTypes(), aoM)
				if diags2.HasError() {
					return mapNull, diags2
				}
				orderElements[wfId] = aoObj
			}
			var diags2 diag.Diagnostics
			orderNull, diags2 = types.MapValue(types.ObjectType{AttrTypes: ActionOrderModel{}.AttributeTypes()}, orderElements)
			if diags2.HasError() {
				return mapNull, diags2
			}
		}
		am := ActionsModel{
			Name:        flatteners.String(a.Name),
			Description: flatteners.StringPtr(a.Description),
			Default:     flatteners.BoolPtr(a.Default),
			Order:       orderNull,
		}
		obj, diags2 := types.ObjectValueFrom(context.Background(), ActionsModel{}.AttributeTypes(), am)
		if diags2.HasError() {
			return mapNull, diags2
		}
		elements[k] = obj
	}
	return types.MapValue(types.ObjectType{AttrTypes: ActionsModel{}.AttributeTypes()}, elements)
}

// ---------------------------------------------------------------------------
// ToAPIModel / ToUpdateAPIModel
// ---------------------------------------------------------------------------

func (m *StackTemplateRevisionResourceModel) ToAPIModel(ctx context.Context, orgName string) (*stacktemplaterevisions.CreateStackTemplateRevisionRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	apiModel := &stacktemplaterevisions.CreateStackTemplateRevisionRequest{}

	if !m.Alias.IsNull() && !m.Alias.IsUnknown() {
		apiModel.Alias = m.Alias.ValueString()
	}
	if !m.Notes.IsNull() && !m.Notes.IsUnknown() {
		apiModel.Notes = m.Notes.ValueString()
	}
	if !m.LongDescription.IsNull() && !m.LongDescription.IsUnknown() {
		apiModel.LongDescription = m.LongDescription.ValueStringPointer()
	}
	if !m.SourceConfigKind.IsNull() && !m.SourceConfigKind.IsUnknown() {
		apiModel.SourceConfigKind = (*stacktemplates.StackTemplateSourceConfigKindEnum)(m.SourceConfigKind.ValueStringPointer())
	}
	if !m.IsActive.IsNull() && !m.IsActive.IsUnknown() {
		apiModel.IsActive = (*sgsdkgo.IsPublicEnum)(m.IsActive.ValueStringPointer())
	}
	if !m.IsPublic.IsNull() && !m.IsPublic.IsUnknown() {
		apiModel.IsPublic = (*sgsdkgo.IsPublicEnum)(m.IsPublic.ValueStringPointer())
	}

	tags, diagsTags := expanders.StringList(ctx, m.Tags)
	diags.Append(diagsTags...)
	if !diags.HasError() {
		apiModel.Tags = tags
	}

	if !m.ContextTags.IsNull() && !m.ContextTags.IsUnknown() {
		contextTags := make(map[string]string)
		diags.Append(m.ContextTags.ElementsAs(ctx, &contextTags, false)...)
		if !diags.HasError() {
			apiModel.ContextTags = contextTags
		}
	}

	if !m.Deprecation.IsNull() && !m.Deprecation.IsUnknown() {
		var dep DeprecationModel
		diags.Append(m.Deprecation.As(ctx, &dep, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
		if !diags.HasError() {
			apiModel.Deprecation = &stacktemplaterevisions.Deprecation{
				EffectiveDate: dep.EffectiveDate.ValueStringPointer(),
				Message:       dep.Message.ValueStringPointer(),
			}
		}
	}

	if !m.WorkflowsConfig.IsNull() && !m.WorkflowsConfig.IsUnknown() {
		wc, diagsWC := convertWorkflowsConfigToAPI(ctx, m.WorkflowsConfig, orgName)
		diags.Append(diagsWC...)
		if !diags.HasError() {
			apiModel.WorkflowsConfig = wc
		}
	}

	if !m.Actions.IsNull() && !m.Actions.IsUnknown() {
		acts, diagsActs := convertActionsToAPI(ctx, m.Actions)
		diags.Append(diagsActs...)
		if !diags.HasError() {
			apiModel.Actions = acts
		}
	}

	return apiModel, diags
}

func (m *StackTemplateRevisionResourceModel) ToUpdateAPIModel(ctx context.Context, orgName string) (*stacktemplaterevisions.UpdateStackTemplateRevisionRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	apiModel := &stacktemplaterevisions.UpdateStackTemplateRevisionRequest{}

	if !m.Alias.IsNull() && !m.Alias.IsUnknown() {
		apiModel.Alias = sgsdkgo.Optional(m.Alias.ValueString())
	} else {
		apiModel.Alias = sgsdkgo.Null[string]()
	}
	if !m.Notes.IsNull() && !m.Notes.IsUnknown() {
		apiModel.Notes = sgsdkgo.Optional(m.Notes.ValueString())
	} else {
		apiModel.Notes = sgsdkgo.Null[string]()
	}
	if !m.SourceConfigKind.IsNull() && !m.SourceConfigKind.IsUnknown() {
		apiModel.SourceConfigKind = sgsdkgo.Optional(stacktemplates.StackTemplateSourceConfigKindEnum(m.SourceConfigKind.ValueString()))
	} else {
		apiModel.SourceConfigKind = sgsdkgo.Null[stacktemplates.StackTemplateSourceConfigKindEnum]()
	}
	if !m.IsActive.IsNull() && !m.IsActive.IsUnknown() {
		apiModel.IsActive = sgsdkgo.Optional(sgsdkgo.IsPublicEnum(m.IsActive.ValueString()))
	}
	if !m.IsPublic.IsNull() && !m.IsPublic.IsUnknown() {
		apiModel.IsPublic = sgsdkgo.Optional(sgsdkgo.IsPublicEnum(m.IsPublic.ValueString()))
	}

	tags, diagsTags := expanders.StringList(ctx, m.Tags)
	diags.Append(diagsTags...)
	if tags != nil {
		apiModel.Tags = sgsdkgo.Optional(tags)
	} else {
		apiModel.Tags = sgsdkgo.Null[[]string]()
	}

	contextTags, diagsCT := expanders.MapStringString(ctx, m.ContextTags)
	diags.Append(diagsCT...)
	if contextTags != nil {
		apiModel.ContextTags = sgsdkgo.Optional(contextTags)
	} else {
		apiModel.ContextTags = sgsdkgo.Null[map[string]string]()
	}

	if !m.Deprecation.IsNull() && !m.Deprecation.IsUnknown() {
		var dep DeprecationModel
		diags.Append(m.Deprecation.As(ctx, &dep, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
		if !diags.HasError() {
			apiModel.Deprecation = sgsdkgo.Optional(stacktemplaterevisions.Deprecation{
				EffectiveDate: dep.EffectiveDate.ValueStringPointer(),
				Message:       dep.Message.ValueStringPointer(),
			})
		}
	} else {
		apiModel.Deprecation = sgsdkgo.Null[stacktemplaterevisions.Deprecation]()
	}

	if !m.WorkflowsConfig.IsNull() && !m.WorkflowsConfig.IsUnknown() {
		wc, diagsWC := convertWorkflowsConfigToAPI(ctx, m.WorkflowsConfig, orgName)
		diags.Append(diagsWC...)
		if !diags.HasError() && wc != nil {
			apiModel.WorkflowsConfig = sgsdkgo.Optional(*wc)
		}
	} else {
		apiModel.WorkflowsConfig = sgsdkgo.Null[stacktemplaterevisions.StackTemplateRevisionWorkflowsConfig]()
	}

	if !m.Actions.IsNull() && !m.Actions.IsUnknown() {
		acts, diagsActs := convertActionsToAPI(ctx, m.Actions)
		diags.Append(diagsActs...)
		if !diags.HasError() && acts != nil {
			apiModel.Actions = sgsdkgo.Optional(acts)
		}
	} else {
		apiModel.Actions = sgsdkgo.Null[map[string]*sgsdkgo.Actions]()
	}

	return apiModel, diags
}

// ---------------------------------------------------------------------------
// BuildAPIModelToStackTemplateRevisionModel
// ---------------------------------------------------------------------------

func BuildAPIModelToStackTemplateRevisionModel(ctx context.Context, apiResponse *stacktemplaterevisions.ReadStackTemplateRevisionModel) (*StackTemplateRevisionResourceModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &StackTemplateRevisionResourceModel{
		Id:              flatteners.StringPtr(apiResponse.Id),
		TemplateId:      flatteners.String(apiResponse.TemplateId),
		Alias:           flatteners.String(apiResponse.Alias),
		Notes:           flatteners.String(apiResponse.Notes),
		LongDescription: flatteners.StringPtr(apiResponse.LongDescription),
		IsPublic:        flatteners.StringPtr((*string)(apiResponse.IsPublic)),
	}

	if apiResponse.SourceConfigKind != nil {
		model.SourceConfigKind = flatteners.String(string(*apiResponse.SourceConfigKind))
	} else {
		model.SourceConfigKind = types.StringNull()
	}
	if apiResponse.IsActive != nil {
		model.IsActive = flatteners.String(string(*apiResponse.IsActive))
	} else {
		model.IsActive = types.StringNull()
	}

	// Tags
	if apiResponse.Tags != nil {
		tagsList, diagsTags := types.ListValueFrom(context.Background(), types.StringType, apiResponse.Tags)
		diags.Append(diagsTags...)
		model.Tags = tagsList
	} else {
		model.Tags = types.ListNull(types.StringType)
	}

	// ContextTags
	if apiResponse.ContextTags != nil {
		contextTags := make(map[string]types.String)
		for k, v := range apiResponse.ContextTags {
			contextTags[k] = flatteners.String(v)
		}
		contextTagsMap, diagsCT := types.MapValueFrom(context.Background(), types.StringType, contextTags)
		diags.Append(diagsCT...)
		model.ContextTags = contextTagsMap
	} else {
		model.ContextTags = types.MapNull(types.StringType)
	}

	// Deprecation
	if apiResponse.Deprecation != nil {
		depM := DeprecationModel{
			EffectiveDate: flatteners.StringPtr(apiResponse.Deprecation.EffectiveDate),
			Message:       flatteners.StringPtr(apiResponse.Deprecation.Message),
		}
		depObj, diagsDep := types.ObjectValueFrom(context.Background(), DeprecationModel{}.AttributeTypes(), depM)
		diags.Append(diagsDep...)
		model.Deprecation = depObj
	} else {
		model.Deprecation = types.ObjectNull(DeprecationModel{}.AttributeTypes())
	}

	// WorkflowsConfig
	wcObj, diagsWC := workflowsConfigFromAPI(apiResponse.WorkflowsConfig)
	diags.Append(diagsWC...)
	model.WorkflowsConfig = wcObj

	// Actions
	actsMap, diagsActs := actionsFromAPI(apiResponse.Actions)
	diags.Append(diagsActs...)
	model.Actions = actsMap

	return model, diags
}

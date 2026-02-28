package workflowtemplaterevision

import (
	"context"
	"encoding/json"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/sg-sdk-go/workflowtemplaterevisions"
	"github.com/StackGuardian/sg-sdk-go/workflowtemplates"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/expanders"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/flatteners"
	workflowtemplate "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/workflow_template"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// WorkflowTemplateRevisionResourceModel describes the resource data model based on schema.go
type WorkflowTemplateRevisionResourceModel struct {
	Id                        types.String `tfsdk:"id"`
	TemplateId                types.String `tfsdk:"template_id"`
	LongDescription           types.String `tfsdk:"description"`
	Alias                     types.String `tfsdk:"alias"`
	Notes                     types.String `tfsdk:"notes"`
	SourceConfigKind          types.String `tfsdk:"source_config_kind"`
	IsPublic                  types.String `tfsdk:"is_public"`
	Deprecation               types.Object `tfsdk:"deprecation"`
	EnvironmentVariables      types.List   `tfsdk:"environment_variables"`
	InputSchemas              types.List   `tfsdk:"input_schemas"`
	MiniSteps                 types.Object `tfsdk:"mini_steps"`
	RunnerConstraints         types.Object `tfsdk:"runner_constraints"`
	Tags                      types.List   `tfsdk:"tags"`
	UserSchedules             types.List   `tfsdk:"user_schedules"`
	ContextTags               types.Map    `tfsdk:"context_tags"`
	Approvers                 types.List   `tfsdk:"approvers"`
	NumberOfApprovalsRequired types.Int64  `tfsdk:"number_of_approvals_required"`
	UserJobCPU                types.Int64  `tfsdk:"user_job_cpu"`
	UserJobMemory             types.Int64  `tfsdk:"user_job_memory"`
	RuntimeSource             types.Object `tfsdk:"runtime_source"`
	TerraformConfig           types.Object `tfsdk:"terraform_config"`
	DeploymentPlatformConfig  types.Object `tfsdk:"deployment_platform_config"`
	WfStepsConfig             types.List   `tfsdk:"wf_steps_config"`
}

// Supporting model structs based on schema structure

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

type EnvironmentVariableConfigModel struct {
	VarName   types.String `tfsdk:"var_name"`
	SecretId  types.String `tfsdk:"secret_id"`
	TextValue types.String `tfsdk:"text_value"`
}

func (EnvironmentVariableConfigModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"var_name":   types.StringType,
		"secret_id":  types.StringType,
		"text_value": types.StringType,
	}
}

type EnvironmentVariableModel struct {
	Config types.Object `tfsdk:"config"`
	Kind   types.String `tfsdk:"kind"`
}

func (EnvironmentVariableModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"config": types.ObjectType{AttrTypes: EnvironmentVariableConfigModel{}.AttributeTypes()},
		"kind":   types.StringType,
	}
}

type InputSchemaModel struct {
	Type         types.String `tfsdk:"type"`
	EncodedData  types.String `tfsdk:"encoded_data"`
	UISchemaData types.String `tfsdk:"ui_schema_data"`
}

func (InputSchemaModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":           types.StringType,
		"encoded_data":   types.StringType,
		"ui_schema_data": types.StringType,
	}
}

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

type RuntimeSourceConfigModel struct {
	IsPrivate               types.Bool   `tfsdk:"is_private"`
	Auth                    types.String `tfsdk:"auth"`
	GitCoreAutoCrlf         types.Bool   `tfsdk:"git_core_auto_crlf"`
	GitSparseCheckoutConfig types.String `tfsdk:"git_sparse_checkout_config"`
	IncludeSubModule        types.Bool   `tfsdk:"include_sub_module"`
	Ref                     types.String `tfsdk:"ref"`
	Repo                    types.String `tfsdk:"repo"`
	WorkingDir              types.String `tfsdk:"working_dir"`
}

func (RuntimeSourceConfigModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"is_private":                 types.BoolType,
		"auth":                       types.StringType,
		"git_core_auto_crlf":         types.BoolType,
		"git_sparse_checkout_config": types.StringType,
		"include_sub_module":         types.BoolType,
		"ref":                        types.StringType,
		"repo":                       types.StringType,
		"working_dir":                types.StringType,
	}
}

type RuntimeSourceModel struct {
	SourceConfigDestKind types.String `tfsdk:"source_config_dest_kind"`
	Config               types.Object `tfsdk:"config"`
}

func (RuntimeSourceModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"source_config_dest_kind": types.StringType,
		"config": types.ObjectType{
			AttrTypes: RuntimeSourceConfigModel{}.AttributeTypes(),
		},
	}
}

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
		"post_apply_wf_steps_config":  types.ListType{ElemType: types.ObjectType{AttrTypes: WfStepsConfigModel{}.AttributeTypes()}},
		"pre_apply_wf_steps_config":   types.ListType{ElemType: types.ObjectType{AttrTypes: WfStepsConfigModel{}.AttributeTypes()}},
		"pre_plan_wf_steps_config":    types.ListType{ElemType: types.ObjectType{AttrTypes: WfStepsConfigModel{}.AttributeTypes()}},
		"post_plan_wf_steps_config":   types.ListType{ElemType: types.ObjectType{AttrTypes: WfStepsConfigModel{}.AttributeTypes()}},
		"pre_init_hooks":              types.ListType{ElemType: types.StringType},
		"pre_plan_hooks":              types.ListType{ElemType: types.StringType},
		"post_plan_hooks":             types.ListType{ElemType: types.StringType},
		"pre_apply_hooks":             types.ListType{ElemType: types.StringType},
		"post_apply_hooks":            types.ListType{ElemType: types.StringType},
		"run_pre_init_hooks_on_drift": types.BoolType,
	}
}

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

type DeploymentPlatformConfigModel struct {
	Kind   types.String `tfsdk:"kind"`
	Config types.Object `tfsdk:"config"`
}

func (DeploymentPlatformConfigModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"kind": types.StringType,
		"config": types.ObjectType{
			AttrTypes: DeploymentPlatformConfigConfigModel{}.AttributeTypes(),
		},
	}
}


type UserSchedulesModel struct {
	Cron  types.String `tfsdk:"cron"`
	State types.String `tfsdk:"state"`
	Desc  types.String `tfsdk:"desc"`
	Name  types.String `tfsdk:"name"`
}

func (UserSchedulesModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"cron":  types.StringType,
		"state": types.StringType,
		"desc":  types.StringType,
		"name":  types.StringType,
	}
}

func convertDeprecationToAPIModel(ctx context.Context, deprecationObj types.Object) (*workflowtemplaterevisions.Deprecation, diag.Diagnostics) {
	if deprecationObj.IsNull() || deprecationObj.IsUnknown() {
		return nil, nil
	}

	var deprecationTerraModel DeprecationModel
	diags := deprecationObj.As(ctx, &deprecationTerraModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	return &workflowtemplaterevisions.Deprecation{
		EffectiveDate: deprecationTerraModel.EffectiveDate.ValueStringPointer(),
		Message:       deprecationTerraModel.Message.ValueStringPointer(),
	}, nil
}

func convertRunnerConstraintsToAPIModel(ctx context.Context, runnerConstraintsTerraType types.Object) (*sgsdkgo.RunnerConstraints, diag.Diagnostics) {
	if runnerConstraintsTerraType.IsNull() || runnerConstraintsTerraType.IsUnknown() {
		return nil, nil
	}

	var runnerConstraintsModel RunnerConstraintsModel
	diags := runnerConstraintsTerraType.As(ctx, &runnerConstraintsModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	names, diags := expanders.StringList(ctx, runnerConstraintsModel.Names)
	if diags.HasError() {
		return nil, diags
	}

	return &sgsdkgo.RunnerConstraints{
		Type:  sgsdkgo.RunnerConstraintsTypeEnum(runnerConstraintsModel.Type.ValueString()),
		Names: names,
	}, nil
}

func convertUserSchedulesToAPIModel(ctx context.Context, userSchedulesList types.List) ([]workflowtemplaterevisions.UserSchedules, diag.Diagnostics) {
	if userSchedulesList.IsNull() || userSchedulesList.IsUnknown() {
		return nil, nil
	}

	var models []UserSchedulesModel
	diags := userSchedulesList.ElementsAs(ctx, &models, false)
	if diags.HasError() {
		return nil, diags
	}

	result := make([]workflowtemplaterevisions.UserSchedules, len(models))
	for i, m := range models {
		schedule := workflowtemplaterevisions.UserSchedules{
			Cron:  m.Cron.ValueString(),
			State: workflowtemplaterevisions.UserSchedulesStateEnum(m.State.ValueString()),
			Desc:  m.Desc.ValueStringPointer(),
			Name:  m.Name.ValueStringPointer(),
		}

		result[i] = schedule
	}
	return result, nil
}

// ToAPIModel converts the Terraform model to the API request model
func (m *WorkflowTemplateRevisionResourceModel) ToAPIModel(ctx context.Context) (*workflowtemplaterevisions.CreateWorkflowTemplateRevisionsRequest, diag.Diagnostics) {
	apiModel := &workflowtemplaterevisions.CreateWorkflowTemplateRevisionsRequest{
		LongDescription:           m.LongDescription.ValueStringPointer(),
		SourceConfigKind:          (*workflowtemplates.WorkflowTemplateSourceConfigKindEnum)(m.SourceConfigKind.ValueStringPointer()),
		Alias:                     m.Alias.ValueString(),
		Notes:                     m.Notes.ValueString(),
		TemplateType:              "IAC",
		IsPublic:                  (*sgsdkgo.IsPublicEnum)(m.IsPublic.ValueStringPointer()),
		NumberOfApprovalsRequired: expanders.IntPtr(m.NumberOfApprovalsRequired.ValueInt64Pointer()),
		UserJobCPU:                expanders.IntPtr(m.UserJobCPU.ValueInt64Pointer()),
		UserJobMemory:             expanders.IntPtr(m.UserJobMemory.ValueInt64Pointer()),
	}

	// TODO:
	deprecation, diags := convertDeprecationToAPIModel(ctx, m.Deprecation)
	if diags.HasError() {
		return nil, diags
	}
	apiModel.Deprecation = deprecation

	// handle runner constraints
	runnerConstraints, diags := convertRunnerConstraintsToAPIModel(ctx, m.RunnerConstraints)
	if diags.HasError() {
		return nil, diags
	}
	apiModel.RunnerConstraints = runnerConstraints

	// Handle UserSchedules
	userSchedules, diags := convertUserSchedulesToAPIModel(ctx, m.UserSchedules)
	if diags.HasError() {
		return nil, diags
	}
	apiModel.UserSchedules = userSchedules

	// Handle Tags
	if !m.Tags.IsNull() && !m.Tags.IsUnknown() {
		tags, diags := expanders.StringList(ctx, m.Tags)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.Tags = tags
	}

	// Handle ContextTags
	if !m.ContextTags.IsNull() && !m.ContextTags.IsUnknown() {
		contextTags := make(map[string]string)
		diags := m.ContextTags.ElementsAs(ctx, &contextTags, false)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.ContextTags = contextTags
	}

	// Handle Approvers
	if !m.Approvers.IsNull() && !m.Approvers.IsUnknown() {
		approvers, diags := expanders.StringList(ctx, m.Approvers)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.Approvers = approvers
	}

	// Handle EnvironmentVariables
	if !m.EnvironmentVariables.IsNull() && !m.EnvironmentVariables.IsUnknown() {
		envVars, diags := convertEnvironmentVariablesToAPI(ctx, m.EnvironmentVariables)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.EnvironmentVariables = envVars
	}

	// Handle InputSchemas
	if !m.InputSchemas.IsNull() && !m.InputSchemas.IsUnknown() {
		inputSchemas, diags := convertInputSchemasToAPI(ctx, m.InputSchemas)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.InputSchemas = inputSchemas
	}

	// Handle MiniSteps
	if !m.MiniSteps.IsNull() && !m.MiniSteps.IsUnknown() {
		miniSteps, diags := convertMinistepsToAPI(ctx, m.MiniSteps)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.Ministeps = miniSteps
	}

	// Handle RuntimeSource
	if !m.RuntimeSource.IsNull() && !m.RuntimeSource.IsUnknown() {
		var runtimeSourceModel workflowtemplate.RuntimeSourceModel
		diags := m.RuntimeSource.As(ctx, &runtimeSourceModel, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		if diags.HasError() {
			return nil, diags
		}
		runtimeSource, diags := runtimeSourceModel.ToAPIModel(ctx)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.RuntimeSource = runtimeSource
	}

	// Handle TerraformConfig
	if !m.TerraformConfig.IsNull() && !m.TerraformConfig.IsUnknown() {
		terraformConfig, diags := convertTerraformConfigToAPI(ctx, m.TerraformConfig)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.TerraformConfig = terraformConfig
	}

	// Handle DeploymentPlatformConfig
	if !m.DeploymentPlatformConfig.IsNull() && !m.DeploymentPlatformConfig.IsUnknown() {
		deploymentConfig, diags := convertDeploymentPlatformConfigToAPI(ctx, m.DeploymentPlatformConfig)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.DeploymentPlatformConfig = deploymentConfig
	}

	// Handle WfStepsConfig
	if !m.WfStepsConfig.IsNull() && !m.WfStepsConfig.IsUnknown() {
		wfStepsConfigs, diags := convertWfStepsConfigListToAPI(ctx, m.WfStepsConfig)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.WfStepsConfig = wfStepsConfigs
	}

	return apiModel, nil
}

// TODO: Review Update function
func (m *WorkflowTemplateRevisionResourceModel) ToUpdateAPIModel(ctx context.Context) (*workflowtemplaterevisions.UpdateWorkflowTemplateRevisionRequest, diag.Diagnostics) {
	diagn := diag.Diagnostics{}

	apiModel := &workflowtemplaterevisions.UpdateWorkflowTemplateRevisionRequest{
		IsPublic:      sgsdkgo.Optional(sgsdkgo.IsPublicEnum(m.IsPublic.ValueString())),
		UserJobCPU:    sgsdkgo.Optional(int(m.UserJobCPU.ValueInt64())),
		UserJobMemory: sgsdkgo.Optional(int(m.UserJobMemory.ValueInt64())),
	}

	// Handle LongDescription
	if !m.LongDescription.IsNull() && !m.LongDescription.IsUnknown() {
		apiModel.LongDescription = sgsdkgo.Optional(m.LongDescription.ValueString())
	} else {
		apiModel.LongDescription = sgsdkgo.Null[string]()
	}

	// Handle SourceConfigKind
	if !m.SourceConfigKind.IsNull() && !m.SourceConfigKind.IsUnknown() {
		sourceConfigKind := workflowtemplates.WorkflowTemplateSourceConfigKindEnum(m.SourceConfigKind.ValueString())
		apiModel.SourceConfigKind = sgsdkgo.Optional(sourceConfigKind)
	} else {
		apiModel.SourceConfigKind = sgsdkgo.Null[workflowtemplates.WorkflowTemplateSourceConfigKindEnum]()
	}

	// Handle Alias
	if !m.Alias.IsNull() && !m.Alias.IsUnknown() {
		apiModel.Alias = sgsdkgo.Optional(m.Alias.ValueString())
	} else {
		apiModel.Alias = sgsdkgo.Null[string]()
	}

	// Handle Notes
	if !m.Notes.IsNull() && !m.Notes.IsUnknown() {
		apiModel.Notes = sgsdkgo.Optional(m.Notes.ValueString())
	} else {
		apiModel.Notes = sgsdkgo.Null[string]()
	}

	// Handle NumberOfApprovalsRequired
	if !m.NumberOfApprovalsRequired.IsNull() && !m.NumberOfApprovalsRequired.IsUnknown() {
		numApprovals := int(m.NumberOfApprovalsRequired.ValueInt64())
		apiModel.NumberOfApprovalsRequired = sgsdkgo.Optional(numApprovals)
	}

	// handle deprecation
	deprecation, diags := convertDeprecationToAPIModel(ctx, m.Deprecation)
	if diags.HasError() {
		return nil, diags
	}
	if deprecation != nil {
		apiModel.Deprecation = sgsdkgo.Optional(*deprecation)
	} else {
		apiModel.Deprecation = sgsdkgo.Null[workflowtemplaterevisions.Deprecation]()
	}

	// handle runner constraints
	runnerConstraints, diags := convertRunnerConstraintsToAPIModel(ctx, m.RunnerConstraints)
	if diags.HasError() {
		return nil, diags
	}
	if runnerConstraints != nil {
		apiModel.RunnerConstraints = sgsdkgo.Optional(*runnerConstraints)
	} else {
		apiModel.RunnerConstraints = sgsdkgo.Null[sgsdkgo.RunnerConstraints]()
	}

	// handle user schedules
	userSchedules, diags := convertUserSchedulesToAPIModel(ctx, m.UserSchedules)
	if diags.HasError() {
		return nil, diags
	}
	if userSchedules != nil {
		apiModel.UserSchedules = sgsdkgo.Optional(userSchedules)
	} else {
		apiModel.UserSchedules = sgsdkgo.Null[[]workflowtemplaterevisions.UserSchedules]()
	}

	// Handle Tags
	if !m.Tags.IsNull() && !m.Tags.IsUnknown() {
		tags, diags := expanders.StringList(ctx, m.Tags)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.Tags = sgsdkgo.Optional(tags)
	} else {
		apiModel.Tags = sgsdkgo.Null[[]string]()
	}

	// Handle ContextTags
	if !m.ContextTags.IsNull() && !m.ContextTags.IsUnknown() {
		contextTags := make(map[string]string)
		diags := m.ContextTags.ElementsAs(ctx, &contextTags, false)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.ContextTags = sgsdkgo.Optional(contextTags)
	} else {
		apiModel.ContextTags = sgsdkgo.Null[map[string]string]()
	}

	// Handle Approvers
	if !m.Approvers.IsNull() && !m.Approvers.IsUnknown() {
		approvers, diags := expanders.StringList(ctx, m.Approvers)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.Approvers = sgsdkgo.Optional(approvers)
	} else {
		apiModel.Approvers = sgsdkgo.Null[[]string]()
	}

	// Handle RuntimeSource
	if !m.RuntimeSource.IsNull() && !m.RuntimeSource.IsUnknown() {
		runtimeSource, diags := convertRuntimeSourceToAPI(ctx, m.RuntimeSource)
		if diags.HasError() {
			return nil, diags
		}
		runtimeSourceUpdate := workflowtemplates.RuntimeSourceUpdate{
			SourceConfigDestKind: runtimeSource.SourceConfigDestKind,
		}
		apiModel.RuntimeSource = sgsdkgo.Optional(runtimeSourceUpdate)
	}

	// Handle TerraformConfig
	if !m.TerraformConfig.IsNull() && !m.TerraformConfig.IsUnknown() {
		terraformConfig, diags := convertTerraformConfigToAPI(ctx, m.TerraformConfig)
		diagn.Append(diags...)
		if !diagn.HasError() && terraformConfig != nil {
			apiModel.TerraformConfig = sgsdkgo.Optional(*terraformConfig)
		}
	}

	// Handle DeploymentPlatformConfig
	if !m.DeploymentPlatformConfig.IsNull() && !m.DeploymentPlatformConfig.IsUnknown() {
		deploymentConfig, diags := convertDeploymentPlatformConfigToAPI(ctx, m.DeploymentPlatformConfig)
		diagn.Append(diags...)
		if !diagn.HasError() && deploymentConfig != nil {
			apiModel.DeploymentPlatformConfig = sgsdkgo.Optional(*deploymentConfig)
		}
	}

	// Handle WfStepsConfig at root level
	if !m.WfStepsConfig.IsNull() && !m.WfStepsConfig.IsUnknown() {
		wfStepsConfigs, diags := convertWfStepsConfigListToAPI(ctx, m.WfStepsConfig)
		diagn.Append(diags...)
		if !diagn.HasError() {
			apiModel.WfStepsConfig = sgsdkgo.Optional(wfStepsConfigs)
		}
	}

	// Handle EnvironmentVariables
	envVars, diags := convertEnvironmentVariablesToAPI(ctx, m.EnvironmentVariables)
	if diags.HasError() {
		return nil, diags
	}
	if envVars != nil {
		apiModel.EnvironmentVariables = sgsdkgo.Optional(envVars)
	} else {
		apiModel.EnvironmentVariables = sgsdkgo.Null[[]sgsdkgo.EnvVars]()
	}

	// Handle InputSchemas
	inputSchemas, diags := convertInputSchemasToAPI(ctx, m.InputSchemas)
	if diags.HasError() {
		return nil, diags
	}
	if inputSchemas != nil {
		apiModel.InputSchemas = sgsdkgo.Optional(inputSchemas)
	} else {
		apiModel.InputSchemas = sgsdkgo.Null[[]sgsdkgo.InputSchemas]()
	}

	// Handle Ministeps
	ministeps, diags := convertMinistepsToAPI(ctx, m.MiniSteps)
	if diags.HasError() {
		return nil, diags
	}
	if ministeps != nil {
		apiModel.Ministeps = sgsdkgo.Optional(*ministeps)
	} else {
		apiModel.Ministeps = sgsdkgo.Null[workflowtemplaterevisions.Ministeps]()
	}

	return apiModel, diagn
}

// Helper functions for type conversion
func convertRuntimeSourceToAPI(ctx context.Context, runtimeSourceObj types.Object) (*workflowtemplates.RuntimeSource, diag.Diagnostics) {
	if runtimeSourceObj.IsNull() || runtimeSourceObj.IsUnknown() {
		return nil, nil
	}

	var runtimeSourceModel RuntimeSourceModel
	diags := runtimeSourceObj.As(ctx, &runtimeSourceModel, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return nil, diags
	}

	runtimeSource := &workflowtemplates.RuntimeSource{
		SourceConfigDestKind: workflowtemplates.SourceConfigDestKindEnum(runtimeSourceModel.SourceConfigDestKind.ValueString()).Ptr(),
	}

	// Convert config
	if !runtimeSourceModel.Config.IsNull() && !runtimeSourceModel.Config.IsUnknown() {
		var configModel RuntimeSourceConfigModel
		diag_cfg := runtimeSourceModel.Config.As(ctx, &configModel, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		if diag_cfg.HasError() {
			return nil, diag_cfg
		}

		runtimeSource.Config = &workflowtemplates.RuntimeSourceConfig{
			IsPrivate:               configModel.IsPrivate.ValueBoolPointer(),
			Auth:                    configModel.Auth.ValueStringPointer(),
			GitCoreAutoCRLF:         configModel.GitCoreAutoCrlf.ValueBoolPointer(),
			GitSparseCheckoutConfig: configModel.GitSparseCheckoutConfig.ValueStringPointer(),
			IncludeSubModule:        configModel.IncludeSubModule.ValueBoolPointer(),
			Ref:                     configModel.Ref.ValueStringPointer(),
			Repo:                    configModel.Repo.ValueString(),
			WorkingDir:              configModel.WorkingDir.ValueStringPointer(),
		}
	}

	return runtimeSource, nil
}

func convertTerraformConfigToAPI(ctx context.Context, terraformConfigObj types.Object) (*sgsdkgo.TerraformConfig, diag.Diagnostics) {
	diagn := diag.Diagnostics{}

	var terraformConfigModel TerraformConfigModel
	diag_tc := terraformConfigObj.As(ctx, &terraformConfigModel, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diag_tc.HasError() {
		return nil, diag_tc
	}

	terraformConfig := &sgsdkgo.TerraformConfig{
		TerraformVersion:       terraformConfigModel.TerraformVersion.ValueStringPointer(),
		DriftCheck:             terraformConfigModel.DriftCheck.ValueBoolPointer(),
		DriftCron:              terraformConfigModel.DriftCron.ValueStringPointer(),
		ManagedTerraformState:  terraformConfigModel.ManagedTerraformState.ValueBoolPointer(),
		ApprovalPreApply:       terraformConfigModel.ApprovalPreApply.ValueBoolPointer(),
		TerraformPlanOptions:   terraformConfigModel.TerraformPlanOptions.ValueStringPointer(),
		TerraformInitOptions:   terraformConfigModel.TerraformInitOptions.ValueStringPointer(),
		Timeout:                expanders.IntPtr(terraformConfigModel.Timeout.ValueInt64Pointer()),
		RunPreInitHooksOnDrift: terraformConfigModel.RunPreInitHooksOnDrift.ValueBoolPointer(),
	}

	// Convert TerraformBinPath (MountPoints)
	if !terraformConfigModel.TerraformBinPath.IsNull() && !terraformConfigModel.TerraformBinPath.IsUnknown() {
		mountPoints, diags := convertMountPointsListToAPI(ctx, terraformConfigModel.TerraformBinPath)
		if diags.HasError() {
			return nil, diags
		}
		terraformConfig.TerraformBinPath = mountPoints
	}

	// Convert PostApplyWfStepsConfig
	if !terraformConfigModel.PostApplyWfStepsConfig.IsNull() && !terraformConfigModel.PostApplyWfStepsConfig.IsUnknown() {
		wfSteps, diags := convertWfStepsConfigListToAPI(ctx, terraformConfigModel.PostApplyWfStepsConfig)
		if diags.HasError() {
			return nil, diags
		}
		terraformConfig.PostApplyWfStepsConfig = wfSteps
	}

	// Convert PreApplyWfStepsConfig
	if !terraformConfigModel.PreApplyWfStepsConfig.IsNull() && !terraformConfigModel.PreApplyWfStepsConfig.IsUnknown() {
		wfSteps, diags := convertWfStepsConfigListToAPI(ctx, terraformConfigModel.PreApplyWfStepsConfig)
		if diags.HasError() {
			return nil, diags
		}
		terraformConfig.PreApplyWfStepsConfig = wfSteps
	}

	// Convert PrePlanWfStepsConfig
	if !terraformConfigModel.PrePlanWfStepsConfig.IsNull() && !terraformConfigModel.PrePlanWfStepsConfig.IsUnknown() {
		wfSteps, diags := convertWfStepsConfigListToAPI(ctx, terraformConfigModel.PrePlanWfStepsConfig)
		if diags.HasError() {
			return nil, diags
		}
		terraformConfig.PrePlanWfStepsConfig = wfSteps
	}

	// Convert PostPlanWfStepsConfig
	if !terraformConfigModel.PostPlanWfStepsConfig.IsNull() && !terraformConfigModel.PostPlanWfStepsConfig.IsUnknown() {
		wfSteps, diags := convertWfStepsConfigListToAPI(ctx, terraformConfigModel.PostPlanWfStepsConfig)
		if diags.HasError() {
			return nil, diags
		}
		terraformConfig.PostPlanWfStepsConfig = wfSteps
	}

	// Convert PreInitHooks
	if !terraformConfigModel.PreInitHooks.IsNull() && !terraformConfigModel.PreInitHooks.IsUnknown() {
		preInitHooks, diags := expanders.StringList(ctx, terraformConfigModel.PreInitHooks)
		if diags.HasError() {
			return nil, diags
		}
		terraformConfig.PreInitHooks = preInitHooks
	}

	// Convert PrePlanHooks
	if !terraformConfigModel.PrePlanHooks.IsNull() && !terraformConfigModel.PrePlanHooks.IsUnknown() {
		prePlanHooks, diags := expanders.StringList(ctx, terraformConfigModel.PrePlanHooks)
		if diags.HasError() {
			return nil, diags
		}
		terraformConfig.PrePlanHooks = prePlanHooks
	}

	// Convert PostPlanHooks
	if !terraformConfigModel.PostPlanHooks.IsNull() && !terraformConfigModel.PostPlanHooks.IsUnknown() {
		postPlanHooks, diags := expanders.StringList(ctx, terraformConfigModel.PostPlanHooks)
		if diags.HasError() {
			return nil, diags
		}
		terraformConfig.PostPlanHooks = postPlanHooks
	}

	// Convert PreApplyHooks
	if !terraformConfigModel.PreApplyHooks.IsNull() && !terraformConfigModel.PreApplyHooks.IsUnknown() {
		preApplyHooks, diags := expanders.StringList(ctx, terraformConfigModel.PreApplyHooks)
		if diags.HasError() {
			return nil, diags
		}
		terraformConfig.PreApplyHooks = preApplyHooks
	}

	// Convert PostApplyHooks
	if !terraformConfigModel.PostApplyHooks.IsNull() && !terraformConfigModel.PostApplyHooks.IsUnknown() {
		postApplyHooks, diags := expanders.StringList(ctx, terraformConfigModel.PostApplyHooks)
		if diags.HasError() {
			return nil, diags
		}
		terraformConfig.PostApplyHooks = postApplyHooks
	}

	return terraformConfig, diagn
}

func convertDeploymentPlatformConfigToAPI(ctx context.Context, deploymentConfigObj types.Object) (*workflowtemplaterevisions.DeploymentPlatformConfig, diag.Diagnostics) {
	var deploymentModel DeploymentPlatformConfigModel
	diags := deploymentConfigObj.As(ctx, &deploymentModel, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return nil, diags
	}

	deploymentConfig := &workflowtemplaterevisions.DeploymentPlatformConfig{
		Kind: workflowtemplaterevisions.DeploymentPlatformConfigKindEnum(deploymentModel.Kind.ValueString()),
	}

	// Convert config
	if !deploymentModel.Config.IsNull() && !deploymentModel.Config.IsUnknown() {
		var configModel DeploymentPlatformConfigConfigModel
		diags := deploymentModel.Config.As(ctx, &configModel, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		if diags.HasError() {
			return nil, diags
		}

		deploymentConfig.Config = workflowtemplaterevisions.DeploymentPlatformConfigConfig{
			IntegrationId: configModel.IntegrationId.ValueString(),
			ProfileName:   configModel.ProfileName.ValueStringPointer(),
		}
	}

	return deploymentConfig, nil
}

func convertMountPointsListToAPI(ctx context.Context, mountPointsList types.List) ([]sgsdkgo.MountPoint, diag.Diagnostics) {
	var mountPointModels []MountPointModel
	diags := mountPointsList.ElementsAs(ctx, &mountPointModels, false)
	if diags.HasError() {
		return nil, diags
	}

	mountPoints := make([]sgsdkgo.MountPoint, len(mountPointModels))
	for i, mp := range mountPointModels {
		mountPoints[i] = sgsdkgo.MountPoint{
			Source:   mp.Source.ValueString(),
			Target:   mp.Target.ValueString(),
			ReadOnly: mp.ReadOnly.ValueBoolPointer(),
		}
	}
	return mountPoints, nil
}

func convertWfStepsConfigListToAPI(ctx context.Context, wfStepsConfigList types.List) ([]sgsdkgo.WfStepsConfig, diag.Diagnostics) {
	var wfStepModels []WfStepsConfigModel
	diags := wfStepsConfigList.ElementsAs(ctx, &wfStepModels, false)
	if diags.HasError() {
		return nil, diags
	}

	wfStepsConfigs := make([]sgsdkgo.WfStepsConfig, len(wfStepModels))
	for i, wfStep := range wfStepModels {
		wfStepsConfigs[i] = sgsdkgo.WfStepsConfig{
			Name:             wfStep.Name.ValueString(),
			Approval:         wfStep.Approval.ValueBoolPointer(),
			Timeout:          expanders.IntPtr(wfStep.Timeout.ValueInt64Pointer()),
			WfStepTemplateId: wfStep.WfStepTemplateId.ValueStringPointer(),
			CmdOverride:      wfStep.CmdOverride.ValueStringPointer(),
		}

		// Handle EnvironmentVariables
		// TODO: replace this with the convertEnvironmentVariablesToAPI
		if !wfStep.EnvironmentVariables.IsNull() && !wfStep.EnvironmentVariables.IsUnknown() {
			var envVarModels []EnvironmentVariableModel
			diags := wfStep.EnvironmentVariables.ElementsAs(ctx, &envVarModels, false)
			if diags.HasError() {
				return nil, diags
			}

			envVars := make([]sgsdkgo.EnvVars, len(envVarModels))
			for j, envVar := range envVarModels {
				var configModel EnvironmentVariableConfigModel
				diags := envVar.Config.As(ctx, &configModel, basetypes.ObjectAsOptions{
					UnhandledNullAsEmpty:    true,
					UnhandledUnknownAsEmpty: true,
				})
				if diags.HasError() {
					return nil, diags
				}
				envVars[j] = sgsdkgo.EnvVars{
					Kind: sgsdkgo.EnvVarsKindEnum(envVar.Kind.ValueString()),
					Config: &sgsdkgo.EnvVarConfig{
						VarName:   configModel.VarName.ValueString(),
						SecretId:  configModel.SecretId.ValueStringPointer(),
						TextValue: configModel.TextValue.ValueStringPointer(),
					},
				}
			}
			wfStepsConfigs[i].EnvironmentVariables = envVars
		}

		// Handle MountPoints
		if !wfStep.MountPoints.IsNull() && !wfStep.MountPoints.IsUnknown() {
			mountPoints, diags := convertMountPointsListToAPI(ctx, wfStep.MountPoints)
			if diags.HasError() {
				return nil, diags
			}
			wfStepsConfigs[i].MountPoints = mountPoints
		}

		// Handle WfStepInputData
		if !wfStep.WfStepInputData.IsNull() && !wfStep.WfStepInputData.IsUnknown() {
			var inputDataModel WfStepInputDataModel
			diags := wfStep.WfStepInputData.As(ctx, &inputDataModel, basetypes.ObjectAsOptions{
				UnhandledNullAsEmpty:    true,
				UnhandledUnknownAsEmpty: true,
			})
			if diags.HasError() {
				return nil, diags
			}
			wfStepsConfigs[i].WfStepInputData = &sgsdkgo.WfStepInputData{
				SchemaType: sgsdkgo.WfStepInputDataSchemaTypeEnum(inputDataModel.SchemaType.ValueString()),
				Data:       parseJSONToMap(inputDataModel.Data.ValueString()),
			}
		}
	}

	return wfStepsConfigs, nil
}

// convertEnvironmentVariablesToAPI converts a list of terraform EnvironmentVariableModel to API EnvVars
func convertEnvironmentVariablesToAPI(ctx context.Context, envVarsList types.List) ([]sgsdkgo.EnvVars, diag.Diagnostics) {
	if envVarsList.IsNull() || envVarsList.IsUnknown() {
		return nil, nil
	}

	var envVarModels []EnvironmentVariableModel
	diags := envVarsList.ElementsAs(ctx, &envVarModels, false)
	if diags.HasError() {
		return nil, diags
	}

	envVars := make([]sgsdkgo.EnvVars, len(envVarModels))
	for i, envVar := range envVarModels {
		var configModel EnvironmentVariableConfigModel
		configDiags := envVar.Config.As(ctx, &configModel, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		if configDiags.HasError() {
			return nil, configDiags
		}

		envVars[i] = sgsdkgo.EnvVars{
			Kind: sgsdkgo.EnvVarsKindEnum(envVar.Kind.ValueString()),
			Config: &sgsdkgo.EnvVarConfig{
				VarName:   configModel.VarName.ValueString(),
				SecretId:  configModel.SecretId.ValueStringPointer(),
				TextValue: configModel.TextValue.ValueStringPointer(),
			},
		}
	}

	return envVars, nil
}

// convertWfStepInputDataToAPI converts terraform WfStepInputDataModel to API WfStepInputData
func convertWfStepInputDataToAPI(ctx context.Context, inputDataObj types.Object) (*sgsdkgo.WfStepInputData, diag.Diagnostics) {
	if inputDataObj.IsNull() || inputDataObj.IsUnknown() {
		return nil, nil
	}

	var inputDataModel WfStepInputDataModel
	diags := inputDataObj.As(ctx, &inputDataModel, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return nil, diags
	}

	return &sgsdkgo.WfStepInputData{
		SchemaType: sgsdkgo.WfStepInputDataSchemaTypeEnum(inputDataModel.SchemaType.ValueString()),
		Data:       parseJSONToMap(inputDataModel.Data.ValueString()),
	}, nil
}

// convertWfStepsConfigToAPI converts a single terraform WfStepsConfigModel to API WfStepsConfig
func convertWfStepsConfigToAPI(ctx context.Context, wfStepModel WfStepsConfigModel) (*sgsdkgo.WfStepsConfig, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	wfStepConfig := &sgsdkgo.WfStepsConfig{
		Name:             wfStepModel.Name.ValueString(),
		Approval:         wfStepModel.Approval.ValueBoolPointer(),
		Timeout:          expanders.IntPtr(wfStepModel.Timeout.ValueInt64Pointer()),
		WfStepTemplateId: wfStepModel.WfStepTemplateId.ValueStringPointer(),
		CmdOverride:      wfStepModel.CmdOverride.ValueStringPointer(),
	}

	// Handle EnvironmentVariables
	if !wfStepModel.EnvironmentVariables.IsNull() && !wfStepModel.EnvironmentVariables.IsUnknown() {
		envVars, envDiags := convertEnvironmentVariablesToAPI(ctx, wfStepModel.EnvironmentVariables)
		if envDiags.HasError() {
			return nil, envDiags
		}
		wfStepConfig.EnvironmentVariables = envVars
	}

	// Handle MountPoints
	if !wfStepModel.MountPoints.IsNull() && !wfStepModel.MountPoints.IsUnknown() {
		mountPoints, mountDiags := convertMountPointsListToAPI(ctx, wfStepModel.MountPoints)
		if mountDiags.HasError() {
			return nil, mountDiags
		}
		wfStepConfig.MountPoints = mountPoints
	}

	// Handle WfStepInputData
	if !wfStepModel.WfStepInputData.IsNull() && !wfStepModel.WfStepInputData.IsUnknown() {
		inputData, inputDiags := convertWfStepInputDataToAPI(ctx, wfStepModel.WfStepInputData)
		if inputDiags.HasError() {
			return nil, inputDiags
		}
		wfStepConfig.WfStepInputData = inputData
	}

	return wfStepConfig, diags
}

// convertWfStepInputDataFromAPI converts API WfStepInputData to terraform WfStepInputDataModel
func convertWfStepInputDataFromAPI(ctx context.Context, inputData *sgsdkgo.WfStepInputData) (types.Object, diag.Diagnostics) {
	nullObject := types.ObjectNull(WfStepInputDataModel{}.AttributeTypes())
	if inputData == nil {
		return nullObject, nil
	}

	// Convert map back to JSON string
	dataJSON, err := json.Marshal(inputData.Data)
	if err != nil {
		dataJSON = []byte("{}")
	}

	inputDataModel := WfStepInputDataModel{
		SchemaType: flatteners.String((string)(inputData.SchemaType)),
		Data:       types.StringValue(string(dataJSON)),
	}

	inputDataObj, diags := types.ObjectValueFrom(ctx, WfStepInputDataModel{}.AttributeTypes(), inputDataModel)
	return inputDataObj, diags
}

// convertWfStepsConfigFromAPI converts API WfStepsConfig to terraform WfStepsConfigModel
func convertWfStepsConfigFromAPI(ctx context.Context, wfStepConfig *sgsdkgo.WfStepsConfig) (types.Object, diag.Diagnostics) {
	nullObject := types.ObjectNull(WfStepsConfigModel{}.AttributeTypes())
	if wfStepConfig == nil {
		return nullObject, nil
	}

	wfStepModel := WfStepsConfigModel{
		Name:             flatteners.String(wfStepConfig.Name),
		Approval:         flatteners.BoolPtr(wfStepConfig.Approval),
		Timeout:          flatteners.Int64Ptr(wfStepConfig.Timeout),
		WfStepTemplateId: flatteners.StringPtr(wfStepConfig.WfStepTemplateId),
		CmdOverride:      flatteners.StringPtr(wfStepConfig.CmdOverride),
	}

	// Handle EnvironmentVariables
	envVarsList, envDiags := convertEnvironmentVariablesFromAPI(ctx, wfStepConfig.EnvironmentVariables)
	if envDiags.HasError() {
		return nullObject, envDiags
	}
	wfStepModel.EnvironmentVariables = envVarsList

	// Handle MountPoints
	mountPoints, mountDiags := convertMountPointListFromAPI(ctx, wfStepConfig.MountPoints)
	if mountDiags.HasError() {
		return nullObject, mountDiags
	}
	wfStepModel.MountPoints = mountPoints

	// Handle WfStepInputData
	inputData, inputDiags := convertWfStepInputDataFromAPI(ctx, wfStepConfig.WfStepInputData)
	if inputDiags.HasError() {
		return nullObject, inputDiags
	}
	wfStepModel.WfStepInputData = inputData

	wfStepObj, diags := types.ObjectValueFrom(ctx, WfStepsConfigModel{}.AttributeTypes(), wfStepModel)
	return wfStepObj, diags
}

// convertWfStepsConfigListFromAPI converts API WfStepsConfig list to terraform WfStepsConfigModel list
func convertWfStepsConfigListFromAPI(ctx context.Context, wfStepsConfigList []sgsdkgo.WfStepsConfig) (types.List, diag.Diagnostics) {
	if wfStepsConfigList == nil {
		return types.ListNull(types.ObjectType{AttrTypes: WfStepsConfigModel{}.AttributeTypes()}), nil
	}

	wfStepModels := make([]WfStepsConfigModel, len(wfStepsConfigList))
	for i, wfStep := range wfStepsConfigList {
		wfStepObj, diags := convertWfStepsConfigFromAPI(ctx, &wfStep)
		if diags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: WfStepsConfigModel{}.AttributeTypes()}), diags
		}

		// Extract the WfStepsConfigModel from the object
		var wfStepModel WfStepsConfigModel
		objDiags := wfStepObj.As(ctx, &wfStepModel, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		if objDiags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: WfStepsConfigModel{}.AttributeTypes()}), objDiags
		}

		wfStepModels[i] = wfStepModel
	}

	wfStepsList, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: WfStepsConfigModel{}.AttributeTypes()}, wfStepModels)
	return wfStepsList, diags
}

// convertRunnerConstraintsFromAPI converts API RunnerConstraints to terraform types.Object
func convertRunnerConstraintsFromAPI(ctx context.Context, runnerConstraints *sgsdkgo.RunnerConstraints) (types.Object, diag.Diagnostics) {
	nullObject := types.ObjectNull(RunnerConstraintsModel{}.AttributeTypes())
	if runnerConstraints == nil {
		return nullObject, nil
	}

	namesList, diags := flatteners.ListOfStringToTerraformList(runnerConstraints.Names)
	if diags.HasError() {
		return nullObject, diags
	}

	model := RunnerConstraintsModel{
		Type:  flatteners.String(string(runnerConstraints.Type)),
		Names: namesList,
	}

	obj, diags := types.ObjectValueFrom(ctx, RunnerConstraintsModel{}.AttributeTypes(), model)
	if diags.HasError() {
		return nullObject, diags
	}

	return obj, nil
}

// convertUserSchedulesFromAPI converts API UserSchedules to terraform types.List
func convertUserSchedulesFromAPI(ctx context.Context, userSchedules []workflowtemplaterevisions.UserSchedules) (types.List, diag.Diagnostics) {
	nullList := types.ListNull(types.ObjectType{AttrTypes: UserSchedulesModel{}.AttributeTypes()})
	if userSchedules == nil {
		return nullList, nil
	}

	models := make([]UserSchedulesModel, len(userSchedules))
	for i, us := range userSchedules {
		models[i] = UserSchedulesModel{
			Cron:  flatteners.String(us.Cron),
			State: flatteners.String(string(us.State)),
			Desc:  flatteners.StringPtr(us.Desc),
			Name:  flatteners.StringPtr(us.Name),
		}
	}

	list, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: UserSchedulesModel{}.AttributeTypes()}, models)
	return list, diags
}

// BuildAPIModelToWorkflowTemplateRevisionModel converts API response to Terraform model
func BuildAPIModelToWorkflowTemplateRevisionModel(ctx context.Context, apiResponse *workflowtemplaterevisions.ReadWorkflowTemplateRevisionModel) (*WorkflowTemplateRevisionResourceModel, diag.Diagnostics) {
	if apiResponse == nil {
		return nil, nil
	}

	model := &WorkflowTemplateRevisionResourceModel{
		Id:               flatteners.StringPtr(apiResponse.Id),
		LongDescription:  flatteners.StringPtr(apiResponse.LongDescription),
		Alias:            flatteners.String(apiResponse.Alias),
		Notes:            flatteners.String(apiResponse.Notes),
		SourceConfigKind: flatteners.StringPtr((*string)(apiResponse.SourceConfigKind)),
		IsPublic:         flatteners.StringPtr((*string)(apiResponse.IsPublic)),
	}

	// handle deprecation
	deprecationTerraType, diags := convertDeprecationFromAPI(ctx, apiResponse.Deprecation)
	if diags.HasError() {
		return nil, diags
	}
	model.Deprecation = deprecationTerraType

	// Handle Tags
	tagsTerraType, diags := flatteners.ListOfStringToTerraformList(apiResponse.Tags)
	if diags.HasError() {
		return nil, diags
	}
	model.Tags = tagsTerraType

	// Handle ContextTags
	if apiResponse.ContextTags != nil {
		contextTagsValue, diags := types.MapValueFrom(ctx, types.StringType, apiResponse.ContextTags)
		if diags.HasError() {
			return nil, diags
		}
		model.ContextTags = contextTagsValue
	} else {
		model.ContextTags = types.MapNull(types.StringType)
	}

	// Handle Approvers
	approverstTerraType, diags := flatteners.ListOfStringToTerraformList(apiResponse.Approvers)
	if diags.HasError() {
		return nil, diags
	}
	model.Approvers = approverstTerraType

	// Handle NumberOfApprovalsRequired
	if apiResponse.NumberOfApprovalsRequired != nil {
		model.NumberOfApprovalsRequired = flatteners.Int64Ptr(apiResponse.NumberOfApprovalsRequired)
	}

	// Handle UserJobCPU and UserJobMemory
	if apiResponse.UserJobCPU != nil {
		model.UserJobCPU = flatteners.Int64Ptr(apiResponse.UserJobCPU)
	}

	if apiResponse.UserJobMemory != nil {
		model.UserJobMemory = flatteners.Int64Ptr(apiResponse.UserJobMemory)
	}

	// Handle RuntimeSource
	runtimeSource, diags := convertRuntimeSourceFromAPI(ctx, apiResponse.RuntimeSource)
	if diags.HasError() {
		return nil, diags
	}
	model.RuntimeSource = runtimeSource

	// Handle TerraformConfig
	terraformConfig, diags := convertTerraformConfigFromAPI(ctx, apiResponse.TerraformConfig)
	if diags.HasError() {
		return nil, diags
	}
	model.TerraformConfig = terraformConfig

	// Handle DeploymentPlatformConfig
	deploymentConfig, diags := convertDeploymentPlatformConfigFromAPI(ctx, apiResponse.DeploymentPlatformConfig)
	if diags.HasError() {
		return nil, diags
	}
	model.DeploymentPlatformConfig = deploymentConfig

	// Handle EnvironmentVariables
	envVars, diags := convertEnvironmentVariablesFromAPI(ctx, apiResponse.EnvironmentVariables)
	if diags.HasError() {
		return nil, diags
	}
	model.EnvironmentVariables = envVars

	// Handle InputSchemas
	inputSchemas, diags := convertInputSchemasFromAPI(ctx, apiResponse.InputSchemas)
	if diags.HasError() {
		return nil, diags
	}
	model.InputSchemas = inputSchemas

	// Handle MiniSteps
	miniSteps, diags := convertMinistepsFromAPI(ctx, apiResponse.Ministeps)
	if diags.HasError() {
		return nil, diags
	}
	model.MiniSteps = miniSteps

	// Handle RunnerConstraints
	runnerConstraints, diags := convertRunnerConstraintsFromAPI(ctx, apiResponse.RunnerConstraints)
	if diags.HasError() {
		return nil, diags
	}
	model.RunnerConstraints = runnerConstraints

	// Handle UserSchedules
	userSchedules, diags := convertUserSchedulesFromAPI(ctx, apiResponse.UserSchedules)
	if diags.HasError() {
		return nil, diags
	}
	model.UserSchedules = userSchedules

	// Handle WfStepsConfig
	wfStepsConfig, diags := convertWfStepsConfigListFromAPI(ctx, apiResponse.WfStepsConfig)
	if diags.HasError() {
		return nil, diags
	}
	model.WfStepsConfig = wfStepsConfig

	return model, nil
}

// Helper functions for reverse conversion (API to Terraform)

func convertDeprecationFromAPI(ctx context.Context, deprecationAPIModle *workflowtemplaterevisions.Deprecation) (types.Object, diag.Diagnostics) {
	nullObject := types.ObjectNull(DeprecationModel{}.AttributeTypes())
	if deprecationAPIModle == nil {
		return nullObject, nil
	}

	deprecationTerraModel := DeprecationModel{
		EffectiveDate: flatteners.StringPtr(deprecationAPIModle.EffectiveDate),
		Message:       flatteners.StringPtr(deprecationAPIModle.Message),
	}

	deprecationTerraType, diags := types.ObjectValueFrom(ctx, DeprecationModel{}.AttributeTypes(), deprecationTerraModel)
	if diags.HasError() {
		return nullObject, diags
	}

	return deprecationTerraType, nil
}

func convertRuntimeSourceFromAPI(ctx context.Context, runtimeSource *workflowtemplates.RuntimeSource) (types.Object, diag.Diagnostics) {
	nullObject := types.ObjectNull(RuntimeSourceModel{}.AttributeTypes())
	if runtimeSource == nil {
		return nullObject, nil
	}

	runtimeSourceModel := RuntimeSourceModel{
		SourceConfigDestKind: flatteners.StringPtr((*string)(runtimeSource.SourceConfigDestKind)),
	}

	if runtimeSource.Config != nil {
		configModel := RuntimeSourceConfigModel{
			IsPrivate:               flatteners.BoolPtr(runtimeSource.Config.IsPrivate),
			Auth:                    flatteners.StringPtr(runtimeSource.Config.Auth),
			GitCoreAutoCrlf:         flatteners.BoolPtr(runtimeSource.Config.GitCoreAutoCRLF),
			GitSparseCheckoutConfig: flatteners.StringPtr(runtimeSource.Config.GitSparseCheckoutConfig),
			IncludeSubModule:        flatteners.BoolPtr(runtimeSource.Config.IncludeSubModule),
			Ref:                     flatteners.StringPtr(runtimeSource.Config.Ref),
			Repo:                    flatteners.String(runtimeSource.Config.Repo),
			WorkingDir:              flatteners.StringPtr(runtimeSource.Config.WorkingDir),
		}

		configObj, diags := types.ObjectValueFrom(ctx, RuntimeSourceConfigModel{}.AttributeTypes(), configModel)
		if diags.HasError() {
			return nullObject, diags
		}

		runtimeSourceModel.Config = configObj
	} else {
		runtimeSourceModel.Config = types.ObjectNull(RuntimeSourceConfigModel{}.AttributeTypes())
	}

	runtimeSourceObj, diags := types.ObjectValueFrom(ctx, RuntimeSourceModel{}.AttributeTypes(), runtimeSourceModel)
	if diags.HasError() {
		return nullObject, diags
	}

	return runtimeSourceObj, nil
}

func convertMountPointListFromAPI(ctx context.Context, mountPointListAPI []sgsdkgo.MountPoint) (types.List, diag.Diagnostics) {
	nullObject := types.ListNull(types.ObjectType{AttrTypes: MountPointModel{}.AttributeTypes()})
	if mountPointListAPI == nil {
		return nullObject, nil
	}

	mountPointListTerraModel := []MountPointModel{}
	for _, mountPointAPIModel := range mountPointListAPI {
		mountPointTerraModel := MountPointModel{
			Source:   flatteners.String(mountPointAPIModel.Source),
			Target:   flatteners.String(mountPointAPIModel.Target),
			ReadOnly: flatteners.BoolPtr(mountPointAPIModel.ReadOnly),
		}
		mountPointListTerraModel = append(mountPointListTerraModel, mountPointTerraModel)
	}

	mountPointTerraType, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: MountPointModel{}.AttributeTypes()}, &mountPointListTerraModel)
	if diags.HasError() {
		return nullObject, diags
	}

	return mountPointTerraType, nil
}

func convertTerraformConfigFromAPI(ctx context.Context, terraformConfig *sgsdkgo.TerraformConfig) (types.Object, diag.Diagnostics) {
	nullObject := types.ObjectNull(TerraformConfigModel{}.AttributeTypes())
	if terraformConfig == nil {
		return nullObject, nil
	}

	terraformConfigModel := TerraformConfigModel{
		TerraformVersion:       flatteners.StringPtr(terraformConfig.TerraformVersion),
		DriftCheck:             flatteners.BoolPtr(terraformConfig.DriftCheck),
		DriftCron:              flatteners.StringPtr(terraformConfig.DriftCron),
		ManagedTerraformState:  flatteners.BoolPtr(terraformConfig.ManagedTerraformState),
		ApprovalPreApply:       flatteners.BoolPtr(terraformConfig.ApprovalPreApply),
		TerraformPlanOptions:   flatteners.StringPtr(terraformConfig.TerraformPlanOptions),
		TerraformInitOptions:   flatteners.StringPtr(terraformConfig.TerraformInitOptions),
		Timeout:                flatteners.Int64Ptr(terraformConfig.Timeout),
		RunPreInitHooksOnDrift: flatteners.BoolPtr(terraformConfig.RunPreInitHooksOnDrift),
	}

	// TODO: Convert nested fields (TerraformBinPath, WfStepsConfigs, Hooks, Providers)

	// terraform bin path
	terraformBinTerraType, diags := convertMountPointListFromAPI(ctx, terraformConfig.TerraformBinPath)
	if diags.HasError() {
		return nullObject, diags
	}
	terraformConfigModel.TerraformBinPath = terraformBinTerraType

	// post apply wf steps config
	postApplyWfStepsConfig, diags := convertWfStepsConfigListFromAPI(ctx, terraformConfig.PostApplyWfStepsConfig)
	if diags.HasError() {
		return nullObject, diags
	}
	terraformConfigModel.PostApplyWfStepsConfig = postApplyWfStepsConfig

	// pre apply wf steps config
	preApplyWfStepsConfig, diags := convertWfStepsConfigListFromAPI(ctx, terraformConfig.PreApplyWfStepsConfig)
	if diags.HasError() {
		return nullObject, diags
	}
	terraformConfigModel.PreApplyWfStepsConfig = preApplyWfStepsConfig

	// pre plan wf steps config
	prePlanWfStepsConfig, diags := convertWfStepsConfigListFromAPI(ctx, terraformConfig.PrePlanWfStepsConfig)
	if diags.HasError() {
		return nullObject, diags
	}
	terraformConfigModel.PrePlanWfStepsConfig = prePlanWfStepsConfig

	// post plan wf steps config
	postPlanWfStepsConfig, diags := convertWfStepsConfigListFromAPI(ctx, terraformConfig.PostPlanWfStepsConfig)
	if diags.HasError() {
		return nullObject, diags
	}
	terraformConfigModel.PostPlanWfStepsConfig = postPlanWfStepsConfig

	// pre init hooks
	preInitHooks, diags := flatteners.ListOfStringToTerraformList(terraformConfig.PreInitHooks)
	if diags.HasError() {
		return nullObject, diags
	}
	terraformConfigModel.PreInitHooks = preInitHooks

	// pre plan hooks
	prePlanHooks, diags := flatteners.ListOfStringToTerraformList(terraformConfig.PrePlanHooks)
	if diags.HasError() {
		return nullObject, diags
	}
	terraformConfigModel.PrePlanHooks = prePlanHooks

	// post plan hooks
	postPlanHooks, diags := flatteners.ListOfStringToTerraformList(terraformConfig.PostPlanHooks)
	if diags.HasError() {
		return nullObject, diags
	}
	terraformConfigModel.PostPlanHooks = postPlanHooks

	// pre apply hooks
	preApplyHooks, diags := flatteners.ListOfStringToTerraformList(terraformConfig.PreApplyHooks)
	if diags.HasError() {
		return nullObject, diags
	}
	terraformConfigModel.PreApplyHooks = preApplyHooks

	// post apply hooks
	postApplyHooks, diags := flatteners.ListOfStringToTerraformList(terraformConfig.PostApplyHooks)
	if diags.HasError() {
		return nullObject, diags
	}
	terraformConfigModel.PostApplyHooks = postApplyHooks

	terraformConfigObj, diags := types.ObjectValueFrom(ctx, TerraformConfigModel{}.AttributeTypes(), terraformConfigModel)
	if diags.HasError() {
		return nullObject, diags
	}

	return terraformConfigObj, nil
}

func convertDeploymentPlatformConfigFromAPI(ctx context.Context, deploymentConfig *workflowtemplaterevisions.DeploymentPlatformConfig) (types.Object, diag.Diagnostics) {
	nullObject := types.ObjectNull(DeploymentPlatformConfigModel{}.AttributeTypes())
	if deploymentConfig == nil {
		return nullObject, nil
	}

	configModel := DeploymentPlatformConfigConfigModel{
		IntegrationId: flatteners.String(deploymentConfig.Config.IntegrationId),
		ProfileName:   flatteners.StringPtr(deploymentConfig.Config.ProfileName),
	}

	configObj, diags := types.ObjectValueFrom(ctx, DeploymentPlatformConfigConfigModel{}.AttributeTypes(), configModel)
	if diags.HasError() {
		return nullObject, diags
	}

	deploymentConfigModel := DeploymentPlatformConfigModel{
		Kind:   flatteners.String(string(deploymentConfig.Kind)),
		Config: configObj,
	}

	deploymentConfigObj, diags := types.ObjectValueFrom(ctx, DeploymentPlatformConfigModel{}.AttributeTypes(), deploymentConfigModel)
	if diags.HasError() {
		return nullObject, diags
	}

	return deploymentConfigObj, nil
}

// Conversion helpers for EnvironmentVariables (Flatteners - API to Terraform)
func convertEnvironmentVariablesFromAPI(ctx context.Context, envVars []sgsdkgo.EnvVars) (types.List, diag.Diagnostics) {
	nullObject := types.ListNull(types.ObjectType{AttrTypes: EnvironmentVariableModel{}.AttributeTypes()})

	if envVars == nil {
		return nullObject, nil
	}

	envVarModels := make([]EnvironmentVariableModel, len(envVars))
	for i, envVar := range envVars {
		configModel := EnvironmentVariableConfigModel{
			VarName:   flatteners.String(envVar.Config.VarName),
			SecretId:  flatteners.StringPtr(envVar.Config.SecretId),
			TextValue: flatteners.StringPtr(envVar.Config.TextValue),
		}

		configObj, diags := types.ObjectValueFrom(ctx, EnvironmentVariableConfigModel{}.AttributeTypes(), configModel)
		if diags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: EnvironmentVariableModel{}.AttributeTypes()}), diags
		}

		envVarModels[i] = EnvironmentVariableModel{
			Config: configObj,
			Kind:   flatteners.String(string(envVar.Kind)),
		}
	}

	envVarList, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: EnvironmentVariableModel{}.AttributeTypes()}, envVarModels)
	if diags.HasError() {
		return types.ListNull(types.ObjectType{AttrTypes: EnvironmentVariableModel{}.AttributeTypes()}), diags
	}

	return envVarList, nil
}

// Conversion helpers for InputSchemas (Flatteners - API to Terraform)
func convertInputSchemasFromAPI(ctx context.Context, inputSchemas []sgsdkgo.InputSchemas) (types.List, diag.Diagnostics) {
	if inputSchemas == nil || len(inputSchemas) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: InputSchemaModel{}.AttributeTypes()}), nil
	}

	inputSchemaModels := make([]InputSchemaModel, len(inputSchemas))
	for i, schema := range inputSchemas {
		inputSchemaModels[i] = InputSchemaModel{
			Type:         flatteners.String(string(schema.Type)),
			EncodedData:  flatteners.StringPtr(schema.EncodedData),
			UISchemaData: flatteners.StringPtr(schema.UiSchemaData),
		}
	}

	inputSchemaList, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: InputSchemaModel{}.AttributeTypes()}, inputSchemaModels)
	if diags.HasError() {
		return types.ListNull(types.ObjectType{AttrTypes: InputSchemaModel{}.AttributeTypes()}), diags
	}

	return inputSchemaList, nil
}

// Conversion helpers for MiniSteps (Flatteners - API to Terraform)
func convertMinistepsFromAPI(ctx context.Context, ministeps *workflowtemplaterevisions.Ministeps) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	nullObject := types.ObjectNull(MinistepsModel{}.AttributeTypes())
	if ministeps == nil {
		return nullObject, nil
	}

	ministepsModel := MinistepsModel{}

	// Convert Notifications
	if ministeps.Notifications != nil {
		notificationsModel := MinistepsNotificationsModel{}

		if ministeps.Notifications.Email != nil {
			emailModel := MinistepsEmailModel{}
			emailModel.ApprovalRequired, diags = convertNotificationRecipientsFromAPI(ctx, ministeps.Notifications.Email.APPROVAL_REQUIRED)
			if diags.HasError() {
				return nullObject, diags
			}

			emailModel.Cancelled, diags = convertNotificationRecipientsFromAPI(ctx, ministeps.Notifications.Email.CANCELLED)
			if diags.HasError() {
				return nullObject, diags
			}

			emailModel.Completed, diags = convertNotificationRecipientsFromAPI(ctx, ministeps.Notifications.Email.COMPLETED)
			if diags.HasError() {
				return nullObject, diags
			}

			emailModel.DriftDetected, diags = convertNotificationRecipientsFromAPI(ctx, ministeps.Notifications.Email.DRIFT_DETECTED)
			if diags.HasError() {
				return nullObject, diags
			}

			emailModel.Errored, diags = convertNotificationRecipientsFromAPI(ctx, ministeps.Notifications.Email.ERRORED)
			if diags.HasError() {
				return nullObject, diags
			}

			emailObj, diags := types.ObjectValueFrom(ctx, MinistepsEmailModel{}.AttributeTypes(), emailModel)
			if diags.HasError() {
				return nullObject, diags
			}
			notificationsModel.Email = emailObj
		} else {
			notificationsModel.Email = types.ObjectNull(MinistepsEmailModel{}.AttributeTypes())
		}

		notificationsObj, diags := types.ObjectValueFrom(ctx, MinistepsNotificationsModel{}.AttributeTypes(), notificationsModel)
		if diags.HasError() {
			return nullObject, diags
		}
		ministepsModel.Notifications = notificationsObj
	} else {
		ministepsModel.Notifications = types.ObjectNull(MinistepsNotificationsModel{}.AttributeTypes())
	}

	// Convert Webhooks
	if ministeps.Webhooks != nil {
		webhooksModel := MinistepsWebhooksContainerModel{}
		webhooksModel.ApprovalRequired, diags = convertWebhookFromAPI(ctx, ministeps.Webhooks.APPROVAL_REQUIRED)
		if diags.HasError() {
			return nullObject, diags
		}

		webhooksModel.Cancelled, diags = convertWebhookFromAPI(ctx, ministeps.Webhooks.CANCELLED)
		if diags.HasError() {
			return nullObject, diags
		}

		webhooksModel.Completed, diags = convertWebhookFromAPI(ctx, ministeps.Webhooks.COMPLETED)
		if diags.HasError() {
			return nullObject, diags
		}

		webhooksModel.DriftDetected, diags = convertWebhookFromAPI(ctx, ministeps.Webhooks.DRIFT_DETECTED)
		if diags.HasError() {
			return nullObject, diags
		}

		webhooksModel.Errored, diags = convertWebhookFromAPI(ctx, ministeps.Webhooks.ERRORED)
		if diags.HasError() {
			return nullObject, diags
		}

		webhooksObj, diags := types.ObjectValueFrom(ctx, MinistepsWebhooksContainerModel{}.AttributeTypes(), webhooksModel)
		if diags.HasError() {
			return nullObject, diags
		}
		ministepsModel.Webhooks = webhooksObj
	} else {
		ministepsModel.Webhooks = types.ObjectNull(MinistepsWebhooksContainerModel{}.AttributeTypes())
	}

	// Convert WfChaining
	if ministeps.WfChaining != nil {
		wfChainingModel := MinistepsWfChainingContainerModel{}

		wfChainingModel.Completed, diags = convertWorkflowChainingFromAPI(ctx, ministeps.WfChaining.COMPLETED)
		if diags.HasError() {
			return nullObject, diags
		}
		wfChainingModel.Errored, diags = convertWorkflowChainingFromAPI(ctx, ministeps.WfChaining.ERRORED)

		wfChainingObj, diags := types.ObjectValueFrom(ctx, MinistepsWfChainingContainerModel{}.AttributeTypes(), wfChainingModel)
		if diags.HasError() {
			return nullObject, diags
		}
		ministepsModel.WfChaining = wfChainingObj
	} else {
		ministepsModel.WfChaining = types.ObjectNull(MinistepsWfChainingContainerModel{}.AttributeTypes())
	}

	ministepsObj, diags := types.ObjectValueFrom(ctx, MinistepsModel{}.AttributeTypes(), ministepsModel)
	if diags.HasError() {
		return nullObject, diags
	}

	return ministepsObj, nil
}

// Helper function to convert notification recipients
func convertNotificationRecipientsFromAPI(ctx context.Context, recipients []workflowtemplaterevisions.MinistepsNotificationRecepients) (types.List, diag.Diagnostics) {
	nullObj := types.ListNull(types.ObjectType{AttrTypes: MinistepsNotificationRecipientsModel{}.AttributeTypes()})
	if recipients == nil {
		return nullObj, nil
	}

	recepientsListTerraModel := []MinistepsNotificationRecipientsModel{}
	for _, recepientList := range recipients {
		recipients, diags := types.ListValueFrom(ctx, types.StringType, recepientList.Recipients)
		if diags.HasError() {
			return nullObj, diags
		}

		recepientsTerraModel := MinistepsNotificationRecipientsModel{
			Recipients: recipients,
		}

		recepientsListTerraModel = append(recepientsListTerraModel, recepientsTerraModel)
	}

	obj, diags := types.ListValueFrom(ctx, types.ListType{ElemType: types.StringType}, recepientsListTerraModel)
	if diags.HasError() {
		return nullObj, diags
	}

	return obj, nil
}

// Helper function to convert webhooks
func convertWebhookFromAPI(ctx context.Context, webhooks []workflowtemplaterevisions.MinistepsWebhooksSchema) (types.List, diag.Diagnostics) {
	nullObj := types.ListNull(types.ObjectType{AttrTypes: MinistepsWebhooksModel{}.AttributeTypes()})
	if webhooks == nil {
		return nullObj, nil
	}

	ministepsWebhookTerraModel := []MinistepsWebhooksModel{}
	for _, webhook := range webhooks {
		ministepsWebhookSchemaModel := MinistepsWebhooksModel{
			WebhookName:   flatteners.String(webhook.WebhookName),
			WebhookUrl:    flatteners.String(webhook.WebhookUrl),
			WebhookSecret: flatteners.StringPtr(webhook.WebhookSecret),
		}

		ministepsWebhookTerraModel = append(ministepsWebhookTerraModel, ministepsWebhookSchemaModel)
	}

	ministepsWebhookTerraType, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: MinistepsWebhooksModel{}.AttributeTypes()}, ministepsWebhookTerraModel)
	if diags.HasError() {
		return nullObj, diags
	}

	return ministepsWebhookTerraType, nil
}

// Helper function to convert workflow chaining
func convertWorkflowChainingFromAPI(ctx context.Context, wfChainingList []workflowtemplaterevisions.MinistepsWfChainingSchema) (types.List, diag.Diagnostics) {
	nullObj := types.ListNull(types.ObjectType{AttrTypes: MinistepsWorkflowChainingModel{}.AttributeTypes()})
	if wfChainingList == nil {
		return nullObj, nil
	}

	workflowChainingListTerraModel := []MinistepsWorkflowChainingModel{}
	for _, wfChaining := range wfChainingList {
		model := MinistepsWorkflowChainingModel{
			WorkflowGroupId:    flatteners.String(wfChaining.WorkflowGroupId),
			StackId:            flatteners.StringPtr(wfChaining.StackId),
			StackRunPayload:    flatteners.StringPtr(wfChaining.StackRunPayload),
			WorkflowId:         flatteners.StringPtr(wfChaining.WorkflowId),
			WorkflowRunPayload: flatteners.StringPtr(wfChaining.WorkflowRunPayload),
		}
		workflowChainingListTerraModel = append(workflowChainingListTerraModel, model)
	}

	obj, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: MinistepsWorkflowChainingModel{}.AttributeTypes()}, workflowChainingListTerraModel)
	if diags.HasError() {
		return nullObj, diags
	}

	return obj, nil
}

// Conversion helpers for EnvironmentVariables (Expanders - Terraform to API)
// Conversion helpers for InputSchemas (Expanders - Terraform to API)
func convertInputSchemasToAPI(ctx context.Context, inputSchemasList types.List) ([]sgsdkgo.InputSchemas, diag.Diagnostics) {
	if inputSchemasList.IsNull() || inputSchemasList.IsUnknown() {
		return nil, nil
	}

	var inputSchemaModels []InputSchemaModel
	diags := inputSchemasList.ElementsAs(ctx, &inputSchemaModels, false)
	if diags.HasError() {
		return nil, diags
	}

	inputSchemas := make([]sgsdkgo.InputSchemas, len(inputSchemaModels))
	for i, schema := range inputSchemaModels {
		inputSchemas[i] = sgsdkgo.InputSchemas{
			Type:         sgsdkgo.InputSchemasTypeEnum(schema.Type.ValueString()),
			EncodedData:  schema.EncodedData.ValueStringPointer(),
			UiSchemaData: schema.UISchemaData.ValueStringPointer(),
		}
	}

	return inputSchemas, nil
}

func convertMinistepsToAPI(ctx context.Context, ministepsObj types.Object) (*workflowtemplaterevisions.Ministeps, diag.Diagnostics) {
	if ministepsObj.IsNull() || ministepsObj.IsUnknown() {
		return nil, nil
	}

	var ministepsModel MinistepsModel
	diags := ministepsObj.As(ctx, &ministepsModel, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return nil, diags
	}

	miniSteps := &workflowtemplaterevisions.Ministeps{}

	// Convert Notifications
	if !ministepsModel.Notifications.IsNull() && !ministepsModel.Notifications.IsUnknown() {
		var notificationsModel MinistepsNotificationsModel
		diags := ministepsModel.Notifications.As(ctx, &notificationsModel, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		if diags.HasError() {
			return nil, diags
		}

		miniSteps.Notifications = &workflowtemplaterevisions.MinistepsNotifications{}

		// Convert Email notifications
		if !notificationsModel.Email.IsNull() && !notificationsModel.Email.IsUnknown() {
			var emailModel MinistepsEmailModel
			diags := notificationsModel.Email.As(ctx, &emailModel, basetypes.ObjectAsOptions{
				UnhandledNullAsEmpty:    true,
				UnhandledUnknownAsEmpty: true,
			})
			if diags.HasError() {
				return nil, diags
			}

			miniSteps.Notifications.Email = &workflowtemplaterevisions.MinistepsNotificationsEmail{}
			miniSteps.Notifications.Email.APPROVAL_REQUIRED, diags = convertNotificationRecipientsToAPI(ctx, emailModel.ApprovalRequired)
			if diags.HasError() {
				return nil, diags
			}

			miniSteps.Notifications.Email.CANCELLED, diags = convertNotificationRecipientsToAPI(ctx, emailModel.Cancelled)
			if diags.HasError() {
				return nil, diags
			}

			miniSteps.Notifications.Email.COMPLETED, diags = convertNotificationRecipientsToAPI(ctx, emailModel.Completed)
			if diags.HasError() {
				return nil, diags
			}

			miniSteps.Notifications.Email.DRIFT_DETECTED, diags = convertNotificationRecipientsToAPI(ctx, emailModel.DriftDetected)
			if diags.HasError() {
				return nil, diags
			}

			miniSteps.Notifications.Email.ERRORED, diags = convertNotificationRecipientsToAPI(ctx, emailModel.Errored)
			if diags.HasError() {
				return nil, diags
			}
		}
	}

	// Convert Webhooks
	if !ministepsModel.Webhooks.IsNull() && !ministepsModel.Webhooks.IsUnknown() {
		var webhooksModel MinistepsWebhooksContainerModel
		diags := ministepsModel.Webhooks.As(ctx, &webhooksModel, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		if diags.HasError() {
			return nil, diags
		}

		miniSteps.Webhooks = &workflowtemplaterevisions.MinistepsWebhooks{}

		miniSteps.Webhooks.APPROVAL_REQUIRED, diags = convertWebhookToAPI(ctx, webhooksModel.ApprovalRequired)
		if diags.HasError() {
			return nil, diags
		}
		miniSteps.Webhooks.CANCELLED, diags = convertWebhookToAPI(ctx, webhooksModel.Cancelled)
		if diags.HasError() {
			return nil, diags
		}
		miniSteps.Webhooks.COMPLETED, diags = convertWebhookToAPI(ctx, webhooksModel.Completed)
		if diags.HasError() {
			return nil, diags
		}
		miniSteps.Webhooks.DRIFT_DETECTED, diags = convertWebhookToAPI(ctx, webhooksModel.DriftDetected)
		if diags.HasError() {
			return nil, diags
		}
		miniSteps.Webhooks.ERRORED, diags = convertWebhookToAPI(ctx, webhooksModel.Errored)
		if diags.HasError() {
			return nil, diags
		}
	}

	// Convert WfChaining
	if !ministepsModel.WfChaining.IsNull() && !ministepsModel.WfChaining.IsUnknown() {
		var wfChainingModel MinistepsWfChainingContainerModel
		diags := ministepsModel.WfChaining.As(ctx, &wfChainingModel, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		if diags.HasError() {
			return nil, diags
		}

		miniSteps.WfChaining = &workflowtemplaterevisions.MinistepsWorkflowChaining{}

		miniSteps.WfChaining.COMPLETED, diags = convertWorkflowChainingToAPI(ctx, wfChainingModel.Completed)
		if diags.HasError() {
			return nil, diags
		}
		miniSteps.WfChaining.ERRORED, diags = convertWorkflowChainingToAPI(ctx, wfChainingModel.Errored)
		if diags.HasError() {
			return nil, diags
		}
	}

	return miniSteps, nil
}

// Helper function to convert notification recipients to API
func convertNotificationRecipientsToAPI(ctx context.Context, recepientsObj types.List) ([]workflowtemplaterevisions.MinistepsNotificationRecepients, diag.Diagnostics) {
	if recepientsObj.IsNull() || recepientsObj.IsUnknown() {
		return nil, nil
	}

	var recepientsListModel []MinistepsNotificationRecipientsModel
	diags := recepientsObj.ElementsAs(ctx, &recepientsListModel, true)
	if diags.HasError() {
		return nil, diags
	}

	notificationRecepients := []workflowtemplaterevisions.MinistepsNotificationRecepients{}
	for _, recepientsModel := range recepientsListModel {
		recepients, diags := expanders.StringList(ctx, recepientsModel.Recipients)
		if diags.HasError() {
			return nil, diags
		}

		notificationRecepients = append(notificationRecepients, workflowtemplaterevisions.MinistepsNotificationRecepients{Recipients: recepients})
	}

	return notificationRecepients, nil
}

// Helper function to convert webhook to API
func convertWebhookToAPI(ctx context.Context, webhookObj types.List) ([]workflowtemplaterevisions.MinistepsWebhooksSchema, diag.Diagnostics) {
	if webhookObj.IsNull() || webhookObj.IsUnknown() {
		return nil, nil
	}

	var webhooksModel []MinistepsWebhooksModel
	diags := webhookObj.ElementsAs(ctx, &webhooksModel, true)
	if diags.HasError() {
		return nil, diags
	}

	var ministepsWebhooksList []workflowtemplaterevisions.MinistepsWebhooksSchema
	for _, webhooksList := range webhooksModel {
		webhookAPIModel := workflowtemplaterevisions.MinistepsWebhooksSchema{
			WebhookName:   webhooksList.WebhookName.ValueString(),
			WebhookUrl:    webhooksList.WebhookUrl.ValueString(),
			WebhookSecret: webhooksList.WebhookSecret.ValueStringPointer(),
		}

		ministepsWebhooksList = append(ministepsWebhooksList, webhookAPIModel)
	}

	return ministepsWebhooksList, nil
}

// Helper function to convert workflow chaining to API
func convertWorkflowChainingToAPI(ctx context.Context, chainingObj types.List) ([]workflowtemplaterevisions.MinistepsWfChainingSchema, diag.Diagnostics) {
	if chainingObj.IsNull() || chainingObj.IsUnknown() {
		return nil, nil
	}

	var chainingListModel []MinistepsWorkflowChainingModel
	diags := chainingObj.ElementsAs(ctx, &chainingListModel, true)
	if diags.HasError() {
		return nil, diags
	}

	var wfChainingAPIModel []workflowtemplaterevisions.MinistepsWfChainingSchema
	for _, chainingModel := range chainingListModel {
		wfChainingAPIModel = append(wfChainingAPIModel, workflowtemplaterevisions.MinistepsWfChainingSchema{
			WorkflowGroupId:    chainingModel.WorkflowGroupId.ValueString(),
			StackId:            chainingModel.StackId.ValueStringPointer(),
			StackRunPayload:    chainingModel.StackRunPayload.ValueStringPointer(),
			WorkflowId:         chainingModel.WorkflowId.ValueStringPointer(),
			WorkflowRunPayload: chainingModel.WorkflowRunPayload.ValueStringPointer(),
		})
	}

	return wfChainingAPIModel, nil
}

func parseJSONToMap(jsonStr string) map[string]interface{} {
	var result map[string]interface{}
	if jsonStr == "" {
		return result
	}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return make(map[string]interface{})
	}
	return result
}

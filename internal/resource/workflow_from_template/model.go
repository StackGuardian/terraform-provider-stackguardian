package workflowfromtemplate

import (
	"context"
	"strings"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgworkflows "github.com/StackGuardian/sg-sdk-go/workflows"
	"github.com/StackGuardian/sg-sdk-go/workflowtemplaterevisions"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/expanders"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/flatteners"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ---------------------------------------------------------------------------
// Top-level model
// ---------------------------------------------------------------------------

type WorkflowUsingTemplateResourceModel struct {
	Id                        types.String `tfsdk:"id"`
	WorkflowGroupId           types.String `tfsdk:"workflow_group_id"`
	ResourceName              types.String `tfsdk:"resource_name"`
	Description               types.String `tfsdk:"description"`
	WfType                    types.String `tfsdk:"wf_type"`
	EnvironmentVariables      types.List   `tfsdk:"environment_variables"`
	MiniSteps                 types.Object `tfsdk:"mini_steps"`
	RunnerConstraints         types.Object `tfsdk:"runner_constraints"`
	Tags                      types.List   `tfsdk:"tags"`
	UserSchedules             types.List   `tfsdk:"user_schedules"`
	ContextTags               types.Map    `tfsdk:"context_tags"`
	Approvers                 types.List   `tfsdk:"approvers"`
	NumberOfApprovalsRequired types.Int64  `tfsdk:"number_of_approvals_required"`
	UserJobCpu                types.Int64  `tfsdk:"user_job_cpu"`
	UserJobMemory             types.Int64  `tfsdk:"user_job_memory"`
	VcsConfig                 types.Object `tfsdk:"vcs_config"`
	TerraformConfig           types.Object `tfsdk:"terraform_config"`
	DeploymentPlatformConfig  types.List   `tfsdk:"deployment_platform_config"`
	WfStepsConfig             types.List   `tfsdk:"wf_steps_config"`
}

func (m WorkflowUsingTemplateResourceModel) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"id":                           types.StringType,
		"workflow_group_id":            types.StringType,
		"resource_name":                types.StringType,
		"description":                  types.StringType,
		"wf_type":                      types.StringType,
		"environment_variables":        types.ListType{ElemType: types.ObjectType{AttrTypes: EnvironmentVariableModel{}.AttributeTypes()}},
		"mini_steps":                   types.ObjectType{AttrTypes: MinistepsModel{}.AttributeTypes()},
		"runner_constraints":           types.ObjectType{AttrTypes: RunnerConstraintsModel{}.AttributeTypes()},
		"tags":                         types.ListType{ElemType: types.StringType},
		"user_schedules":               types.ListType{ElemType: types.ObjectType{AttrTypes: UserSchedulesModel{}.AttributeTypes()}},
		"context_tags":                 types.MapType{ElemType: types.StringType},
		"approvers":                    types.ListType{ElemType: types.StringType},
		"number_of_approvals_required": types.Int64Type,
		"user_job_cpu":                 types.Int64Type,
		"user_job_memory":              types.Int64Type,
		"vcs_config":                   types.ObjectType{AttrTypes: VcsConfigModel{}.AttributeTypes(ctx)},
		"terraform_config":             types.ObjectType{AttrTypes: TerraformConfigModel{}.AttributeTypes()},
		"deployment_platform_config":   types.ListType{ElemType: types.ObjectType{AttrTypes: DeploymentPlatformConfigModel{}.AttributeTypes()}},
		"wf_steps_config":              types.ListType{ElemType: types.ObjectType{AttrTypes: WfStepsConfigModel{}.AttributeTypes()}},
	}
}

// ---------------------------------------------------------------------------
// Environment variables
// ---------------------------------------------------------------------------

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

func (m EnvironmentVariableConfigModel) ToAPIModel() *sgsdkgo.EnvVarConfig {
	return &sgsdkgo.EnvVarConfig{
		VarName:   m.VarName.ValueString(),
		SecretId:  m.SecretId.ValueStringPointer(),
		TextValue: m.TextValue.ValueStringPointer(),
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

func (m EnvironmentVariableModel) ToAPIModel(ctx context.Context) (sgsdkgo.EnvVars, diag.Diagnostics) {
	var configModel EnvironmentVariableConfigModel
	diags := m.Config.As(ctx, &configModel, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return sgsdkgo.EnvVars{}, diags
	}
	return sgsdkgo.EnvVars{
		Kind:   sgsdkgo.EnvVarsKindEnum(m.Kind.ValueString()),
		Config: configModel.ToAPIModel(),
	}, nil
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

func (m MinistepsNotificationRecipientsModel) ToAPIModel(ctx context.Context) (workflowtemplaterevisions.MinistepsNotificationRecepients, diag.Diagnostics) {
	recipients, diags := expanders.StringList(ctx, m.Recipients)
	if diags.HasError() {
		return workflowtemplaterevisions.MinistepsNotificationRecepients{}, diags
	}
	return workflowtemplaterevisions.MinistepsNotificationRecepients{Recipients: recipients}, nil
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

func (m MinistepsWebhooksModel) ToAPIModel() workflowtemplaterevisions.MinistepsWebhooksSchema {
	return workflowtemplaterevisions.MinistepsWebhooksSchema{
		WebhookName:   m.WebhookName.ValueString(),
		WebhookUrl:    m.WebhookUrl.ValueString(),
		WebhookSecret: m.WebhookSecret.ValueStringPointer(),
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

func (m MinistepsWorkflowChainingModel) ToAPIModel() workflowtemplaterevisions.MinistepsWfChainingSchema {
	entry := workflowtemplaterevisions.MinistepsWfChainingSchema{
		WorkflowGroupId: m.WorkflowGroupId.ValueString(),
		StackId:         m.StackId.ValueStringPointer(),
		WorkflowId:      m.WorkflowId.ValueStringPointer(),
	}
	if s := m.WorkflowRunPayload.ValueString(); s != "" {
		entry.WorkflowRunPayload = expanders.JSONStringToInterface(s)
	}
	if s := m.StackRunPayload.ValueString(); s != "" {
		entry.StackRunPayload = expanders.JSONStringToInterface(s)
	}
	return entry
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

func (m MinistepsEmailModel) ToAPIModel(ctx context.Context) (*workflowtemplaterevisions.MinistepsNotificationsEmail, diag.Diagnostics) {
	email := &workflowtemplaterevisions.MinistepsNotificationsEmail{}
	var diags diag.Diagnostics

	email.APPROVAL_REQUIRED, diags = convertNotificationRecipientsToAPI(ctx, m.ApprovalRequired)
	if diags.HasError() {
		return nil, diags
	}
	email.CANCELLED, diags = convertNotificationRecipientsToAPI(ctx, m.Cancelled)
	if diags.HasError() {
		return nil, diags
	}
	email.COMPLETED, diags = convertNotificationRecipientsToAPI(ctx, m.Completed)
	if diags.HasError() {
		return nil, diags
	}
	email.DRIFT_DETECTED, diags = convertNotificationRecipientsToAPI(ctx, m.DriftDetected)
	if diags.HasError() {
		return nil, diags
	}
	email.ERRORED, diags = convertNotificationRecipientsToAPI(ctx, m.Errored)
	if diags.HasError() {
		return nil, diags
	}
	return email, nil
}

type MinistepsNotificationsModel struct {
	Email types.Object `tfsdk:"email"`
}

func (MinistepsNotificationsModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"email": types.ObjectType{AttrTypes: MinistepsEmailModel{}.AttributeTypes()},
	}
}

func (m MinistepsNotificationsModel) ToAPIModel(ctx context.Context) (*workflowtemplaterevisions.MinistepsNotifications, diag.Diagnostics) {
	notif := &workflowtemplaterevisions.MinistepsNotifications{}
	if m.Email.IsNull() || m.Email.IsUnknown() {
		return notif, nil
	}
	var emailModel MinistepsEmailModel
	diags := m.Email.As(ctx, &emailModel, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return nil, diags
	}
	email, diags := emailModel.ToAPIModel(ctx)
	if diags.HasError() {
		return nil, diags
	}
	notif.Email = email
	return notif, nil
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

func (m MinistepsWebhooksContainerModel) ToAPIModel(ctx context.Context) (*workflowtemplaterevisions.MinistepsWebhooks, diag.Diagnostics) {
	webhooks := &workflowtemplaterevisions.MinistepsWebhooks{}
	var diags diag.Diagnostics

	webhooks.APPROVAL_REQUIRED, diags = convertWebhookToAPI(ctx, m.ApprovalRequired)
	if diags.HasError() {
		return nil, diags
	}
	webhooks.CANCELLED, diags = convertWebhookToAPI(ctx, m.Cancelled)
	if diags.HasError() {
		return nil, diags
	}
	webhooks.COMPLETED, diags = convertWebhookToAPI(ctx, m.Completed)
	if diags.HasError() {
		return nil, diags
	}
	webhooks.DRIFT_DETECTED, diags = convertWebhookToAPI(ctx, m.DriftDetected)
	if diags.HasError() {
		return nil, diags
	}
	webhooks.ERRORED, diags = convertWebhookToAPI(ctx, m.Errored)
	if diags.HasError() {
		return nil, diags
	}
	return webhooks, nil
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

func (m MinistepsWfChainingContainerModel) ToAPIModel(ctx context.Context) (*workflowtemplaterevisions.MinistepsWorkflowChaining, diag.Diagnostics) {
	chaining := &workflowtemplaterevisions.MinistepsWorkflowChaining{}
	var diags diag.Diagnostics

	chaining.COMPLETED, diags = convertWorkflowChainingToAPI(ctx, m.Completed)
	if diags.HasError() {
		return nil, diags
	}
	chaining.ERRORED, diags = convertWorkflowChainingToAPI(ctx, m.Errored)
	if diags.HasError() {
		return nil, diags
	}
	return chaining, nil
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

func (m MinistepsModel) ToAPIModel(ctx context.Context) (*workflowtemplaterevisions.Ministeps, diag.Diagnostics) {
	miniSteps := &workflowtemplaterevisions.Ministeps{}

	if !m.Notifications.IsNull() && !m.Notifications.IsUnknown() {
		var notifModel MinistepsNotificationsModel
		diags := m.Notifications.As(ctx, &notifModel, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		if diags.HasError() {
			return nil, diags
		}
		notif, diags := notifModel.ToAPIModel(ctx)
		if diags.HasError() {
			return nil, diags
		}
		miniSteps.Notifications = notif
	}

	if !m.Webhooks.IsNull() && !m.Webhooks.IsUnknown() {
		var webhooksModel MinistepsWebhooksContainerModel
		diags := m.Webhooks.As(ctx, &webhooksModel, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		if diags.HasError() {
			return nil, diags
		}
		webhooks, diags := webhooksModel.ToAPIModel(ctx)
		if diags.HasError() {
			return nil, diags
		}
		miniSteps.Webhooks = webhooks
	}

	if !m.WfChaining.IsNull() && !m.WfChaining.IsUnknown() {
		var chainingModel MinistepsWfChainingContainerModel
		diags := m.WfChaining.As(ctx, &chainingModel, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		if diags.HasError() {
			return nil, diags
		}
		chaining, diags := chainingModel.ToAPIModel(ctx)
		if diags.HasError() {
			return nil, diags
		}
		miniSteps.WfChaining = chaining
	}

	return miniSteps, nil
}

// ---------------------------------------------------------------------------
// RunnerConstraints
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

func (m RunnerConstraintsModel) ToAPIModel(ctx context.Context) (*sgsdkgo.RunnerConstraints, diag.Diagnostics) {
	names, diags := expanders.StringList(ctx, m.Names)
	if diags.HasError() {
		return nil, diags
	}
	return &sgsdkgo.RunnerConstraints{
		Type:  (*sgsdkgo.RunnerConstraintsTypeEnum)(m.Type.ValueStringPointer()),
		Names: names,
	}, nil
}

// ---------------------------------------------------------------------------
// UserSchedules
// ---------------------------------------------------------------------------

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

func (m UserSchedulesModel) ToAPIModel() sgsdkgo.UserSchedules {
	state := sgsdkgo.StateEnum(m.State.ValueString())
	return sgsdkgo.UserSchedules{
		Cron:  m.Cron.ValueStringPointer(),
		State: &state,
		Desc:  m.Desc.ValueStringPointer(),
		Name:  m.Name.ValueStringPointer(),
	}
}

// ---------------------------------------------------------------------------
// VcsConfig — key difference: IacVcsConfigModel uses IacTemplateId, not CustomSource
// ---------------------------------------------------------------------------

type IacInputDataModel struct {
	SchemaId   types.String `tfsdk:"schema_id"`
	SchemaType types.String `tfsdk:"schema_type"`
	Data       types.String `tfsdk:"data"`
}

func (IacInputDataModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"schema_id":   types.StringType,
		"schema_type": types.StringType,
		"data":        types.StringType,
	}
}

func (m IacInputDataModel) ToAPIModel() *sgsdkgo.IacInputData {
	return &sgsdkgo.IacInputData{
		SchemaId:   m.SchemaId.ValueStringPointer(),
		SchemaType: sgsdkgo.IacInputDataSchemaTypeEnum(m.SchemaType.ValueString()).Ptr(),
		Data:       expanders.JSONStringToMap(m.Data.ValueString()),
	}
}

type IacVcsConfigModel struct {
	IacTemplateId types.String `tfsdk:"iac_template_id"`
}

func (IacVcsConfigModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"iac_template_id": types.StringType,
	}
}

func (m IacVcsConfigModel) ToAPIModel() *sgsdkgo.IacvcsConfig {
	return &sgsdkgo.IacvcsConfig{
		UseMarketplaceTemplate: expanders.BoolPtr(true),
		IacTemplateId:          m.IacTemplateId.ValueStringPointer(),
	}
}

type VcsConfigModel struct {
	IacVcsConfig types.Object `tfsdk:"iac_vcs_config"`
	IacInputData types.Object `tfsdk:"iac_input_data"`
}

func (m VcsConfigModel) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"iac_vcs_config": types.ObjectType{AttrTypes: IacVcsConfigModel{}.AttributeTypes()},
		"iac_input_data": types.ObjectType{AttrTypes: IacInputDataModel{}.AttributeTypes()},
	}
}

func (m VcsConfigModel) ToAPIModel(ctx context.Context) (*sgsdkgo.VcsConfig, diag.Diagnostics) {
	result := &sgsdkgo.VcsConfig{}

	if !m.IacVcsConfig.IsNull() && !m.IacVcsConfig.IsUnknown() {
		var iacVcsModel IacVcsConfigModel
		diags := m.IacVcsConfig.As(ctx, &iacVcsModel, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, diags
		}
		result.IacVcsConfig = iacVcsModel.ToAPIModel()
	}

	if !m.IacInputData.IsNull() && !m.IacInputData.IsUnknown() {
		var iacInputDataModel IacInputDataModel
		diags := m.IacInputData.As(ctx, &iacInputDataModel, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, diags
		}
		result.IacInputData = iacInputDataModel.ToAPIModel()
	}

	return result, nil
}

// ---------------------------------------------------------------------------
// TerraformConfig
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

func (m MountPointModel) ToAPIModel() sgsdkgo.MountPoint {
	return sgsdkgo.MountPoint{
		Source:   m.Source.ValueString(),
		Target:   m.Target.ValueString(),
		ReadOnly: m.ReadOnly.ValueBoolPointer(),
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

func (m WfStepInputDataModel) ToAPIModel() (*sgsdkgo.WfStepInputData, diag.Diagnostics) {
	schemaType, err := sgsdkgo.NewWfStepInputDataSchemaTypeEnumFromString(m.SchemaType.ValueString())
	if err != nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Invalid schema type", "The provided schema type is invalid: "+err.Error())}
	}
	return &sgsdkgo.WfStepInputData{
		SchemaType: schemaType.Ptr(),
		Data:       expanders.JSONStringToMap(m.Data.ValueString()),
	}, nil
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

func (m WfStepsConfigModel) ToAPIModel(ctx context.Context) (*sgsdkgo.WfStepsConfig, diag.Diagnostics) {
	result := sgsdkgo.WfStepsConfig{
		Name:             m.Name.ValueStringPointer(),
		Approval:         m.Approval.ValueBoolPointer(),
		Timeout:          expanders.IntPtr(m.Timeout.ValueInt64Pointer()),
		WfStepTemplateId: m.WfStepTemplateId.ValueStringPointer(),
		CmdOverride:      m.CmdOverride.ValueStringPointer(),
	}

	envVars, diags := convertEnvironmentVariablesToAPI(ctx, m.EnvironmentVariables)
	if diags.HasError() {
		return nil, diags
	}
	result.EnvironmentVariables = envVars

	if !m.MountPoints.IsNull() && !m.MountPoints.IsUnknown() {
		mountPoints, diags := convertMountPointsToAPI(ctx, m.MountPoints)
		if diags.HasError() {
			return nil, diags
		}
		result.MountPoints = mountPoints
	}

	if !m.WfStepInputData.IsNull() && !m.WfStepInputData.IsUnknown() {
		var inputDataModel WfStepInputDataModel
		diags := m.WfStepInputData.As(ctx, &inputDataModel, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		if diags.HasError() {
			return nil, diags
		}
		wfStepInputData, diags := inputDataModel.ToAPIModel()
		if diags.HasError() {
			return nil, diags
		}
		result.WfStepInputData = wfStepInputData
	}

	return &result, nil
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
	RunPrePlanHooksOnDrift  types.Bool   `tfsdk:"run_pre_plan_hooks_on_drift"`
	RunPostPlanHooksOnDrift types.Bool   `tfsdk:"run_post_plan_hooks_on_drift"`
}

func (TerraformConfigModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"terraform_version":            types.StringType,
		"drift_check":                  types.BoolType,
		"drift_cron":                   types.StringType,
		"managed_terraform_state":      types.BoolType,
		"approval_pre_apply":           types.BoolType,
		"terraform_plan_options":       types.StringType,
		"terraform_init_options":       types.StringType,
		"terraform_bin_path":           types.ListType{ElemType: types.ObjectType{AttrTypes: MountPointModel{}.AttributeTypes()}},
		"timeout":                      types.Int64Type,
		"post_apply_wf_steps_config":   types.ListType{ElemType: types.ObjectType{AttrTypes: WfStepsConfigModel{}.AttributeTypes()}},
		"pre_apply_wf_steps_config":    types.ListType{ElemType: types.ObjectType{AttrTypes: WfStepsConfigModel{}.AttributeTypes()}},
		"pre_plan_wf_steps_config":     types.ListType{ElemType: types.ObjectType{AttrTypes: WfStepsConfigModel{}.AttributeTypes()}},
		"post_plan_wf_steps_config":    types.ListType{ElemType: types.ObjectType{AttrTypes: WfStepsConfigModel{}.AttributeTypes()}},
		"pre_init_hooks":               types.ListType{ElemType: types.StringType},
		"pre_plan_hooks":               types.ListType{ElemType: types.StringType},
		"post_plan_hooks":              types.ListType{ElemType: types.StringType},
		"pre_apply_hooks":              types.ListType{ElemType: types.StringType},
		"post_apply_hooks":             types.ListType{ElemType: types.StringType},
		"run_pre_init_hooks_on_drift":  types.BoolType,
		"run_pre_plan_hooks_on_drift":  types.BoolType,
		"run_post_plan_hooks_on_drift": types.BoolType,
	}
}

// isSet reports whether a types.String holds a real (non-null, non-unknown) value.
func isSet(s types.String) bool { return !s.IsNull() && !s.IsUnknown() }

// isNonEmpty reports whether a types.String holds a real, non-empty value. Used for
// allow_blank=False API fields where an empty string means "unset" and must be omitted.
func isNonEmpty(s types.String) bool { return isSet(s) && s.ValueString() != "" }

// isSetBool reports whether a types.Bool holds a real (non-null, non-unknown) value.
func isSetBool(b types.Bool) bool { return !b.IsNull() && !b.IsUnknown() }

func (m TerraformConfigModel) ToAPIModel(ctx context.Context) (*sgsdkgo.TerraformConfig, diag.Diagnostics) {
	// Each field is Optional+Computed: an attribute the user omitted is null/unknown.
	// Guard against both so the field stays nil (rather than &"" / &false from a bare
	// ValueXxxPointer on an unknown value) — nil lets mergeTerraformConfig fill it from
	// the template instead of sending a blank the API rejects (e.g. driftCron).
	cfg := &sgsdkgo.TerraformConfig{}
	if !m.Timeout.IsNull() && !m.Timeout.IsUnknown() {
		cfg.Timeout = expanders.IntPtr(m.Timeout.ValueInt64Pointer())
	}
	// For allow_blank=False string fields, treat empty string as unset (omit) — a known
	// "" stored for plan stability must not be sent as a blank the API rejects.
	if isNonEmpty(m.TerraformVersion) {
		cfg.TerraformVersion = m.TerraformVersion.ValueStringPointer()
	}
	if isSetBool(m.DriftCheck) {
		cfg.DriftCheck = m.DriftCheck.ValueBoolPointer()
	}
	if isNonEmpty(m.DriftCron) {
		cfg.DriftCron = m.DriftCron.ValueStringPointer()
	}
	if isSetBool(m.ManagedTerraformState) {
		cfg.ManagedTerraformState = m.ManagedTerraformState.ValueBoolPointer()
	}
	if isSetBool(m.ApprovalPreApply) {
		cfg.ApprovalPreApply = m.ApprovalPreApply.ValueBoolPointer()
	}
	if isNonEmpty(m.TerraformPlanOptions) {
		cfg.TerraformPlanOptions = m.TerraformPlanOptions.ValueStringPointer()
	}
	if isNonEmpty(m.TerraformInitOptions) {
		cfg.TerraformInitOptions = m.TerraformInitOptions.ValueStringPointer()
	}
	if !m.RunPreInitHooksOnDrift.IsNull() && !m.RunPreInitHooksOnDrift.IsUnknown() {
		cfg.RunPreInitHooksOnDrift = m.RunPreInitHooksOnDrift.ValueBoolPointer()
	}
	if !m.RunPrePlanHooksOnDrift.IsNull() && !m.RunPrePlanHooksOnDrift.IsUnknown() {
		cfg.RunPrePlanHooksOnDrift = m.RunPrePlanHooksOnDrift.ValueBoolPointer()
	}
	if !m.RunPostPlanHooksOnDrift.IsNull() && !m.RunPostPlanHooksOnDrift.IsUnknown() {
		cfg.RunPostPlanHooksOnDrift = m.RunPostPlanHooksOnDrift.ValueBoolPointer()
	}

	if !m.TerraformBinPath.IsNull() && !m.TerraformBinPath.IsUnknown() {
		mountPoints, diags := convertMountPointsToAPI(ctx, m.TerraformBinPath)
		if diags.HasError() {
			return nil, diags
		}
		cfg.TerraformBinPath = mountPoints
	}

	for _, pair := range []struct {
		src  types.List
		dest *[]sgsdkgo.WfStepsConfig
	}{
		{m.PostApplyWfStepsConfig, &cfg.PostApplyWfStepsConfig},
		{m.PreApplyWfStepsConfig, &cfg.PreApplyWfStepsConfig},
		{m.PrePlanWfStepsConfig, &cfg.PrePlanWfStepsConfig},
		{m.PostPlanWfStepsConfig, &cfg.PostPlanWfStepsConfig},
	} {
		if !pair.src.IsNull() && !pair.src.IsUnknown() {
			steps, diags := convertWfStepsConfigListToAPI(ctx, pair.src)
			if diags.HasError() {
				return nil, diags
			}
			*pair.dest = steps
		}
	}

	for _, pair := range []struct {
		src  types.List
		dest *[]string
	}{
		{m.PreInitHooks, &cfg.PreInitHooks},
		{m.PrePlanHooks, &cfg.PrePlanHooks},
		{m.PostPlanHooks, &cfg.PostPlanHooks},
		{m.PreApplyHooks, &cfg.PreApplyHooks},
		{m.PostApplyHooks, &cfg.PostApplyHooks},
	} {
		if !pair.src.IsNull() && !pair.src.IsUnknown() {
			hooks, diags := expanders.StringList(ctx, pair.src)
			if diags.HasError() {
				return nil, diags
			}
			*pair.dest = hooks
		}
	}

	return cfg, nil
}

// ---------------------------------------------------------------------------
// DeploymentPlatformConfig
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

func (m DeploymentPlatformConfigConfigModel) ToAPIModel() workflowtemplaterevisions.DeploymentPlatformConfigConfig {
	cfg := workflowtemplaterevisions.DeploymentPlatformConfigConfig{}
	if !m.IntegrationId.IsNull() {
		cfg.IntegrationId = m.IntegrationId.ValueString()
	}
	if !m.ProfileName.IsNull() {
		cfg.ProfileName = m.ProfileName.ValueStringPointer()
	}
	return cfg
}

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

func (m DeploymentPlatformConfigModel) ToAPIModel(ctx context.Context) (*workflowtemplaterevisions.DeploymentPlatformConfig, diag.Diagnostics) {
	cfg := &workflowtemplaterevisions.DeploymentPlatformConfig{}
	if !m.Kind.IsNull() {
		cfg.Kind = workflowtemplaterevisions.DeploymentPlatformConfigKindEnum(m.Kind.ValueString())
	}
	if !m.Config.IsNull() && !m.Config.IsUnknown() {
		var configModel DeploymentPlatformConfigConfigModel
		diags := m.Config.As(ctx, &configModel, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		if diags.HasError() {
			return nil, diags
		}
		cfg.Config = configModel.ToAPIModel()
	}
	return cfg, nil
}

// ---------------------------------------------------------------------------
// ToAPIModel
// ---------------------------------------------------------------------------

func (m WorkflowUsingTemplateResourceModel) ToAPIModel(ctx context.Context, tpl *workflowtemplaterevisions.ReadWorkflowTemplateRevisionModel) (*sgworkflows.Workflow, diag.Diagnostics) {
	tags, diags := expanders.StringList(ctx, m.Tags)
	if diags.HasError() {
		return nil, diags
	}

	approvers, diags := expanders.StringList(ctx, m.Approvers)
	if diags.HasError() {
		return nil, diags
	}

	contextTagsMap, diags := expanders.MapStringString(ctx, m.ContextTags)
	if diags.HasError() {
		return nil, diags
	}

	envVars, diags := convertEnvironmentVariablesToAPI(ctx, m.EnvironmentVariables)
	if diags.HasError() {
		return nil, diags
	}
	envVarPtrs := make([]*sgsdkgo.EnvVars, len(envVars))
	for i := range envVars {
		envVarPtrs[i] = &envVars[i]
	}

	terraformConfig, diags := convertTerraformConfigToAPI(ctx, m.TerraformConfig)
	if diags.HasError() {
		return nil, diags
	}

	runnerConstraints, diags := convertRunnerConstraintsToAPI(ctx, m.RunnerConstraints)
	if diags.HasError() {
		return nil, diags
	}

	wfStepsConfig, diags := convertWfStepsConfigListToAPI(ctx, m.WfStepsConfig)
	if diags.HasError() {
		return nil, diags
	}
	wfStepsConfigPtrs := make([]*sgsdkgo.WfStepsConfig, len(wfStepsConfig))
	for i := range wfStepsConfig {
		wfStepsConfigPtrs[i] = &wfStepsConfig[i]
	}

	miniSteps, diags := convertMinistepsToAPI(ctx, m.MiniSteps)
	if diags.HasError() {
		return nil, diags
	}

	userSchedules, diags := convertUserSchedulesToAPI(ctx, m.UserSchedules)
	if diags.HasError() {
		return nil, diags
	}

	deploymentPlatformConfig, diags := convertDeploymentPlatformConfigToAPI(ctx, m.DeploymentPlatformConfig)
	if diags.HasError() {
		return nil, diags
	}

	vcsConfig, diags := convertVcsConfigToAPI(ctx, m.VcsConfig)
	if diags.HasError() {
		return nil, diags
	}

	var wfType *sgsdkgo.WfTypeEnum
	if !m.WfType.IsNull() && !m.WfType.IsUnknown() {
		t := sgsdkgo.WfTypeEnum(m.WfType.ValueString())
		wfType = &t
	}

	var numberOfApprovalsRequired *int
	if !m.NumberOfApprovalsRequired.IsNull() && !m.NumberOfApprovalsRequired.IsUnknown() {
		v := int(m.NumberOfApprovalsRequired.ValueInt64())
		numberOfApprovalsRequired = &v
	}

	var userJobCpu *int
	if !m.UserJobCpu.IsNull() && !m.UserJobCpu.IsUnknown() {
		v := int(m.UserJobCpu.ValueInt64())
		userJobCpu = &v
	}

	var userJobMemory *int
	if !m.UserJobMemory.IsNull() && !m.UserJobMemory.IsUnknown() {
		v := int(m.UserJobMemory.ValueInt64())
		userJobMemory = &v
	}

	var resourceName *string
	if !m.ResourceName.IsNull() && !m.ResourceName.IsUnknown() {
		resourceName = m.ResourceName.ValueStringPointer()
	}

	wf := &sgworkflows.Workflow{
		Id:                        m.Id.ValueStringPointer(),
		ResourceName:              resourceName,
		Description:               m.Description.ValueStringPointer(),
		WfType:                    wfType,
		Tags:                      tags,
		Approvers:                 approvers,
		NumberOfApprovalsRequired: numberOfApprovalsRequired,
		UserJobCpu:                userJobCpu,
		UserJobMemory:             userJobMemory,
		ContextTags:               contextTagsMap,
		EnvironmentVariables:      envVarPtrs,
		TerraformConfig:           terraformConfig,
		RunnerConstraints:         runnerConstraints,
		WfStepsConfig:             wfStepsConfigPtrs,
		MiniSteps:                 miniSteps,
		UserSchedules:             userSchedules,
		DeploymentPlatformConfig:  deploymentPlatformConfig,
		VcsConfig:                 vcsConfig,
	}

	// Provider-side resolution: fill any field the user did not set from the
	// workflow template revision, so config/state/reality line up field-for-field.
	mergeTemplateDefaults(wf, tpl)

	return wf, nil
}

// ---------------------------------------------------------------------------
// ToUpdateAPIModel
// ---------------------------------------------------------------------------

func (m WorkflowUsingTemplateResourceModel) ToUpdateAPIModel(ctx context.Context, tpl *workflowtemplaterevisions.ReadWorkflowTemplateRevisionModel) (*sgworkflows.PatchedWorkflow, diag.Diagnostics) {
	workflow, diags := m.ToAPIModel(ctx, tpl)
	if diags.HasError() {
		return nil, diags
	}

	patched := &sgworkflows.PatchedWorkflow{}

	if workflow.ResourceName != nil {
		patched.ResourceName = sgsdkgo.Optional(*workflow.ResourceName)
	} else {
		patched.ResourceName = sgsdkgo.Null[string]()
	}

	if workflow.Description != nil {
		patched.Description = sgsdkgo.Optional(*workflow.Description)
	} else {
		patched.Description = sgsdkgo.Null[string]()
	}

	if workflow.WfType != nil {
		patched.WfType = sgsdkgo.Optional(*workflow.WfType)
	} else {
		patched.WfType = sgsdkgo.Null[sgsdkgo.WfTypeEnum]()
	}

	if workflow.Tags != nil {
		patched.Tags = sgsdkgo.Optional(workflow.Tags)
	} else {
		patched.Tags = sgsdkgo.Null[[]string]()
	}

	if workflow.Approvers != nil {
		patched.Approvers = sgsdkgo.Optional(workflow.Approvers)
	} else {
		patched.Approvers = sgsdkgo.Null[[]string]()
	}

	if workflow.ContextTags != nil {
		patched.ContextTags = sgsdkgo.Optional(workflow.ContextTags)
	} else {
		patched.ContextTags = sgsdkgo.Null[map[string]string]()
	}

	if workflow.NumberOfApprovalsRequired != nil {
		patched.NumberOfApprovalsRequired = sgsdkgo.Optional(*workflow.NumberOfApprovalsRequired)
	} else {
		patched.NumberOfApprovalsRequired = sgsdkgo.Null[int]()
	}

	if workflow.UserJobCpu != nil {
		patched.UserJobCpu = sgsdkgo.Optional(*workflow.UserJobCpu)
	} else {
		patched.UserJobCpu = sgsdkgo.Null[int]()
	}

	if workflow.UserJobMemory != nil {
		patched.UserJobMemory = sgsdkgo.Optional(*workflow.UserJobMemory)
	} else {
		patched.UserJobMemory = sgsdkgo.Null[int]()
	}

	if workflow.EnvironmentVariables != nil {
		patched.EnvironmentVariables = sgsdkgo.Optional(workflow.EnvironmentVariables)
	} else {
		patched.EnvironmentVariables = sgsdkgo.Null[[]*sgsdkgo.EnvVars]()
	}

	if workflow.WfStepsConfig != nil {
		patched.WfStepsConfig = sgsdkgo.Optional(workflow.WfStepsConfig)
	} else {
		patched.WfStepsConfig = sgsdkgo.Null[[]*sgsdkgo.WfStepsConfig]()
	}

	if workflow.TerraformConfig != nil {
		patched.TerraformConfig = sgsdkgo.Optional(*workflow.TerraformConfig)
	} else {
		patched.TerraformConfig = sgsdkgo.Null[sgsdkgo.TerraformConfig]()
	}

	if workflow.RunnerConstraints != nil {
		patched.RunnerConstraints = sgsdkgo.Optional(*workflow.RunnerConstraints)
	} else {
		patched.RunnerConstraints = sgsdkgo.Null[sgsdkgo.RunnerConstraints]()
	}

	if workflow.VcsConfig != nil {
		patched.VcsConfig = sgsdkgo.Optional(*workflow.VcsConfig)
	} else {
		patched.VcsConfig = sgsdkgo.Null[sgsdkgo.VcsConfig]()
	}

	if workflow.MiniSteps != nil {
		patched.MiniSteps = sgsdkgo.Optional(*workflow.MiniSteps)
	} else {
		patched.MiniSteps = sgsdkgo.Null[workflowtemplaterevisions.Ministeps]()
	}

	if workflow.DeploymentPlatformConfig != nil {
		patched.DeploymentPlatformConfig = sgsdkgo.Optional(workflow.DeploymentPlatformConfig)
	} else {
		patched.DeploymentPlatformConfig = sgsdkgo.Null[[]*workflowtemplaterevisions.DeploymentPlatformConfig]()
	}

	if workflow.UserSchedules != nil {
		userSchedulesPtrs := make([]*sgsdkgo.UserSchedules, len(workflow.UserSchedules))
		for i := range workflow.UserSchedules {
			userSchedulesPtrs[i] = &workflow.UserSchedules[i]
		}
		patched.UserSchedules = sgsdkgo.Optional(userSchedulesPtrs)
	} else {
		patched.UserSchedules = sgsdkgo.Null[[]*sgsdkgo.UserSchedules]()
	}

	return patched, diags
}

// ---------------------------------------------------------------------------
// ConvertWorkflowUsingTemplateFromAPI
// ---------------------------------------------------------------------------

// ConvertWorkflowUsingTemplateFromAPI builds the final state model by mapping the
// FULL API reality (the fully-merged workflow record) onto every attribute. Because
// the provider sent a resolved payload (user config merged with template defaults),
// state now mirrors reality field-for-field — which is what lets Terraform detect
// drift. WorkflowGroupId is not part of the workflow record, so it is preserved from
// source.
func ConvertWorkflowUsingTemplateFromAPI(ctx context.Context, response *sgworkflows.WorkflowReadResponse, source WorkflowUsingTemplateResourceModel) (WorkflowUsingTemplateResourceModel, diag.Diagnostics) {
	var allDiags diag.Diagnostics

	model := WorkflowUsingTemplateResourceModel{
		WorkflowGroupId: source.WorkflowGroupId,
	}

	wf := response.Msg
	if wf == nil {
		return source, allDiags
	}

	model.Id = flatteners.StringPtr(wf.Id)
	model.ResourceName = flatteners.StringPtr(wf.ResourceName)
	// Description is Optional+Computed; store a known value (empty string when the API
	// returns none) so UseStateForUnknown holds it stable instead of re-planning as
	// "known after apply".
	model.Description = flatteners.StringPtrDefault(wf.Description)
	model.NumberOfApprovalsRequired = flatteners.Int64Ptr(wf.NumberOfApprovalsRequired)
	model.UserJobCpu = flatteners.Int64Ptr(wf.UserJobCpu)
	model.UserJobMemory = flatteners.Int64Ptr(wf.UserJobMemory)

	if wf.WfType != nil {
		model.WfType = types.StringValue(string(*wf.WfType))
	} else {
		model.WfType = source.WfType
	}

	tags, diags := flatteners.ListOfStringToTerraformList(wf.Tags)
	allDiags.Append(diags...)
	if allDiags.HasError() {
		return source, allDiags
	}
	model.Tags = knownEmptyListIfNull(tags, types.StringType)

	approvers, diags := flatteners.ListOfStringToTerraformList(wf.Approvers)
	allDiags.Append(diags...)
	if allDiags.HasError() {
		return source, allDiags
	}
	model.Approvers = knownEmptyListIfNull(approvers, types.StringType)

	contextTags, diags := flatteners.MapStringString(ctx, wf.ContextTags)
	allDiags.Append(diags...)
	if allDiags.HasError() {
		return source, allDiags
	}
	model.ContextTags = knownEmptyMapIfNull(contextTags)

	envVars := make([]sgsdkgo.EnvVars, len(wf.EnvironmentVariables))
	for i, ptr := range wf.EnvironmentVariables {
		if ptr != nil {
			envVars[i] = *ptr
		}
	}
	envVarsList, diags := convertEnvironmentVariablesFromAPI(ctx, envVars)
	allDiags.Append(diags...)
	if allDiags.HasError() {
		return source, allDiags
	}
	model.EnvironmentVariables = knownEmptyListIfNull(envVarsList, types.ObjectType{AttrTypes: EnvironmentVariableModel{}.AttributeTypes()})

	terraformConfig, diags := convertTerraformConfigFromAPI(ctx, wf.TerraformConfig)
	allDiags.Append(diags...)
	if allDiags.HasError() {
		return source, allDiags
	}
	model.TerraformConfig = terraformConfig

	runnerConstraints, diags := convertRunnerConstraintsFromAPI(ctx, wf.RunnerConstraints)
	allDiags.Append(diags...)
	if allDiags.HasError() {
		return source, allDiags
	}
	model.RunnerConstraints = runnerConstraints

	wfStepsConfig := make([]sgsdkgo.WfStepsConfig, len(wf.WfStepsConfig))
	for i, ptr := range wf.WfStepsConfig {
		if ptr != nil {
			wfStepsConfig[i] = *ptr
		}
	}
	wfStepsConfigList, diags := convertWfStepsConfigListFromAPI(ctx, wfStepsConfig)
	allDiags.Append(diags...)
	if allDiags.HasError() {
		return source, allDiags
	}
	model.WfStepsConfig = knownEmptyListIfNull(wfStepsConfigList, types.ObjectType{AttrTypes: WfStepsConfigModel{}.AttributeTypes()})

	miniSteps, diags := convertMinistepsFromAPI(ctx, wf.MiniSteps)
	allDiags.Append(diags...)
	if allDiags.HasError() {
		return source, allDiags
	}
	model.MiniSteps = knownEmptyObjectIfNull(miniSteps, MinistepsModel{}.AttributeTypes())

	userSchedules, diags := convertUserSchedulesFromAPI(ctx, wf.UserSchedules)
	allDiags.Append(diags...)
	if allDiags.HasError() {
		return source, allDiags
	}
	model.UserSchedules = knownEmptyListIfNull(userSchedules, types.ObjectType{AttrTypes: UserSchedulesModel{}.AttributeTypes()})

	deploymentPlatformConfig, diags := convertDeploymentPlatformConfigFromAPI(ctx, wf.DeploymentPlatformConfig)
	allDiags.Append(diags...)
	if allDiags.HasError() {
		return source, allDiags
	}
	model.DeploymentPlatformConfig = knownEmptyListIfNull(deploymentPlatformConfig, types.ObjectType{AttrTypes: DeploymentPlatformConfigModel{}.AttributeTypes()})

	vcsConfig, diags := convertVcsConfigFromAPI(ctx, wf.VcsConfig)
	allDiags.Append(diags...)
	if allDiags.HasError() {
		return source, allDiags
	}
	model.VcsConfig = vcsConfig

	return model, allDiags
}

// knownEmptyListIfNull returns a known empty list (of elemType) when in is null,
// otherwise returns in unchanged. Computed list attributes must hold a known value
// in state so that UseStateForUnknown engages on subsequent plans — a null value is
// skipped by that plan modifier and would otherwise re-plan as "known after apply",
// producing a perpetual diff.
func knownEmptyListIfNull(in types.List, elemType attr.Type) types.List {
	if in.IsNull() {
		return types.ListValueMust(elemType, []attr.Value{})
	}
	return in
}

// knownEmptyStringIfNull returns a known empty string when in is null, otherwise in.
// Same rationale as knownEmptyListIfNull for Computed string attributes.
func knownEmptyStringIfNull(in types.String) types.String {
	if in.IsNull() {
		return types.StringValue("")
	}
	return in
}

// knownFalseIfNull returns a known false when in is null, otherwise in. Same rationale
// as knownEmptyListIfNull for Computed bool attributes the API returns empty.
func knownFalseIfNull(in types.Bool) types.Bool {
	if in.IsNull() {
		return types.BoolValue(false)
	}
	return in
}

// knownEmptyMapIfNull is the map equivalent of knownEmptyListIfNull.
func knownEmptyMapIfNull(in types.Map) types.Map {
	if in.IsNull() {
		return types.MapValueMust(types.StringType, map[string]attr.Value{})
	}
	return in
}

// knownEmptyObjectIfNull returns a known object with all-null attributes (of attrTypes)
// when in is null, otherwise returns in unchanged. Same rationale as
// knownEmptyListIfNull: a Computed object attribute must hold a known value in state for
// UseStateForUnknown to prevent a perpetual "known after apply" diff.
func knownEmptyObjectIfNull(in types.Object, attrTypes map[string]attr.Type) types.Object {
	if in.IsNull() {
		values := make(map[string]attr.Value, len(attrTypes))
		for name, t := range attrTypes {
			values[name] = newNullValue(t)
		}
		return types.ObjectValueMust(attrTypes, values)
	}
	return in
}

// newNullValue returns a typed null attr.Value for the given attr.Type.
func newNullValue(t attr.Type) attr.Value {
	switch tt := t.(type) {
	case types.ObjectType:
		return types.ObjectNull(tt.AttrTypes)
	case types.ListType:
		return types.ListNull(tt.ElemType)
	case types.MapType:
		return types.MapNull(tt.ElemType)
	case types.SetType:
		return types.SetNull(tt.ElemType)
	case basetypes.BoolType:
		return types.BoolNull()
	case basetypes.Int64Type:
		return types.Int64Null()
	case basetypes.Float64Type:
		return types.Float64Null()
	case basetypes.NumberType:
		return types.NumberNull()
	default:
		return types.StringNull()
	}
}

// ---------------------------------------------------------------------------
// ToAPI helpers
// ---------------------------------------------------------------------------

func convertEnvironmentVariablesToAPI(ctx context.Context, list types.List) ([]sgsdkgo.EnvVars, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var models []EnvironmentVariableModel
	diags := list.ElementsAs(ctx, &models, false)
	if diags.HasError() {
		return nil, diags
	}
	result := make([]sgsdkgo.EnvVars, len(models))
	for i, m := range models {
		r, d := m.ToAPIModel(ctx)
		if d.HasError() {
			return nil, d
		}
		result[i] = r
	}
	return result, nil
}

func convertRunnerConstraintsToAPI(ctx context.Context, obj types.Object) (*sgsdkgo.RunnerConstraints, diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil
	}
	var m RunnerConstraintsModel
	diags := obj.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}
	return m.ToAPIModel(ctx)
}

func convertUserSchedulesToAPI(ctx context.Context, list types.List) ([]sgsdkgo.UserSchedules, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var models []UserSchedulesModel
	diags := list.ElementsAs(ctx, &models, false)
	if diags.HasError() {
		return nil, diags
	}
	result := make([]sgsdkgo.UserSchedules, len(models))
	for i, m := range models {
		result[i] = m.ToAPIModel()
	}
	return result, nil
}

func convertNotificationRecipientsToAPI(ctx context.Context, list types.List) ([]workflowtemplaterevisions.MinistepsNotificationRecepients, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var models []MinistepsNotificationRecipientsModel
	diags := list.ElementsAs(ctx, &models, true)
	if diags.HasError() {
		return nil, diags
	}
	result := make([]workflowtemplaterevisions.MinistepsNotificationRecepients, len(models))
	for i, m := range models {
		r, d := m.ToAPIModel(ctx)
		if d.HasError() {
			return nil, d
		}
		result[i] = r
	}
	return result, nil
}

func convertWebhookToAPI(ctx context.Context, list types.List) ([]workflowtemplaterevisions.MinistepsWebhooksSchema, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var models []MinistepsWebhooksModel
	diags := list.ElementsAs(ctx, &models, true)
	if diags.HasError() {
		return nil, diags
	}
	result := make([]workflowtemplaterevisions.MinistepsWebhooksSchema, len(models))
	for i, m := range models {
		result[i] = m.ToAPIModel()
	}
	return result, nil
}

func convertWorkflowChainingToAPI(ctx context.Context, list types.List) ([]workflowtemplaterevisions.MinistepsWfChainingSchema, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var models []MinistepsWorkflowChainingModel
	diags := list.ElementsAs(ctx, &models, true)
	if diags.HasError() {
		return nil, diags
	}
	result := make([]workflowtemplaterevisions.MinistepsWfChainingSchema, len(models))
	for i, m := range models {
		result[i] = m.ToAPIModel()
	}
	return result, nil
}

func convertMinistepsToAPI(ctx context.Context, obj types.Object) (*workflowtemplaterevisions.Ministeps, diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil
	}
	var m MinistepsModel
	diags := obj.As(ctx, &m, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return nil, diags
	}
	return m.ToAPIModel(ctx)
}

func convertMountPointsToAPI(ctx context.Context, list types.List) ([]sgsdkgo.MountPoint, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var models []MountPointModel
	diags := list.ElementsAs(ctx, &models, false)
	if diags.HasError() {
		return nil, diags
	}
	result := make([]sgsdkgo.MountPoint, len(models))
	for i, m := range models {
		result[i] = m.ToAPIModel()
	}
	return result, nil
}

func convertWfStepsConfigListToAPI(ctx context.Context, list types.List) ([]sgsdkgo.WfStepsConfig, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var models []WfStepsConfigModel
	diags := list.ElementsAs(ctx, &models, false)
	if diags.HasError() {
		return nil, diags
	}
	result := make([]sgsdkgo.WfStepsConfig, len(models))
	for i, m := range models {
		r, d := m.ToAPIModel(ctx)
		if d.HasError() {
			return nil, d
		}
		result[i] = *r
	}
	return result, nil
}

func convertTerraformConfigToAPI(ctx context.Context, obj types.Object) (*sgsdkgo.TerraformConfig, diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil
	}
	var m TerraformConfigModel
	diags := obj.As(ctx, &m, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return nil, diags
	}
	return m.ToAPIModel(ctx)
}

func convertDeploymentPlatformConfigToAPI(ctx context.Context, list types.List) ([]*workflowtemplaterevisions.DeploymentPlatformConfig, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	var models []DeploymentPlatformConfigModel
	diags := list.ElementsAs(ctx, &models, false)
	if diags.HasError() {
		return nil, diags
	}
	result := make([]*workflowtemplaterevisions.DeploymentPlatformConfig, len(models))
	for i, m := range models {
		r, d := m.ToAPIModel(ctx)
		if d.HasError() {
			return nil, d
		}
		result[i] = r
	}
	return result, nil
}

func convertVcsConfigToAPI(ctx context.Context, obj types.Object) (*sgsdkgo.VcsConfig, diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil
	}
	var m VcsConfigModel
	diags := obj.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}
	return m.ToAPIModel(ctx)
}

// ---------------------------------------------------------------------------
// FromAPI converters
// ---------------------------------------------------------------------------

func convertEnvironmentVariablesFromAPI(ctx context.Context, envVars []sgsdkgo.EnvVars) (types.List, diag.Diagnostics) {
	nullList := types.ListNull(types.ObjectType{AttrTypes: EnvironmentVariableModel{}.AttributeTypes()})
	if len(envVars) == 0 {
		return nullList, nil
	}

	models := make([]EnvironmentVariableModel, len(envVars))
	for i, envVar := range envVars {
		configModel := EnvironmentVariableConfigModel{
			VarName:   flatteners.String(envVar.Config.VarName),
			SecretId:  flatteners.StringPtr(envVar.Config.SecretId),
			TextValue: flatteners.StringPtr(envVar.Config.TextValue),
		}
		configObj, diags := types.ObjectValueFrom(ctx, EnvironmentVariableConfigModel{}.AttributeTypes(), configModel)
		if diags.HasError() {
			return nullList, diags
		}
		models[i] = EnvironmentVariableModel{
			Config: configObj,
			Kind:   flatteners.String(string(envVar.Kind)),
		}
	}

	list, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: EnvironmentVariableModel{}.AttributeTypes()}, models)
	if diags.HasError() {
		return nullList, diags
	}
	return list, nil
}

func convertRunnerConstraintsFromAPI(ctx context.Context, rc *sgsdkgo.RunnerConstraints) (types.Object, diag.Diagnostics) {
	nullObj := types.ObjectNull(RunnerConstraintsModel{}.AttributeTypes())
	if rc == nil {
		return nullObj, nil
	}

	namesList, diags := flatteners.ListOfStringToTerraformList(rc.Names)
	if diags.HasError() {
		return nullObj, diags
	}

	obj, diags := types.ObjectValueFrom(ctx, RunnerConstraintsModel{}.AttributeTypes(), RunnerConstraintsModel{
		Type:  flatteners.StringPtr((*string)(rc.Type)),
		Names: namesList,
	})
	if diags.HasError() {
		return nullObj, diags
	}
	return obj, nil
}

func convertUserSchedulesFromAPI(ctx context.Context, schedules []sgsdkgo.UserSchedules) (types.List, diag.Diagnostics) {
	nullList := types.ListNull(types.ObjectType{AttrTypes: UserSchedulesModel{}.AttributeTypes()})
	if len(schedules) == 0 {
		return nullList, nil
	}

	models := make([]UserSchedulesModel, 0, len(schedules))
	for _, s := range schedules {
		if flatteners.IsEmptyObject(s) {
			continue
		}
		models = append(models, UserSchedulesModel{
			Cron:  flatteners.StringPtr(s.Cron),
			State: flatteners.StringPtr((*string)(s.State)),
			Desc:  flatteners.StringPtr(s.Desc),
			Name:  flatteners.StringPtr(s.Name),
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

func convertNotificationRecipientsFromAPI(ctx context.Context, recipients []workflowtemplaterevisions.MinistepsNotificationRecepients) (types.List, diag.Diagnostics) {
	nullList := types.ListNull(types.ObjectType{AttrTypes: MinistepsNotificationRecipientsModel{}.AttributeTypes()})
	if len(recipients) == 0 {
		return nullList, nil
	}

	models := []MinistepsNotificationRecipientsModel{}
	for _, r := range recipients {
		if flatteners.IsEmptyObject(r) {
			continue
		}
		recipientsList, diags := types.ListValueFrom(ctx, types.StringType, r.Recipients)
		if diags.HasError() {
			return nullList, diags
		}
		models = append(models, MinistepsNotificationRecipientsModel{Recipients: recipientsList})
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

func convertWebhookFromAPI(ctx context.Context, webhooks []workflowtemplaterevisions.MinistepsWebhooksSchema) (types.List, diag.Diagnostics) {
	nullList := types.ListNull(types.ObjectType{AttrTypes: MinistepsWebhooksModel{}.AttributeTypes()})
	if len(webhooks) == 0 {
		return nullList, nil
	}

	models := []MinistepsWebhooksModel{}
	for _, w := range webhooks {
		if flatteners.IsEmptyObject(w) {
			continue
		}
		models = append(models, MinistepsWebhooksModel{
			WebhookName:   flatteners.String(w.WebhookName),
			WebhookUrl:    flatteners.String(w.WebhookUrl),
			WebhookSecret: flatteners.StringPtr(w.WebhookSecret),
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

func convertWorkflowChainingFromAPI(ctx context.Context, chainingList []workflowtemplaterevisions.MinistepsWfChainingSchema) (types.List, diag.Diagnostics) {
	nullList := types.ListNull(types.ObjectType{AttrTypes: MinistepsWorkflowChainingModel{}.AttributeTypes()})
	if len(chainingList) == 0 {
		return nullList, nil
	}

	models := []MinistepsWorkflowChainingModel{}
	for _, c := range chainingList {
		if flatteners.IsEmptyObject(c) {
			continue
		}
		models = append(models, MinistepsWorkflowChainingModel{
			WorkflowGroupId:    flatteners.String(c.WorkflowGroupId),
			StackId:            flatteners.StringPtr(c.StackId),
			WorkflowId:         flatteners.StringPtr(c.WorkflowId),
			WorkflowRunPayload: flatteners.JSONInterfaceToString(c.WorkflowRunPayload),
			StackRunPayload:    flatteners.JSONInterfaceToString(c.StackRunPayload),
		})
	}
	if len(models) == 0 {
		return nullList, nil
	}

	list, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: MinistepsWorkflowChainingModel{}.AttributeTypes()}, models)
	if diags.HasError() {
		return nullList, diags
	}
	return list, nil
}

func convertMinistepsFromAPI(ctx context.Context, ministeps *workflowtemplaterevisions.Ministeps) (types.Object, diag.Diagnostics) {
	nullObj := types.ObjectNull(MinistepsModel{}.AttributeTypes())
	if ministeps == nil || flatteners.IsEmptyObject(ministeps) {
		return nullObj, nil
	}

	model := MinistepsModel{}

	if ministeps.Notifications != nil {
		notifModel := MinistepsNotificationsModel{}

		if ministeps.Notifications.Email != nil {
			emailModel := MinistepsEmailModel{}
			var diags diag.Diagnostics

			emailModel.ApprovalRequired, diags = convertNotificationRecipientsFromAPI(ctx, ministeps.Notifications.Email.APPROVAL_REQUIRED)
			if diags.HasError() {
				return nullObj, diags
			}
			emailModel.Cancelled, diags = convertNotificationRecipientsFromAPI(ctx, ministeps.Notifications.Email.CANCELLED)
			if diags.HasError() {
				return nullObj, diags
			}
			emailModel.Completed, diags = convertNotificationRecipientsFromAPI(ctx, ministeps.Notifications.Email.COMPLETED)
			if diags.HasError() {
				return nullObj, diags
			}
			emailModel.DriftDetected, diags = convertNotificationRecipientsFromAPI(ctx, ministeps.Notifications.Email.DRIFT_DETECTED)
			if diags.HasError() {
				return nullObj, diags
			}
			emailModel.Errored, diags = convertNotificationRecipientsFromAPI(ctx, ministeps.Notifications.Email.ERRORED)
			if diags.HasError() {
				return nullObj, diags
			}

			emailObj, diags := types.ObjectValueFrom(ctx, MinistepsEmailModel{}.AttributeTypes(), emailModel)
			if diags.HasError() {
				return nullObj, diags
			}
			notifModel.Email = emailObj
		} else {
			notifModel.Email = types.ObjectNull(MinistepsEmailModel{}.AttributeTypes())
		}

		notifObj, diags := types.ObjectValueFrom(ctx, MinistepsNotificationsModel{}.AttributeTypes(), notifModel)
		if diags.HasError() {
			return nullObj, diags
		}
		model.Notifications = notifObj
	} else {
		model.Notifications = types.ObjectNull(MinistepsNotificationsModel{}.AttributeTypes())
	}

	if ministeps.Webhooks != nil {
		webhooksModel := MinistepsWebhooksContainerModel{}
		var diags diag.Diagnostics

		webhooksModel.ApprovalRequired, diags = convertWebhookFromAPI(ctx, ministeps.Webhooks.APPROVAL_REQUIRED)
		if diags.HasError() {
			return nullObj, diags
		}
		webhooksModel.Cancelled, diags = convertWebhookFromAPI(ctx, ministeps.Webhooks.CANCELLED)
		if diags.HasError() {
			return nullObj, diags
		}
		webhooksModel.Completed, diags = convertWebhookFromAPI(ctx, ministeps.Webhooks.COMPLETED)
		if diags.HasError() {
			return nullObj, diags
		}
		webhooksModel.DriftDetected, diags = convertWebhookFromAPI(ctx, ministeps.Webhooks.DRIFT_DETECTED)
		if diags.HasError() {
			return nullObj, diags
		}
		webhooksModel.Errored, diags = convertWebhookFromAPI(ctx, ministeps.Webhooks.ERRORED)
		if diags.HasError() {
			return nullObj, diags
		}

		webhooksObj, diags := types.ObjectValueFrom(ctx, MinistepsWebhooksContainerModel{}.AttributeTypes(), webhooksModel)
		if diags.HasError() {
			return nullObj, diags
		}
		model.Webhooks = webhooksObj
	} else {
		model.Webhooks = types.ObjectNull(MinistepsWebhooksContainerModel{}.AttributeTypes())
	}

	if ministeps.WfChaining != nil {
		chainingModel := MinistepsWfChainingContainerModel{}
		var diags diag.Diagnostics

		chainingModel.Completed, diags = convertWorkflowChainingFromAPI(ctx, ministeps.WfChaining.COMPLETED)
		if diags.HasError() {
			return nullObj, diags
		}
		chainingModel.Errored, diags = convertWorkflowChainingFromAPI(ctx, ministeps.WfChaining.ERRORED)
		if diags.HasError() {
			return nullObj, diags
		}

		chainingObj, diags := types.ObjectValueFrom(ctx, MinistepsWfChainingContainerModel{}.AttributeTypes(), chainingModel)
		if diags.HasError() {
			return nullObj, diags
		}
		model.WfChaining = chainingObj
	} else {
		model.WfChaining = types.ObjectNull(MinistepsWfChainingContainerModel{}.AttributeTypes())
	}

	obj, diags := types.ObjectValueFrom(ctx, MinistepsModel{}.AttributeTypes(), model)
	if diags.HasError() {
		return nullObj, diags
	}
	return obj, nil
}

func convertMountPointsFromAPI(ctx context.Context, mountPoints []sgsdkgo.MountPoint) (types.List, diag.Diagnostics) {
	nullList := types.ListNull(types.ObjectType{AttrTypes: MountPointModel{}.AttributeTypes()})
	if len(mountPoints) == 0 {
		return nullList, nil
	}

	models := make([]MountPointModel, len(mountPoints))
	for i, mp := range mountPoints {
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

func convertWfStepFromAPI(ctx context.Context, step *sgsdkgo.WfStepsConfig) (types.Object, diag.Diagnostics) {
	nullObj := types.ObjectNull(WfStepsConfigModel{}.AttributeTypes())
	if step == nil {
		return nullObj, nil
	}

	m := WfStepsConfigModel{
		Name:             flatteners.StringPtr(step.Name),
		Approval:         flatteners.BoolPtr(step.Approval),
		Timeout:          flatteners.Int64Ptr(step.Timeout),
		WfStepTemplateId: flatteners.StringPtr(step.WfStepTemplateId),
		CmdOverride:      flatteners.StringPtr(step.CmdOverride),
	}

	envVarsList, diags := convertEnvironmentVariablesFromAPI(ctx, step.EnvironmentVariables)
	if diags.HasError() {
		return nullObj, diags
	}
	m.EnvironmentVariables = envVarsList

	mountPoints, diags := convertMountPointsFromAPI(ctx, step.MountPoints)
	if diags.HasError() {
		return nullObj, diags
	}
	m.MountPoints = mountPoints

	if step.WfStepInputData != nil {
		inputDataModel := WfStepInputDataModel{
			SchemaType: flatteners.String(string(*step.WfStepInputData.SchemaType)),
			Data:       flatteners.JSONInterfaceToString(step.WfStepInputData.Data),
		}
		inputDataObj, diags := types.ObjectValueFrom(ctx, WfStepInputDataModel{}.AttributeTypes(), inputDataModel)
		if diags.HasError() {
			return nullObj, diags
		}
		m.WfStepInputData = inputDataObj
	} else {
		m.WfStepInputData = types.ObjectNull(WfStepInputDataModel{}.AttributeTypes())
	}

	obj, diags := types.ObjectValueFrom(ctx, WfStepsConfigModel{}.AttributeTypes(), m)
	if diags.HasError() {
		return nullObj, diags
	}
	return obj, nil
}

func convertWfStepsConfigListFromAPI(ctx context.Context, steps []sgsdkgo.WfStepsConfig) (types.List, diag.Diagnostics) {
	nullList := types.ListNull(types.ObjectType{AttrTypes: WfStepsConfigModel{}.AttributeTypes()})
	if len(steps) == 0 {
		return nullList, nil
	}

	models := make([]WfStepsConfigModel, len(steps))
	for i, step := range steps {
		obj, diags := convertWfStepFromAPI(ctx, &step)
		if diags.HasError() {
			return nullList, diags
		}
		var m WfStepsConfigModel
		diags = obj.As(ctx, &m, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		if diags.HasError() {
			return nullList, diags
		}
		models[i] = m
	}

	list, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: WfStepsConfigModel{}.AttributeTypes()}, models)
	if diags.HasError() {
		return nullList, diags
	}
	return list, nil
}

// normalizeTerraformVersion strips the engine prefix the API prepends to the version
// string ("TERRAFORM-1.5.0" / "OPENTOFU-1.6.0") so state stores the bare version the
// user declares (e.g. "1.5.0"). Without this, config ("1.5.0") and the API-returned
// value ("TERRAFORM-1.5.0") never match, producing a perpetual diff on terraform_version.
func normalizeTerraformVersion(v *string) types.String {
	if v == nil {
		return types.StringNull()
	}
	s := *v
	for _, prefix := range []string{"TERRAFORM-", "OPENTOFU-"} {
		if strings.HasPrefix(strings.ToUpper(s), prefix) {
			s = s[len(prefix):]
			break
		}
	}
	return flatteners.StringPtr(&s)
}

func convertTerraformConfigFromAPI(ctx context.Context, cfg *sgsdkgo.TerraformConfig) (types.Object, diag.Diagnostics) {
	nullObj := types.ObjectNull(TerraformConfigModel{}.AttributeTypes())
	if cfg == nil || flatteners.IsEmptyObject(cfg) {
		return nullObj, nil
	}

	m := TerraformConfigModel{
		TerraformVersion:        normalizeTerraformVersion(cfg.TerraformVersion),
		DriftCheck:              flatteners.BoolPtr(cfg.DriftCheck),
		DriftCron:               flatteners.StringPtr(cfg.DriftCron),
		ManagedTerraformState:   flatteners.BoolPtr(cfg.ManagedTerraformState),
		ApprovalPreApply:        flatteners.BoolPtr(cfg.ApprovalPreApply),
		TerraformPlanOptions:    flatteners.StringPtr(cfg.TerraformPlanOptions),
		TerraformInitOptions:    flatteners.StringPtr(cfg.TerraformInitOptions),
		Timeout:                 flatteners.Int64Ptr(cfg.Timeout),
		RunPreInitHooksOnDrift:  flatteners.BoolPtr(cfg.RunPreInitHooksOnDrift),
		RunPrePlanHooksOnDrift:  flatteners.BoolPtr(cfg.RunPrePlanHooksOnDrift),
		RunPostPlanHooksOnDrift: flatteners.BoolPtr(cfg.RunPostPlanHooksOnDrift),
	}

	terraformBinPath, diags := convertMountPointsFromAPI(ctx, cfg.TerraformBinPath)
	if diags.HasError() {
		return nullObj, diags
	}
	m.TerraformBinPath = terraformBinPath

	postApply, diags := convertWfStepsConfigListFromAPI(ctx, cfg.PostApplyWfStepsConfig)
	if diags.HasError() {
		return nullObj, diags
	}
	m.PostApplyWfStepsConfig = postApply

	preApply, diags := convertWfStepsConfigListFromAPI(ctx, cfg.PreApplyWfStepsConfig)
	if diags.HasError() {
		return nullObj, diags
	}
	m.PreApplyWfStepsConfig = preApply

	prePlan, diags := convertWfStepsConfigListFromAPI(ctx, cfg.PrePlanWfStepsConfig)
	if diags.HasError() {
		return nullObj, diags
	}
	m.PrePlanWfStepsConfig = prePlan

	postPlan, diags := convertWfStepsConfigListFromAPI(ctx, cfg.PostPlanWfStepsConfig)
	if diags.HasError() {
		return nullObj, diags
	}
	m.PostPlanWfStepsConfig = postPlan

	preInitHooks, diags := flatteners.ListOfStringToTerraformList(cfg.PreInitHooks)
	if diags.HasError() {
		return nullObj, diags
	}
	m.PreInitHooks = preInitHooks

	prePlanHooks, diags := flatteners.ListOfStringToTerraformList(cfg.PrePlanHooks)
	if diags.HasError() {
		return nullObj, diags
	}
	m.PrePlanHooks = prePlanHooks

	postPlanHooks, diags := flatteners.ListOfStringToTerraformList(cfg.PostPlanHooks)
	if diags.HasError() {
		return nullObj, diags
	}
	m.PostPlanHooks = postPlanHooks

	preApplyHooks, diags := flatteners.ListOfStringToTerraformList(cfg.PreApplyHooks)
	if diags.HasError() {
		return nullObj, diags
	}
	m.PreApplyHooks = preApplyHooks

	postApplyHooks, diags := flatteners.ListOfStringToTerraformList(cfg.PostApplyHooks)
	if diags.HasError() {
		return nullObj, diags
	}
	m.PostApplyHooks = postApplyHooks

	// These nested fields are Optional+Computed; the API may return them empty (→ null
	// from the flatteners above). A null value on a Computed attribute is skipped by
	// UseStateForUnknown and re-plans as "known after apply" forever, so coerce empties
	// to known values to keep plans clean.
	wfStepElem := types.ObjectType{AttrTypes: WfStepsConfigModel{}.AttributeTypes()}
	mountElem := types.ObjectType{AttrTypes: MountPointModel{}.AttributeTypes()}
	m.PreInitHooks = knownEmptyListIfNull(m.PreInitHooks, types.StringType)
	m.PrePlanHooks = knownEmptyListIfNull(m.PrePlanHooks, types.StringType)
	m.PostPlanHooks = knownEmptyListIfNull(m.PostPlanHooks, types.StringType)
	m.PreApplyHooks = knownEmptyListIfNull(m.PreApplyHooks, types.StringType)
	m.PostApplyHooks = knownEmptyListIfNull(m.PostApplyHooks, types.StringType)
	m.TerraformBinPath = knownEmptyListIfNull(m.TerraformBinPath, mountElem)
	m.PostApplyWfStepsConfig = knownEmptyListIfNull(m.PostApplyWfStepsConfig, wfStepElem)
	m.PreApplyWfStepsConfig = knownEmptyListIfNull(m.PreApplyWfStepsConfig, wfStepElem)
	m.PrePlanWfStepsConfig = knownEmptyListIfNull(m.PrePlanWfStepsConfig, wfStepElem)
	m.PostPlanWfStepsConfig = knownEmptyListIfNull(m.PostPlanWfStepsConfig, wfStepElem)

	// Scalars the API returns empty are coerced to known values for plan stability
	// (UseStateForUnknown skips null state). ToAPIModel treats an empty string / false
	// as "unset" and omits it, so a known-empty in state never produces a blank payload
	// the API rejects (e.g. driftCron is allow_blank=False).
	m.TerraformPlanOptions = knownEmptyStringIfNull(m.TerraformPlanOptions)
	m.TerraformInitOptions = knownEmptyStringIfNull(m.TerraformInitOptions)
	m.DriftCron = knownEmptyStringIfNull(m.DriftCron)
	m.DriftCheck = knownFalseIfNull(m.DriftCheck)
	m.ManagedTerraformState = knownFalseIfNull(m.ManagedTerraformState)
	m.ApprovalPreApply = knownFalseIfNull(m.ApprovalPreApply)
	m.RunPreInitHooksOnDrift = knownFalseIfNull(m.RunPreInitHooksOnDrift)
	m.RunPrePlanHooksOnDrift = knownFalseIfNull(m.RunPrePlanHooksOnDrift)
	m.RunPostPlanHooksOnDrift = knownFalseIfNull(m.RunPostPlanHooksOnDrift)
	if m.Timeout.IsNull() {
		m.Timeout = types.Int64Value(0)
	}

	obj, diags := types.ObjectValueFrom(ctx, TerraformConfigModel{}.AttributeTypes(), m)
	if diags.HasError() {
		return nullObj, diags
	}
	return obj, nil
}

func convertDeploymentPlatformConfigFromAPI(ctx context.Context, configs []*workflowtemplaterevisions.DeploymentPlatformConfig) (types.List, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: DeploymentPlatformConfigModel{}.AttributeTypes()}
	nullList := types.ListNull(elemType)
	if len(configs) == 0 {
		return nullList, nil
	}

	models := make([]DeploymentPlatformConfigModel, 0, len(configs))
	for _, cfg := range configs {
		configModel := DeploymentPlatformConfigConfigModel{
			IntegrationId: flatteners.String(cfg.Config.IntegrationId),
			ProfileName:   flatteners.StringPtr(cfg.Config.ProfileName),
		}
		configObj, diags := types.ObjectValueFrom(ctx, DeploymentPlatformConfigConfigModel{}.AttributeTypes(), configModel)
		if diags.HasError() {
			return nullList, diags
		}
		models = append(models, DeploymentPlatformConfigModel{
			Kind:   flatteners.String(string(cfg.Kind)),
			Config: configObj,
		})
	}

	list, diags := types.ListValueFrom(ctx, elemType, models)
	if diags.HasError() {
		return nullList, diags
	}
	return list, nil
}

func convertVcsConfigFromAPI(ctx context.Context, vcsConfig *sgsdkgo.VcsConfig) (types.Object, diag.Diagnostics) {
	nullObj := types.ObjectNull(VcsConfigModel{}.AttributeTypes(ctx))
	if vcsConfig == nil || flatteners.IsEmptyObject(vcsConfig) {
		return nullObj, nil
	}

	var iacVcsConfigObj types.Object
	if vcsConfig.IacVcsConfig != nil {
		iacVcsModel := IacVcsConfigModel{
			IacTemplateId: flatteners.StringPtr(vcsConfig.IacVcsConfig.IacTemplateId),
		}
		var diags diag.Diagnostics
		iacVcsConfigObj, diags = types.ObjectValueFrom(ctx, IacVcsConfigModel{}.AttributeTypes(), iacVcsModel)
		if diags.HasError() {
			return nullObj, diags
		}
	} else {
		iacVcsConfigObj = types.ObjectNull(IacVcsConfigModel{}.AttributeTypes())
	}

	var iacInputDataObj types.Object
	if vcsConfig.IacInputData != nil {
		iacInputDataModel := IacInputDataModel{
			SchemaId:   flatteners.StringPtr(vcsConfig.IacInputData.SchemaId),
			SchemaType: types.StringValue(string(*vcsConfig.IacInputData.SchemaType)),
			Data:       flatteners.JSONInterfaceToString(vcsConfig.IacInputData.Data),
		}
		var diags diag.Diagnostics
		iacInputDataObj, diags = types.ObjectValueFrom(ctx, IacInputDataModel{}.AttributeTypes(), iacInputDataModel)
		if diags.HasError() {
			return nullObj, diags
		}
	} else {
		iacInputDataObj = types.ObjectNull(IacInputDataModel{}.AttributeTypes())
	}

	vcsModel := VcsConfigModel{
		IacVcsConfig: iacVcsConfigObj,
		IacInputData: iacInputDataObj,
	}
	return types.ObjectValueFrom(ctx, VcsConfigModel{}.AttributeTypes(ctx), vcsModel)
}

// ---------------------------------------------------------------------------
// Template merge (provider-side resolution)
//
// The platform persists a workflow as a fully-merged record (user input +
// template + org + platform defaults). The user's Terraform config only carries
// the fields they declared. To make Terraform's diff engine work, the provider
// fetches the workflow template revision and merges it into the payload: any
// field the user did NOT set is filled from the template. The resulting resolved
// payload is what we send to the API and store in state, so config, state, and
// reality line up field-for-field.
//
// Org defaults are already baked into the template at template-create time, so a
// template revision carries org-resolved values — fetching the template is
// sufficient; no separate org-settings call is needed.
// ---------------------------------------------------------------------------

// mergeTemplateDefaults fills fields the user left unset on wf with the values
// from the template revision tpl. User-provided values always win.
func mergeTemplateDefaults(wf *sgworkflows.Workflow, tpl *workflowtemplaterevisions.ReadWorkflowTemplateRevisionModel) {
	if wf == nil || tpl == nil {
		return
	}

	if wf.ResourceName == nil && tpl.Alias != "" {
		alias := tpl.Alias
		wf.ResourceName = &alias
	}
	if wf.Description == nil && tpl.LongDescription != nil {
		wf.Description = tpl.LongDescription
	}
	if wf.NumberOfApprovalsRequired == nil && tpl.NumberOfApprovalsRequired != nil {
		wf.NumberOfApprovalsRequired = tpl.NumberOfApprovalsRequired
	}
	if wf.UserJobCpu == nil && tpl.UserJobCPU != nil {
		wf.UserJobCpu = tpl.UserJobCPU
	}
	if wf.UserJobMemory == nil && tpl.UserJobMemory != nil {
		wf.UserJobMemory = tpl.UserJobMemory
	}
	// TerraformConfig is deep-merged field-by-field: a field the user set wins; any
	// field the user left unset is filled from the template. (Whole-object replacement
	// would send the user's partial block with blanks the API rejects, e.g. driftCron.)
	wf.TerraformConfig = mergeTerraformConfig(wf.TerraformConfig, tpl.TerraformConfig)
	if wf.RunnerConstraints == nil && tpl.RunnerConstraints != nil {
		wf.RunnerConstraints = tpl.RunnerConstraints
	}
	if wf.MiniSteps == nil && tpl.Ministeps != nil {
		wf.MiniSteps = tpl.Ministeps
	}

	if len(wf.Tags) == 0 && len(tpl.Tags) > 0 {
		wf.Tags = tpl.Tags
	}
	if len(wf.Approvers) == 0 && len(tpl.Approvers) > 0 {
		wf.Approvers = tpl.Approvers
	}
	if len(wf.ContextTags) == 0 && len(tpl.ContextTags) > 0 {
		wf.ContextTags = tpl.ContextTags
	}
	if len(wf.UserSchedules) == 0 && len(tpl.UserSchedules) > 0 {
		schedules := make([]sgsdkgo.UserSchedules, len(tpl.UserSchedules))
		for i := range tpl.UserSchedules {
			t := tpl.UserSchedules[i]
			cron := t.Cron
			state := sgsdkgo.StateEnum(t.State)
			schedules[i] = sgsdkgo.UserSchedules{
				Name:  t.Name,
				Desc:  t.Desc,
				Cron:  &cron,
				State: &state,
			}
		}
		wf.UserSchedules = schedules
	}
	if len(wf.DeploymentPlatformConfig) == 0 && len(tpl.DeploymentPlatformConfig) > 0 {
		wf.DeploymentPlatformConfig = tpl.DeploymentPlatformConfig
	}

	// EnvironmentVariables: template stores value slice, workflow uses pointer slice.
	if len(wf.EnvironmentVariables) == 0 && len(tpl.EnvironmentVariables) > 0 {
		ptrs := make([]*sgsdkgo.EnvVars, len(tpl.EnvironmentVariables))
		for i := range tpl.EnvironmentVariables {
			ptrs[i] = &tpl.EnvironmentVariables[i]
		}
		wf.EnvironmentVariables = ptrs
	}

	// WfStepsConfig: template stores value slice, workflow uses pointer slice.
	if len(wf.WfStepsConfig) == 0 && len(tpl.WfStepsConfig) > 0 {
		ptrs := make([]*sgsdkgo.WfStepsConfig, len(tpl.WfStepsConfig))
		for i := range tpl.WfStepsConfig {
			ptrs[i] = &tpl.WfStepsConfig[i]
		}
		wf.WfStepsConfig = ptrs
	}
}

// mergeTerraformConfig deep-merges a TerraformConfig field-by-field: the user's value
// is kept when set, otherwise the template's value fills it. Returns the user config
// unchanged if the template has none, and the template's config if the user has none.
func mergeTerraformConfig(user, tpl *sgsdkgo.TerraformConfig) *sgsdkgo.TerraformConfig {
	if tpl == nil {
		return user
	}
	if user == nil {
		return tpl
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

	return user
}

// IacTemplateId extracts the workflow template revision id from the model's
// vcs_config.iac_vcs_config.iac_template_id. Returns "" if not set.
func (m WorkflowUsingTemplateResourceModel) IacTemplateId(ctx context.Context) (string, diag.Diagnostics) {
	if m.VcsConfig.IsNull() || m.VcsConfig.IsUnknown() {
		return "", nil
	}
	var vcs VcsConfigModel
	diags := m.VcsConfig.As(ctx, &vcs, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return "", diags
	}
	if vcs.IacVcsConfig.IsNull() || vcs.IacVcsConfig.IsUnknown() {
		return "", nil
	}
	var iac IacVcsConfigModel
	diags = vcs.IacVcsConfig.As(ctx, &iac, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return "", diags
	}
	return iac.IacTemplateId.ValueString(), nil
}

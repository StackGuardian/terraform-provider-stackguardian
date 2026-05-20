package workflowgit

import (
	"context"

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

type WorkflowGitResourceModel struct {
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
	VcsTriggers               types.Object `tfsdk:"vcs_triggers"`
}

func (m WorkflowGitResourceModel) AttributeTypes(ctx context.Context) map[string]attr.Type {
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
		"vcs_triggers":                 types.ObjectType{AttrTypes: VcsTriggersModel{}.AttributeTypes()},
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
// VcsTriggers
// ---------------------------------------------------------------------------

type VcsTriggerActionConfigModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

func (VcsTriggerActionConfigModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled": types.BoolType,
	}
}

type VcsTriggersModel struct {
	TrackedBranch           types.String `tfsdk:"tracked_branch"`
	ApprovalPreApply        types.Bool   `tfsdk:"approval_pre_apply"`
	PlanOnly                types.Bool   `tfsdk:"plan_only"`
	FileTriggersEnabled     types.Bool   `tfsdk:"file_triggers_enabled"`
	FileTriggerPatterns     types.List   `tfsdk:"file_trigger_patterns"`
	GhWebhookUrl            types.String `tfsdk:"gh_webhook_url"`
	AllPullRequests         types.Map    `tfsdk:"all_pull_requests"`
	PullRequestOpened       types.Map    `tfsdk:"pull_request_opened"`
	PullRequestModified     types.Map    `tfsdk:"pull_request_modified"`
	CreateTag               types.Map    `tfsdk:"create_tag"`
	Push                    types.Map    `tfsdk:"push"`
}

func (VcsTriggersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"tracked_branch":             types.StringType,
		"approval_pre_apply":         types.BoolType,
		"plan_only":                  types.BoolType,
		"file_triggers_enabled":      types.BoolType,
		"file_trigger_patterns":      types.ListType{ElemType: types.StringType},
		"gh_webhook_url":             types.StringType,
		"all_pull_requests":          types.MapType{ElemType: types.ObjectType{AttrTypes: VcsTriggerActionConfigModel{}.AttributeTypes()}},
		"pull_request_opened":        types.MapType{ElemType: types.ObjectType{AttrTypes: VcsTriggerActionConfigModel{}.AttributeTypes()}},
		"pull_request_modified":      types.MapType{ElemType: types.ObjectType{AttrTypes: VcsTriggerActionConfigModel{}.AttributeTypes()}},
		"create_tag":                 types.MapType{ElemType: types.ObjectType{AttrTypes: VcsTriggerActionConfigModel{}.AttributeTypes()}},
		"push":                       types.MapType{ElemType: types.ObjectType{AttrTypes: VcsTriggerActionConfigModel{}.AttributeTypes()}},
	}
}

func (m VcsTriggersModel) ToAPIModel(ctx context.Context) (*sgsdkgo.VcsTriggers, diag.Diagnostics) {
	result := &sgsdkgo.VcsTriggers{}

	if !m.TrackedBranch.IsNull() && !m.TrackedBranch.IsUnknown() {
		result.TrackedBranch = m.TrackedBranch.ValueStringPointer()
	}
	if !m.ApprovalPreApply.IsNull() && !m.ApprovalPreApply.IsUnknown() {
		result.ApprovalPreApply = m.ApprovalPreApply.ValueBoolPointer()
	}
	if !m.PlanOnly.IsNull() && !m.PlanOnly.IsUnknown() {
		result.PlanOnly = m.PlanOnly.ValueBoolPointer()
	}
	if !m.FileTriggersEnabled.IsNull() && !m.FileTriggersEnabled.IsUnknown() {
		result.FileTriggersEnabled = m.FileTriggersEnabled.ValueBoolPointer()
	}
	if !m.FileTriggerPatterns.IsNull() && !m.FileTriggerPatterns.IsUnknown() {
		patterns, diags := expanders.StringList(ctx, m.FileTriggerPatterns)
		if diags.HasError() {
			return nil, diags
		}
		result.FileTriggerPatterns = patterns
	}


	for _, pair := range []struct {
		src  types.Map
		dest *map[string]sgsdkgo.VcsTriggerActionConfig
	}{
		{m.AllPullRequests, &result.AllPullRequests},
		{m.PullRequestOpened, &result.PullRequestOpened},
		{m.PullRequestModified, &result.PullRequestModified},
		{m.CreateTag, &result.CreateTag},
		{m.Push, &result.Push},
	} {
		if !pair.src.IsNull() && !pair.src.IsUnknown() {
			var models map[string]VcsTriggerActionConfigModel
			diags := pair.src.ElementsAs(ctx, &models, false)
			if diags.HasError() {
				return nil, diags
			}
			v := make(map[string]sgsdkgo.VcsTriggerActionConfig, len(models))
			for k, cfg := range models {
				v[k] = sgsdkgo.VcsTriggerActionConfig{Enabled: cfg.Enabled.ValueBool()}
			}
			*pair.dest = v
		}
	}

	return result, nil
}

func convertVcsTriggersToAPI(ctx context.Context, obj types.Object) (*sgsdkgo.VcsTriggers, diag.Diagnostics) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil
	}
	var m VcsTriggersModel
	diags := obj.As(ctx, &m, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}
	return m.ToAPIModel(ctx)
}

func convertVcsTriggersFromAPI(ctx context.Context, vt *sgsdkgo.VcsTriggers) (types.Object, diag.Diagnostics) {
	nullObj := types.ObjectNull(VcsTriggersModel{}.AttributeTypes())
	if vt == nil || flatteners.IsEmptyObject(vt) {
		return nullObj, nil
	}

	fileTriggerPatterns, diags := flatteners.ListOfStringToTerraformList(vt.FileTriggerPatterns)
	if diags.HasError() {
		return nullObj, diags
	}

	flattenActionMap := func(v map[string]sgsdkgo.VcsTriggerActionConfig) (types.Map, diag.Diagnostics) {
		elemType := types.ObjectType{AttrTypes: VcsTriggerActionConfigModel{}.AttributeTypes()}
		if len(v) == 0 {
			return types.MapNull(elemType), nil
		}
		models := make(map[string]VcsTriggerActionConfigModel, len(v))
		for k, cfg := range v {
			models[k] = VcsTriggerActionConfigModel{Enabled: types.BoolValue(cfg.Enabled)}
		}
		return types.MapValueFrom(ctx, elemType, models)
	}

	allPullRequests, diags := flattenActionMap(vt.AllPullRequests)
	if diags.HasError() {
		return nullObj, diags
	}
	pullRequestOpened, diags := flattenActionMap(vt.PullRequestOpened)
	if diags.HasError() {
		return nullObj, diags
	}
	pullRequestModified, diags := flattenActionMap(vt.PullRequestModified)
	if diags.HasError() {
		return nullObj, diags
	}
	createTag, diags := flattenActionMap(vt.CreateTag)
	if diags.HasError() {
		return nullObj, diags
	}
	push, diags := flattenActionMap(vt.Push)
	if diags.HasError() {
		return nullObj, diags
	}

	m := VcsTriggersModel{
		TrackedBranch:           flatteners.StringPtr(vt.TrackedBranch),
		ApprovalPreApply:        flatteners.BoolPtr(vt.ApprovalPreApply),
		PlanOnly:                flatteners.BoolPtr(vt.PlanOnly),
		FileTriggersEnabled:     flatteners.BoolPtr(vt.FileTriggersEnabled),
		FileTriggerPatterns:     fileTriggerPatterns,
		GhWebhookUrl:            flatteners.StringPtrDefault(vt.GhWebhookUrl),
		AllPullRequests:         allPullRequests,
		PullRequestOpened:       pullRequestOpened,
		PullRequestModified:     pullRequestModified,
		CreateTag:               createTag,
		Push:                    push,
	}
	obj, d := types.ObjectValueFrom(ctx, VcsTriggersModel{}.AttributeTypes(), m)
	if d.HasError() {
		return nullObj, d
	}
	return obj, nil
}

// ---------------------------------------------------------------------------
// VcsConfig
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

func (m CustomSourceConfigModel) ToAPIModel() *sgsdkgo.CustomSourceConfig {
	cfg := &sgsdkgo.CustomSourceConfig{
		IsPrivate:               m.IsPrivate.ValueBoolPointer(),
		Auth:                    m.Auth.ValueStringPointer(),
		WorkingDir:              m.WorkingDir.ValueStringPointer(),
		GitSparseCheckoutConfig: m.GitSparseCheckoutConfig.ValueStringPointer(),
		Ref:                     m.Ref.ValueStringPointer(),
		Repo:                    m.Repo.ValueStringPointer(),
		IncludeSubModule:        m.IncludeSubModule.ValueBoolPointer(),
	}
	if !m.GitCoreAutoCrlf.IsNull() && !m.GitCoreAutoCrlf.IsUnknown() {
		cfg.GitCoreAutoCrlf = m.GitCoreAutoCrlf.ValueBoolPointer()
	}
	return cfg
}

type CustomSourceModel struct {
	SourceConfigDestKind types.String `tfsdk:"source_config_dest_kind"`
	Config               types.Object `tfsdk:"config"`
}

func (m CustomSourceModel) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"source_config_dest_kind": types.StringType,
		"config":                  types.ObjectType{AttrTypes: CustomSourceConfigModel{}.AttributeTypes()},
	}
}

func (m CustomSourceModel) ToAPIModel(ctx context.Context) (*sgsdkgo.CustomSource, diag.Diagnostics) {
	src := &sgsdkgo.CustomSource{
		SourceConfigDestKind: sgsdkgo.CustomSourceSourceConfigDestKindEnum(m.SourceConfigDestKind.ValueString()).Ptr(),
	}
	if !m.Config.IsNull() && !m.Config.IsUnknown() {
		var configModel CustomSourceConfigModel
		diags := m.Config.As(ctx, &configModel, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, diags
		}
		src.Config = configModel.ToAPIModel()
	}
	return src, nil
}

type IacVcsConfigModel struct {
	CustomSource types.Object `tfsdk:"custom_source"`
}

func (m IacVcsConfigModel) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"custom_source": types.ObjectType{AttrTypes: CustomSourceModel{}.AttributeTypes(ctx)},
	}
}

func (m IacVcsConfigModel) ToAPIModel(ctx context.Context) (*sgsdkgo.IacvcsConfig, diag.Diagnostics) {
	cfg := &sgsdkgo.IacvcsConfig{}
	if !m.CustomSource.IsNull() && !m.CustomSource.IsUnknown() {
		var customSourceModel CustomSourceModel
		diags := m.CustomSource.As(ctx, &customSourceModel, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, diags
		}
		src, diags := customSourceModel.ToAPIModel(ctx)
		if diags.HasError() {
			return nil, diags
		}
		cfg.CustomSource = src
	}
	return cfg, nil
}

type VcsConfigModel struct {
	IacVcsConfig types.Object `tfsdk:"iac_vcs_config"`
	IacInputData types.Object `tfsdk:"iac_input_data"`
}

func (m VcsConfigModel) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"iac_vcs_config": types.ObjectType{AttrTypes: IacVcsConfigModel{}.AttributeTypes(ctx)},
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
		iacVcsCfg, diags := iacVcsModel.ToAPIModel(ctx)
		if diags.HasError() {
			return nil, diags
		}
		result.IacVcsConfig = iacVcsCfg
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

		WfStepInputData, diags := inputDataModel.ToAPIModel()
		if diags.HasError() {
			return nil, diags
		}
		result.WfStepInputData = WfStepInputData
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

func (m TerraformConfigModel) ToAPIModel(ctx context.Context) (*sgsdkgo.TerraformConfig, diag.Diagnostics) {
	cfg := &sgsdkgo.TerraformConfig{
		TerraformVersion:      m.TerraformVersion.ValueStringPointer(),
		DriftCheck:            m.DriftCheck.ValueBoolPointer(),
		DriftCron:             m.DriftCron.ValueStringPointer(),
		ManagedTerraformState: m.ManagedTerraformState.ValueBoolPointer(),
		ApprovalPreApply:      m.ApprovalPreApply.ValueBoolPointer(),
		TerraformPlanOptions:  m.TerraformPlanOptions.ValueStringPointer(),
		TerraformInitOptions:  m.TerraformInitOptions.ValueStringPointer(),
		Timeout:               expanders.IntPtr(m.Timeout.ValueInt64Pointer()),
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

func (m WorkflowGitResourceModel) ToAPIModel(ctx context.Context) (*sgworkflows.Workflow, diag.Diagnostics) {
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

	vcsTriggers, diags := convertVcsTriggersToAPI(ctx, m.VcsTriggers)
	if diags.HasError() {
		return nil, diags
	}

	if vcsTriggers != nil && vcsConfig != nil && vcsConfig.IacVcsConfig != nil &&
		vcsConfig.IacVcsConfig.CustomSource != nil && vcsConfig.IacVcsConfig.CustomSource.SourceConfigDestKind != nil {
		t := sgsdkgo.VcsTriggersTypeEnum(*vcsConfig.IacVcsConfig.CustomSource.SourceConfigDestKind)
		vcsTriggers.Type = &t
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

	return &sgworkflows.Workflow{
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
		VcsTriggers:               vcsTriggers,
	}, nil
}

// ---------------------------------------------------------------------------
// ToUpdateAPIModel
// ---------------------------------------------------------------------------

func (m WorkflowGitResourceModel) ToUpdateAPIModel(ctx context.Context) (*sgworkflows.PatchedWorkflow, diag.Diagnostics) {
	workflow, diags := m.ToAPIModel(ctx)
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

	if workflow.VcsTriggers != nil {
		patched.VcsTriggers = sgsdkgo.Optional(*workflow.VcsTriggers)
	} else {
		patched.VcsTriggers = sgsdkgo.Null[sgsdkgo.VcsTriggers]()
	}

	return patched, diags
}

// ---------------------------------------------------------------------------
// convertWorkflowGitFromAPI
// ---------------------------------------------------------------------------

func ConvertWorkflowGitFromAPI(ctx context.Context, response *sgworkflows.WorkflowReadResponse) (WorkflowGitResourceModel, diag.Diagnostics) {
	var allDiags diag.Diagnostics
	model := WorkflowGitResourceModel{}

	wf := response.Msg
	if wf == nil {
		return model, allDiags
	}

	model.Id = flatteners.StringPtr(wf.Id)
	model.ResourceName = flatteners.StringPtr(wf.ResourceName)
	model.Description = flatteners.StringPtrDefaultNull(wf.Description)
	model.NumberOfApprovalsRequired = flatteners.Int64Ptr(wf.NumberOfApprovalsRequired)
	model.UserJobCpu = flatteners.Int64Ptr(wf.UserJobCpu)
	model.UserJobMemory = flatteners.Int64Ptr(wf.UserJobMemory)

	if wf.WfType != nil {
		model.WfType = flatteners.String(string(*wf.WfType))
	} else {
		model.WfType = types.StringNull()
	}

	tags, diags := flatteners.ListOfStringToTerraformList(wf.Tags)
	allDiags.Append(diags...)
	model.Tags = tags

	approvers, diags := flatteners.ListOfStringToTerraformList(wf.Approvers)
	allDiags.Append(diags...)
	model.Approvers = approvers

	contextTags, diags := flatteners.MapStringString(ctx, wf.ContextTags)
	allDiags.Append(diags...)
	model.ContextTags = contextTags

	envVars := make([]sgsdkgo.EnvVars, len(wf.EnvironmentVariables))
	for i, ptr := range wf.EnvironmentVariables {
		if ptr != nil {
			envVars[i] = *ptr
		}
	}
	envVarsList, diags := convertEnvironmentVariablesFromAPI(ctx, envVars)
	allDiags.Append(diags...)
	model.EnvironmentVariables = envVarsList

	terraformConfig, diags := convertTerraformConfigFromAPI(ctx, wf.TerraformConfig)
	allDiags.Append(diags...)
	model.TerraformConfig = terraformConfig

	runnerConstraints, diags := convertRunnerConstraintsFromAPI(ctx, wf.RunnerConstraints)
	allDiags.Append(diags...)
	model.RunnerConstraints = runnerConstraints

	wfStepsConfig := make([]sgsdkgo.WfStepsConfig, len(wf.WfStepsConfig))
	for i, ptr := range wf.WfStepsConfig {
		if ptr != nil {
			wfStepsConfig[i] = *ptr
		}
	}
	wfStepsConfigList, diags := convertWfStepsConfigListFromAPI(ctx, wfStepsConfig)
	allDiags.Append(diags...)
	model.WfStepsConfig = wfStepsConfigList

	miniSteps, diags := convertMinistepsFromAPI(ctx, wf.MiniSteps)
	allDiags.Append(diags...)
	model.MiniSteps = miniSteps

	userSchedules, diags := convertUserSchedulesFromAPI(ctx, wf.UserSchedules)
	allDiags.Append(diags...)
	model.UserSchedules = userSchedules

	deploymentPlatformConfig, diags := convertDeploymentPlatformConfigFromAPI(ctx, wf.DeploymentPlatformConfig)
	allDiags.Append(diags...)
	model.DeploymentPlatformConfig = deploymentPlatformConfig

	vcsConfig, diags := convertVcsConfigFromAPI(ctx, wf.VcsConfig)
	allDiags.Append(diags...)
	model.VcsConfig = vcsConfig

	vcsTriggers, diags := convertVcsTriggersFromAPI(ctx, wf.VcsTriggers)
	allDiags.Append(diags...)
	model.VcsTriggers = vcsTriggers

	return model, allDiags
}

// ---------------------------------------------------------------------------
// ToAPI helpers — thin wrappers that handle null/unknown + obj.As / list iteration
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

func convertTerraformConfigFromAPI(ctx context.Context, cfg *sgsdkgo.TerraformConfig) (types.Object, diag.Diagnostics) {
	nullObj := types.ObjectNull(TerraformConfigModel{}.AttributeTypes())
	if cfg == nil || flatteners.IsEmptyObject(cfg) {
		return nullObj, nil
	}

	m := TerraformConfigModel{
		TerraformVersion:        flatteners.StringPtr(cfg.TerraformVersion),
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
		var customSourceObj types.Object
		if vcsConfig.IacVcsConfig.CustomSource != nil {
			cs := vcsConfig.IacVcsConfig.CustomSource
			var configObj types.Object
			if cs.Config != nil && !flatteners.IsEmptyObject(cs.Config) {
				configModel := CustomSourceConfigModel{
					IsPrivate:               types.BoolPointerValue(cs.Config.IsPrivate),
					Auth:                    flatteners.StringPtr(cs.Config.Auth),
					WorkingDir:              flatteners.StringPtr(cs.Config.WorkingDir),
					GitSparseCheckoutConfig: flatteners.StringPtr(cs.Config.GitSparseCheckoutConfig),
					GitCoreAutoCrlf:         types.BoolPointerValue(cs.Config.GitCoreAutoCrlf),
					Ref:                     flatteners.StringPtr(cs.Config.Ref),
					Repo:                    flatteners.StringPtr(cs.Config.Repo),
					IncludeSubModule:        types.BoolPointerValue(cs.Config.IncludeSubModule),
				}
				var diags diag.Diagnostics
				configObj, diags = types.ObjectValueFrom(ctx, CustomSourceConfigModel{}.AttributeTypes(), configModel)
				if diags.HasError() {
					return nullObj, diags
				}
			} else {
				configObj = types.ObjectNull(CustomSourceConfigModel{}.AttributeTypes())
			}

			customSourceModel := CustomSourceModel{
				SourceConfigDestKind: types.StringValue(string(*cs.SourceConfigDestKind)),
				Config:               configObj,
			}
			var diags diag.Diagnostics
			customSourceObj, diags = types.ObjectValueFrom(ctx, CustomSourceModel{}.AttributeTypes(ctx), customSourceModel)
			if diags.HasError() {
				return nullObj, diags
			}
		} else {
			customSourceObj = types.ObjectNull(CustomSourceModel{}.AttributeTypes(ctx))
		}

		iacVcsModel := IacVcsConfigModel{
			CustomSource: customSourceObj,
		}
		var diags diag.Diagnostics
		iacVcsConfigObj, diags = types.ObjectValueFrom(ctx, IacVcsConfigModel{}.AttributeTypes(ctx), iacVcsModel)
		if diags.HasError() {
			return nullObj, diags
		}
	} else {
		iacVcsConfigObj = types.ObjectNull(IacVcsConfigModel{}.AttributeTypes(ctx))
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

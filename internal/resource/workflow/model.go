package workflow

import (
	"context"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgworkflows "github.com/StackGuardian/sg-sdk-go/workflows"
	"github.com/StackGuardian/sg-sdk-go/workflowtemplaterevisions"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/expanders"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/flatteners"
	workflowtemplaterevision "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/workflow_template_revision"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type WorkflowResourceModel struct {
	Id                        types.String `tfsdk:"id"`
	WorkflowGroupId           types.String `tfsdk:"workflow_group_id"`
	ResourceName              types.String `tfsdk:"resource_name"`
	Description               types.String `tfsdk:"description"`
	WfType                    types.String `tfsdk:"wf_type"`
	UpgradeMode               types.String `tfsdk:"upgrade_mode"`
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

func (m WorkflowResourceModel) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"id":                           types.StringType,
		"workflow_group_id":            types.StringType,
		"resource_name":                types.StringType,
		"description":                  types.StringType,
		"wf_type":                      types.StringType,
		"upgrade_mode":                 types.StringType,
		"environment_variables":        types.ListType{ElemType: types.ObjectType{AttrTypes: workflowtemplaterevision.EnvironmentVariableModel{}.AttributeTypes()}},
		"mini_steps":                   types.ObjectType{AttrTypes: workflowtemplaterevision.MinistepsModel{}.AttributeTypes()},
		"runner_constraints":           types.ObjectType{AttrTypes: workflowtemplaterevision.RunnerConstraintsModel{}.AttributeTypes()},
		"tags":                         types.ListType{ElemType: types.StringType},
		"user_schedules":               types.ListType{ElemType: types.ObjectType{AttrTypes: workflowtemplaterevision.UserSchedulesModel{}.AttributeTypes()}},
		"context_tags":                 types.MapType{ElemType: types.StringType},
		"approvers":                    types.ListType{ElemType: types.StringType},
		"number_of_approvals_required": types.Int64Type,
		"user_job_cpu":                 types.Int64Type,
		"user_job_memory":              types.Int64Type,
		"vcs_config":                   types.ObjectType{AttrTypes: VcsConfigModel{}.AttributeTypes(ctx)},
		"terraform_config":             types.ObjectType{AttrTypes: workflowtemplaterevision.TerraformConfigModel{}.AttributeTypes()},
		"deployment_platform_config":   types.ListType{ElemType: types.ObjectType{AttrTypes: workflowtemplaterevision.DeploymentPlatformConfigModel{}.AttributeTypes()}},
		"wf_steps_config":              types.ListType{ElemType: types.ObjectType{AttrTypes: workflowtemplaterevision.WfStepsConfigModel{}.AttributeTypes()}},
	}
}

// ---------------------------------------------------------------------------
// VcsConfig
// ---------------------------------------------------------------------------

type IacInputDataModel struct {
	SchemaId   types.String `tfsdk:"schema_id"`
	SchemaType types.String `tfsdk:"schema_type"`
	Data       types.String `tfsdk:"data"`
}

func (m IacInputDataModel) AttributeTypes(_ context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"schema_id":   types.StringType,
		"schema_type": types.StringType,
		"data":        types.StringType,
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

func (m CustomSourceConfigModel) AttributeTypes(_ context.Context) map[string]attr.Type {
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

func (m CustomSourceModel) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"source_config_dest_kind": types.StringType,
		"config":                  types.ObjectType{AttrTypes: CustomSourceConfigModel{}.AttributeTypes(ctx)},
	}
}

type IacVcsConfigModel struct {
	UseMarketplaceTemplate types.Bool   `tfsdk:"use_marketplace_template"`
	IacTemplateId          types.String `tfsdk:"iac_template_id"`
	CustomSource           types.Object `tfsdk:"custom_source"`
}

func (m IacVcsConfigModel) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"use_marketplace_template": types.BoolType,
		"iac_template_id":          types.StringType,
		"custom_source":            types.ObjectType{AttrTypes: CustomSourceModel{}.AttributeTypes(ctx)},
	}
}

type VcsConfigModel struct {
	IacVcsConfig types.Object `tfsdk:"iac_vcs_config"`
	IacInputData types.Object `tfsdk:"iac_input_data"`
}

func (m VcsConfigModel) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"iac_vcs_config": types.ObjectType{AttrTypes: IacVcsConfigModel{}.AttributeTypes(ctx)},
		"iac_input_data": types.ObjectType{AttrTypes: IacInputDataModel{}.AttributeTypes(ctx)},
	}
}

// ---------------------------------------------------------------------------
// ToAPIModel
// ---------------------------------------------------------------------------

func (m WorkflowResourceModel) ToAPIModel(ctx context.Context) (*sgworkflows.Workflow, diag.Diagnostics) {
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

	envVars, diags := workflowtemplaterevision.ConvertEnvironmentVariablesToAPI(ctx, m.EnvironmentVariables)
	if diags.HasError() {
		return nil, diags
	}
	envVarPtrs := make([]*sgsdkgo.EnvVars, len(envVars))
	for i := range envVars {
		envVarPtrs[i] = &envVars[i]
	}

	terraformConfig, diags := workflowtemplaterevision.ConvertTerraformConfigToAPI(ctx, m.TerraformConfig)
	if diags.HasError() {
		return nil, diags
	}

	runnerConstraints, diags := workflowtemplaterevision.ConvertRunnerConstraintsToAPIModel(ctx, m.RunnerConstraints)
	if diags.HasError() {
		return nil, diags
	}

	wfStepsConfig, diags := workflowtemplaterevision.ConvertWfStepsConfigListToAPI(ctx, m.WfStepsConfig)
	if diags.HasError() {
		return nil, diags
	}
	wfStepsConfigPtrs := make([]*sgsdkgo.WfStepsConfig, len(wfStepsConfig))
	for i := range wfStepsConfig {
		wfStepsConfigPtrs[i] = &wfStepsConfig[i]
	}

	miniSteps, diags := workflowtemplaterevision.ConvertMinistepsToAPI(ctx, m.MiniSteps)
	if diags.HasError() {
		return nil, diags
	}

	// This ensures that "" is not sent in the payload while making a call to
	// update the workflow since webhook secret cannot be null and it is not allowed in
	// the api.
	//
	// Since in workflow webhook secret is a optional & computed field in provider schema
	// it's has to be set after apply else it will complain that the attribute is unknown.
	// Therefore if the API returns nil for this attribute we set it to "".
	if miniSteps != nil && miniSteps.Webhooks != nil {
		w := miniSteps.Webhooks
		w.APPROVAL_REQUIRED = nilifyEmptyWebhookSecrets(w.APPROVAL_REQUIRED)
		w.CANCELLED = nilifyEmptyWebhookSecrets(w.CANCELLED)
		w.COMPLETED = nilifyEmptyWebhookSecrets(w.COMPLETED)
		w.DRIFT_DETECTED = nilifyEmptyWebhookSecrets(w.DRIFT_DETECTED)
		w.ERRORED = nilifyEmptyWebhookSecrets(w.ERRORED)
	}

	userSchedules, diags := workflowtemplaterevision.ConvertUserSchedulesToAPIModel(ctx, m.UserSchedules)
	if diags.HasError() {
		return nil, diags
	}

	deploymentPlatformConfig, diags := workflowtemplaterevision.ConvertDeploymentPlatformConfigToAPI(ctx, m.DeploymentPlatformConfig)
	if diags.HasError() {
		return nil, diags
	}

	vcsConfig, diags := convertVcsConfigToAPIModel(ctx, m.VcsConfig)
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

	return &sgworkflows.Workflow{
		ResourceName:              m.ResourceName.ValueStringPointer(),
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
	}, nil
}

func (m WorkflowResourceModel) ToUpdateAPIModel(ctx context.Context) (*sgworkflows.PatchedWorkflow, diag.Diagnostics) {
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
		userSchedulesPtrs := make([]*workflowtemplaterevisions.UserSchedules, len(workflow.UserSchedules))
		for i := range workflow.UserSchedules {
			userSchedulesPtrs[i] = &workflow.UserSchedules[i]
		}
		patched.UserSchedules = sgsdkgo.Optional(userSchedulesPtrs)
	} else {
		patched.UserSchedules = sgsdkgo.Null[[]*workflowtemplaterevisions.UserSchedules]()
	}

	return patched, diags
}

// ---------------------------------------------------------------------------
// convertVcsConfigToAPIModel
// ---------------------------------------------------------------------------

func convertVcsConfigToAPIModel(ctx context.Context, vcsConfigObj types.Object) (*sgsdkgo.VcsConfig, diag.Diagnostics) {
	if vcsConfigObj.IsNull() || vcsConfigObj.IsUnknown() {
		return nil, nil
	}

	var vcsModel VcsConfigModel
	diags := vcsConfigObj.As(ctx, &vcsModel, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		return nil, diags
	}

	result := &sgsdkgo.VcsConfig{}

	if !vcsModel.IacVcsConfig.IsNull() && !vcsModel.IacVcsConfig.IsUnknown() {
		var iacVcsModel IacVcsConfigModel
		diags = vcsModel.IacVcsConfig.As(ctx, &iacVcsModel, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, diags
		}

		iacVcsCfg := &sgsdkgo.IacvcsConfig{
			UseMarketplaceTemplate: iacVcsModel.UseMarketplaceTemplate.ValueBool(),
			IacTemplateId:          iacVcsModel.IacTemplateId.ValueStringPointer(),
		}

		if !iacVcsModel.CustomSource.IsNull() && !iacVcsModel.CustomSource.IsUnknown() {
			var customSourceModel CustomSourceModel
			diags = iacVcsModel.CustomSource.As(ctx, &customSourceModel, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				return nil, diags
			}

			customSrc := &sgsdkgo.CustomSource{
				SourceConfigDestKind: sgsdkgo.CustomSourceSourceConfigDestKindEnum(customSourceModel.SourceConfigDestKind.ValueString()),
			}

			if !customSourceModel.Config.IsNull() && !customSourceModel.Config.IsUnknown() {
				var configModel CustomSourceConfigModel
				diags = customSourceModel.Config.As(ctx, &configModel, basetypes.ObjectAsOptions{})
				if diags.HasError() {
					return nil, diags
				}

				customSrc.Config = &sgsdkgo.CustomSourceConfig{
					IsPrivate:               configModel.IsPrivate.ValueBoolPointer(),
					Auth:                    configModel.Auth.ValueStringPointer(),
					WorkingDir:              configModel.WorkingDir.ValueStringPointer(),
					GitSparseCheckoutConfig: configModel.GitSparseCheckoutConfig.ValueStringPointer(),
					GitCoreAutoCrlf:         configModel.GitCoreAutoCrlf.ValueBoolPointer(),
					Ref:                     configModel.Ref.ValueStringPointer(),
					Repo:                    configModel.Repo.ValueStringPointer(),
					IncludeSubModule:        configModel.IncludeSubModule.ValueBoolPointer(),
				}
			}

			iacVcsCfg.CustomSource = customSrc
		}

		result.IacVcsConfig = iacVcsCfg
	}

	if !vcsModel.IacInputData.IsNull() && !vcsModel.IacInputData.IsUnknown() {
		var iacInputDataModel IacInputDataModel
		diags = vcsModel.IacInputData.As(ctx, &iacInputDataModel, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			return nil, diags
		}

		result.IacInputData = &sgsdkgo.IacInputData{
			SchemaId:   iacInputDataModel.SchemaId.ValueStringPointer(),
			SchemaType: sgsdkgo.IacInputDataSchemaTypeEnum(iacInputDataModel.SchemaType.ValueString()),
			Data:       expanders.ParseJSONToMap(iacInputDataModel.Data.ValueString()),
		}
	}

	return result, nil
}

// ---------------------------------------------------------------------------
// convertMinistepsFromAPI (and helpers)
// ---------------------------------------------------------------------------

func nilifyEmptyWebhookSecrets(webhooks []workflowtemplaterevisions.MinistepsWebhooksSchema) []workflowtemplaterevisions.MinistepsWebhooksSchema {
	for i := range webhooks {
		if webhooks[i].WebhookSecret != nil && *webhooks[i].WebhookSecret == "" {
			webhooks[i].WebhookSecret = nil
		}
	}
	return webhooks
}

func convertNotificationRecipientsFromAPI(ctx context.Context, recipients []workflowtemplaterevisions.MinistepsNotificationRecepients) (types.List, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: workflowtemplaterevision.MinistepsNotificationRecipientsModel{}.AttributeTypes()}
	if recipients == nil {
		return types.ListValueMust(elemType, []attr.Value{}), nil
	}

	models := []workflowtemplaterevision.MinistepsNotificationRecipientsModel{}
	for _, r := range recipients {
		recipientsList, diags := types.ListValueFrom(ctx, types.StringType, r.Recipients)
		if diags.HasError() {
			return types.ListNull(elemType), diags
		}
		models = append(models, workflowtemplaterevision.MinistepsNotificationRecipientsModel{
			Recipients: recipientsList,
		})
	}

	list, diags := types.ListValueFrom(ctx, elemType, models)
	if diags.HasError() {
		return types.ListNull(elemType), diags
	}
	return list, nil
}

func convertWebhookFromAPI(ctx context.Context, webhooks []workflowtemplaterevisions.MinistepsWebhooksSchema) (types.List, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: workflowtemplaterevision.MinistepsWebhooksModel{}.AttributeTypes()}
	if webhooks == nil {
		return types.ListValueMust(elemType, []attr.Value{}), nil
	}

	models := []workflowtemplaterevision.MinistepsWebhooksModel{}
	for _, w := range webhooks {
		models = append(models, workflowtemplaterevision.MinistepsWebhooksModel{
			WebhookName:   flatteners.String(w.WebhookName),
			WebhookUrl:    flatteners.String(w.WebhookUrl),
			WebhookSecret: flatteners.StringPtrDefault(w.WebhookSecret),
		})
	}

	list, diags := types.ListValueFrom(ctx, elemType, models)
	if diags.HasError() {
		return types.ListNull(elemType), diags
	}
	return list, nil
}

func convertWorkflowChainingFromAPI(ctx context.Context, wfChainingList []workflowtemplaterevisions.MinistepsWfChainingSchema) (types.List, diag.Diagnostics) {
	elemType := types.ObjectType{AttrTypes: workflowtemplaterevision.MinistepsWorkflowChainingModel{}.AttributeTypes()}
	if wfChainingList == nil {
		return types.ListValueMust(elemType, []attr.Value{}), nil
	}

	models := []workflowtemplaterevision.MinistepsWorkflowChainingModel{}
	for _, c := range wfChainingList {
		models = append(models, workflowtemplaterevision.MinistepsWorkflowChainingModel{
			WorkflowGroupId:    flatteners.String(c.WorkflowGroupId),
			StackId:            flatteners.StringPtr(c.StackId),
			WorkflowId:         flatteners.StringPtr(c.WorkflowId),
			WorkflowRunPayload: flatteners.JSONInterfaceToStringDefault(c.WorkflowRunPayload),
			StackRunPayload:    flatteners.JSONInterfaceToStringDefault(c.StackRunPayload),
		})
	}

	list, diags := types.ListValueFrom(ctx, elemType, models)
	if diags.HasError() {
		return types.ListNull(elemType), diags
	}
	return list, nil
}

func convertMinistepsFromAPI(ctx context.Context, ministeps *workflowtemplaterevisions.Ministeps) (types.Object, diag.Diagnostics) {
	ministepsModel := workflowtemplaterevision.MinistepsModel{}

	buildEmailModel := func(email *workflowtemplaterevisions.MinistepsNotificationsEmail) (workflowtemplaterevision.MinistepsEmailModel, diag.Diagnostics) {
		m := workflowtemplaterevision.MinistepsEmailModel{}
		var src workflowtemplaterevisions.MinistepsNotificationsEmail
		if email != nil {
			src = *email
		}
		var d diag.Diagnostics
		var err diag.Diagnostics
		m.ApprovalRequired, err = convertNotificationRecipientsFromAPI(ctx, src.APPROVAL_REQUIRED)
		d.Append(err...)
		m.Cancelled, err = convertNotificationRecipientsFromAPI(ctx, src.CANCELLED)
		d.Append(err...)
		m.Completed, err = convertNotificationRecipientsFromAPI(ctx, src.COMPLETED)
		d.Append(err...)
		m.DriftDetected, err = convertNotificationRecipientsFromAPI(ctx, src.DRIFT_DETECTED)
		d.Append(err...)
		m.Errored, err = convertNotificationRecipientsFromAPI(ctx, src.ERRORED)
		d.Append(err...)
		return m, d
	}

	// Notifications
	{
		notificationsModel := workflowtemplaterevision.MinistepsNotificationsModel{}
		var notifSrc *workflowtemplaterevisions.MinistepsNotificationsEmail
		if ministeps != nil && ministeps.Notifications != nil {
			notifSrc = ministeps.Notifications.Email
		}
		emailModel, diags := buildEmailModel(notifSrc)
		if diags.HasError() {
			return types.ObjectNull(workflowtemplaterevision.MinistepsModel{}.AttributeTypes()), diags
		}
		emailObj, diags := types.ObjectValueFrom(ctx, workflowtemplaterevision.MinistepsEmailModel{}.AttributeTypes(), emailModel)
		if diags.HasError() {
			return types.ObjectNull(workflowtemplaterevision.MinistepsModel{}.AttributeTypes()), diags
		}
		notificationsModel.Email = emailObj
		notificationsObj, diags := types.ObjectValueFrom(ctx, workflowtemplaterevision.MinistepsNotificationsModel{}.AttributeTypes(), notificationsModel)
		if diags.HasError() {
			return types.ObjectNull(workflowtemplaterevision.MinistepsModel{}.AttributeTypes()), diags
		}
		ministepsModel.Notifications = notificationsObj
	}

	// Webhooks
	{
		webhooksModel := workflowtemplaterevision.MinistepsWebhooksContainerModel{}
		var webhooksSrc workflowtemplaterevisions.MinistepsWebhooks
		if ministeps != nil && ministeps.Webhooks != nil {
			webhooksSrc = *ministeps.Webhooks
		}
		var d diag.Diagnostics
		var err diag.Diagnostics
		webhooksModel.ApprovalRequired, err = convertWebhookFromAPI(ctx, webhooksSrc.APPROVAL_REQUIRED)
		d.Append(err...)
		webhooksModel.Cancelled, err = convertWebhookFromAPI(ctx, webhooksSrc.CANCELLED)
		d.Append(err...)
		webhooksModel.Completed, err = convertWebhookFromAPI(ctx, webhooksSrc.COMPLETED)
		d.Append(err...)
		webhooksModel.DriftDetected, err = convertWebhookFromAPI(ctx, webhooksSrc.DRIFT_DETECTED)
		d.Append(err...)
		webhooksModel.Errored, err = convertWebhookFromAPI(ctx, webhooksSrc.ERRORED)
		d.Append(err...)
		if d.HasError() {
			return types.ObjectNull(workflowtemplaterevision.MinistepsModel{}.AttributeTypes()), d
		}
		webhooksObj, diags := types.ObjectValueFrom(ctx, workflowtemplaterevision.MinistepsWebhooksContainerModel{}.AttributeTypes(), webhooksModel)
		if diags.HasError() {
			return types.ObjectNull(workflowtemplaterevision.MinistepsModel{}.AttributeTypes()), diags
		}
		ministepsModel.Webhooks = webhooksObj
	}

	// WfChaining
	{
		wfChainingModel := workflowtemplaterevision.MinistepsWfChainingContainerModel{}
		var chainingSrc workflowtemplaterevisions.MinistepsWorkflowChaining
		if ministeps != nil && ministeps.WfChaining != nil {
			chainingSrc = *ministeps.WfChaining
		}
		var d diag.Diagnostics
		var err diag.Diagnostics
		wfChainingModel.Completed, err = convertWorkflowChainingFromAPI(ctx, chainingSrc.COMPLETED)
		d.Append(err...)
		wfChainingModel.Errored, err = convertWorkflowChainingFromAPI(ctx, chainingSrc.ERRORED)
		d.Append(err...)
		if d.HasError() {
			return types.ObjectNull(workflowtemplaterevision.MinistepsModel{}.AttributeTypes()), d
		}
		wfChainingObj, diags := types.ObjectValueFrom(ctx, workflowtemplaterevision.MinistepsWfChainingContainerModel{}.AttributeTypes(), wfChainingModel)
		if diags.HasError() {
			return types.ObjectNull(workflowtemplaterevision.MinistepsModel{}.AttributeTypes()), diags
		}
		ministepsModel.WfChaining = wfChainingObj
	}

	ministepsObj, diags := types.ObjectValueFrom(ctx, workflowtemplaterevision.MinistepsModel{}.AttributeTypes(), ministepsModel)
	if diags.HasError() {
		return types.ObjectNull(workflowtemplaterevision.MinistepsModel{}.AttributeTypes()), diags
	}
	return ministepsObj, nil
}

// ---------------------------------------------------------------------------
// convertTerraformConfigFromAPI
// ---------------------------------------------------------------------------

func convertTerraformConfigFromAPI(ctx context.Context, terraformConfig *sgsdkgo.TerraformConfig) (types.Object, diag.Diagnostics) {
	nullObject := types.ObjectNull(workflowtemplaterevision.TerraformConfigModel{}.AttributeTypes())
	if terraformConfig == nil {
		return nullObject, nil
	}

	terraformVersion := flatteners.StringPtr(terraformConfig.TerraformVersion)
	if terraformVersion.IsNull() || terraformVersion.IsUnknown() {
		terraformVersion = types.StringValue("")
	}

	driftCheck := flatteners.BoolPtr(terraformConfig.DriftCheck)
	if driftCheck.IsNull() || driftCheck.IsUnknown() {
		driftCheck = types.BoolValue(false)
	}

	driftCron := flatteners.StringPtr(terraformConfig.DriftCron)
	if driftCron.IsNull() || driftCron.IsUnknown() {
		driftCron = types.StringValue("0 */6 * * ? *")
	}

	managedTerraformState := flatteners.BoolPtr(terraformConfig.ManagedTerraformState)
	if managedTerraformState.IsNull() || managedTerraformState.IsUnknown() {
		managedTerraformState = types.BoolValue(false)
	}

	approvalPreApply := flatteners.BoolPtr(terraformConfig.ApprovalPreApply)
	if approvalPreApply.IsNull() || approvalPreApply.IsUnknown() {
		approvalPreApply = types.BoolValue(false)
	}

	terraformPlanOptions := flatteners.StringPtr(terraformConfig.TerraformPlanOptions)
	if terraformPlanOptions.IsNull() || terraformPlanOptions.IsUnknown() {
		terraformPlanOptions = types.StringValue("")
	}

	terraformInitOptions := flatteners.StringPtr(terraformConfig.TerraformInitOptions)
	if terraformInitOptions.IsNull() || terraformInitOptions.IsUnknown() {
		terraformInitOptions = types.StringValue("")
	}

	timeout := flatteners.Int64Ptr(terraformConfig.Timeout)
	if timeout.IsNull() || timeout.IsUnknown() {
		timeout = types.Int64Value(0)
	}

	runPreInitHooksOnDrift := flatteners.BoolPtr(terraformConfig.RunPreInitHooksOnDrift)
	if runPreInitHooksOnDrift.IsNull() || runPreInitHooksOnDrift.IsUnknown() {
		runPreInitHooksOnDrift = types.BoolValue(false)
	}

	mountPointElemType := types.ObjectType{AttrTypes: workflowtemplaterevision.MountPointModel{}.AttributeTypes()}
	wfStepsElemType := types.ObjectType{AttrTypes: workflowtemplaterevision.WfStepsConfigModel{}.AttributeTypes()}

	terraformBinPath, diags := workflowtemplaterevision.ConvertMountPointListFromAPI(ctx, terraformConfig.TerraformBinPath)
	if diags.HasError() {
		return nullObject, diags
	}
	if terraformBinPath.IsNull() || terraformBinPath.IsUnknown() {
		terraformBinPath = types.ListValueMust(mountPointElemType, []attr.Value{})
	}

	postApplyWfStepsConfig, diags := workflowtemplaterevision.ConvertWfStepsConfigListFromAPI(ctx, terraformConfig.PostApplyWfStepsConfig)
	if diags.HasError() {
		return nullObject, diags
	}
	if postApplyWfStepsConfig.IsNull() || postApplyWfStepsConfig.IsUnknown() {
		postApplyWfStepsConfig = types.ListValueMust(wfStepsElemType, []attr.Value{})
	}

	preApplyWfStepsConfig, diags := workflowtemplaterevision.ConvertWfStepsConfigListFromAPI(ctx, terraformConfig.PreApplyWfStepsConfig)
	if diags.HasError() {
		return nullObject, diags
	}
	if preApplyWfStepsConfig.IsNull() || preApplyWfStepsConfig.IsUnknown() {
		preApplyWfStepsConfig = types.ListValueMust(wfStepsElemType, []attr.Value{})
	}

	prePlanWfStepsConfig, diags := workflowtemplaterevision.ConvertWfStepsConfigListFromAPI(ctx, terraformConfig.PrePlanWfStepsConfig)
	if diags.HasError() {
		return nullObject, diags
	}
	if prePlanWfStepsConfig.IsNull() || prePlanWfStepsConfig.IsUnknown() {
		prePlanWfStepsConfig = types.ListValueMust(wfStepsElemType, []attr.Value{})
	}

	postPlanWfStepsConfig, diags := workflowtemplaterevision.ConvertWfStepsConfigListFromAPI(ctx, terraformConfig.PostPlanWfStepsConfig)
	if diags.HasError() {
		return nullObject, diags
	}
	if postPlanWfStepsConfig.IsNull() || postPlanWfStepsConfig.IsUnknown() {
		postPlanWfStepsConfig = types.ListValueMust(wfStepsElemType, []attr.Value{})
	}

	preInitHooks, diags := flatteners.ListOfStringToTerraformList(terraformConfig.PreInitHooks)
	if diags.HasError() {
		return nullObject, diags
	}
	if preInitHooks.IsNull() || preInitHooks.IsUnknown() {
		preInitHooks = types.ListValueMust(types.StringType, []attr.Value{})
	}

	prePlanHooks, diags := flatteners.ListOfStringToTerraformList(terraformConfig.PrePlanHooks)
	if diags.HasError() {
		return nullObject, diags
	}
	if prePlanHooks.IsNull() || prePlanHooks.IsUnknown() {
		prePlanHooks = types.ListValueMust(types.StringType, []attr.Value{})
	}

	postPlanHooks, diags := flatteners.ListOfStringToTerraformList(terraformConfig.PostPlanHooks)
	if diags.HasError() {
		return nullObject, diags
	}
	if postPlanHooks.IsNull() || postPlanHooks.IsUnknown() {
		postPlanHooks = types.ListValueMust(types.StringType, []attr.Value{})
	}

	preApplyHooks, diags := flatteners.ListOfStringToTerraformList(terraformConfig.PreApplyHooks)
	if diags.HasError() {
		return nullObject, diags
	}
	if preApplyHooks.IsNull() || preApplyHooks.IsUnknown() {
		preApplyHooks = types.ListValueMust(types.StringType, []attr.Value{})
	}

	postApplyHooks, diags := flatteners.ListOfStringToTerraformList(terraformConfig.PostApplyHooks)
	if diags.HasError() {
		return nullObject, diags
	}
	if postApplyHooks.IsNull() || postApplyHooks.IsUnknown() {
		postApplyHooks = types.ListValueMust(types.StringType, []attr.Value{})
	}

	terraformConfigModel := workflowtemplaterevision.TerraformConfigModel{
		TerraformVersion:       terraformVersion,
		DriftCheck:             driftCheck,
		DriftCron:              driftCron,
		ManagedTerraformState:  managedTerraformState,
		ApprovalPreApply:       approvalPreApply,
		TerraformPlanOptions:   terraformPlanOptions,
		TerraformInitOptions:   terraformInitOptions,
		TerraformBinPath:       terraformBinPath,
		Timeout:                timeout,
		PostApplyWfStepsConfig: postApplyWfStepsConfig,
		PreApplyWfStepsConfig:  preApplyWfStepsConfig,
		PrePlanWfStepsConfig:   prePlanWfStepsConfig,
		PostPlanWfStepsConfig:  postPlanWfStepsConfig,
		PreInitHooks:           preInitHooks,
		PrePlanHooks:           prePlanHooks,
		PostPlanHooks:          postPlanHooks,
		PreApplyHooks:          preApplyHooks,
		PostApplyHooks:         postApplyHooks,
		RunPreInitHooksOnDrift: runPreInitHooksOnDrift,
	}

	return types.ObjectValueFrom(ctx, workflowtemplaterevision.TerraformConfigModel{}.AttributeTypes(), terraformConfigModel)
}

// ---------------------------------------------------------------------------
// convertWorkflowFromAPI
// ---------------------------------------------------------------------------

func ConvertWorkflowFromAPI(ctx context.Context, response *sgworkflows.WorkflowReadResponse, workflowGroupId string) (WorkflowResourceModel, diag.Diagnostics) {
	return convertWorkflowFromAPI(ctx, response, workflowGroupId)
}

func convertWorkflowFromAPI(ctx context.Context, response *sgworkflows.WorkflowReadResponse, workflowGroupId string) (WorkflowResourceModel, diag.Diagnostics) {
	var allDiags diag.Diagnostics
	model := WorkflowResourceModel{}

	wf := response.Msg
	if wf == nil {
		return model, allDiags
	}

	model.Id = flatteners.StringPtr(wf.Id)
	model.WorkflowGroupId = types.StringValue(workflowGroupId)
	model.ResourceName = flatteners.StringPtr(wf.ResourceName)
	model.Description = flatteners.StringPtr(wf.Description)
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

	contextTagsElems := make(map[string]attr.Value, len(wf.ContextTags))
	contextTags, diags := types.MapValue(types.StringType, contextTagsElems)
	allDiags.Append(diags...)
	model.ContextTags = contextTags

	envVars := make([]sgsdkgo.EnvVars, len(wf.EnvironmentVariables))
	for i, ptr := range wf.EnvironmentVariables {
		if ptr != nil {
			envVars[i] = *ptr
		}
	}
	envVarsList, diags := workflowtemplaterevision.ConvertEnvironmentVariablesFromAPI(ctx, envVars)
	allDiags.Append(diags...)
	model.EnvironmentVariables = envVarsList

	terraformConfig, diags := convertTerraformConfigFromAPI(ctx, wf.TerraformConfig)
	allDiags.Append(diags...)
	model.TerraformConfig = terraformConfig

	runnerConstraints, diags := workflowtemplaterevision.ConvertRunnerConstraintsFromAPI(ctx, wf.RunnerConstraints)
	allDiags.Append(diags...)
	model.RunnerConstraints = runnerConstraints

	wfStepsConfig := make([]sgsdkgo.WfStepsConfig, len(wf.WfStepsConfig))
	for i, ptr := range wf.WfStepsConfig {
		if ptr != nil {
			wfStepsConfig[i] = *ptr
		}
	}
	wfStepsConfigList, diags := workflowtemplaterevision.ConvertWfStepsConfigListFromAPI(ctx, wfStepsConfig)
	allDiags.Append(diags...)
	model.WfStepsConfig = wfStepsConfigList

	miniSteps, diags := convertMinistepsFromAPI(ctx, wf.MiniSteps)
	allDiags.Append(diags...)
	model.MiniSteps = miniSteps

	userSchedules, diags := workflowtemplaterevision.ConvertUserSchedulesFromAPI(ctx, wf.UserSchedules)
	allDiags.Append(diags...)
	model.UserSchedules = userSchedules

	deploymentPlatformConfig, diags := workflowtemplaterevision.ConvertDeploymentPlatformConfigFromAPI(ctx, wf.DeploymentPlatformConfig)
	allDiags.Append(diags...)
	model.DeploymentPlatformConfig = deploymentPlatformConfig

	vcsConfig, diags := convertVcsConfigFromAPIModel(ctx, wf.VcsConfig)
	allDiags.Append(diags...)
	model.VcsConfig = vcsConfig

	return model, allDiags
}

// ---------------------------------------------------------------------------
// convertVcsConfigFromAPIModel
// ---------------------------------------------------------------------------

func convertVcsConfigFromAPIModel(ctx context.Context, vcsConfig *sgsdkgo.VcsConfig) (types.Object, diag.Diagnostics) {
	nullObj := types.ObjectNull(VcsConfigModel{}.AttributeTypes(ctx))
	if vcsConfig == nil {
		return nullObj, nil
	}

	var iacVcsConfigObj types.Object
	if vcsConfig.IacVcsConfig != nil {
		var customSourceObj types.Object
		if vcsConfig.IacVcsConfig.CustomSource != nil {
			cs := vcsConfig.IacVcsConfig.CustomSource
			var configObj types.Object
			if cs.Config != nil {
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
				configObj, diags = types.ObjectValueFrom(ctx, CustomSourceConfigModel{}.AttributeTypes(ctx), configModel)
				if diags.HasError() {
					return nullObj, diags
				}
			} else {
				configObj = types.ObjectNull(CustomSourceConfigModel{}.AttributeTypes(ctx))
			}

			customSourceModel := CustomSourceModel{
				SourceConfigDestKind: types.StringValue(string(cs.SourceConfigDestKind)),
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
			UseMarketplaceTemplate: types.BoolValue(vcsConfig.IacVcsConfig.UseMarketplaceTemplate),
			IacTemplateId:          flatteners.StringPtr(vcsConfig.IacVcsConfig.IacTemplateId),
			CustomSource:           customSourceObj,
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
			SchemaType: types.StringValue(string(vcsConfig.IacInputData.SchemaType)),
			Data:       types.StringNull(),
		}
		var diags diag.Diagnostics
		iacInputDataObj, diags = types.ObjectValueFrom(ctx, IacInputDataModel{}.AttributeTypes(ctx), iacInputDataModel)
		if diags.HasError() {
			return nullObj, diags
		}
	} else {
		iacInputDataObj = types.ObjectNull(IacInputDataModel{}.AttributeTypes(ctx))
	}

	vcsModel := VcsConfigModel{
		IacVcsConfig: iacVcsConfigObj,
		IacInputData: iacInputDataObj,
	}
	return types.ObjectValueFrom(ctx, VcsConfigModel{}.AttributeTypes(ctx), vcsModel)
}

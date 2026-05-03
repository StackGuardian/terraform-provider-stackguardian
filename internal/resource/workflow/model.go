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

type workflowResourceModel struct {
	Id                        types.String `tfsdk:"id"`
	WorkflowGroupId           types.String `tfsdk:"workflow_group_id"`
	ResourceName              types.String `tfsdk:"resource_name"`
	Description               types.String `tfsdk:"description"`
	WfType                    types.String `tfsdk:"wf_type"`
	EnvironmentVariables      types.List   `tfsdk:"environment_variables"`
	InputSchemas              types.List   `tfsdk:"input_schemas"`
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
	DeploymentPlatformConfig  types.Object `tfsdk:"deployment_platform_config"`
	WfStepsConfig             types.List   `tfsdk:"wf_steps_config"`
}

func (m workflowResourceModel) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"id":                           types.StringType,
		"workflow_group_id":            types.StringType,
		"resource_name":                types.StringType,
		"description":                  types.StringType,
		"wf_type":                      types.StringType,
		"environment_variables":        types.ListType{ElemType: types.ObjectType{AttrTypes: workflowtemplaterevision.EnvironmentVariableModel{}.AttributeTypes()}},
		"input_schemas":                types.ListType{ElemType: types.ObjectType{AttrTypes: workflowtemplaterevision.InputSchemaModel{}.AttributeTypes()}},
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
		"deployment_platform_config":   types.ObjectType{AttrTypes: workflowtemplaterevision.DeploymentPlatformConfigModel{}.AttributeTypes()},
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

func (m workflowResourceModel) ToAPIModel(ctx context.Context) (*sgworkflows.Workflow, diag.Diagnostics) {
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

func (m workflowResourceModel) ToUpdateAPIModel(ctx context.Context) (*sgworkflows.PatchedWorkflow, diag.Diagnostics) {
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
			Data:       workflowtemplaterevision.ParseJSONToMap(iacInputDataModel.Data.ValueString()),
		}
	}

	return result, nil
}

// ---------------------------------------------------------------------------
// convertWorkflowFromAPI
// ---------------------------------------------------------------------------

func convertWorkflowFromAPI(ctx context.Context, response *sgworkflows.WorkflowReadResponse, workflowGroupId string) (workflowResourceModel, diag.Diagnostics) {
	var allDiags diag.Diagnostics
	model := workflowResourceModel{}

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

	terraformConfig, diags := workflowtemplaterevision.ConvertTerraformConfigFromAPI(ctx, wf.TerraformConfig)
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

	miniSteps, diags := workflowtemplaterevision.ConvertMinistepsFromAPI(ctx, wf.MiniSteps)
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

	model.InputSchemas = types.ListNull(types.ObjectType{AttrTypes: workflowtemplaterevision.InputSchemaModel{}.AttributeTypes()})

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

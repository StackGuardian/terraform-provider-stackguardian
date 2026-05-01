package stack

import (
	"context"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/expanders"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/flatteners"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type StackResourceModel struct {
	Id                       types.String `tfsdk:"id"`
	ResourceName             types.String `tfsdk:"resource_name"`
	Description              types.String `tfsdk:"description"`
	Tags                     types.List   `tfsdk:"tags"`
	WorkflowGroupId          types.String `tfsdk:"workflow_group_id"`
	EnvironmentVariables     types.List   `tfsdk:"environment_variables"`
	DeploymentPlatformConfig types.List   `tfsdk:"deployment_platform_config"`
	Actions                  types.Map    `tfsdk:"actions"`
	TemplateGroupId          types.String `tfsdk:"template_group_id"`
	WorkflowsConfig          types.Object `tfsdk:"workflows_config"`
	UserSchedules            types.List   `tfsdk:"user_schedules"`
	ContextTags              types.Map    `tfsdk:"context_tags"`
	MiniSteps                types.Object `tfsdk:"mini_steps"`
}

func (m *StackResourceModel) ToAPIModel(ctx context.Context) (*sgsdkgo.Stack, diag.Diagnostics) {
	var diags diag.Diagnostics
	apiModel := &sgsdkgo.Stack{
		Id:           m.Id.ValueStringPointer(),
		ResourceName: m.ResourceName.ValueStringPointer(),
	}

	if !m.Description.IsUnknown() && !m.Description.IsNull() {
		apiModel.Description = m.Description.ValueStringPointer()
	}

	if !m.Tags.IsUnknown() && !m.Tags.IsNull() {
		tags, tagDiags := expanders.StringList(ctx, m.Tags)
		diags.Append(tagDiags...)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.Tags = tags
	}

	if !m.EnvironmentVariables.IsUnknown() && !m.EnvironmentVariables.IsNull() {
		envVars, envDiags := expandEnvironmentVariables(ctx, m.EnvironmentVariables)
		diags.Append(envDiags...)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.EnvironmentVariables = envVars
	}

	if !m.DeploymentPlatformConfig.IsUnknown() && !m.DeploymentPlatformConfig.IsNull() {
		dpc, dpcDiags := expandDeploymentPlatformConfig(ctx, m.DeploymentPlatformConfig)
		diags.Append(dpcDiags...)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.DeploymentPlatformConfig = dpc
	}

	if !m.Actions.IsUnknown() && !m.Actions.IsNull() {
		actions, actionDiags := expandActionsMap(ctx, m.Actions)
		diags.Append(actionDiags...)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.Actions = actions
	}

	if !m.TemplateGroupId.IsUnknown() && !m.TemplateGroupId.IsNull() {
		apiModel.TemplateGroupId = m.TemplateGroupId.ValueStringPointer()
	}

	if !m.WorkflowsConfig.IsUnknown() && !m.WorkflowsConfig.IsNull() {
		wfc, wfcDiags := expandWorkflowsConfig(ctx, m.WorkflowsConfig)
		diags.Append(wfcDiags...)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.WorkflowsConfig = wfc
	}

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
	}

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

func (m *StackResourceModel) ToUpdateAPIModel(ctx context.Context) (*sgsdkgo.PatchedStack, diag.Diagnostics) {
	var diags diag.Diagnostics
	apiModel := &sgsdkgo.PatchedStack{}

	if !m.ResourceName.IsUnknown() && !m.ResourceName.IsNull() {
		apiModel.ResourceName = sgsdkgo.Optional(m.ResourceName.ValueString())
	}

	if !m.Description.IsUnknown() && !m.Description.IsNull() {
		apiModel.Description = sgsdkgo.Optional(m.Description.ValueString())
	} else if m.Description.IsNull() {
		apiModel.Description = sgsdkgo.Null[string]()
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
	}

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

	if !m.Actions.IsUnknown() && !m.Actions.IsNull() {
		actions, actionDiags := expandActionsMap(ctx, m.Actions)
		diags.Append(actionDiags...)
		if diags.HasError() {
			return nil, diags
		}
		if actions != nil {
			apiModel.Actions = sgsdkgo.Optional(actions)
		} else {
			apiModel.Actions = sgsdkgo.Null[map[string]*sgsdkgo.Actions]()
		}
	} else if m.Actions.IsNull() {
		apiModel.Actions = sgsdkgo.Null[map[string]*sgsdkgo.Actions]()
	}

	if !m.TemplateGroupId.IsUnknown() && !m.TemplateGroupId.IsNull() {
		apiModel.TemplateGroupId = sgsdkgo.Optional(m.TemplateGroupId.ValueString())
	} else if m.TemplateGroupId.IsNull() {
		apiModel.TemplateGroupId = sgsdkgo.Null[string]()
	}

	if !m.WorkflowsConfig.IsUnknown() && !m.WorkflowsConfig.IsNull() {
		wfc, wfcDiags := expandWorkflowsConfig(ctx, m.WorkflowsConfig)
		diags.Append(wfcDiags...)
		if diags.HasError() {
			return nil, diags
		}
		if wfc != nil {
			apiModel.WorkflowsConfig = sgsdkgo.Optional(*wfc)
		} else {
			apiModel.WorkflowsConfig = sgsdkgo.Null[sgsdkgo.WorkflowsConfig]()
		}
	} else if m.WorkflowsConfig.IsNull() {
		apiModel.WorkflowsConfig = sgsdkgo.Null[sgsdkgo.WorkflowsConfig]()
	}

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
	}

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

func BuildAPIModelToStackModel(ctx context.Context, apiResponse *sgsdkgo.GeneratedStackGetResponseMsg) (*StackResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	stackModel := &StackResourceModel{
		Id:              flatteners.String(apiResponse.ResourceId),
		ResourceName:    flatteners.String(*apiResponse.ResourceName),
		Description:     flatteners.String(*apiResponse.Description),
		TemplateGroupId: flatteners.String(*apiResponse.TemplateGroupId),
	}

	// Convert Tags
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

	// Convert EnvironmentVariables
	if apiResponse.EnvironmentVariables != nil {
		envVarsList, envDiags := flattenEnvironmentVariables(ctx, apiResponse.EnvironmentVariables)
		diags.Append(envDiags...)
		if diags.HasError() {
			return nil, diags
		}
		stackModel.EnvironmentVariables = envVarsList
	} else {
		stackModel.EnvironmentVariables = types.ListNull(types.ObjectType{AttrTypes: envVarElementAttrTypes})
	}

	// Convert DeploymentPlatformConfig
	if apiResponse.DeploymentPlatformConfig != nil {
		dpcList, dpcDiags := flattenDeploymentPlatformConfig(ctx, apiResponse.DeploymentPlatformConfig)
		diags.Append(dpcDiags...)
		if diags.HasError() {
			return nil, diags
		}
		stackModel.DeploymentPlatformConfig = dpcList
	} else {
		stackModel.DeploymentPlatformConfig = types.ListNull(types.ObjectType{AttrTypes: deploymentPlatformConfigModelAttrs})
	}

	// Convert Actions
	if apiResponse.Actions != nil {
		actionMap, actionDiags := flattenActionsMap(ctx, apiResponse.Actions)
		diags.Append(actionDiags...)
		if diags.HasError() {
			return nil, diags
		}
		stackModel.Actions = actionMap
	} else {
		stackModel.Actions = types.MapNull(types.ObjectType{})
	}

	// Convert WorkflowsConfig
	if apiResponse.WorkflowsConfig != nil {
		wfcObj, wfcDiags := flattenWorkflowsConfig(ctx, apiResponse.WorkflowsConfig)
		diags.Append(wfcDiags...)
		if diags.HasError() {
			return nil, diags
		}
		stackModel.WorkflowsConfig = wfcObj
	} else {
		stackModel.WorkflowsConfig = types.ObjectNull(workflowsConfigAttrTypes())
	}

	// Convert UserSchedules
	if apiResponse.UserSchedules != nil {
		userSchedulesList, usDiags := flattenUserSchedules(ctx, apiResponse.UserSchedules)
		diags.Append(usDiags...)
		if diags.HasError() {
			return nil, diags
		}
		stackModel.UserSchedules = userSchedulesList
	} else {
		stackModel.UserSchedules = types.ListNull(types.ObjectType{AttrTypes: userSchedulesAttrTypes()})
	}

	// Convert ContextTags
	if apiResponse.ContextTags != nil {
		ctElements := make(map[string]attr.Value)
		for k, v := range apiResponse.ContextTags {
			if v != nil {
				ctElements[k] = types.StringValue(*v)
			}
		}
		ctMap := types.MapValueMust(types.StringType, ctElements)
		stackModel.ContextTags = ctMap
	} else {
		stackModel.ContextTags = types.MapNull(types.StringType)
	}

	// Convert MiniSteps
	if apiResponse.MiniSteps != nil {
		msObj, msDiags := flattenMiniSteps(ctx, apiResponse.MiniSteps)
		diags.Append(msDiags...)
		if diags.HasError() {
			return nil, diags
		}
		stackModel.MiniSteps = msObj
	} else {
		stackModel.MiniSteps = types.ObjectNull(miniStepsAttrTypes())
	}

	return stackModel, diags
}

// EnvironmentVariableModel represents a single environment variable configuration
type EnvironmentVariableModel struct {
	Config types.Object `tfsdk:"config"`
	Kind   types.String `tfsdk:"kind"`
}

func (m EnvironmentVariableModel) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return envVarElementAttrTypes
}

// DeploymentPlatformConfigModel represents deployment platform configuration
type DeploymentPlatformConfigModel struct {
	Kind   types.String `tfsdk:"kind"`
	Config types.Object `tfsdk:"config"`
}

// UserSchedulesModel represents user schedule configuration
type UserSchedulesModel struct {
	Name   types.String `tfsdk:"name"`
	Desc   types.String `tfsdk:"desc"`
	Cron   types.String `tfsdk:"cron"`
	State  types.String `tfsdk:"state"`
	Inputs types.Object `tfsdk:"inputs"`
}

// StackActionParametersModel represents action parameters
type StackActionParametersModel struct {
	TerraformAction          types.Object `tfsdk:"terraform_action"`
	DeploymentPlatformConfig types.List   `tfsdk:"deployment_platform_config"`
	WfStepsConfig            types.List   `tfsdk:"wf_steps_config"`
	EnvironmentVariables     types.List   `tfsdk:"environment_variables"`
}

// TerraformActionModel represents terraform action configuration
type TerraformActionModel struct {
	Action types.String `tfsdk:"action"`
}

// ActionDependencyConditionModel represents action dependency condition
type ActionDependencyConditionModel struct {
	LatestStatus types.String `tfsdk:"latest_status"`
}

// ActionDependencyModel represents action dependency
type ActionDependencyModel struct {
	Id        types.String `tfsdk:"id"`
	Condition types.Object `tfsdk:"condition"`
}

// ActionOrderModel represents order configuration in actions
type ActionOrderModel struct {
	Parameters   types.Object `tfsdk:"parameters"`
	Dependencies types.List   `tfsdk:"dependencies"`
}

// ActionsModel represents an action configuration
type ActionsModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Default     types.Bool   `tfsdk:"default"`
	Order       types.Map    `tfsdk:"order"`
}

// expandEnvironmentVariables converts Terraform environment variables list to API format
func expandEnvironmentVariables(ctx context.Context, envVars types.List) ([]*sgsdkgo.EnvVars, diag.Diagnostics) {
	if envVars.IsNull() || envVars.IsUnknown() {
		return nil, nil
	}

	var envVarModels []EnvironmentVariableModel
	diags := envVars.ElementsAs(ctx, &envVarModels, false)
	if diags.HasError() {
		return nil, diags
	}

	result := make([]*sgsdkgo.EnvVars, len(envVarModels))
	for i, envVar := range envVarModels {
		var configModel struct {
			VarName   types.String `tfsdk:"var_name"`
			SecretId  types.String `tfsdk:"secret_id"`
			TextValue types.String `tfsdk:"text_value"`
		}

		result[i] = &sgsdkgo.EnvVars{
			Kind: sgsdkgo.EnvVarsKindEnum(envVar.Kind.ValueString()),
		}

		if !envVar.Config.IsNull() && !envVar.Config.IsUnknown() {
			configDiags := envVar.Config.As(ctx, &configModel, basetypes.ObjectAsOptions{
				UnhandledNullAsEmpty:    true,
				UnhandledUnknownAsEmpty: true,
			})
			if configDiags.HasError() {
				return nil, configDiags
			}
			result[i].Config = &sgsdkgo.EnvVarConfig{
				VarName:   configModel.VarName.ValueString(),
				SecretId:  configModel.SecretId.ValueStringPointer(),
				TextValue: configModel.TextValue.ValueStringPointer(),
			}
		}
	}

	return result, nil
}

// flattenEnvironmentVariables converts API environment variables to Terraform format
func flattenEnvironmentVariables(ctx context.Context, envVars []*sgsdkgo.EnvVars) (types.List, diag.Diagnostics) {
	nullList := types.ListNull(types.ObjectType{AttrTypes: envVarElementAttrTypes})
	if envVars == nil {
		return nullList, nil
	}

	envVarModels := make([]EnvironmentVariableModel, len(envVars))
	for i, envVar := range envVars {
		configObj := types.ObjectNull(envVarConfigAttrs)
		if envVar.Config != nil {
			var objDiags diag.Diagnostics
			configObj, objDiags = types.ObjectValue(envVarConfigAttrs, map[string]attr.Value{
				"var_name":   types.StringValue(envVar.Config.VarName),
				"secret_id":  flatteners.StringPtr(envVar.Config.SecretId),
				"text_value": flatteners.StringPtr(envVar.Config.TextValue),
			})
			if objDiags.HasError() {
				return nullList, objDiags
			}
		}
		envVarModels[i] = EnvironmentVariableModel{
			Config: configObj,
			Kind:   types.StringValue(string(envVar.Kind)),
		}
	}

	return types.ListValueFrom(ctx, types.ObjectType{AttrTypes: envVarElementAttrTypes}, envVarModels)
}

// expandDeploymentPlatformConfig converts Terraform deployment platform config to API format
func expandDeploymentPlatformConfig(ctx context.Context, dpc types.List) ([]*sgsdkgo.DeploymentPlatformConfig, diag.Diagnostics) {
	if dpc.IsNull() || dpc.IsUnknown() {
		return nil, nil
	}

	var dpcModels []DeploymentPlatformConfigModel
	diags := dpc.ElementsAs(ctx, &dpcModels, false)
	if diags.HasError() {
		return nil, diags
	}

	result := make([]*sgsdkgo.DeploymentPlatformConfig, len(dpcModels))
	for i, cfg := range dpcModels {
		kind := sgsdkgo.DeploymentPlatformConfigKindEnum(cfg.Kind.ValueString())
		dpcItem := &sgsdkgo.DeploymentPlatformConfig{
			Kind: &kind,
		}

		if !cfg.Config.IsNull() && !cfg.Config.IsUnknown() {
			var configModel struct {
				IntegrationId types.String `tfsdk:"integration_id"`
				ProfileName   types.String `tfsdk:"profile_name"`
			}
			configDiags := cfg.Config.As(ctx, &configModel, basetypes.ObjectAsOptions{
				UnhandledNullAsEmpty:    true,
				UnhandledUnknownAsEmpty: true,
			})
			if configDiags.HasError() {
				return nil, configDiags
			}
			dpcItem.Config = &sgsdkgo.DeploymentPlatformConfigConfig{
				IntegrationId: configModel.IntegrationId.ValueStringPointer(),
				ProfileName:   configModel.ProfileName.ValueStringPointer(),
			}
		}

		result[i] = dpcItem
	}

	return result, nil
}

// flattenDeploymentPlatformConfig converts API deployment platform config to Terraform format
func flattenDeploymentPlatformConfig(ctx context.Context, dpc []*sgsdkgo.DeploymentPlatformConfig) (types.List, diag.Diagnostics) {
	if dpc == nil {
		return types.ListNull(types.ObjectType{AttrTypes: deploymentPlatformConfigModelAttrs}), nil
	}

	dpcModels := make([]DeploymentPlatformConfigModel, len(dpc))
	for i, cfg := range dpc {
		configObj := types.ObjectNull(deploymentPlatformConfigConfigAttrs)
		if cfg.Config != nil {
			var objDiags diag.Diagnostics
			configObj, objDiags = types.ObjectValue(deploymentPlatformConfigConfigAttrs, map[string]attr.Value{
				"integration_id": flatteners.StringPtr(cfg.Config.IntegrationId),
				"profile_name":   flatteners.StringPtr(cfg.Config.ProfileName),
			})
			if objDiags.HasError() {
				return types.ListNull(types.ObjectType{AttrTypes: deploymentPlatformConfigModelAttrs}), objDiags
			}
		}

		kindStr := ""
		if cfg.Kind != nil {
			kindStr = string(*cfg.Kind)
		}

		dpcModels[i] = DeploymentPlatformConfigModel{
			Kind:   types.StringValue(kindStr),
			Config: configObj,
		}
	}

	return types.ListValueFrom(ctx, types.ObjectType{AttrTypes: deploymentPlatformConfigModelAttrs}, dpcModels)
}

// expandActionsMap converts Terraform actions map to API format
func expandActionsMap(ctx context.Context, actions types.Map) (map[string]*sgsdkgo.Actions, diag.Diagnostics) {
	if actions.IsNull() || actions.IsUnknown() {
		return nil, nil
	}

	var models map[string]ActionsModel
	if diags := actions.ElementsAs(ctx, &models, false); diags.HasError() {
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
						actionEnum := sgsdkgo.ActionEnum(tam.Action.ValueString())
						params.TerraformAction = &sgsdkgo.TerraformAction{Action: &actionEnum}
					}

					if !pm.DeploymentPlatformConfig.IsNull() && !pm.DeploymentPlatformConfig.IsUnknown() {
						dpcs, diags := expandDeploymentPlatformConfig(ctx, pm.DeploymentPlatformConfig)
						if diags.HasError() {
							return nil, diags
						}
						params.DeploymentPlatformConfig = dpcs
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

// flattenActionsMap converts API actions map to Terraform format
func flattenActionsMap(ctx context.Context, actions map[string]*sgsdkgo.Actions) (types.Map, diag.Diagnostics) {
	if actions == nil {
		return types.MapNull(types.ObjectType{}), nil
	}

	elements := make(map[string]attr.Value)
	for k, v := range actions {
		actionObj, _ := types.ObjectValue(
			map[string]attr.Type{
				"name":        types.StringType,
				"description": types.StringType,
				"default":     types.BoolType,
				"order":       types.MapType{ElemType: types.ObjectType{}},
			},
			map[string]attr.Value{
				"name":        types.StringValue(v.Name),
				"description": flatteners.StringPtr(v.Description),
				"default":     flatteners.BoolPtr(v.Default),
				"order":       types.MapNull(types.ObjectType{}),
			},
		)

		elements[k] = actionObj
	}

	return types.MapValue(types.ObjectType{}, elements)
}

// expandWorkflowsConfig converts Terraform workflows config to API format
func expandWorkflowsConfig(ctx context.Context, wfc types.Object) (*sgsdkgo.WorkflowsConfig, diag.Diagnostics) {
	if wfc.IsNull() || wfc.IsUnknown() {
		return nil, nil
	}

	var model struct {
		Workflows types.List `tfsdk:"workflows"`
	}
	diags := wfc.As(ctx, &model, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return nil, diags
	}

	if model.Workflows.IsNull() || model.Workflows.IsUnknown() {
		return &sgsdkgo.WorkflowsConfig{}, nil
	}

	var wfModels []struct {
		Id                        types.String `tfsdk:"id"`
		ResourceName              types.String `tfsdk:"resource_name"`
		WfType                    types.String `tfsdk:"wf_type"`
		NumberOfApprovalsRequired types.Int64  `tfsdk:"number_of_approvals_required"`
		UserJobCpu                types.Int64  `tfsdk:"user_job_cpu"`
		UserJobMemory             types.Int64  `tfsdk:"user_job_memory"`
		TerraformConfig           types.Object `tfsdk:"terraform_config"`
		EnvironmentVariables      types.List   `tfsdk:"environment_variables"`
		DeploymentPlatformConfig  types.List   `tfsdk:"deployment_platform_config"`
		VcsConfig                 types.Object `tfsdk:"vcs_config"`
		UserSchedules             types.List   `tfsdk:"user_schedules"`
		MiniSteps                 types.Object `tfsdk:"mini_steps"`
		RunnerConstraints         types.List   `tfsdk:"runner_constraints"`
		ContextTags               types.Map    `tfsdk:"context_tags"`
		CacheConfig               types.Object `tfsdk:"cache_config"`
		WfStepsConfig             types.List   `tfsdk:"wf_steps_config"`
		Approvers                 types.List   `tfsdk:"approvers"`
		ParallelExecution         types.Bool   `tfsdk:"parallel_execution"`
		InputSchemas              types.List   `tfsdk:"input_schemas"`
	}

	if diags := model.Workflows.ElementsAs(ctx, &wfModels, false); diags.HasError() {
		return nil, diags
	}

	workflows := make([]*sgsdkgo.WorkflowsConfigWorkflow, len(wfModels))
	for i, wm := range wfModels {
		var wfTypeEnum *sgsdkgo.WfTypeEnum
		if !wm.WfType.IsNull() && !wm.WfType.IsUnknown() {
			wfType := sgsdkgo.WfTypeEnum(wm.WfType.ValueString())
			wfTypeEnum = &wfType
		}

		wf := &sgsdkgo.WorkflowsConfigWorkflow{
			Id:                        wm.Id.ValueStringPointer(),
			ResourceName:              wm.ResourceName.ValueStringPointer(),
			WfType:                    wfTypeEnum,
			NumberOfApprovalsRequired: expanders.IntPtr(wm.NumberOfApprovalsRequired.ValueInt64Pointer()),
			UserJobCpu:                expanders.IntPtr(wm.UserJobCpu.ValueInt64Pointer()),
			UserJobMemory:             expanders.IntPtr(wm.UserJobMemory.ValueInt64Pointer()),
		}

		// Handle other nested fields similar to stack_template_revision
		if !wm.EnvironmentVariables.IsNull() && !wm.EnvironmentVariables.IsUnknown() {
			envVars, envDiags := expandEnvironmentVariables(ctx, wm.EnvironmentVariables)
			if envDiags.HasError() {
				return nil, envDiags
			}
			wf.EnvironmentVariables = envVars
		}

		if !wm.DeploymentPlatformConfig.IsNull() && !wm.DeploymentPlatformConfig.IsUnknown() {
			dpc, dpcDiags := expandDeploymentPlatformConfig(ctx, wm.DeploymentPlatformConfig)
			if dpcDiags.HasError() {
				return nil, dpcDiags
			}
			wf.DeploymentPlatformConfig = dpc
		}

		if !wm.Approvers.IsNull() && !wm.Approvers.IsUnknown() {
			approvers, appDiags := expanders.StringList(ctx, wm.Approvers)
			if appDiags.HasError() {
				return nil, appDiags
			}
			wf.Approvers = approvers
		}

		workflows[i] = wf
	}

	return &sgsdkgo.WorkflowsConfig{Workflows: workflows}, nil
}

// flattenWorkflowsConfig converts API workflows config to Terraform format
func flattenWorkflowsConfig(ctx context.Context, wfc *sgsdkgo.WorkflowsConfig) (types.Object, diag.Diagnostics) {
	if wfc == nil {
		return types.ObjectNull(workflowsConfigAttrTypes()), nil
	}

	return types.ObjectValue(workflowsConfigAttrTypes(), map[string]attr.Value{
		"workflows": types.ListNull(types.ObjectType{}),
	})
}

// expandUserSchedules converts Terraform user schedules list to API format
func expandUserSchedules(ctx context.Context, us types.List) ([]*sgsdkgo.StackUserSchedules, diag.Diagnostics) {
	if us.IsNull() || us.IsUnknown() {
		return nil, nil
	}

	var usModels []UserSchedulesModel
	diags := us.ElementsAs(ctx, &usModels, false)
	if diags.HasError() {
		return nil, diags
	}

	result := make([]*sgsdkgo.StackUserSchedules, len(usModels))
	for i, schedule := range usModels {
		result[i] = &sgsdkgo.StackUserSchedules{
			Name:  schedule.Name.ValueStringPointer(),
			Desc:  schedule.Desc.ValueStringPointer(),
			Cron:  schedule.Cron.ValueString(),
			State: sgsdkgo.StateEnum(schedule.State.ValueString()),
		}
	}

	return result, nil
}

// flattenUserSchedules converts API user schedules list to Terraform format
func flattenUserSchedules(ctx context.Context, us []*sgsdkgo.StackUserSchedules) (types.List, diag.Diagnostics) {
	if us == nil {
		return types.ListNull(types.ObjectType{AttrTypes: userSchedulesAttrTypes()}), nil
	}

	usModels := make([]UserSchedulesModel, len(us))
	for i, schedule := range us {
		usModels[i] = UserSchedulesModel{
			Name:   flatteners.StringPtr(schedule.Name),
			Desc:   flatteners.StringPtr(schedule.Desc),
			Cron:   types.StringValue(schedule.Cron),
			State:  types.StringValue(string(schedule.State)),
			Inputs: types.ObjectNull(map[string]attr.Type{}),
		}
	}

	return types.ListValueFrom(ctx, types.ObjectType{AttrTypes: userSchedulesAttrTypes()}, usModels)
}

// expandContextTags converts Terraform context tags map to API format
func expandContextTags(ctx context.Context, ct types.Map) (map[string]*string, diag.Diagnostics) {
	if ct.IsNull() || ct.IsUnknown() {
		return nil, nil
	}

	result := make(map[string]*string)
	elements := ct.Elements()
	for k, v := range elements {
		if strValue, ok := v.(types.String); ok && !strValue.IsNull() {
			val := strValue.ValueString()
			result[k] = &val
		}
	}

	return result, nil
}

// expandMiniSteps converts Terraform mini steps to API format
func expandMiniSteps(ctx context.Context, ms types.Object) (*sgsdkgo.MiniStepsSchema, diag.Diagnostics) {
	if ms.IsNull() || ms.IsUnknown() {
		return nil, nil
	}

	var model struct {
		Notifications types.List `tfsdk:"notifications"`
		Webhooks      types.List `tfsdk:"webhooks"`
		Chaining      types.List `tfsdk:"chaining"`
	}
	diags := ms.As(ctx, &model, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return nil, diags
	}

	return &sgsdkgo.MiniStepsSchema{}, nil
}

// flattenMiniSteps converts API mini steps to Terraform format
func flattenMiniSteps(ctx context.Context, ms *sgsdkgo.MiniStepsSchema) (types.Object, diag.Diagnostics) {
	if ms == nil {
		return types.ObjectNull(miniStepsAttrTypes()), nil
	}

	return types.ObjectValue(miniStepsAttrTypes(), map[string]attr.Value{
		"notifications": types.ListNull(types.ObjectType{}),
		"webhooks":      types.ListNull(types.ObjectType{}),
		"chaining":      types.ListNull(types.ObjectType{}),
	})
}

// Helper attribute type maps
var envVarConfigAttrs = map[string]attr.Type{
	"var_name":   types.StringType,
	"secret_id":  types.StringType,
	"text_value": types.StringType,
}

var envVarElementAttrTypes = map[string]attr.Type{
	"config": types.ObjectType{AttrTypes: envVarConfigAttrs},
	"kind":   types.StringType,
}

var deploymentPlatformConfigConfigAttrs = map[string]attr.Type{
	"integration_id": types.StringType,
	"profile_name":   types.StringType,
}

var deploymentPlatformConfigModelAttrs = map[string]attr.Type{
	"kind":   types.StringType,
	"config": types.ObjectType{AttrTypes: deploymentPlatformConfigConfigAttrs},
}

func workflowsConfigAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"workflows": types.ListType{ElemType: types.ObjectType{}},
	}
}

func userSchedulesAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":   types.StringType,
		"desc":   types.StringType,
		"cron":   types.StringType,
		"state":  types.StringType,
		"inputs": types.ObjectType{},
	}
}

func miniStepsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"notifications": types.ListType{ElemType: types.ObjectType{}},
		"webhooks":      types.ListType{ElemType: types.ObjectType{}},
		"chaining":      types.ListType{ElemType: types.ObjectType{}},
	}
}

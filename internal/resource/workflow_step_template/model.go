package workflowsteptemplate

import (
	"context"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/sg-sdk-go/workflowsteptemplate"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/expanders"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/flatteners"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// WorkflowStepTemplateResourceModel represents the Terraform resource model
type WorkflowStepTemplateResourceModel struct {
	Id               types.String `tfsdk:"id"`
	TemplateName     types.String `tfsdk:"template_name"`
	TemplateType     types.String `tfsdk:"template_type"`
	IsActive         types.String `tfsdk:"is_active"`
	IsPublic         types.String `tfsdk:"is_public"`
	Description      types.String `tfsdk:"description"`
	Tags             types.List   `tfsdk:"tags"`
	ContextTags      types.Map    `tfsdk:"context_tags"`
	SharedOrgsList   types.List   `tfsdk:"shared_orgs_list"`
	RuntimeSource    types.Object `tfsdk:"runtime_source"`
	SourceConfigKind types.String `tfsdk:"source_config_kind"`
	LatestRevision   types.Int32  `tfsdk:"latest_revision"`
	NextRevision     types.Int32  `tfsdk:"next_revision"`
}

// RuntimeSourceModel represents the runtime source nested object
type RuntimeSourceModel struct {
	SourceConfigDestKind types.String `tfsdk:"source_config_dest_kind"`
	Config               types.Object `tfsdk:"config"`
	AdditionalConfig     types.Map    `tfsdk:"additional_config"`
}

func (RuntimeSourceModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"source_config_dest_kind": types.StringType,
		"config":                  types.ObjectType{AttrTypes: RuntimeSourceConfigModel{}.AttributeTypes()},
		"additional_config":       types.MapType{ElemType: types.StringType},
	}
}

// RuntimeSourceConfigModel represents the config nested within runtime source
type RuntimeSourceConfigModel struct {
	IsPrivate              types.Bool   `tfsdk:"is_private"`
	Auth                   types.String `tfsdk:"auth"`
	DockerImage            types.String `tfsdk:"docker_image"`
	DockerRegistryUsername types.String `tfsdk:"docker_registry_username"`
	LocalWorkspaceDir      types.String `tfsdk:"local_workspace_dir"`
}

func (RuntimeSourceConfigModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"is_private":               types.BoolType,
		"auth":                     types.StringType,
		"docker_image":             types.StringType,
		"docker_registry_username": types.StringType,
		"local_workspace_dir":      types.StringType,
	}
}

// ToAPIModel converts the RuntimeSourceConfigModel to the SDK config model
func (m *RuntimeSourceConfigModel) ToAPIModel() *workflowsteptemplate.WorkflowStepRuntimeSourceConfig {
	return &workflowsteptemplate.WorkflowStepRuntimeSourceConfig{
		DockerImage:            m.DockerImage.ValueString(),
		IsPrivate:              m.IsPrivate.ValueBoolPointer(),
		Auth:                   m.Auth.ValueStringPointer(),
		DockerRegistryUsername: m.DockerRegistryUsername.ValueStringPointer(),
		LocalWorkspaceDir:      m.LocalWorkspaceDir.ValueStringPointer(),
	}
}

// ToAPIModel converts the RuntimeSourceModel to the SDK runtime source model
func (m *RuntimeSourceModel) ToAPIModel(ctx context.Context) (*workflowsteptemplate.WorkflowStepRuntimeSource, diag.Diagnostics) {
	apiRuntimeSource := &workflowsteptemplate.WorkflowStepRuntimeSource{
		SourceConfigDestKind: workflowsteptemplate.SourceConfigDestKindContainerRegistryEnum,
	}

	if !m.Config.IsUnknown() && !m.Config.IsNull() {
		var configModel RuntimeSourceConfigModel
		diags := m.Config.As(ctx, &configModel, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    false,
			UnhandledUnknownAsEmpty: false,
		})
		if diags.HasError() {
			return nil, diags
		}
		apiRuntimeSource.Config = configModel.ToAPIModel()
	}

	if !m.AdditionalConfig.IsUnknown() && !m.AdditionalConfig.IsNull() {
		var additionalConfig map[string]interface{}
		diags := m.AdditionalConfig.ElementsAs(ctx, &additionalConfig, false)
		if diags.HasError() {
			return nil, diags
		}
		apiRuntimeSource.AdditionalConfig = additionalConfig
	}

	return apiRuntimeSource, nil
}

// RuntimeSourceConfigToTerraType converts the SDK config to a Terraform object
func RuntimeSourceConfigToTerraType(config *workflowsteptemplate.WorkflowStepRuntimeSourceConfig) (types.Object, diag.Diagnostics) {
	if config == nil {
		return types.ObjectNull(RuntimeSourceConfigModel{}.AttributeTypes()), nil
	}

	configModel := RuntimeSourceConfigModel{
		DockerImage:            flatteners.String(config.DockerImage),
		IsPrivate:              types.BoolValue(*config.IsPrivate),
		Auth:                   flatteners.StringPtr(config.Auth),
		DockerRegistryUsername: flatteners.StringPtr(config.DockerRegistryUsername),
		LocalWorkspaceDir:      flatteners.StringPtr(config.LocalWorkspaceDir),
	}

	return types.ObjectValueFrom(context.Background(), RuntimeSourceConfigModel{}.AttributeTypes(), configModel)
}

// RuntimeSourceToTerraType converts the SDK runtime source to a Terraform object
func RuntimeSourceToTerraType(runtimeSource *workflowsteptemplate.WorkflowStepRuntimeSource) (types.Object, diag.Diagnostics) {
	objectNull := types.ObjectNull(RuntimeSourceModel{}.AttributeTypes())
	if runtimeSource == nil {
		return objectNull, nil
	}

	runtimeSourceModel := &RuntimeSourceModel{
		SourceConfigDestKind: flatteners.String(string(runtimeSource.SourceConfigDestKind)),
	}

	configObj, diags := RuntimeSourceConfigToTerraType(runtimeSource.Config)
	if diags.HasError() {
		return objectNull, diags
	}
	runtimeSourceModel.Config = configObj

	if runtimeSource.AdditionalConfig != nil {
		acMap, acDiags := types.MapValueFrom(context.Background(), types.StringType, runtimeSource.AdditionalConfig)
		diags.Append(acDiags...)
		if diags.HasError() {
			return objectNull, diags
		}
		runtimeSourceModel.AdditionalConfig = acMap
	} else {
		runtimeSourceModel.AdditionalConfig = types.MapNull(types.StringType)
	}

	return types.ObjectValueFrom(context.Background(), RuntimeSourceModel{}.AttributeTypes(), runtimeSourceModel)
}

// ToAPIModel converts the Terraform model to the SDK create request model
func (m *WorkflowStepTemplateResourceModel) ToAPIModel(ctx context.Context) (*workflowsteptemplate.CreateWorkflowStepTemplate, diag.Diagnostics) {
	apiModel := &workflowsteptemplate.CreateWorkflowStepTemplate{
		TemplateName:     m.TemplateName.ValueString(),
		TemplateType:     workflowsteptemplate.TemplateTypeWorkflowStepEnum,
		ShortDescription: m.Description.ValueStringPointer(),
	}

	// Set optional ID
	if !m.Id.IsUnknown() && !m.Id.IsNull() {
		idStr := m.Id.ValueString()
		apiModel.Id = &idStr
	}

	// Set IsActive
	if !m.IsActive.IsUnknown() && !m.IsActive.IsNull() {
		apiModel.IsActive = (*workflowsteptemplate.IsPublicEnum)(m.IsActive.ValueStringPointer())
	}

	// Set IsPublic
	if !m.IsPublic.IsUnknown() && !m.IsPublic.IsNull() {
		apiModel.IsPublic = (*workflowsteptemplate.IsPublicEnum)(m.IsPublic.ValueStringPointer())
	}

	// Parse tags
	tags, diags := expanders.StringList(ctx, m.Tags)
	if diags.HasError() {
		return nil, diags
	}
	apiModel.Tags = tags

	// Parse context tags
	if !m.ContextTags.IsUnknown() && !m.ContextTags.IsNull() {
		var contextTags map[string]string
		ctDiags := m.ContextTags.ElementsAs(ctx, &contextTags, false)
		diags.Append(ctDiags...)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.ContextTags = contextTags
	}

	// Parse shared orgs list
	sharedOrgs, diags := expanders.StringList(ctx, m.SharedOrgsList)
	if diags.HasError() {
		return nil, diags
	}
	apiModel.SharedOrgsList = sharedOrgs

	// Set source config kind
	if !m.SourceConfigKind.IsUnknown() && !m.SourceConfigKind.IsNull() {
		apiModel.SourceConfigKind = workflowsteptemplate.WorkflowStepTemplateSourceConfigKindDockerImageEnum
	}

	// Parse runtime source
	if !m.RuntimeSource.IsUnknown() && !m.RuntimeSource.IsNull() {
		var runtimeSourceModel RuntimeSourceModel
		rsDiags := m.RuntimeSource.As(ctx, &runtimeSourceModel, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    false,
			UnhandledUnknownAsEmpty: false,
		})
		diags.Append(rsDiags...)
		if diags.HasError() {
			return nil, diags
		}

		runtimeSource, rsApiDiags := runtimeSourceModel.ToAPIModel(ctx)
		diags.Append(rsApiDiags...)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.RuntimeSource = runtimeSource
	}

	return apiModel, diags
}

// ToUpdateAPIModel converts the Terraform model to the SDK update request model
func (m *WorkflowStepTemplateResourceModel) ToPatchedAPIModel(ctx context.Context) (*workflowsteptemplate.UpdateWorkflowStepTemplateRequestModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	apiModel := &workflowsteptemplate.UpdateWorkflowStepTemplateRequestModel{}

	// Set template name
	if !m.TemplateName.IsUnknown() && !m.TemplateName.IsNull() {
		templateName := m.TemplateName.ValueString()
		// Using core.Optional pattern from the sgsdkgo package
		apiModel.TemplateName = sgsdkgo.Optional(templateName)
	}

	// Set description
	if !m.Description.IsUnknown() && !m.Description.IsNull() {
		apiModel.ShortDescription = sgsdkgo.Optional(m.Description.ValueString())
	}

	// Set IsActive
	if !m.IsActive.IsUnknown() && !m.IsActive.IsNull() {
		apiModel.IsActive = sgsdkgo.Optional(workflowsteptemplate.IsPublicEnum(m.IsActive.ValueString()))
	}

	// Set IsPublic
	if !m.IsPublic.IsUnknown() && !m.IsPublic.IsNull() {
		apiModel.IsPublic = sgsdkgo.Optional(workflowsteptemplate.IsPublicEnum(m.IsPublic.ValueString()))
	}

	// Parse tags
	if !m.Tags.IsUnknown() && !m.Tags.IsNull() {
		tags, tagDiags := expanders.StringList(ctx, m.Tags)
		diags.Append(tagDiags...)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.Tags = sgsdkgo.Optional(tags)
	}

	// Parse context tags
	if !m.ContextTags.IsUnknown() && !m.ContextTags.IsNull() {
		var contextTags map[string]string
		ctDiags := m.ContextTags.ElementsAs(ctx, &contextTags, false)
		diags.Append(ctDiags...)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.ContextTags = sgsdkgo.Optional(contextTags)
	}

	// Parse shared orgs list
	if !m.SharedOrgsList.IsUnknown() && !m.SharedOrgsList.IsNull() {
		sharedOrgs, soDiags := expanders.StringList(ctx, m.SharedOrgsList)
		diags.Append(soDiags...)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.SharedOrgsList = sgsdkgo.Optional(sharedOrgs)
	}

	// Parse runtime source
	if !m.RuntimeSource.IsUnknown() && !m.RuntimeSource.IsNull() {
		var runtimeSourceModel RuntimeSourceModel
		diags := m.RuntimeSource.As(ctx, &runtimeSourceModel, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    false,
			UnhandledUnknownAsEmpty: false,
		})
		if diags.HasError() {
			return nil, diags
		}

		runtimeSource, diags := runtimeSourceModel.ToAPIModel(ctx)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.RuntimeSource = sgsdkgo.Optional(*runtimeSource)
	} else {
		apiModel.RuntimeSource = sgsdkgo.Null[workflowsteptemplate.WorkflowStepRuntimeSource]()
	}

	return apiModel, diags
}

// BuildAPIModelToWorkflowStepTemplateModel converts the SDK response to the Terraform model
func BuildAPIModelToWorkflowStepTemplateModel(apiResponse *workflowsteptemplate.UpdateWorkflowStepTemplateResponse) (*WorkflowStepTemplateResourceModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &WorkflowStepTemplateResourceModel{
		Id:               flatteners.String(apiResponse.Id),
		TemplateName:     flatteners.String(apiResponse.TemplateName),
		TemplateType:     flatteners.String(string(apiResponse.TemplateType)),
		Description:      flatteners.StringPtr(apiResponse.ShortDescription),
		SourceConfigKind: flatteners.String(string(apiResponse.SourceConfigKind)),
		IsActive:         flatteners.StringPtr((*string)(apiResponse.IsActive)),
		IsPublic:         flatteners.StringPtr((*string)(apiResponse.IsPublic)),
	}

	// Handle tags
	if apiResponse.Tags != nil {
		var tags []types.String
		for _, tag := range apiResponse.Tags {
			tags = append(tags, flatteners.String(tag))
		}
		tagsList, tagDiags := types.ListValueFrom(context.Background(), types.StringType, tags)
		diags.Append(tagDiags...)
		if diags.HasError() {
			return nil, diags
		}
		model.Tags = tagsList
	} else {
		model.Tags = types.ListNull(types.StringType)
	}

	// Handle context tags
	if apiResponse.ContextTags != nil {
		contextTagsMap, ctDiags := types.MapValueFrom(context.Background(), types.StringType, apiResponse.ContextTags)
		diags.Append(ctDiags...)
		if diags.HasError() {
			return nil, diags
		}
		model.ContextTags = contextTagsMap
	} else {
		model.ContextTags = types.MapNull(types.StringType)
	}

	// Handle shared orgs list
	if apiResponse.SharedOrgsList != nil {
		var orgs []types.String
		for _, org := range apiResponse.SharedOrgsList {
			orgs = append(orgs, flatteners.String(org))
		}
		orgsList, oDiags := types.ListValueFrom(context.Background(), types.StringType, orgs)
		diags.Append(oDiags...)
		if diags.HasError() {
			return nil, diags
		}
		model.SharedOrgsList = orgsList
	} else {
		model.SharedOrgsList = types.ListNull(types.StringType)
	}

	// Handle runtime source
	runtimeSourceObj, rsDiags := RuntimeSourceToTerraType(apiResponse.RuntimeSource)
	diags.Append(rsDiags...)
	if diags.HasError() {
		return nil, diags
	}
	model.RuntimeSource = runtimeSourceObj

	// Handle revisions
	if apiResponse.LatestRevision != nil {
		model.LatestRevision = flatteners.Int32(int(*apiResponse.LatestRevision))
	} else {
		model.LatestRevision = types.Int32Null()
	}

	if apiResponse.NextRevision != nil {
		model.NextRevision = flatteners.Int32(int(*apiResponse.NextRevision))
	} else {
		model.NextRevision = types.Int32Null()
	}

	return model, diags
}

package workflowsteptemplaterevision

import (
	"context"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/sg-sdk-go/workflowsteptemplate"
	"github.com/StackGuardian/sg-sdk-go/workflowsteptemplaterevision"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/expanders"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/flatteners"
	workflowsteptemplateresource "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/workflow_step_template"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// WorkflowStepTemplateRevisionResourceModel represents the Terraform resource model
type WorkflowStepTemplateRevisionResourceModel struct {
	Id               types.String `tfsdk:"id"`
	TemplateId       types.String `tfsdk:"template_id"`
	Alias            types.String `tfsdk:"alias"`
	Notes            types.String `tfsdk:"notes"`
	LongDescription  types.String `tfsdk:"description"`
	TemplateType     types.String `tfsdk:"template_type"`
	SourceConfigKind types.String `tfsdk:"source_config_kind"`
	IsActive         types.String `tfsdk:"is_active"`
	IsPublic         types.String `tfsdk:"is_public"`
	Tags             types.List   `tfsdk:"tags"`
	ContextTags      types.Map    `tfsdk:"context_tags"`
	RuntimeSource    types.Object `tfsdk:"runtime_source"`
	Deprecation      types.Object `tfsdk:"deprecation"`
}

// DeprecationModel represents the deprecation nested object
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

// ToAPIModel converts the Terraform model to the SDK create request model
func (m *WorkflowStepTemplateRevisionResourceModel) ToAPIModel(ctx context.Context) (*workflowsteptemplaterevision.CreateWorkflowStepTemplateRevisionModel, diag.Diagnostics) {
	apiModel := &workflowsteptemplaterevision.CreateWorkflowStepTemplateRevisionModel{
		TemplateType: workflowsteptemplate.TemplateTypeWorkflowStepEnum,
	}

	var diags diag.Diagnostics

	// Set Alias
	if !m.Alias.IsUnknown() && !m.Alias.IsNull() {
		apiModel.Alias = m.Alias.ValueStringPointer()
	}

	// Set Notes
	if !m.Notes.IsUnknown() && !m.Notes.IsNull() {
		apiModel.Notes = m.Notes.ValueStringPointer()
	}

	// Set LongDescription
	if !m.LongDescription.IsUnknown() && !m.LongDescription.IsNull() {
		apiModel.LongDescription = m.LongDescription.ValueStringPointer()
	}

	// Set SourceConfigKind
	if !m.SourceConfigKind.IsUnknown() && !m.SourceConfigKind.IsNull() {
		apiModel.SourceConfigKind = workflowsteptemplate.WorkflowStepTemplateSourceConfigKindDockerImageEnum
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
	tags, tagDiags := expanders.StringList(ctx, m.Tags)
	diags.Append(tagDiags...)
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

	// Parse runtime source
	if !m.RuntimeSource.IsUnknown() && !m.RuntimeSource.IsNull() {
		var runtimeSourceModel workflowsteptemplateresource.RuntimeSourceModel
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
		apiModel.RuntimeSource = runtimeSource
	}

	// Parse deprecation
	if !m.Deprecation.IsUnknown() && !m.Deprecation.IsNull() {
		var deprecationModel DeprecationModel
		depDiags := m.Deprecation.As(ctx, &deprecationModel, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    false,
			UnhandledUnknownAsEmpty: false,
		})
		diags.Append(depDiags...)
		if diags.HasError() {
			return nil, diags
		}

		apiModel.Deprecation = &workflowsteptemplaterevision.Deprecation{
			EffectiveDate: deprecationModel.EffectiveDate.ValueStringPointer(),
			Message:       deprecationModel.Message.ValueStringPointer(),
		}
	}

	return apiModel, diags
}

// ToPatchedAPIModel converts the Terraform model to the SDK update request model
func (m *WorkflowStepTemplateRevisionResourceModel) ToPatchedAPIModel(ctx context.Context) (*workflowsteptemplaterevision.UpdateWorkflowStepTemplateRevisionModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	apiModel := &workflowsteptemplaterevision.UpdateWorkflowStepTemplateRevisionModel{}

	// Set Alias
	if !m.Alias.IsUnknown() && !m.Alias.IsNull() {
		apiModel.Alias = sgsdkgo.Optional(m.Alias.ValueString())
	}

	// Set Notes
	if !m.Notes.IsUnknown() && !m.Notes.IsNull() {
		apiModel.Notes = sgsdkgo.Optional(m.Notes.ValueString())
	}

	// Set LongDescription
	if !m.LongDescription.IsUnknown() && !m.LongDescription.IsNull() {
		apiModel.LongDescription = sgsdkgo.Optional(m.LongDescription.ValueString())
	}

	// Set SourceConfigKind
	if !m.SourceConfigKind.IsUnknown() && !m.SourceConfigKind.IsNull() {
		apiModel.SourceConfigKind = sgsdkgo.Optional(workflowsteptemplate.WorkflowStepTemplateSourceConfigKindEnum(m.SourceConfigKind.ValueString()))
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

	// Parse runtime source
	if !m.RuntimeSource.IsUnknown() && !m.RuntimeSource.IsNull() {
		var runtimeSourceModel workflowsteptemplateresource.RuntimeSourceModel
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

	// Parse deprecation
	if !m.Deprecation.IsUnknown() && !m.Deprecation.IsNull() {
		var deprecationModel DeprecationModel
		depDiags := m.Deprecation.As(ctx, &deprecationModel, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    false,
			UnhandledUnknownAsEmpty: false,
		})
		diags.Append(depDiags...)
		if diags.HasError() {
			return nil, diags
		}

		apiModel.Deprecation = sgsdkgo.Optional(workflowsteptemplaterevision.Deprecation{
			EffectiveDate: deprecationModel.EffectiveDate.ValueStringPointer(),
			Message:       deprecationModel.Message.ValueStringPointer(),
		})
	}

	return apiModel, diags
}

// BuildAPIModelToRevisionModel converts the SDK response to the Terraform model
func BuildAPIModelToRevisionModel(apiResponse *workflowsteptemplaterevision.WorkflowStepTemplateRevisionResponseData, id string, templateId string) (*WorkflowStepTemplateRevisionResourceModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &WorkflowStepTemplateRevisionResourceModel{
		Id:               flatteners.String(id),
		TemplateId:       flatteners.String(templateId),
		Alias:            flatteners.StringPtr(apiResponse.Alias),
		Notes:            flatteners.StringPtr(apiResponse.Notes),
		LongDescription:  flatteners.StringPtr(apiResponse.LongDescription),
		TemplateType:     flatteners.String(string(apiResponse.TemplateType)),
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

	// Handle runtime source
	if apiResponse.RuntimeSource != nil {
		runtimeSourceModel := &workflowsteptemplateresource.RuntimeSourceModel{
			SourceConfigDestKind: flatteners.String(string(apiResponse.RuntimeSource.SourceConfigDestKind)),
		}

		// Handle config
		if apiResponse.RuntimeSource.Config != nil {
			configModel := workflowsteptemplateresource.RuntimeSourceConfigModel{
				DockerImage:            flatteners.String(apiResponse.RuntimeSource.Config.DockerImage),
				IsPrivate:              types.BoolValue(*apiResponse.RuntimeSource.Config.IsPrivate),
				Auth:                   flatteners.StringPtr(apiResponse.RuntimeSource.Config.Auth),
				DockerRegistryUsername: flatteners.StringPtr(apiResponse.RuntimeSource.Config.DockerRegistryUsername),
				LocalWorkspaceDir:      flatteners.StringPtr(apiResponse.RuntimeSource.Config.LocalWorkspaceDir),
			}

			configObj, cfgDiags := types.ObjectValueFrom(context.Background(), workflowsteptemplateresource.RuntimeSourceConfigModel{}.AttributeTypes(), configModel)
			diags.Append(cfgDiags...)
			if diags.HasError() {
				return nil, diags
			}
			runtimeSourceModel.Config = configObj
		} else {
			runtimeSourceModel.Config = types.ObjectNull(workflowsteptemplateresource.RuntimeSourceConfigModel{}.AttributeTypes())
		}

		// Handle additional config
		if apiResponse.RuntimeSource.AdditionalConfig != nil {
			acMap, acDiags := types.MapValueFrom(context.Background(), types.StringType, apiResponse.RuntimeSource.AdditionalConfig)
			diags.Append(acDiags...)
			if diags.HasError() {
				return nil, diags
			}
			runtimeSourceModel.AdditionalConfig = acMap
		} else {
			runtimeSourceModel.AdditionalConfig = types.MapNull(types.StringType)
		}

		runtimeSourceObj, rsDiags := types.ObjectValueFrom(context.Background(), workflowsteptemplateresource.RuntimeSourceModel{}.AttributeTypes(), runtimeSourceModel)
		diags.Append(rsDiags...)
		if diags.HasError() {
			return nil, diags
		}
		model.RuntimeSource = runtimeSourceObj
	} else {
		model.RuntimeSource = types.ObjectNull(workflowsteptemplateresource.RuntimeSourceModel{}.AttributeTypes())
	}

	// Handle deprecation
	if apiResponse.Deprecation != nil {
		deprecationModel := DeprecationModel{
			EffectiveDate: flatteners.StringPtr(apiResponse.Deprecation.EffectiveDate),
			Message:       flatteners.StringPtr(apiResponse.Deprecation.Message),
		}

		deprecationObj, depDiags := types.ObjectValueFrom(context.Background(), DeprecationModel{}.AttributeTypes(), deprecationModel)
		diags.Append(depDiags...)
		if diags.HasError() {
			return nil, diags
		}
		model.Deprecation = deprecationObj
	} else {
		model.Deprecation = types.ObjectNull(DeprecationModel{}.AttributeTypes())
	}

	return model, diags
}

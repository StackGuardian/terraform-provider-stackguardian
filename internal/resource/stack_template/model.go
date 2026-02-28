package stacktemplate

import (
	"context"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/sg-sdk-go/stacktemplates"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/expanders"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/flatteners"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type StackTemplateResourceModel struct {
	Id               types.String `tfsdk:"id"`
	TemplateName     types.String `tfsdk:"template_name"`
	OwnerOrg         types.String `tfsdk:"owner_org"`
	SourceConfigKind types.String `tfsdk:"source_config_kind"`
	IsActive         types.String `tfsdk:"is_active"`
	IsPublic         types.String `tfsdk:"is_public"`
	ShortDescription types.String `tfsdk:"description"`
	Tags             types.List   `tfsdk:"tags"`
	ContextTags      types.Map    `tfsdk:"context_tags"`
	SharedOrgsList   types.List   `tfsdk:"shared_orgs_list"`
}

func (m *StackTemplateResourceModel) ToAPIModel(ctx context.Context) (*stacktemplates.CreateStackTemplateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	apiModel := &stacktemplates.CreateStackTemplateRequest{
		TemplateName:     m.TemplateName.ValueString(),
		ShortDescription: m.ShortDescription.ValueStringPointer(),
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

	// Convert Tags
	if !m.Tags.IsNull() && !m.Tags.IsUnknown() {
		tags, diagsTags := expanders.StringList(ctx, m.Tags)
		diags.Append(diagsTags...)
		if !diags.HasError() {
			apiModel.Tags = tags
		}
	}

	// Convert SharedOrgsList
	if !m.SharedOrgsList.IsNull() && !m.SharedOrgsList.IsUnknown() {
		sharedOrgs, diagsShared := expanders.StringList(ctx, m.SharedOrgsList)
		diags.Append(diagsShared...)
		if !diags.HasError() {
			apiModel.SharedOrgsList = sharedOrgs
		}
	}

	// Convert ContextTags
	if !m.ContextTags.IsNull() && !m.ContextTags.IsUnknown() {
		contextTags := make(map[string]string)
		diagsCT := m.ContextTags.ElementsAs(ctx, &contextTags, false)
		diags.Append(diagsCT...)
		if !diags.HasError() && len(contextTags) > 0 {
			apiModel.ContextTags = contextTags
		}
	}

	return apiModel, diags
}

func (m *StackTemplateResourceModel) ToUpdateAPIModel(ctx context.Context) (*stacktemplates.UpdateStackTemplateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	apiModel := &stacktemplates.UpdateStackTemplateRequest{
		TemplateName: sgsdkgo.Optional(m.TemplateName.ValueString()),
	}

	if !m.SourceConfigKind.IsNull() && !m.SourceConfigKind.IsUnknown() {
		apiModel.SourceConfigKind = sgsdkgo.Optional(stacktemplates.StackTemplateSourceConfigKindEnum(m.SourceConfigKind.ValueString()))
	} else {
		apiModel.SourceConfigKind = sgsdkgo.Null[stacktemplates.StackTemplateSourceConfigKindEnum]()
	}

	if !m.ShortDescription.IsNull() && !m.ShortDescription.IsUnknown() {
		apiModel.ShortDescription = sgsdkgo.Optional(m.ShortDescription.ValueString())
	} else {
		apiModel.ShortDescription = sgsdkgo.Null[string]()
	}

	if !m.IsActive.IsNull() && !m.IsActive.IsUnknown() {
		apiModel.IsActive = sgsdkgo.Optional(sgsdkgo.IsPublicEnum(m.IsActive.ValueString()))
	}

	if !m.IsPublic.IsNull() && !m.IsPublic.IsUnknown() {
		apiModel.IsPublic = sgsdkgo.Optional(sgsdkgo.IsPublicEnum(m.IsPublic.ValueString()))
	}

	// Convert Tags
	tags, diagsTags := expanders.StringList(ctx, m.Tags)
	diags.Append(diagsTags...)
	if tags != nil {
		apiModel.Tags = sgsdkgo.Optional(tags)
	} else {
		apiModel.Tags = sgsdkgo.Null[[]string]()
	}

	// Convert ContextTags
	contextTags, diagsCT := expanders.MapStringString(ctx, m.ContextTags)
	diags.Append(diagsCT...)
	if contextTags != nil {
		apiModel.ContextTags = sgsdkgo.Optional(contextTags)
	} else {
		apiModel.ContextTags = sgsdkgo.Null[map[string]string]()
	}

	// Convert SharedOrgsList
	sharedOrgsList, diagsShared := expanders.StringList(ctx, m.SharedOrgsList)
	diags.Append(diagsShared...)
	if sharedOrgsList != nil {
		apiModel.SharedOrgsList = sgsdkgo.Optional(sharedOrgsList)
	} else {
		apiModel.SharedOrgsList = sgsdkgo.Null[[]string]()
	}

	return apiModel, diags
}

func BuildAPIModelToStackTemplateModel(apiResponse *stacktemplates.ReadStackTemplateResponse) (*StackTemplateResourceModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &StackTemplateResourceModel{
		Id:               flatteners.StringPtr(apiResponse.Id),
		TemplateName:     flatteners.StringPtr(apiResponse.TemplateName),
		OwnerOrg:         flatteners.StringPtr(apiResponse.OwnerOrg),
		ShortDescription: flatteners.StringPtr(apiResponse.ShortDescription),
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

	if apiResponse.IsPublic != nil {
		model.IsPublic = flatteners.String(string(*apiResponse.IsPublic))
	} else {
		model.IsPublic = types.StringNull()
	}

	// Convert Tags
	if apiResponse.Tags != nil {
		var tags []types.String
		for _, tag := range apiResponse.Tags {
			tags = append(tags, flatteners.String(tag))
		}
		tagsList, diagsTags := types.ListValueFrom(context.Background(), types.StringType, tags)
		diags.Append(diagsTags...)
		model.Tags = tagsList
	} else {
		model.Tags = types.ListNull(types.StringType)
	}

	// Convert SharedOrgsList
	if apiResponse.SharedOrgsList != nil {
		var sharedOrgs []types.String
		for _, org := range apiResponse.SharedOrgsList {
			sharedOrgs = append(sharedOrgs, flatteners.String(org))
		}
		sharedOrgsList, diagsShared := types.ListValueFrom(context.Background(), types.StringType, sharedOrgs)
		diags.Append(diagsShared...)
		model.SharedOrgsList = sharedOrgsList
	} else {
		model.SharedOrgsList = types.ListNull(types.StringType)
	}

	// Convert ContextTags
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

	return model, diags
}

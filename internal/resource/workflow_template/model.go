package workflowtemplate

import (
	"context"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/sg-sdk-go/workflowtemplates"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/expanders"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/flatteners"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type WorkflowTemplateResourceModel struct {
	Id               types.String `tfsdk:"id"`
	TemplateName     types.String `tfsdk:"template_name"`
	OwnerOrg         types.String `tfsdk:"owner_org"`
	SourceConfigKind types.String `tfsdk:"source_config_kind"`
	IsPublic         types.String `tfsdk:"is_public"`
	ShortDescription types.String `tfsdk:"description"`
	RuntimeSource    types.Object `tfsdk:"runtime_source"`
	SharedOrgsList   types.List   `tfsdk:"shared_orgs_list"`
	Tags             types.List   `tfsdk:"tags"`
	ContextTags      types.Map    `tfsdk:"context_tags"`
	VCSTriggers      types.Object `tfsdk:"vcs_triggers"`
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

type VCSTriggersModel struct {
	Type      types.String `tfsdk:"type"`
	CreateTag types.Object `tfsdk:"create_tag"`
}

type VCSTriggersCreateRevisionModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

func (VCSTriggersCreateRevisionModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled": types.BoolType,
	}
}

type VCSTriggersCreateTagModel struct {
	CreateRevision types.Object `tfsdk:"create_revision"`
}

func (VCSTriggersCreateTagModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"create_revision": types.ObjectType{AttrTypes: VCSTriggersCreateRevisionModel{}.AttributeTypes()},
	}
}

func (VCSTriggersModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type": types.StringType,
		"create_tag": types.ObjectType{
			AttrTypes: VCSTriggersCreateTagModel{}.AttributeTypes(),
		},
	}
}

func VCSTriggersToAPIModel(ctx context.Context, m types.Object) (*workflowtemplates.VCSTriggers, diag.Diagnostics) {
	var vcsTriggersModel VCSTriggersModel
	if m.IsNull() || m.IsUnknown() {
		return nil, nil
	}

	diags := m.As(ctx, &vcsTriggersModel, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return nil, diags
	}

	vcsTriggers := &workflowtemplates.VCSTriggers{
		Type: workflowtemplates.VCSTriggersTypeEnum(vcsTriggersModel.Type.ValueString()).Ptr(),
	}

	// Convert create_tag
	if !vcsTriggersModel.CreateTag.IsNull() && !vcsTriggersModel.CreateTag.IsUnknown() {
		var createTagModel VCSTriggersCreateTagModel
		createTagAPIModel := workflowtemplates.VCSTriggersCreateTag{}
		diags := vcsTriggersModel.CreateTag.As(ctx, &createTagModel, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		if diags.HasError() {
			return nil, diags
		}

		if !createTagModel.CreateRevision.IsNull() && !createTagModel.CreateRevision.IsUnknown() {
			var createRevision VCSTriggersCreateRevisionModel
			diags := createTagModel.CreateRevision.As(ctx, &createRevision, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
			if diags.HasError() {
				return nil, diags
			}

			vcsCreateRevisionAPIModel := workflowtemplates.VCSTriggersCreateTagCreateRevision{
				Enabled: createRevision.Enabled.ValueBoolPointer(),
			}
			createTagAPIModel.CreateRevision = &vcsCreateRevisionAPIModel
		}
		vcsTriggers.CreateTag = &createTagAPIModel
	}

	return vcsTriggers, nil
}

func (m RuntimeSourceModel) ToAPIModel(ctx context.Context) (*workflowtemplates.RuntimeSource, diag.Diagnostics) {
	runtimeSource := &workflowtemplates.RuntimeSource{
		SourceConfigDestKind: workflowtemplates.SourceConfigDestKindEnum(m.SourceConfigDestKind.ValueString()).Ptr(),
	}

	// Convert config
	if !m.Config.IsNull() && !m.Config.IsUnknown() {
		var configModel RuntimeSourceConfigModel
		diag_cfg := m.Config.As(ctx, &configModel, basetypes.ObjectAsOptions{
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

func (m *WorkflowTemplateResourceModel) ToAPIModel(ctx context.Context) (*workflowtemplates.CreateWorkflowTemplateRequest, diag.Diagnostics) {
	diag := diag.Diagnostics{}

	apiModel := &workflowtemplates.CreateWorkflowTemplateRequest{
		TemplateName:     m.TemplateName.ValueString(),
		OwnerOrg:         m.OwnerOrg.ValueString(),
		ShortDescription: m.ShortDescription.ValueStringPointer(),
	}

	if !m.SourceConfigKind.IsNull() && !m.SourceConfigKind.IsUnknown() {
		apiModel.SourceConfigKind = (*workflowtemplates.WorkflowTemplateSourceConfigKindEnum)(m.SourceConfigKind.ValueStringPointer())
	}

	if !m.IsPublic.IsNull() && !m.IsPublic.IsUnknown() {
		apiModel.IsPublic = (*sgsdkgo.IsPublicEnum)(m.IsPublic.ValueStringPointer())
	}

	// Convert Tags from types.List to []string
	if !m.Tags.IsNull() && !m.Tags.IsUnknown() {
		tags, diags_tags := expanders.StringList(ctx, m.Tags)
		diag.Append(diags_tags...)
		if !diag.HasError() {
			apiModel.Tags = tags
		}
	}

	// Convert SharedOrgsList
	if !m.SharedOrgsList.IsNull() && !m.SharedOrgsList.IsUnknown() {
		sharedOrgs, diags_shared := expanders.StringList(ctx, m.SharedOrgsList)
		diag.Append(diags_shared...)
		if !diag.HasError() {
			apiModel.SharedOrgsList = sharedOrgs
		}
	}

	// Convert ContextTags from types.Map to map[string]string
	if !m.ContextTags.IsNull() && !m.ContextTags.IsUnknown() {
		contextTags := make(map[string]string)
		diag_ct := m.ContextTags.ElementsAs(ctx, &contextTags, false)
		diag.Append(diag_ct...)
		if !diag.HasError() && len(contextTags) > 0 {
			apiModel.ContextTags = contextTags
		}
	}

	// Convert RuntimeSource
	if !m.RuntimeSource.IsNull() && !m.RuntimeSource.IsUnknown() {
		var runtimeSourceModel RuntimeSourceModel
		diags := m.RuntimeSource.As(ctx, &runtimeSourceModel, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		if diags.HasError() {
			return nil, diags
		}
		runtimeSourceApiModel, diags := runtimeSourceModel.ToAPIModel(ctx)
		if diags.HasError() {
			return nil, diags
		}
		apiModel.RuntimeSource = runtimeSourceApiModel
	}

	// Convert VCSTriggers
	vcsTriggersAPIModel, diags := VCSTriggersToAPIModel(ctx, m.VCSTriggers)
	if diags.HasError() {
		return nil, diags
	}
	apiModel.VCSTriggers = vcsTriggersAPIModel

	return apiModel, diag
}

func (m *WorkflowTemplateResourceModel) ToUpdateAPIModel(ctx context.Context) (*workflowtemplates.UpdateWorkflowTemplateRequest, diag.Diagnostics) {
	diag := diag.Diagnostics{}

	apiModel := &workflowtemplates.UpdateWorkflowTemplateRequest{
		TemplateName:     sgsdkgo.Optional(m.TemplateName.ValueString()),
		SourceConfigKind: sgsdkgo.Optional(workflowtemplates.WorkflowTemplateSourceConfigKindEnum(m.SourceConfigKind.ValueString())),
	}

	if !m.ShortDescription.IsNull() && !m.ShortDescription.IsUnknown() {
		apiModel.ShortDescription = sgsdkgo.Optional(m.ShortDescription.ValueString())
	} else {
		apiModel.ShortDescription = sgsdkgo.Null[string]()
	}

	if !m.IsPublic.IsNull() && !m.IsPublic.IsUnknown() {
		apiModel.IsPublic = sgsdkgo.Optional(sgsdkgo.IsPublicEnum(m.IsPublic.ValueString()))
	}

	// Convert Tags
	tags, diags := expanders.StringList(ctx, m.Tags)
	if diags.HasError() {
		return nil, diags
	}
	if tags != nil {
		apiModel.Tags = sgsdkgo.Optional(tags)
	} else {
		apiModel.Tags = sgsdkgo.Null[[]string]()
	}

	// Convert ContextTags
	contextTags, diags := expanders.MapStringString(ctx, m.ContextTags)
	if contextTags != nil {
		apiModel.ContextTags = sgsdkgo.Optional(contextTags)
	} else {
		apiModel.ContextTags = sgsdkgo.Null[map[string]string]()
	}

	// Convert RuntimeSource
	if !m.RuntimeSource.IsNull() && !m.RuntimeSource.IsUnknown() {
		var runtimeSourceModel RuntimeSourceModel
		diag_rt := m.RuntimeSource.As(ctx, &runtimeSourceModel, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		if diag_rt.HasError() {
			return nil, diag_rt
		}

		runtimeSource := &workflowtemplates.RuntimeSourceUpdate{
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

			runtimeSource.Config = &workflowtemplates.RuntimeSourceConfigUpdate{
				IsPrivate:               configModel.IsPrivate.ValueBoolPointer(),
				GitCoreAutoCRLF:         configModel.GitCoreAutoCrlf.ValueBoolPointer(),
				GitSparseCheckoutConfig: configModel.GitSparseCheckoutConfig.ValueStringPointer(),
				IncludeSubModule:        configModel.IncludeSubModule.ValueBoolPointer(),
				Ref:                     configModel.Ref.ValueStringPointer(),
				WorkingDir:              configModel.WorkingDir.ValueStringPointer(),
			}
		}
		apiModel.RuntimeSource = sgsdkgo.Optional(*runtimeSource)
	} else {
		apiModel.RuntimeSource = sgsdkgo.Null[workflowtemplates.RuntimeSourceUpdate]()
	}

	// convert SharedOrgsList
	sharedOrgsList, diags := expanders.StringList(ctx, m.SharedOrgsList)
	if diags.HasError() {
		return nil, diags
	}
	if sharedOrgsList != nil {
		apiModel.SharedOrgsList = sgsdkgo.Optional(sharedOrgsList)
	} else {
		apiModel.SharedOrgsList = sgsdkgo.Null[[]string]()
	}

	// convert VCSTriggers
	vcsTriggersAPIModel, diags := VCSTriggersToAPIModel(ctx, m.VCSTriggers)
	if diags.HasError() {
		return nil, diags
	}
	if vcsTriggersAPIModel != nil {
		apiModel.VCSTriggers = sgsdkgo.Optional(*vcsTriggersAPIModel)
	} else {
		apiModel.VCSTriggers = sgsdkgo.Null[workflowtemplates.VCSTriggers]()
	}

	return apiModel, diag
}

func VCSTriggersToTerraType(vcsTriggers *workflowtemplates.VCSTriggers) (types.Object, diag.Diagnostics) {
	nullObject := types.ObjectNull(VCSTriggersModel{}.AttributeTypes())
	if vcsTriggers == nil {
		return nullObject, nil
	}

	vcsTriggersModel := VCSTriggersModel{}
	if vcsTriggers.Type != nil {
		vcsTriggersModel.Type = flatteners.String(string(*vcsTriggers.Type))
	} else {
		vcsTriggersModel.Type = types.StringNull()
	}

	if vcsTriggers.CreateTag != nil {
		createTagModel := VCSTriggersCreateTagModel{}
		if vcsTriggers.CreateTag.CreateRevision != nil {
			createRevisionModel := VCSTriggersCreateRevisionModel{}
			if vcsTriggers.CreateTag.CreateRevision.Enabled != nil {
				createRevisionModel.Enabled = flatteners.BoolPtr(vcsTriggers.CreateTag.CreateRevision.Enabled)
			} else {
				createRevisionModel.Enabled = types.BoolNull()
			}
			createRevisionTerraType, diags := types.ObjectValueFrom(context.TODO(), VCSTriggersCreateRevisionModel{}.AttributeTypes(), &createRevisionModel)
			if diags.HasError() {
				return nullObject, diags
			}

			createTagModel = VCSTriggersCreateTagModel{
				CreateRevision: createRevisionTerraType,
			}
		} else {
			createTagModel.CreateRevision = types.ObjectNull(VCSTriggersCreateRevisionModel{}.AttributeTypes())
		}

		createTagTerraType, diags := types.ObjectValueFrom(context.Background(), VCSTriggersCreateTagModel{}.AttributeTypes(), &createTagModel)
		if diags.HasError() {
			return nullObject, diags
		}

		vcsTriggersModel.CreateTag = createTagTerraType
	} else {
		vcsTriggersModel.CreateTag = types.ObjectNull(VCSTriggersCreateTagModel{}.AttributeTypes())
	}

	vcsTriggersTerraType, diags := types.ObjectValueFrom(context.Background(), VCSTriggersModel{}.AttributeTypes(), vcsTriggersModel)
	if diags.HasError() {
		return nullObject, diags
	}

	return vcsTriggersTerraType, nil
}

func RuntimeSourceToTerraType(runtimeSource *workflowtemplates.RuntimeSource) (types.Object, diag.Diagnostics) {
	nullObject := types.ObjectNull(RuntimeSourceModel{}.AttributeTypes())
	if runtimeSource == nil {
		return nullObject, nil
	}

	runtimeSourceModel := RuntimeSourceModel{}

	if runtimeSource.SourceConfigDestKind != nil {
		runtimeSourceModel.SourceConfigDestKind = flatteners.String(string(*runtimeSource.SourceConfigDestKind))
	} else {
		runtimeSourceModel.SourceConfigDestKind = types.StringNull()
	}

	if runtimeSource.Config != nil {
		configModel := &RuntimeSourceConfigModel{
			IsPrivate:               flatteners.BoolPtr(runtimeSource.Config.IsPrivate),
			Auth:                    flatteners.StringPtr(runtimeSource.Config.Auth),
			GitCoreAutoCrlf:         flatteners.BoolPtr(runtimeSource.Config.GitCoreAutoCRLF),
			GitSparseCheckoutConfig: flatteners.StringPtr(runtimeSource.Config.GitSparseCheckoutConfig),
			IncludeSubModule:        flatteners.BoolPtr(runtimeSource.Config.IncludeSubModule),
			Ref:                     flatteners.StringPtr(runtimeSource.Config.Ref),
			Repo:                    flatteners.String(runtimeSource.Config.Repo),
			WorkingDir:              flatteners.StringPtr(runtimeSource.Config.WorkingDir),
		}

		configObj, diags := types.ObjectValueFrom(context.Background(), RuntimeSourceConfigModel{}.AttributeTypes(), configModel)
		if diags.HasError() {
			return nullObject, diags
		}
		runtimeSourceModel.Config = configObj
	} else {
		runtimeSourceModel.Config = types.ObjectNull(RuntimeSourceConfigModel{}.AttributeTypes())
	}

	var runtimeSourceTerraType types.Object
	runtimeSourceTerraType, diags := types.ObjectValueFrom(context.Background(), RuntimeSourceModel{}.AttributeTypes(), runtimeSourceModel)
	if diags.HasError() {
		return nullObject, diags
	}

	return runtimeSourceTerraType, nil
}

func BuildAPIModelToWorkflowTemplateModel(apiResponse *workflowtemplates.ReadWorkflowTemplateResponse) (*WorkflowTemplateResourceModel, diag.Diagnostics) {
	diag := diag.Diagnostics{}

	model := &WorkflowTemplateResourceModel{
		Id:               flatteners.StringPtr(apiResponse.Id),
		TemplateName:     flatteners.StringPtr(apiResponse.TemplateName),
		OwnerOrg:         flatteners.StringPtr(apiResponse.OwnerOrg),
		SourceConfigKind: flatteners.String(string(*apiResponse.SourceConfigKind)),
		IsPublic:         flatteners.String(string(*apiResponse.IsPublic)),
		ShortDescription: flatteners.StringPtr(apiResponse.ShortDescription),
	}

	// Convert Tags
	if apiResponse.Tags != nil {
		var tags []types.String
		for _, tag := range apiResponse.Tags {
			tags = append(tags, flatteners.String(tag))
		}
		tagsList, diags_tags := types.ListValueFrom(context.Background(), types.StringType, tags)
		diag.Append(diags_tags...)
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
		sharedOrgsList, diags_shared := types.ListValueFrom(context.Background(), types.StringType, sharedOrgs)
		diag.Append(diags_shared...)
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
		contextTagsMap, diags_ct := types.MapValueFrom(context.Background(), types.StringType, contextTags)
		diag.Append(diags_ct...)
		model.ContextTags = contextTagsMap
	} else {
		model.ContextTags = types.MapNull(types.StringType)
	}

	// Convert RuntimeSource
	runtimeSourceTerraType, diags := RuntimeSourceToTerraType(apiResponse.RuntimeSource)
	if diags.HasError() {
		return nil, diags
	}
	model.RuntimeSource = runtimeSourceTerraType

	// Convert VCSTriggers
	vcsTriggersTerraType, diags := VCSTriggersToTerraType(apiResponse.VCSTriggers)
	if diags.HasError() {
		return nil, diags
	}
	model.VCSTriggers = vcsTriggersTerraType

	return model, diag
}

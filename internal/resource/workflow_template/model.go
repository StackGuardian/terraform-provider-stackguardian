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

func (m RuntimeSourceModel) ToAPIModel(ctx context.Context) (*workflowtemplates.RuntimeSource, diag.Diagnostics) {
	runtimeSource := &workflowtemplates.RuntimeSource{}

	// source_config_dest_kind is Optional (not Computed), but ValueString() still
	// returns "" for a Null value, and .Ptr() always produces a non-nil pointer — so
	// without this guard an omitted value would be sent to the API as an explicit
	// empty string instead of being left out of the payload.
	if !m.SourceConfigDestKind.IsNull() && !m.SourceConfigDestKind.IsUnknown() {
		runtimeSource.SourceConfigDestKind = workflowtemplates.SourceConfigDestKindEnum(m.SourceConfigDestKind.ValueString()).Ptr()
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
			Auth:                    configModel.Auth.ValueStringPointer(),
			GitSparseCheckoutConfig: configModel.GitSparseCheckoutConfig.ValueStringPointer(),
			IncludeSubModule:        configModel.IncludeSubModule.ValueBoolPointer(),
			Repo:                    configModel.Repo.ValueString(),
			WorkingDir:              configModel.WorkingDir.ValueStringPointer(),
		}

		// is_private, git_core_auto_crlf, and ref are Optional+Computed: when unset in
		// config they are Unknown (not Null) on Create, and ValueBoolPointer()/
		// ValueStringPointer() return a pointer to the zero value for Unknown rather than
		// nil. Since these API fields are pointer types with `omitempty`, a non-nil
		// zero-value pointer is still marshaled, so the guard is required to actually omit
		// the field.
		if !configModel.IsPrivate.IsNull() && !configModel.IsPrivate.IsUnknown() {
			runtimeSource.Config.IsPrivate = configModel.IsPrivate.ValueBoolPointer()
		}
		if !configModel.GitCoreAutoCrlf.IsNull() && !configModel.GitCoreAutoCrlf.IsUnknown() {
			runtimeSource.Config.GitCoreAutoCRLF = configModel.GitCoreAutoCrlf.ValueBoolPointer()
		}
		if !configModel.Ref.IsNull() && !configModel.Ref.IsUnknown() {
			runtimeSource.Config.Ref = configModel.Ref.ValueStringPointer()
		}
	}
	return runtimeSource, nil
}

// runtimeSourceConfigToUpdate maps a Create-shaped *RuntimeSourceConfig (already
// null/unknown-guarded by RuntimeSourceModel.ToAPIModel above) to the Update-shaped
// *RuntimeSourceConfigUpdate. No guards needed here — cfg's fields are already nil
// exactly where they should be omitted. repo has no Update-shaped equivalent (the SDK
// has no way to change it — see ValidateRuntimeSourceRepoUnchanged), so it's dropped.
func runtimeSourceConfigToUpdate(cfg *workflowtemplates.RuntimeSourceConfig) *workflowtemplates.RuntimeSourceConfigUpdate {
	if cfg == nil {
		return nil
	}
	return &workflowtemplates.RuntimeSourceConfigUpdate{
		Auth:                    cfg.Auth,
		GitCoreAutoCRLF:         cfg.GitCoreAutoCRLF,
		GitSparseCheckoutConfig: cfg.GitSparseCheckoutConfig,
		IncludeSubModule:        cfg.IncludeSubModule,
		IsPrivate:               cfg.IsPrivate,
		Ref:                     cfg.Ref,
		WorkingDir:              cfg.WorkingDir,
	}
}

// ConvertRuntimeSourceToUpdateAPI converts a runtime_source types.Object to
// *workflowtemplates.RuntimeSourceUpdate for an Update request, by reusing
// RuntimeSourceModel.ToAPIModel's guarded Create-path conversion (see its own doc
// comment for why the is_private/git_core_auto_crlf/ref guards are needed) and mapping
// the result through runtimeSourceConfigToUpdate, rather than re-deriving the same
// guards a second time in a separately-maintained function. Shared by workflow_template
// and workflow_template_revision's ToUpdateAPIModel.
func ConvertRuntimeSourceToUpdateAPI(ctx context.Context, runtimeSourceObj types.Object) (*workflowtemplates.RuntimeSourceUpdate, diag.Diagnostics) {
	if runtimeSourceObj.IsNull() || runtimeSourceObj.IsUnknown() {
		return nil, nil
	}

	var m RuntimeSourceModel
	diags := runtimeSourceObj.As(ctx, &m, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	if diags.HasError() {
		return nil, diags
	}

	rs, diags := m.ToAPIModel(ctx)
	if diags.HasError() {
		return nil, diags
	}

	return &workflowtemplates.RuntimeSourceUpdate{
		SourceConfigDestKind: rs.SourceConfigDestKind,
		Config:               runtimeSourceConfigToUpdate(rs.Config),
	}, diags
}

func (m *WorkflowTemplateResourceModel) ToAPIModel(ctx context.Context) (*workflowtemplates.CreateWorkflowTemplateRequest, diag.Diagnostics) {
	diag := diag.Diagnostics{}

	apiModel := &workflowtemplates.CreateWorkflowTemplateRequest{
		TemplateName:     m.TemplateName.ValueString(),
		OwnerOrg:         m.OwnerOrg.ValueString(),
		ShortDescription: m.ShortDescription.ValueStringPointer(),
	}

	// id is Optional+Computed: a practitioner-supplied value must reach the API so the
	// created resource's id matches what was planned. Without this, the server always
	// assigns its own id, and Terraform's post-apply consistency check fails whenever
	// the config sets id explicitly (it differs from the auto-assigned value).
	if !m.Id.IsNull() && !m.Id.IsUnknown() {
		apiModel.Id = m.Id.ValueStringPointer()
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
	if diags.HasError() {
		return nil, diags
	}
	if contextTags != nil {
		apiModel.ContextTags = sgsdkgo.Optional(contextTags)
	} else {
		apiModel.ContextTags = sgsdkgo.Null[map[string]string]()
	}

	// Convert RuntimeSource
	if !m.RuntimeSource.IsNull() && !m.RuntimeSource.IsUnknown() {
		runtimeSource, diags := ConvertRuntimeSourceToUpdateAPI(ctx, m.RuntimeSource)
		if diags.HasError() {
			return nil, diags
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

	return apiModel, diag
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

	return model, diag
}

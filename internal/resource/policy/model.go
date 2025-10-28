package policy

import (
	"context"
	"encoding/json"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/expanders"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/flatteners"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type PolicyResourceModel struct {
	Id                        types.String `tfsdk:"id"`
	ResourceName              types.String `tfsdk:"resource_name"`
	PolicyType                types.String `tfsdk:"policy_type"`
	Description               types.String `tfsdk:"description"`
	NumberOfApprovalsRequired types.Int32  `tfsdk:"number_of_approvals_required"`
	Approvers                 types.List   `tfsdk:"approvers"`
	EnforcedOn                types.List   `tfsdk:"enforced_on"`
	Tags                      types.List   `tfsdk:"tags"`
	PoliciesConfig            types.List   `tfsdk:"policies_config"`
}

type policyPoliciesConfigModel struct {
	Name            types.String `tfsdk:"name"`
	Skip            types.Bool   `tfsdk:"skip"`
	OnFail          types.String `tfsdk:"on_fail"`
	OnPass          types.String `tfsdk:"on_pass"`
	PolicyInputData types.Object `tfsdk:"policy_input_data"`
	PolicyVCSConfig types.Object `tfsdk:"policy_vcs_config"`
}

func (m policyPoliciesConfigModel) AttributeTypes() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":              types.StringType,
			"skip":              types.BoolType,
			"on_fail":           types.StringType,
			"on_pass":           types.StringType,
			"policy_input_data": types.ObjectType{AttrTypes: policyInputDataModel{}.AttributeTypes()},
			"policy_vcs_config": types.ObjectType{AttrTypes: policyVCSConfigModel{}.AttributeTypes()},
		},
	}
}

func (m *policyPoliciesConfigModel) ToAPIModel() (*sgsdkgo.PoliciesConfig, diag.Diagnostics) {
	policyPoliciesConfigAPIModel := &sgsdkgo.PoliciesConfig{
		Name: m.Name.ValueString(),
		Skip: m.Skip.ValueBoolPointer(),
	}

	if !m.OnFail.IsNull() && !m.OnFail.IsUnknown() {
		onFailEnum, err := sgsdkgo.NewOnFailEnumFromString(m.OnFail.ValueString())
		if err != nil {
			return nil, []diag.Diagnostic{diag.NewErrorDiagnostic("Fail to convert policyConfig terraform type to go type", err.Error())}
		}
		policyPoliciesConfigAPIModel.OnFail = onFailEnum
	}

	if !m.OnPass.IsNull() && !m.OnPass.IsUnknown() {
		onPassEnum, err := sgsdkgo.NewOnPassEnumFromString(m.OnPass.ValueString())
		if err != nil {
			return nil, []diag.Diagnostic{diag.NewErrorDiagnostic("Fail to convert policyConfig terraform type to go type", err.Error())}
		}
		policyPoliciesConfigAPIModel.OnPass = onPassEnum
	}

	if !m.PolicyInputData.IsNull() && !m.PolicyInputData.IsUnknown() {
		var policyInputDataModel policyInputDataModel

		diags := m.PolicyInputData.As(context.TODO(), &policyInputDataModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
		if diags.HasError() {
			return nil, diags
		}

		policyPoliciesConfigAPIModel.PolicyInputData, diags = policyInputDataModel.ToAPIModel()
		if diags.HasError() {
			return nil, diags
		}
	}

	if !m.PolicyVCSConfig.IsNull() && !m.PolicyVCSConfig.IsUnknown() {
		var policyVCSConfigModel policyVCSConfigModel
		diags := m.PolicyVCSConfig.As(context.TODO(), &policyVCSConfigModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
		if diags.HasError() {
			return nil, diags
		}

		policyPoliciesConfigAPIModel.PolicyVcsConfig, diags = policyVCSConfigModel.ToAPIModel()
		if diags.HasError() {
			return nil, diags
		}
	}

	return policyPoliciesConfigAPIModel, nil
}

type policyInputDataModel struct {
	SchemaType types.String `tfsdk:"schema_type"`
	Data       types.String `tfsdk:"data"`
}

func (m policyInputDataModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"schema_type": types.StringType,
		"data":        types.StringType,
	}
}

func (m *policyInputDataModel) ToAPIModel() (*sgsdkgo.InputData, diag.Diagnostics) {
	policyInputData := &sgsdkgo.InputData{}

	if !m.SchemaType.IsNull() && !m.SchemaType.IsUnknown() {
		schemaType, err := sgsdkgo.NewInputDataSchemaTypeEnumFromString(m.SchemaType.ValueString())
		if err != nil {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Fail to convert policyConfig terraform type to go type", err.Error())}
		}
		policyInputData.SchemaType = schemaType
	}

	if !m.Data.IsNull() && !m.Data.IsUnknown() {
		dataString := m.Data.ValueString()

		var dataMap map[string]interface{}
		err := json.Unmarshal([]byte(dataString), &dataMap)
		if err != nil {
			return nil, []diag.Diagnostic{diag.NewErrorDiagnostic("Fail to convert policyConfig terraform type to go type", err.Error())}
		}
		policyInputData.Data = dataMap
	}
	return policyInputData, nil
}

type policyVCSConfigModel struct {
	UseMarketplaceTemplate types.Bool   `tfsdk:"use_marketplace_template"`
	PolicyTemplateId       types.String `tfsdk:"policy_template_id"`
	CustomSource           types.Object `tfsdk:"custom_source"`
}

func (m policyVCSConfigModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"use_marketplace_template": types.BoolType,
		"policy_template_id":       types.StringType,
		"custom_source":            types.ObjectType{AttrTypes: policyCustomSourceModel{}.AttributeTypes()},
	}
}

func (m *policyVCSConfigModel) ToAPIModel() (*sgsdkgo.PolicyVcsConfig, diag.Diagnostics) {
	policyVCSConfigAPIModel := sgsdkgo.PolicyVcsConfig{
		UseMarketplaceTemplate: m.UseMarketplaceTemplate.ValueBool(),
		PolicyTemplateId:       m.PolicyTemplateId.ValueStringPointer(),
	}

	if !m.CustomSource.IsNull() && !m.CustomSource.IsUnknown() {
		var policyCustomSourceModel policyCustomSourceModel
		diags := m.CustomSource.As(context.TODO(), &policyCustomSourceModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
		if diags.HasError() {
			return nil, diags
		}

		// TODO: Implement logic for additional config

		policyVCSConfigAPIModel.CustomSource, diags = policyCustomSourceModel.ToAPIModel()
		if diags.HasError() {
			return nil, diags
		}
	}

	return &policyVCSConfigAPIModel, nil
}

type policyCustomSourceModel struct {
	SourceConfigDestKind types.String `tfsdk:"source_config_dest_kind"`
	SourceConfigKind     types.String `tfsdk:"source_config_kind"`
	Config               types.Object `tfsdk:"config"`
	AdditionalConfig     types.String `tfsdk:"additional_config"`
}

func (m policyCustomSourceModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"source_config_dest_kind": types.StringType,
		"source_config_kind":      types.StringType,
		"additional_config":       types.StringType,
		"config":                  types.ObjectType{AttrTypes: policyCustomSourceConfigModel{}.AttributeTypes()},
	}
}

func (m *policyCustomSourceModel) ToAPIModel() (*sgsdkgo.CustomSourcePolicy, diag.Diagnostics) {
	policyCustomSourceAPIModel := &sgsdkgo.CustomSourcePolicy{}

	if !m.SourceConfigDestKind.IsNull() && !m.SourceConfigDestKind.IsUnknown() {
		sourceConfigDestKind, err := sgsdkgo.NewCustomSourcePolicySourceConfigDestKindEnumFromString(m.SourceConfigDestKind.ValueString())
		if err != nil {
			return nil, []diag.Diagnostic{diag.NewErrorDiagnostic("Fail to convert policyConfig terraform type to go type", err.Error())}
		}
		policyCustomSourceAPIModel.SourceConfigDestKind = sourceConfigDestKind
	}

	if !m.SourceConfigKind.IsNull() && !m.SourceConfigKind.IsUnknown() {
		sourceConfigKind, err := sgsdkgo.NewSourceConfigKindEnumFromString(m.SourceConfigKind.ValueString())
		if err != nil {
			return nil, []diag.Diagnostic{diag.NewErrorDiagnostic("Fail to convert policyConfig terraform type to go type", err.Error())}
		}
		policyCustomSourceAPIModel.SourceConfigKind = sourceConfigKind
	}

	if !m.Config.IsNull() && !m.Config.IsUnknown() {
		var policyConfigModel policyCustomSourceConfigModel
		diags := m.Config.As(context.TODO(), &policyConfigModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
		if diags.HasError() {
			return nil, diags
		}

		policyCustomSourceAPIModel.Config, diags = policyConfigModel.ToAPIModel()
		if diags.HasError() {
			return nil, diags
		}
	}

	if !m.AdditionalConfig.IsNull() && !m.AdditionalConfig.IsUnknown() {
		var additionalConfig map[string]interface{}
		err := json.Unmarshal([]byte(m.AdditionalConfig.ValueString()), &additionalConfig)
		if err != nil {
			return nil, diag.Diagnostics{diag.NewErrorDiagnostic("Fail to unmarshal additional config", err.Error())}
		}
		policyCustomSourceAPIModel.AdditionalConfig = additionalConfig
	}

	return policyCustomSourceAPIModel, nil
}

type policyCustomSourceConfigModel struct {
	IsPrivate               types.Bool   `tfsdk:"is_private"`
	Auth                    types.String `tfsdk:"auth"`
	WorkingDir              types.String `tfsdk:"working_dir"`
	Ref                     types.String `tfsdk:"ref"`
	Repo                    types.String `tfsdk:"repo"`
	IncludeSubModule        types.Bool   `tfsdk:"include_submodule"`
	GitSparseCheckoutConfig types.String `tfsdk:"git_sparse_checkout_config"`
	GitCoreAutoCRLF         types.Bool   `tfsdk:"git_core_auto_crlf"`
}

func (m policyCustomSourceConfigModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"is_private":                 types.BoolType,
		"auth":                       types.StringType,
		"working_dir":                types.StringType,
		"ref":                        types.StringType,
		"repo":                       types.StringType,
		"include_submodule":          types.BoolType,
		"git_sparse_checkout_config": types.StringType,
		"git_core_auto_crlf":         types.BoolType,
	}
}

func (m *policyCustomSourceConfigModel) ToAPIModel() (*sgsdkgo.CustomSourcePolicyConfig, diag.Diagnostics) {
	policyConfigAPIModel := &sgsdkgo.CustomSourcePolicyConfig{
		IsPrivate:               m.IsPrivate.ValueBoolPointer(),
		IncludeSubModule:        m.IncludeSubModule.ValueBoolPointer(),
		Ref:                     m.Ref.ValueStringPointer(),
		GitCoreAutoCrlf:         m.GitCoreAutoCRLF.ValueBoolPointer(),
		GitSparseCheckoutConfig: m.GitSparseCheckoutConfig.ValueStringPointer(),
		Auth:                    m.Auth.ValueStringPointer(),
		WorkingDir:              m.WorkingDir.ValueStringPointer(),
		Repo:                    m.Repo.ValueStringPointer(),
	}

	return policyConfigAPIModel, nil
}

func policiesConfigModelToAPIModel(policiesConfig types.List) ([]*sgsdkgo.PoliciesConfig, diag.Diagnostics) {
	if policiesConfig.IsNull() || policiesConfig.IsUnknown() {
		return nil, nil
	}
	var policiesConfigModel []*policyPoliciesConfigModel

	diags := policiesConfig.ElementsAs(context.TODO(), &policiesConfigModel, true)
	if diags.HasError() {
		return nil, diags
	}

	policiesConfigAPIModel := []*sgsdkgo.PoliciesConfig{}
	for _, policyConfig := range policiesConfigModel {
		policyConfigAPIModel, diags := policyConfig.ToAPIModel()
		if diags.HasError() {
			return nil, diags
		}
		policiesConfigAPIModel = append(policiesConfigAPIModel, policyConfigAPIModel)
	}

	return policiesConfigAPIModel, nil
}

func (m *PolicyResourceModel) ToAPIModel() (*sgsdkgo.PolicyGeneral, diag.Diagnostics) {
	policyAPIModel := &sgsdkgo.PolicyGeneral{
		ResourceName: m.ResourceName.ValueStringPointer(),
	}

	if !m.Description.IsNull() {
		policyAPIModel.Description = m.Description.ValueStringPointer()
	}

	if !m.NumberOfApprovalsRequired.IsNull() {
		policyAPIModel.NumberOfApprovalsRequired = expanders.IntPtr(m.NumberOfApprovalsRequired.ValueInt32Pointer())
	}

	// Approvers
	approvers, diags := expanders.StringList(context.TODO(), m.Approvers)
	if diags.HasError() {
		return nil, diags
	} else if approvers != nil {
		policyAPIModel.Approvers = approvers
	}

	// Enforced On
	enforcedOn, diags := expanders.StringList(context.TODO(), m.EnforcedOn)
	if diags.HasError() {
		return nil, diags
	} else if enforcedOn != nil {
		policyAPIModel.EnforcedOn = enforcedOn
	}

	// Tags
	tags, diags := expanders.StringList(context.TODO(), m.Tags)
	if diags.HasError() {
		return nil, diags
	} else if tags != nil {
		policyAPIModel.Tags = tags
	}

	policiesConfig, diags := policiesConfigModelToAPIModel(m.PoliciesConfig)
	if diags.HasError() {
		return nil, diags
	} else if policiesConfig != nil {
		policyAPIModel.PoliciesConfig = policiesConfig
	}

	return policyAPIModel, nil
}

func (m *PolicyResourceModel) ToPatchedAPIModel() (*sgsdkgo.PatchedPolicyGeneral, diag.Diagnostics) {
	policyAPIModel := &sgsdkgo.PatchedPolicyGeneral{
		ResourceName: sgsdkgo.Optional(m.ResourceName.ValueString()),
	}

	if !m.Description.IsNull() {
		policyAPIModel.Description = sgsdkgo.Optional(m.Description.ValueString())
	} else {
		policyAPIModel.Description = sgsdkgo.Null[string]()
	}

	if !m.NumberOfApprovalsRequired.IsNull() {
		policyAPIModel.NumberOfApprovalsRequired = sgsdkgo.Optional(*expanders.IntPtr(m.NumberOfApprovalsRequired.ValueInt32Pointer()))
	}

	// Approvers
	approvers, diags := expanders.StringList(context.TODO(), m.Approvers)
	if diags.HasError() {
		return nil, diags
	} else if approvers != nil {
		policyAPIModel.Approvers = sgsdkgo.Optional(approvers)
	} else {
		policyAPIModel.Approvers = sgsdkgo.Null[[]string]()
	}

	// Enforced On
	enforcedOn, diags := expanders.StringList(context.TODO(), m.EnforcedOn)
	if diags.HasError() {
		return nil, diags
	} else if enforcedOn != nil {
		policyAPIModel.EnforcedOn = sgsdkgo.Optional(enforcedOn)
	} else {
		policyAPIModel.EnforcedOn = sgsdkgo.Null[[]string]()
	}

	// Tags
	tags, diags := expanders.StringList(context.TODO(), m.Tags)
	if diags.HasError() {
		return nil, diags
	} else if tags != nil {
		policyAPIModel.Tags = sgsdkgo.Optional(tags)
	} else {
		policyAPIModel.Tags = sgsdkgo.Null[[]string]()
	}

	policiesConfig, diags := policiesConfigModelToAPIModel(m.PoliciesConfig)
	if diags.HasError() {
		return nil, diags
	} else if policiesConfig != nil {
		policyAPIModel.PoliciesConfig = sgsdkgo.Optional(policiesConfig)
	} else {
		policyAPIModel.PoliciesConfig = sgsdkgo.Null[[]*sgsdkgo.PoliciesConfig]()
	}

	return policyAPIModel, nil
}

func BuildAPIModelToPolicyModel(apiResponse *sgsdkgo.PolicyGeneralResponse) (*PolicyResourceModel, diag.Diagnostics) {
	policyConfigModel := &PolicyResourceModel{
		Id:                        flatteners.String(apiResponse.Id),
		ResourceName:              flatteners.StringPtr(apiResponse.ResourceName),
		PolicyType:                flatteners.String("GENERAL"),
		Description:               flatteners.StringPtr(apiResponse.Description),
		NumberOfApprovalsRequired: flatteners.Int32Ptr(apiResponse.NumberOfApprovalsRequired),
	}

	// Approvers
	approvers, diags := flatteners.ListOfStringToTerraformList(apiResponse.Approvers)
	if diags.HasError() {
		return nil, diags
	}
	policyConfigModel.Approvers = approvers

	// Tags
	tags, diags := flatteners.ListOfStringToTerraformList(apiResponse.Tags)
	if diags.HasError() {
		return nil, diags
	}
	policyConfigModel.Tags = tags

	// EnforcedOn
	enforcedOn, diags := flatteners.ListOfStringToTerraformList(apiResponse.EnforcedOn)
	if diags.HasError() {
		return nil, diags
	}
	policyConfigModel.EnforcedOn = enforcedOn

	if apiResponse.PoliciesConfig != nil {
		var policiesConfigs []*policyPoliciesConfigModel
		for _, policiesConfig := range apiResponse.PoliciesConfig {
			policiesConfigModel, diags := policiesConfigToTerraType(policiesConfig)
			if diags.HasError() {
				return nil, diags
			}
			policiesConfigs = append(policiesConfigs, policiesConfigModel)
		}

		terraType, diags := types.ListValueFrom(context.TODO(), policyPoliciesConfigModel{}.AttributeTypes(), &policiesConfigs)
		if diags.HasError() {
			return nil, diags
		}

		policyConfigModel.PoliciesConfig = terraType
	} else {
		policyConfigModel.PoliciesConfig = types.ListNull(policyPoliciesConfigModel{}.AttributeTypes())
	}

	return policyConfigModel, nil
}

func policiesConfigToTerraType(policiesConfig *sgsdkgo.PoliciesConfig) (*policyPoliciesConfigModel, diag.Diagnostics) {

	if policiesConfig == nil {
		return nil, nil
	}

	policyPoliciesConfigModel := policyPoliciesConfigModel{
		Name:   flatteners.String(policiesConfig.Name),
		Skip:   flatteners.BoolPtr(policiesConfig.Skip),
		OnFail: flatteners.String(string(policiesConfig.OnFail)),
		OnPass: flatteners.String(string(policiesConfig.OnPass)),
	}

	//PolicyInputData types.Object `tfsdk:"policy_input_data"`
	policyInputDataModel, diags := InputDataToTerraType(policiesConfig.PolicyInputData)
	if diags.HasError() {
		return nil, diags
	}
	policyPoliciesConfigModel.PolicyInputData = policyInputDataModel

	//PolicyVCSConfig types.Object `tfsdk:"policy_vcs_config"`
	policyVCSConfigModel, diags := VCSConfigToTerraType(policiesConfig.PolicyVcsConfig)
	if diags.HasError() {
		return nil, diags
	}
	policyPoliciesConfigModel.PolicyVCSConfig = policyVCSConfigModel

	return &policyPoliciesConfigModel, nil
}

func InputDataToTerraType(policyInputData *sgsdkgo.InputData) (types.Object, diag.Diagnostics) {
	nullObject := types.ObjectNull(policyInputDataModel{}.AttributeTypes())
	if policyInputData == nil {
		return nullObject, nil
	}

	policyInputDataModel := &policyInputDataModel{
		SchemaType: flatteners.String(string(policyInputData.SchemaType)),
	}

	dataByte, err := json.Marshal(policyInputData.Data)
	if err != nil {
		return nullObject, diag.Diagnostics{diag.NewErrorDiagnostic("Fail to marshal create data in InputData", err.Error())}
	}
	policyInputDataModel.Data = types.StringValue(string(dataByte))

	terraType, diags := types.ObjectValueFrom(context.TODO(), policyInputDataModel.AttributeTypes(), &policyInputDataModel)
	if diags.HasError() {
		return nullObject, diags
	}

	return terraType, nil
}

func VCSConfigToTerraType(VCSConfig *sgsdkgo.PolicyVcsConfig) (types.Object, diag.Diagnostics) {
	objectNull := types.ObjectNull(policyVCSConfigModel{}.AttributeTypes())
	if VCSConfig == nil {
		return objectNull, nil
	}

	VCSConfigModel := policyVCSConfigModel{
		UseMarketplaceTemplate: flatteners.BoolPtr(&VCSConfig.UseMarketplaceTemplate),
		PolicyTemplateId:       flatteners.StringPtr(VCSConfig.PolicyTemplateId),
	}

	customSourceModel, diags := CustomSourceToTerraType(VCSConfig.CustomSource)
	if diags.HasError() {
		return objectNull, diags
	}

	VCSConfigModel.CustomSource = customSourceModel

	terraType, diags := types.ObjectValueFrom(context.TODO(), VCSConfigModel.AttributeTypes(), &VCSConfigModel)
	if diags.HasError() {
		return objectNull, diags
	}

	return terraType, nil
}

func CustomSourceToTerraType(customSourcePolicy *sgsdkgo.CustomSourcePolicy) (types.Object, diag.Diagnostics) {
	objectNull := types.ObjectNull(policyCustomSourceModel{}.AttributeTypes())
	if customSourcePolicy == nil {
		return objectNull, nil
	}

	customSourcePolicyModel := policyCustomSourceModel{
		SourceConfigDestKind: flatteners.String(string(customSourcePolicy.SourceConfigDestKind)),
		SourceConfigKind:     flatteners.String(string(customSourcePolicy.SourceConfigKind)),
	}

	// Config
	configModel, diags := CustomSourceConfigToTerraType(customSourcePolicy.Config)
	if diags.HasError() {
		return objectNull, nil
	}
	customSourcePolicyModel.Config = configModel

	//AdditionalConfig
	if !customSourcePolicyModel.AdditionalConfig.IsNull() && !customSourcePolicyModel.AdditionalConfig.IsUnknown() {
		additionalConfigModel, err := json.Marshal(customSourcePolicy.AdditionalConfig)
		if err != nil {
			return objectNull, diag.Diagnostics{diag.NewErrorDiagnostic("Fail to marshal addtional config in Policy resource", err.Error())}
		}
		customSourcePolicyModel.AdditionalConfig = types.StringValue(string(additionalConfigModel))
	} else {
		customSourcePolicyModel.AdditionalConfig = types.StringNull()
	}

	terraType, diags := types.ObjectValueFrom(context.TODO(), customSourcePolicyModel.AttributeTypes(), customSourcePolicyModel)
	if diags.HasError() {
		return objectNull, diags
	}

	return terraType, nil
}

func CustomSourceConfigToTerraType(customSourcePolicyConfig *sgsdkgo.CustomSourcePolicyConfig) (types.Object, diag.Diagnostics) {
	objectNull := types.ObjectNull(policyCustomSourceConfigModel{}.AttributeTypes())
	if customSourcePolicyConfig == nil {
		return objectNull, nil
	}

	customSourceConfigModel := policyCustomSourceConfigModel{
		IsPrivate:               flatteners.BoolPtr(customSourcePolicyConfig.IsPrivate),
		Auth:                    flatteners.StringPtr(customSourcePolicyConfig.Auth),
		WorkingDir:              flatteners.StringPtr(customSourcePolicyConfig.WorkingDir),
		Ref:                     flatteners.StringPtr(customSourcePolicyConfig.Ref),
		Repo:                    flatteners.StringPtr(customSourcePolicyConfig.Repo),
		IncludeSubModule:        flatteners.BoolPtr(customSourcePolicyConfig.IncludeSubModule),
		GitSparseCheckoutConfig: flatteners.StringPtr(customSourcePolicyConfig.GitSparseCheckoutConfig),
		GitCoreAutoCRLF:         flatteners.BoolPtr(customSourcePolicyConfig.GitCoreAutoCrlf),
	}

	terraType, diags := types.ObjectValueFrom(context.TODO(), customSourceConfigModel.AttributeTypes(), &customSourceConfigModel)
	if diags.HasError() {
		return objectNull, nil
	}

	return terraType, diags
}

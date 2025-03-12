package runnergroup

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

type RunnerGroupResourceModel struct {
	ResourceName               types.String `tfsdk:"resource_name"`
	Description                types.String `tfsdk:"description"`
	RunnerToken                types.String `tfsdk:"runner_token"`
	MaxNumberOfRunners         types.Int32  `tfsdk:"max_number_of_runners"`
	Tags                       types.List   `tfsdk:"tags"`
	StorageBackendConfig       types.Object `tfsdk:"storage_backend_config"`
	RunControllerRuntimeSource types.Object `tfsdk:"run_controller_runtime_source"`
}

func (m *RunnerGroupResourceModel) ToAPIModel() (*sgsdkgo.RunnerGroup, diag.Diagnostics) {
	runnerGroupAPIModel := &sgsdkgo.RunnerGroup{
		ResourceName: m.ResourceName.ValueStringPointer(),
	}

	if !m.RunnerToken.IsUnknown() && !m.RunnerToken.IsNull() {
		runnerGroupAPIModel.RunnerToken = m.RunnerToken.ValueStringPointer()
	}

	if !m.Description.IsUnknown() && !m.Description.IsNull() {
		runnerGroupAPIModel.Description = m.Description.ValueStringPointer()
	}

	if !m.MaxNumberOfRunners.IsUnknown() && !m.MaxNumberOfRunners.IsNull() {
		runnerGroupAPIModel.MaxNumberOfRunners = expanders.IntPtr(m.MaxNumberOfRunners.ValueInt32Pointer())
	}

	// Tags
	if !m.Tags.IsNull() && !m.Tags.IsUnknown() {
		tags, diags := expanders.StringList(context.TODO(), m.Tags)
		if diags.HasError() {
			return nil, diags
		} else if tags != nil {
			runnerGroupAPIModel.Tags = tags
		}
	}

	if !m.StorageBackendConfig.IsNull() && !m.StorageBackendConfig.IsUnknown() {
		var storageBackendConfigModel storageBackendConfigModel
		diags := m.StorageBackendConfig.As(context.TODO(), &storageBackendConfigModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
		if diags.HasError() {
			return nil, diags
		}
		runnerGroupAPIModel.StorageBackendConfig, diags = storageBackendConfigModel.ToAPIModel()
		if diags.HasError() {
			return nil, diags
		}
	}

	if !m.RunControllerRuntimeSource.IsNull() && !m.RunControllerRuntimeSource.IsUnknown() {
		var runControllerRuntimeSourceModel storageBackendRunControllerRuntimeSource
		diags := m.RunControllerRuntimeSource.As(context.TODO(), &runControllerRuntimeSourceModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		if diags.HasError() {
			return nil, diags
		}
		runnerGroupAPIModel.RunControllerRuntimeSource, diags = runControllerRuntimeSourceModel.ToAPIModel()
		if diags.HasError() {
			return nil, diags
		}
	}

	return runnerGroupAPIModel, nil
}

func (m *RunnerGroupResourceModel) ToPatchedAPIModel() (*sgsdkgo.PatchedRunnerGroup, diag.Diagnostics) {
	patchedRunnerGroupModel := &sgsdkgo.PatchedRunnerGroup{
		ResourceName: sgsdkgo.Optional(m.ResourceName.ValueString()),
	}

	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		patchedRunnerGroupModel.Description = sgsdkgo.Optional(m.Description.ValueString())
	} else {
		patchedRunnerGroupModel.Description = sgsdkgo.Null[string]()
	}

	if !m.RunnerToken.IsNull() && !m.RunnerToken.IsUnknown() {
		patchedRunnerGroupModel.RunnerToken = sgsdkgo.Optional(m.RunnerToken.ValueString())
	}

	if !m.MaxNumberOfRunners.IsNull() && !m.MaxNumberOfRunners.IsUnknown() {
		patchedRunnerGroupModel.MaxNumberOfRunners = sgsdkgo.Optional(int(m.MaxNumberOfRunners.ValueInt32()))
	} else {
		patchedRunnerGroupModel.MaxNumberOfRunners = sgsdkgo.Null[int]()
	}

	// Tags
	if !m.Tags.IsNull() && !m.Tags.IsUnknown() {
		tags, diags := expanders.StringList(context.TODO(), m.Tags)
		if diags.HasError() {
			return nil, diags
		}
		patchedRunnerGroupModel.Tags = sgsdkgo.Optional(tags)
	} else {
		patchedRunnerGroupModel.Tags = sgsdkgo.Null[[]string]()
	}

	if !m.StorageBackendConfig.IsNull() && !m.StorageBackendConfig.IsUnknown() {
		var storageBackendConfigModel storageBackendConfigModel
		diags := m.StorageBackendConfig.As(context.TODO(), &storageBackendConfigModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
		if diags.HasError() {
			return nil, diags
		}
		storageBackendConfig, diags := storageBackendConfigModel.ToAPIModel()
		if diags.HasError() {
			return nil, diags
		}
		patchedRunnerGroupModel.StorageBackendConfig = sgsdkgo.Optional(*storageBackendConfig)
	} else {
		patchedRunnerGroupModel.StorageBackendConfig = sgsdkgo.Null[sgsdkgo.StorageBackendConfig]()
	}

	if !m.RunControllerRuntimeSource.IsNull() && !m.RunControllerRuntimeSource.IsUnknown() {
		var runControllerRuntimeSourceModel storageBackendRunControllerRuntimeSource
		diags := m.StorageBackendConfig.As(context.TODO(), &runControllerRuntimeSourceModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		if diags.HasError() {
			return nil, diags
		}
		runtimeSource, diags := runControllerRuntimeSourceModel.ToAPIModel()
		if diags.HasError() {
			return nil, diags
		}
		patchedRunnerGroupModel.RunControllerRuntimeSource = sgsdkgo.Optional(*runtimeSource)
	} else {
		patchedRunnerGroupModel.RunControllerRuntimeSource = sgsdkgo.Null[sgsdkgo.RuntimeSource]()
	}

	return patchedRunnerGroupModel, nil
}

type storageBackendConfigModel struct {
	Type                        types.String `tfsdk:"type"`
	AzureBlobStorageAccessKey   types.String `tfsdk:"azure_blob_storage_access_key"`
	AzureBlobStorageAccountName types.String `tfsdk:"azure_blob_storage_account_name"`
	AwsRegion                   types.String `tfsdk:"aws_region"`
	S3BucketName                types.String `tfsdk:"s3_bucket_name"`
	Auth                        types.Object `tfsdk:"auth"`
}

func (m storageBackendConfigModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":                            types.StringType,
		"azure_blob_storage_access_key":   types.StringType,
		"azure_blob_storage_account_name": types.StringType,
		"aws_region":                      types.StringType,
		"s3_bucket_name":                  types.StringType,
		"auth":                            types.ObjectType{AttrTypes: storageBackendConfigAuthModel{}.AttributeTypes()},
	}
}

func (m *storageBackendConfigModel) ToAPIModel() (*sgsdkgo.StorageBackendConfig, diag.Diagnostics) {
	runnerGroupAPIModel := &sgsdkgo.StorageBackendConfig{
		Type:                        sgsdkgo.StorageBackendConfigTypeEnum(m.Type.ValueString()),
		AzureBlobStorageAccessKey:   m.AzureBlobStorageAccessKey.ValueStringPointer(),
		AzureBlobStorageAccountName: m.AzureBlobStorageAccountName.ValueStringPointer(),
		AwsRegion:                   m.AwsRegion.ValueStringPointer(),
		S3BucketName:                m.S3BucketName.ValueStringPointer(),
	}

	if !m.Auth.IsNull() && !m.Auth.IsUnknown() {
		var auth storageBackendConfigAuthModel
		diags := m.Auth.As(context.TODO(), &auth, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
		if diags.HasError() {
			return nil, diags
		}
		runnerGroupAPIModel.Auth = auth.ToAPIModel()
	}

	return runnerGroupAPIModel, nil
}

type storageBackendConfigAuthModel struct {
	IntegrationId types.String `tfsdk:"integration_id"`
}

func (m *storageBackendConfigAuthModel) ToAPIModel() *sgsdkgo.StorageBackendConfigAuth {
	return &sgsdkgo.StorageBackendConfigAuth{
		IntegrationId: m.IntegrationId.ValueString(),
	}
}

func (m storageBackendConfigAuthModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"integration_id": types.StringType,
	}
}

type storageBackendRunControllerRuntimeSource struct {
	SourceConfigDestKind types.String `tfsdk:"source_config_dest_kind"`
	Config               types.Object `tfsdk:"config"`
}

func (m storageBackendRunControllerRuntimeSource) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"source_config_dest_kind": types.StringType,
		"config":                  types.ObjectType{AttrTypes: storageBackendRunControllerRuntimeSourceConfigModel{}.AttributeTypes()},
	}
}

func (m *storageBackendRunControllerRuntimeSource) ToAPIModel() (*sgsdkgo.RuntimeSource, diag.Diagnostics) {
	runtimeSourceAPIModel := sgsdkgo.RuntimeSource{
		SourceConfigDestKind: m.SourceConfigDestKind.ValueStringPointer(),
	}

	if !m.Config.IsNull() && !m.Config.IsUnknown() {
		var runtimeSourceConfigModel storageBackendRunControllerRuntimeSourceConfigModel
		diags := m.Config.As(context.TODO(), &runtimeSourceConfigModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
		if diags.HasError() {
			return nil, diags
		}
	}

	return &runtimeSourceAPIModel, nil
}

type storageBackendRunControllerRuntimeSourceConfigModel struct {
	IncludeSubModule        types.Bool   `tfsdk:"include_sub_module"`
	Ref                     types.String `tfsdk:"ref"`
	GitCoreAutoCRLF         types.Bool   `tfsdk:"git_core_auto_crlf"`
	GitSparseCheckoutConfig types.String `tfsdk:"git_sparse_checkout_config"`
	Auth                    types.String `tfsdk:"auth"`
	WorkingDir              types.String `tfsdk:"working_dir"`
	Repo                    types.String `tfsdk:"repo"`
	DockerImage             types.String `tfsdk:"docker_image"`
	DockerRegistryUsername  types.String `tfsdk:"docker_registry_username"`
	LocalWorkspaceDir       types.String `tfsdk:"local_workspace_dir"`
}

func (m storageBackendRunControllerRuntimeSourceConfigModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"include_sub_module":         types.BoolType,
		"ref":                        types.StringType,
		"git_core_auto_crlf":         types.BoolType,
		"git_sparse_checkout_config": types.StringType,
		"auth":                       types.StringType,
		"working_dir":                types.StringType,
		"repo":                       types.StringType,
		"docker_image":               types.StringType,
		"docker_registry_username":   types.StringType,
		"local_workspace_dir":        types.StringType,
	}
}

func (m *storageBackendRunControllerRuntimeSourceConfigModel) ToAPIModel() *sgsdkgo.RuntimeSourceConfig {
	return &sgsdkgo.RuntimeSourceConfig{
		IncludeSubModule:        m.IncludeSubModule.ValueBoolPointer(),
		Ref:                     m.Ref.ValueStringPointer(),
		WorkingDir:              m.WorkingDir.ValueStringPointer(),
		LocalWorkspaceDir:       m.LocalWorkspaceDir.ValueStringPointer(),
		GitSparseCheckoutConfig: m.GitSparseCheckoutConfig.ValueStringPointer(),
		GitCoreAutoCrlf:         m.GitCoreAutoCRLF.ValueBoolPointer(),
		DockerImage:             m.DockerImage.ValueStringPointer(),
		DockerRegistryUsername:  m.DockerImage.ValueStringPointer(),
	}
}

func storageBackendConfigToTerraType(storageBackendConfig *sgsdkgo.StorageBackendConfig) (types.Object, diag.Diagnostics) {
	objectNull := types.ObjectNull(storageBackendConfigModel{}.AttributeTypes())
	if storageBackendConfig == nil {
		return objectNull, nil
	}

	storageBackendConfigModelValue := &storageBackendConfigModel{
		Type: flatteners.String(string(storageBackendConfig.Type)),

		AzureBlobStorageAccessKey:   flatteners.StringPtr(storageBackendConfig.AzureBlobStorageAccessKey),
		AzureBlobStorageAccountName: flatteners.StringPtr(storageBackendConfig.AzureBlobStorageAccountName),
		AwsRegion:                   flatteners.StringPtr(storageBackendConfig.AwsRegion),
		S3BucketName:                flatteners.StringPtr(storageBackendConfig.S3BucketName),
	}

	if storageBackendConfig.Auth != nil {
		authModel := storageBackendConfigAuthModel{
			IntegrationId: flatteners.String(storageBackendConfig.Auth.IntegrationId),
		}

		authTerraType, diags := types.ObjectValueFrom(context.TODO(), storageBackendConfigAuthModel{}.AttributeTypes(), &authModel)
		if diags.HasError() {
			return objectNull, diags
		}

		storageBackendConfigModelValue.Auth = authTerraType
	} else {
		storageBackendConfigModelValue.Auth = types.ObjectNull(storageBackendConfigAuthModel{}.AttributeTypes())
	}

	storageBackendConfigTerraType, diags := types.ObjectValueFrom(context.TODO(), storageBackendConfigModel{}.AttributeTypes(), &storageBackendConfigModelValue)
	if diags.HasError() {
		return objectNull, diags
	}

	return storageBackendConfigTerraType, nil
}

func BuildAPIModelToRunnerGroupModel(apiResponse *sgsdkgo.RunnerGroup) (*RunnerGroupResourceModel, diag.Diagnostics) {
	runnerGroupModel := &RunnerGroupResourceModel{
		ResourceName:       flatteners.StringPtr(apiResponse.ResourceName),
		Description:        flatteners.StringPtr(apiResponse.Description),
		RunnerToken:        flatteners.StringPtr(apiResponse.RunnerToken),
		MaxNumberOfRunners: flatteners.Int32Ptr(apiResponse.MaxNumberOfRunners),
	}

	// Tags
	tags, diags := flatteners.ListOfStringToTerraformList(apiResponse.Tags)
	if diags.HasError() {
		return nil, diags
	}
	runnerGroupModel.Tags = tags

	if apiResponse.RunControllerRuntimeSource != nil {
		runtimeSourceModelValue := storageBackendRunControllerRuntimeSource{
			SourceConfigDestKind: flatteners.String(string(*apiResponse.RunControllerRuntimeSource.SourceConfigDestKind)),
		}
		if apiResponse.RunControllerRuntimeSource.Config != nil {
			configModel := storageBackendRunControllerRuntimeSourceConfigModel{
				IncludeSubModule: flatteners.BoolPtr(apiResponse.RunControllerRuntimeSource.Config.IncludeSubModule),
				Ref:              flatteners.StringPtr(apiResponse.RunControllerRuntimeSource.Config.Ref),
				GitCoreAutoCRLF:  flatteners.BoolPtr(apiResponse.RunControllerRuntimeSource.Config.GitCoreAutoCrlf),
				Auth:             flatteners.StringPtr(apiResponse.RunControllerRuntimeSource.Config.Auth),
				WorkingDir:       flatteners.StringPtr(apiResponse.RunControllerRuntimeSource.Config.WorkingDir),
				Repo:             flatteners.StringPtr(apiResponse.RunControllerRuntimeSource.Config.Repo),
			}
			runtimeSourceModelValue.Config, diags = types.ObjectValueFrom(context.TODO(), configModel.AttributeTypes(), &configModel)
			if diags.HasError() {
				return nil, diags
			}
		} else {
			runtimeSourceModelValue.Config = types.ObjectNull(storageBackendRunControllerRuntimeSourceConfigModel{}.AttributeTypes())
		}
		runnerGroupModel.RunControllerRuntimeSource, diags = types.ObjectValueFrom(context.TODO(), runtimeSourceModelValue.AttributeTypes(), &runtimeSourceModelValue)
		if diags.HasError() {
			return nil, diags
		}
	} else {
		runnerGroupModel.RunControllerRuntimeSource = types.ObjectNull(storageBackendRunControllerRuntimeSource{}.AttributeTypes())
	}

	runnerGroupModel.StorageBackendConfig, diags = storageBackendConfigToTerraType(apiResponse.StorageBackendConfig)
	if diags.HasError() {
		return nil, diags
	}

	return runnerGroupModel, nil
}

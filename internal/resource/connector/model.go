package connector

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

type ConnectorResourceModel struct {
	Id                types.String `tfsdk:"id"`
	ResourceName      types.String `tfsdk:"resource_name"`
	Description       types.String `tfsdk:"description"`
	Settings          types.Object `tfsdk:"settings"`
	DiscoverySettings types.Object `tfsdk:"discovery_settings"`
	Tags              types.List   `tfsdk:"tags"`
}

type ConnectorSettingsModel struct {
	Kind   types.String `tfsdk:"kind"`
	Config types.List   `tfsdk:"config"`
}

func (ConnectorSettingsModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"kind":   types.StringType,
		"config": types.ListType{ElemType: types.ObjectType{AttrTypes: ConnectorSettingsConfigModel{}.AttributeTypes()}},
	}
}

type ConnectorSettingsConfigModel struct {
	RoleArn                 types.String `tfsdk:"role_arn"`
	ExternalId              types.String `tfsdk:"external_id"`
	DurationSeconds         types.String `tfsdk:"duration_seconds"`
	InstallationId          types.String `tfsdk:"installation_id"`
	GithubAppId             types.String `tfsdk:"github_app_id"`
	GithubAppWebhookSecret  types.String `tfsdk:"github_app_webhook_secret"`
	GithubApiUrl            types.String `tfsdk:"github_api_url"`
	GithubHttpUrl           types.String `tfsdk:"github_http_url"`
	GithubAppClientId       types.String `tfsdk:"github_app_client_id"`
	GithubAppClientSecret   types.String `tfsdk:"github_app_client_secret"`
	GithubAppPemFileContent types.String `tfsdk:"github_app_pem_file_content"`
	GithubAppWebhookURL     types.String `tfsdk:"github_app_webhook_url"`
	GitlabCreds             types.String `tfsdk:"gitlab_creds"`
	GitlabHttpUrl           types.String `tfsdk:"gitlab_http_url"`
	GitlabApiUrl            types.String `tfsdk:"gitlab_api_url"`
	AzureCreds              types.String `tfsdk:"azure_creds"`
	AzureDevopsHttpUrl      types.String `tfsdk:"azure_devops_http_url"`
	AzureDevopsApiUrl       types.String `tfsdk:"azure_devops_api_url"`
	BitbucketCreds          types.String `tfsdk:"bitbucket_creds"`
	AwsAccessKeyId          types.String `tfsdk:"aws_access_key_id"`
	AwsSecretAccessKey      types.String `tfsdk:"aws_secret_access_key"`
	AwsDefaultRegion        types.String `tfsdk:"aws_default_region"`
	ArmTenantId             types.String `tfsdk:"arm_tenant_id"`
	ArmSubscriptionId       types.String `tfsdk:"arm_subscription_id"`
	ArmClientId             types.String `tfsdk:"arm_client_id"`
	ArmClientSecret         types.String `tfsdk:"arm_client_secret"`
	GcpConfigFileContent    types.String `tfsdk:"gcp_config_file_content"`
}

func (m ConnectorSettingsConfigModel) toAPIModel() *sgsdkgo.SettingsConfig {
	return &sgsdkgo.SettingsConfig{
		RoleArn:                 m.RoleArn.ValueStringPointer(),
		ExternalId:              m.ExternalId.ValueStringPointer(),
		DurationSeconds:         m.DurationSeconds.ValueStringPointer(),
		InstallationId:          m.InstallationId.ValueStringPointer(),
		GithubAppId:             m.GithubAppId.ValueStringPointer(),
		GithubAppWebhookSecret:  m.GithubAppWebhookSecret.ValueStringPointer(),
		GithubApiUrl:            m.GithubApiUrl.ValueStringPointer(),
		GithubHttpUrl:           m.GithubHttpUrl.ValueStringPointer(),
		GithubAppClientId:       m.GithubAppClientId.ValueStringPointer(),
		GithubAppClientSecret:   m.GithubAppClientSecret.ValueStringPointer(),
		GithubAppPemFileContent: m.GithubAppPemFileContent.ValueStringPointer(),
		GithubAppWebhookUrl:     m.GithubAppWebhookURL.ValueStringPointer(),
		GitlabCreds:             m.GitlabCreds.ValueStringPointer(),
		GitlabHttpUrl:           m.GitlabHttpUrl.ValueStringPointer(),
		GitlabApiUrl:            m.GitlabApiUrl.ValueStringPointer(),
		AzureCreds:              m.AzureCreds.ValueStringPointer(),
		AzureDevopsHttpUrl:      m.AzureDevopsHttpUrl.ValueStringPointer(),
		AzureDevopsApiUrl:       m.AzureDevopsApiUrl.ValueStringPointer(),
		BitbucketCreds:          m.BitbucketCreds.ValueStringPointer(),
		AwsAccessKeyId:          m.AwsAccessKeyId.ValueStringPointer(),
		AwsSecretAccessKey:      m.AwsSecretAccessKey.ValueStringPointer(),
		AwsDefaultRegion:        m.AwsDefaultRegion.ValueStringPointer(),
		ArmTenantId:             m.ArmTenantId.ValueStringPointer(),
		ArmSubscriptionId:       m.ArmSubscriptionId.ValueStringPointer(),
		ArmClientId:             m.ArmClientId.ValueStringPointer(),
		ArmClientSecret:         m.ArmClientSecret.ValueStringPointer(),
		GcpConfigFileContent:    m.GcpConfigFileContent.ValueStringPointer(),
	}
}

func (m ConnectorSettingsConfigModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"role_arn":                    types.StringType,
		"external_id":                 types.StringType,
		"duration_seconds":            types.StringType,
		"installation_id":             types.StringType,
		"github_app_id":               types.StringType,
		"github_app_webhook_secret":   types.StringType,
		"github_api_url":              types.StringType,
		"github_http_url":             types.StringType,
		"github_app_client_id":        types.StringType,
		"github_app_client_secret":    types.StringType,
		"github_app_pem_file_content": types.StringType,
		"github_app_webhook_url":      types.StringType,
		"gitlab_creds":                types.StringType,
		"gitlab_http_url":             types.StringType,
		"gitlab_api_url":              types.StringType,
		"azure_creds":                 types.StringType,
		"azure_devops_http_url":       types.StringType,
		"azure_devops_api_url":        types.StringType,
		"bitbucket_creds":             types.StringType,
		"aws_access_key_id":           types.StringType,
		"aws_secret_access_key":       types.StringType,
		"aws_default_region":          types.StringType,
		"arm_tenant_id":               types.StringType,
		"arm_subscription_id":         types.StringType,
		"arm_client_id":               types.StringType,
		"arm_client_secret":           types.StringType,
		"gcp_config_file_content":     types.StringType,
	}
}

type ConnectorDiscoverySettingsModel struct {
	Benchmarks types.Map `tfsdk:"benchmarks"`
}

func (ConnectorDiscoverySettingsModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"benchmarks": types.MapType{ElemType: ConnectorDiscoverySettingsBenchmarksModel{}.AttributeTypes()},
	}
}

type ConnectorDiscoverySettingsRegionModel struct {
	Region types.String `tfsdk:"region"`
}

func (ConnectorDiscoverySettingsRegionModel) AttributeTypes() attr.Type {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"region": types.StringType,
	}}
}

type ConnectorDiscoverySettingsBenchmarksModel struct {
	Description        types.String `tfsdk:"description"`
	Label              types.String `tfsdk:"label"`
	RuntimeSource      types.Object `tfsdk:"runtime_source"`
	SummaryDescription types.String `tfsdk:"summary_description"`
	SummaryTitle       types.String `tfsdk:"summary_title"`
	DiscoveryInterval  types.Int64  `tfsdk:"discovery_interval"`
	IsCustomCheck      types.Bool   `tfsdk:"is_custom_check"`
	Active             types.Bool   `tfsdk:"active"`
	Checks             types.List   `tfsdk:"checks"`
	Regions            types.Map    `tfsdk:"regions"`
}

func (m *ConnectorDiscoverySettingsBenchmarksModel) toAPIModel() (*sgsdkgo.DiscoveryBenchmark, diag.Diagnostics) {
	benchmarkAPIModel := &sgsdkgo.DiscoveryBenchmark{
		Description:   m.Description.ValueStringPointer(),
		SummaryDesc:   m.SummaryDescription.ValueStringPointer(),
		SummaryTitle:  m.SummaryTitle.ValueString(),
		Label:         m.Label.ValueString(),
		Active:        m.Active.ValueBoolPointer(),
		IsCustomCheck: m.IsCustomCheck.ValueBoolPointer(),
	}

	// checks
	var benchmarkChecksModel []types.String
	diags := m.Checks.ElementsAs(context.Background(), &benchmarkChecksModel, false)
	if diags.HasError() {
		return nil, diags
	}

	var benchmarkChecks []string
	for _, check := range benchmarkChecksModel {
		benchmarkChecks = append(benchmarkChecks, check.ValueString())
	}

	benchmarkAPIModel.Checks = benchmarkChecks

	// runtime resource
	if !m.RuntimeSource.IsNull() {
		var runtimeSourceModel ConnectorDiscoverySettingsBenchmarksRuntimeSourceModel
		diags = m.RuntimeSource.As(context.TODO(), &runtimeSourceModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
		if diags.HasError() {
			return nil, diags
		}

		destKind, err := sgsdkgo.NewCustomSourceSourceConfigDestKindEnumFromString(runtimeSourceModel.SourceConfigDestKind.ValueString())
		if err != nil {
			return nil, []diag.Diagnostic{diag.NewErrorDiagnostic("Error in converting sourceConfigDestKind", err.Error())}
		}
		benchmarkRuntimeResource := sgsdkgo.CustomSource{
			SourceConfigDestKind: destKind,
		}

		if runtimeSourceModel.Config.IsUnknown() {
			var customSourceConfigModel ConnectorDiscoverySettingsBenchmarksRuntimeSourceCustomSourceConfigModel
			diags = runtimeSourceModel.Config.As(context.TODO(), &customSourceConfigModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
			if diags.HasError() {
				return nil, diags
			}

			configAPIModel := &sgsdkgo.CustomSourceConfig{
				Auth:             customSourceConfigModel.Auth.ValueStringPointer(),
				IncludeSubModule: customSourceConfigModel.IncludeSubModule.ValueBoolPointer(),
				Ref:              customSourceConfigModel.Ref.ValueStringPointer(),
				GitCoreAutoCrlf:  customSourceConfigModel.GitCoreAutoCRLF.ValueBoolPointer(),
				WorkingDir:       customSourceConfigModel.WorkingDir.ValueStringPointer(),
				Repo:             customSourceConfigModel.Repo.ValueStringPointer(),
				IsPrivate:        customSourceConfigModel.IsPrivate.ValueBoolPointer(),
			}

			benchmarkRuntimeResource.Config = configAPIModel
		}
		benchmarkAPIModel.RuntimeSource = &benchmarkRuntimeResource
	}

	// regions
	if !m.Regions.IsUnknown() {
		var benchmarkRegionsModel map[string]*ConnectorDiscoverySettingsBenchmarksRegionsModel
		diags = m.Regions.ElementsAs(context.Background(), &benchmarkRegionsModel, false)
		if diags.HasError() {
			return nil, diags
		}

		benchmarkRegions := map[string]*sgsdkgo.DiscoveryRegion{}
		for region, regionValue := range benchmarkRegionsModel {
			benchmarkRegions[region] = &sgsdkgo.DiscoveryRegion{}
			if !regionValue.Emails.IsUnknown() {
				var emailsModel []types.String
				var emailsAPIModel []string
				diags = regionValue.Emails.ElementsAs(context.Background(), &emailsModel, false)
				if diags.HasError() {
					return nil, diags
				}
				for _, email := range emailsModel {
					if email.ValueString() != "" {
						emailsAPIModel = append(emailsAPIModel, email.ValueString())
					}
				}
				benchmarkRegions[region].Emails = emailsAPIModel
			}
		}
		benchmarkAPIModel.Regions = benchmarkRegions
	}

	if !m.DiscoveryInterval.IsUnknown() {
		intValue := int(m.DiscoveryInterval.ValueInt64())
		benchmarkAPIModel.DiscoveryInterval = &intValue
	}

	return benchmarkAPIModel, nil
}

func (ConnectorDiscoverySettingsBenchmarksModel) AttributeTypes() attr.Type {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"description":         types.StringType,
			"label":               types.StringType,
			"runtime_source":      types.ObjectType{AttrTypes: ConnectorDiscoverySettingsBenchmarksRuntimeSourceModel{}.AttributeTypes()},
			"summary_description": types.StringType,
			"summary_title":       types.StringType,
			"discovery_interval":  types.Int64Type,
			"is_custom_check":     types.BoolType,
			"active":              types.BoolType,
			"checks":              types.ListType{ElemType: types.StringType},
			"regions":             types.MapType{ElemType: types.ObjectType{AttrTypes: ConnectorDiscoverySettingsBenchmarksRegionsModel{}.AttributeTypes()}},
		},
	}
}

type ConnectorDiscoverySettingsBenchmarksRuntimeSourceModel struct {
	SourceConfigDestKind types.String `tfsdk:"source_config_dest_kind"`
	Config               types.Object `tfsdk:"config"`
}

func (m ConnectorDiscoverySettingsBenchmarksRuntimeSourceModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"source_config_dest_kind": types.StringType,
		"config":                  types.ObjectType{AttrTypes: ConnectorDiscoverySettingsBenchmarksRuntimeSourceCustomSourceConfigModel{}.AttributeTypes()},
	}
}

type ConnectorDiscoverySettingsBenchmarksRuntimeSourceCustomSourceConfigModel struct {
	IncludeSubModule types.Bool   `tfsdk:"include_sub_module"`
	Ref              types.String `tfsdk:"ref"`
	GitCoreAutoCRLF  types.Bool   `tfsdk:"git_core_auto_crlf"`
	Auth             types.String `tfsdk:"auth"`
	WorkingDir       types.String `tfsdk:"working_dir"`
	Repo             types.String `tfsdk:"repo"`
	IsPrivate        types.Bool   `tfsdk:"is_private"`
}

func (m ConnectorDiscoverySettingsBenchmarksRuntimeSourceCustomSourceConfigModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"include_sub_module": types.BoolType,
		"ref":                types.StringType,
		"git_core_auto_crlf": types.BoolType,
		"auth":               types.StringType,
		"working_dir":        types.StringType,
		"repo":               types.StringType,
		"is_private":         types.BoolType,
	}
}

type ConnectorDiscoverySettingsBenchmarksRegionsModel struct {
	Emails types.List `tfsdk:"emails"`
}

func (ConnectorDiscoverySettingsBenchmarksRegionsModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"emails": types.ListType{ElemType: types.StringType},
	}
}

func settingsToAPIModel(m types.Object) (*sgsdkgo.IntegrationsSettings, diag.Diagnostics) {
	// Set kind and config in Settings
	var settingsModelValue *ConnectorSettingsModel
	diags := m.As(context.Background(), &settingsModelValue, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
	if diags.HasError() {
		return nil, diags
	}

	var settingsConfigModel []*ConnectorSettingsConfigModel
	diags = settingsModelValue.Config.ElementsAs(context.TODO(), &settingsConfigModel, false)
	if diags.HasError() {
		return nil, diags
	}

	settings := &sgsdkgo.IntegrationsSettings{
		Kind: sgsdkgo.IntegrationsSettingsKindEnum(settingsModelValue.Kind.ValueString()),
	}

	settingsConfigAPIValue := []*sgsdkgo.SettingsConfig{settingsConfigModel[0].toAPIModel()}

	settings.Config = settingsConfigAPIValue

	return settings, nil
}

func discoverSettingsToAPIModel(m types.Object) (*sgsdkgo.Discoverysettings, diag.Diagnostics) {
	// Parse discovery settings
	if m.IsUnknown() {
		return nil, nil
	}

	discoverySettingsAPIModel := &sgsdkgo.Discoverysettings{}
	var discoverySettingsModel *ConnectorDiscoverySettingsModel
	diags := m.As(context.Background(), &discoverySettingsModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
	if diags.HasError() {
		return nil, diags
	}

	// Parse benchmarks
	if !discoverySettingsModel.Benchmarks.IsUnknown() {
		var benchmarksModel map[string]*ConnectorDiscoverySettingsBenchmarksModel
		diags = discoverySettingsModel.Benchmarks.ElementsAs(context.Background(), &benchmarksModel, false)
		if diags.HasError() {
			return nil, diags
		}

		benchmarksAPIModel := map[string]*sgsdkgo.DiscoveryBenchmark{}
		for benchmarkName, benchmarkModel := range benchmarksModel {
			benchmarksAPIModel[benchmarkName], diags = benchmarkModel.toAPIModel()
			if diags.HasError() {
				return nil, diags
			}
		}
		discoverySettingsAPIModel.Benchmarks = benchmarksAPIModel
	}

	return discoverySettingsAPIModel, nil
}

func (m *ConnectorResourceModel) ToAPIModel(ctx context.Context) (*sgsdkgo.Integration, diag.Diagnostics) {
	apiModel := sgsdkgo.Integration{
		ResourceName: m.ResourceName.ValueStringPointer(),
	}

	if !m.Id.IsUnknown() {
		apiModel.Id = m.Id.ValueStringPointer()
	}

	if !m.Description.IsUnknown() {
		apiModel.Description = m.Description.ValueStringPointer()
	}

	settings, diags := settingsToAPIModel(m.Settings)
	if diags.HasError() {
		return nil, diags
	}
	if settings != nil {
		apiModel.Settings = settings
	}

	discoverySettings, diags := discoverSettingsToAPIModel(m.DiscoverySettings)
	if diags.HasError() {
		return nil, diags
	}
	if discoverySettings != nil {
		apiModel.DiscoverySettings = discoverySettings
	}

	// Parse tags
	tags, diags := expanders.StringList(context.TODO(), m.Tags)
	if diags.HasError() {
		return nil, diags
	}
	if tags != nil {
		apiModel.Tags = tags
	}

	return &apiModel, nil
}

func (m *ConnectorResourceModel) ToAPIPatchedModel(ctx context.Context) (*sgsdkgo.PatchedIntegration, diag.Diagnostics) {
	apiPatchedModel := &sgsdkgo.PatchedIntegration{
		ResourceName: sgsdkgo.Optional(m.ResourceName.ValueStringPointer()),
	}

	if !m.Description.IsUnknown() {
		apiPatchedModel.Description = sgsdkgo.Optional(m.Description.ValueStringPointer())
	} else {
		apiPatchedModel.Description = sgsdkgo.Null[*string]()
	}

	// Parse tags
	tags, diags := expanders.StringList(context.TODO(), m.Tags)
	if diags.HasError() {
		return nil, diags
	}
	if tags == nil {
		apiPatchedModel.Tags = sgsdkgo.Null[[]string]()
	} else {
		apiPatchedModel.Tags = sgsdkgo.Optional(tags)
	}

	// Parse Settings
	settings, diags := settingsToAPIModel(m.Settings)
	if diags.HasError() {
		return nil, diags
	}
	apiPatchedModel.Settings = sgsdkgo.Optional(settings)

	// Parse discovery settings
	discoverySettings, diags := discoverSettingsToAPIModel(m.DiscoverySettings)
	if diags.HasError() {
		return nil, diags
	}
	if discoverySettings != nil {
		apiPatchedModel.DiscoverySettings = sgsdkgo.Optional(discoverySettings)
	} else {
		apiPatchedModel.DiscoverySettings = sgsdkgo.Null[*sgsdkgo.Discoverysettings]()
	}

	return apiPatchedModel, nil
}

func BuildAPIModelToConnectorModel(apiResponse *sgsdkgo.GeneratedConnectorReadResponseMsg) (*ConnectorResourceModel, diag.Diagnostics) {
	connectorModel := &ConnectorResourceModel{
		Id:           flatteners.String(apiResponse.Id),
		ResourceName: flatteners.String(apiResponse.ResourceName),
		Description:  flatteners.String(apiResponse.Description),
	}

	settingsConfigModel := []*ConnectorSettingsConfigModel{
		{
			RoleArn:                 flatteners.StringPtr(apiResponse.Settings.Config[0].RoleArn),
			ExternalId:              flatteners.StringPtr(apiResponse.Settings.Config[0].ExternalId),
			DurationSeconds:         flatteners.StringPtr(apiResponse.Settings.Config[0].DurationSeconds),
			InstallationId:          flatteners.StringPtr(apiResponse.Settings.Config[0].InstallationId),
			GithubAppId:             flatteners.StringPtr(apiResponse.Settings.Config[0].GithubAppId),
			GithubAppWebhookSecret:  flatteners.StringPtr(apiResponse.Settings.Config[0].GithubAppWebhookSecret),
			GithubApiUrl:            flatteners.StringPtr(apiResponse.Settings.Config[0].GithubApiUrl),
			GithubHttpUrl:           flatteners.StringPtr(apiResponse.Settings.Config[0].GithubHttpUrl),
			GithubAppClientId:       flatteners.StringPtr(apiResponse.Settings.Config[0].GithubAppClientId),
			GithubAppClientSecret:   flatteners.StringPtr(apiResponse.Settings.Config[0].GithubAppClientSecret),
			GithubAppPemFileContent: flatteners.StringPtr(apiResponse.Settings.Config[0].GithubAppPemFileContent),
			GithubAppWebhookURL:     flatteners.StringPtr(apiResponse.Settings.Config[0].GithubAppWebhookUrl),
			GitlabCreds:             flatteners.StringPtr(apiResponse.Settings.Config[0].GitlabCreds),
			GitlabHttpUrl:           flatteners.StringPtr(apiResponse.Settings.Config[0].GitlabHttpUrl),
			GitlabApiUrl:            flatteners.StringPtr(apiResponse.Settings.Config[0].GitlabApiUrl),
			AzureCreds:              flatteners.StringPtr(apiResponse.Settings.Config[0].AzureCreds),
			AzureDevopsHttpUrl:      flatteners.StringPtr(apiResponse.Settings.Config[0].AzureDevopsHttpUrl),
			AzureDevopsApiUrl:       flatteners.StringPtr(apiResponse.Settings.Config[0].AzureDevopsApiUrl),
			BitbucketCreds:          flatteners.StringPtr(apiResponse.Settings.Config[0].BitbucketCreds),
			AwsAccessKeyId:          flatteners.StringPtr(apiResponse.Settings.Config[0].AwsAccessKeyId),
			AwsSecretAccessKey:      flatteners.StringPtr(apiResponse.Settings.Config[0].AwsSecretAccessKey),
			AwsDefaultRegion:        flatteners.StringPtr(apiResponse.Settings.Config[0].AwsDefaultRegion),
			ArmTenantId:             flatteners.StringPtr(apiResponse.Settings.Config[0].ArmTenantId),
			ArmSubscriptionId:       flatteners.StringPtr(apiResponse.Settings.Config[0].ArmSubscriptionId),
			ArmClientId:             flatteners.StringPtr(apiResponse.Settings.Config[0].ArmClientId),
			ArmClientSecret:         flatteners.StringPtr(apiResponse.Settings.Config[0].ArmClientSecret),
			GcpConfigFileContent:    flatteners.StringPtr(apiResponse.Settings.Config[0].GcpConfigFileContent),
		},
	}

	settingsConfigTerraType, diags := types.ListValueFrom(context.TODO(), types.ObjectType{AttrTypes: ConnectorSettingsConfigModel{}.AttributeTypes()}, &settingsConfigModel)
	if diags.HasError() {
		return nil, diags
	}
	connectorSettingsModel := ConnectorSettingsModel{
		Kind:   flatteners.String(apiResponse.Settings.Kind),
		Config: settingsConfigTerraType,
	}
	settings, diags := types.ObjectValueFrom(context.Background(), connectorSettingsModel.AttributeTypes(), connectorSettingsModel)
	if diags.HasError() {
		return nil, diags
	}
	connectorModel.Settings = settings

	// Discovery Settings
	if apiResponse.DiscoverySettings == nil {
		connectorModel.DiscoverySettings = types.ObjectNull(ConnectorDiscoverySettingsModel{}.AttributeTypes())
	} else {
		DiscoverySettingsModel := &ConnectorDiscoverySettingsModel{}

		// benchmarks
		if apiResponse.DiscoverySettings.Benchmarks == nil {
			DiscoverySettingsModel.Benchmarks = types.MapNull(ConnectorDiscoverySettingsBenchmarksModel{}.AttributeTypes())
		} else {
			// if benchmarks is not nil
			benchmarks := make(map[string]*ConnectorDiscoverySettingsBenchmarksModel, len(apiResponse.DiscoverySettings.Benchmarks))
			for benchmarkKey, benchmark := range apiResponse.DiscoverySettings.Benchmarks {
				benchmarksModel := &ConnectorDiscoverySettingsBenchmarksModel{}
				benchmarksModel.Description = flatteners.StringPtr(benchmark.Description)
				benchmarksModel.Label = flatteners.StringPtr(benchmark.Label)
				benchmarksModel.SummaryDescription = flatteners.StringPtr(benchmark.SummaryDesc)
				benchmarksModel.SummaryTitle = flatteners.StringPtr(benchmark.SummaryTitle)
				benchmarksModel.DiscoveryInterval = flatteners.Int64(int64(*benchmark.DiscoveryInterval))
				benchmarksModel.IsCustomCheck = types.BoolPointerValue(benchmark.IsCustomCheck)
				benchmarksModel.Active = types.BoolValue(benchmark.Active)

				if benchmark.RuntimeSource != nil {
					runtimeSourceModelValue := ConnectorDiscoverySettingsBenchmarksRuntimeSourceModel{
						SourceConfigDestKind: flatteners.String(string(benchmark.RuntimeSource.SourceConfigDestKind)),
					}
					if benchmark.RuntimeSource.Config != nil {
						configModel := ConnectorDiscoverySettingsBenchmarksRuntimeSourceCustomSourceConfigModel{
							IncludeSubModule: flatteners.BoolPtr(benchmark.RuntimeSource.Config.IncludeSubModule),
							Ref:              flatteners.StringPtr(benchmark.RuntimeSource.Config.Ref),
							GitCoreAutoCRLF:  flatteners.BoolPtr(benchmark.RuntimeSource.Config.GitCoreAutoCrlf),
							Auth:             flatteners.StringPtr(benchmark.RuntimeSource.Config.Auth),
							WorkingDir:       flatteners.StringPtr(benchmark.RuntimeSource.Config.WorkingDir),
							Repo:             flatteners.StringPtr(benchmark.RuntimeSource.Config.Repo),
							IsPrivate:        flatteners.BoolPtr(benchmark.RuntimeSource.Config.IsPrivate),
						}
						runtimeSourceModelValue.Config, diags = types.ObjectValueFrom(context.TODO(), configModel.AttributeTypes(), &configModel)
						if diags.HasError() {
							return nil, diags
						}
					} else {
						runtimeSourceModelValue.Config = types.ObjectNull(ConnectorDiscoverySettingsBenchmarksRuntimeSourceCustomSourceConfigModel{}.AttributeTypes())
					}
					benchmarksModel.RuntimeSource, diags = types.ObjectValueFrom(context.TODO(), runtimeSourceModelValue.AttributeTypes(), &runtimeSourceModelValue)
					if diags.HasError() {
						return nil, diags
					}
				} else {
					benchmarksModel.RuntimeSource = types.ObjectNull(ConnectorDiscoverySettingsBenchmarksRuntimeSourceModel{}.AttributeTypes())
				}

				// regions
				if benchmark.Regions == nil {
					benchmarksModel.Regions = types.MapNull(types.ObjectType{AttrTypes: ConnectorDiscoverySettingsBenchmarksRegionsModel{}.AttributeTypes()})
				} else {
					regions := map[string]types.Object{}
					for regionsKey, regionsValue := range benchmark.Regions {
						emailsModel := []types.String{}
						for _, email := range regionsValue.Emails {
							emailsModel = append(emailsModel, flatteners.String(email))
						}
						emailTerraType, diags := types.ListValueFrom(context.Background(), types.StringType, &emailsModel)
						if diags.HasError() {
							return nil, diags
						}
						regionsModel := &ConnectorDiscoverySettingsBenchmarksRegionsModel{
							Emails: emailTerraType,
						}
						regionsTerraObject, diags := types.ObjectValueFrom(context.Background(), regionsModel.AttributeTypes(), &regionsModel)
						if diags.HasError() {
							return nil, diags
						}
						regions[regionsKey] = regionsTerraObject
					}
					regionsTerraType, diags := types.MapValueFrom(context.Background(), types.ObjectType{AttrTypes: ConnectorDiscoverySettingsBenchmarksRegionsModel{}.AttributeTypes()}, &regions)
					if diags.HasError() {
						return nil, diags
					}
					benchmarksModel.Regions = regionsTerraType
				}
				// checks
				checksModel := []types.String{}
				for _, check := range benchmark.Checks {
					checksModel = append(checksModel, types.StringValue(check))
				}
				checkTerraType, diags := types.ListValueFrom(context.TODO(), types.StringType, &checksModel)
				if diags.HasError() {
					return nil, diags
				}
				benchmarksModel.Checks = checkTerraType

				benchmarks[benchmarkKey] = benchmarksModel
			}
			benchmarksTerraType, diags := types.MapValueFrom(context.TODO(), ConnectorDiscoverySettingsBenchmarksModel{}.AttributeTypes(), &benchmarks)
			if diags.HasError() {
				return nil, diags
			}
			DiscoverySettingsModel.Benchmarks = benchmarksTerraType
		}

		connectorModel.DiscoverySettings, diags = types.ObjectValueFrom(context.TODO(), ConnectorDiscoverySettingsModel{}.AttributeTypes(), DiscoverySettingsModel)
		if diags.HasError() {
			return nil, diags
		}
	}

	if apiResponse.Tags == nil {
		connectorModel.Tags = types.ListNull(types.StringType)
	} else {
		tagModel := []types.String{}
		for _, tag := range apiResponse.Tags {
			tagModel = append(tagModel, flatteners.String(tag))
		}
		tagTerraType, diags := types.ListValueFrom(context.TODO(), types.StringType, &tagModel)
		if diags.HasError() {
			return nil, diags
		}
		connectorModel.Tags = tagTerraType
	}

	return connectorModel, nil
}

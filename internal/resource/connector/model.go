package connector

import (
	"context"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	flatteners "github.com/StackGuardian/terraform-provider-stackguardian/internal/flattners"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type ConnectorResourceModel struct {
	ResourceName      types.String `tfsdk:"resource_name"`
	Description       types.String `tfsdk:"description"`
	Settings          types.Object `tfsdk:"settings"`
	DiscoverySettings types.Object `tfsdk:"discovery_settings"`
	IsActive          types.String `tfsdk:"is_active"`
	Scope             types.List   `tfsdk:"scope"`
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

func (m ConnectorSettingsConfigModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
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
	DiscoveryInterval types.Int64 `tfsdk:"discovery_interval"`

	// Convert to []Region
	Regions types.List `tfsdk:"regions"`

	// Convert to map[string]interface{}
	Benchmarks types.Map `tfsdk:"benchmarks"`
}

func (ConnectorDiscoverySettingsModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"discovery_interval": types.Int64Type,
		"regions":            types.ListType{ElemType: ConnectorDiscoverySettingsRegionModel{}.AttributeTypes()},
		"benchmarks":         types.MapType{ElemType: ConnectorDiscoverySettingsBenchmarksModel{}.AttributeTypes()},
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
	LastDiscoveryTime  types.Int64  `tfsdk:"last_discovery_time"`
	IsCustomCheck      types.Bool   `tfsdk:"is_custom_check"`
	Active             types.Bool   `tfsdk:"active"`
	Checks             types.List   `tfsdk:"checks"`
	Regions            types.Map    `tfsdk:"regions"`
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
			"last_discovery_time": types.Int64Type,
			"is_custom_check":     types.BoolType,
			"active":              types.BoolType,
			"checks":              types.ListType{ElemType: types.StringType},
			"regions":             types.MapType{ElemType: types.ObjectType{AttrTypes: ConnectorDiscoverySettingsBenchmarksRegionsModel{}.AttributeTypes()}},
		},
	}
}

type ConnectorDiscoverySettingsBenchmarksRuntimeSourceModel struct {
	CustomSource types.Object `tfsdk:"custom_source"`
}

func (m ConnectorDiscoverySettingsBenchmarksRuntimeSourceModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"custom_source": types.ObjectType{AttrTypes: ConnectorDiscoverySettingsBenchmarksRuntimeSourceCustomSourceModel{}.AttributeTypes()},
	}
}

type ConnectorDiscoverySettingsBenchmarksRuntimeSourceCustomSourceModel struct {
	SourceConfigDestKind types.String `tfsdk:"source_config_dest_kind"`
	Config               types.Object `tfsdk:"config"`
}

func (m ConnectorDiscoverySettingsBenchmarksRuntimeSourceCustomSourceModel) AttributeTypes() map[string]attr.Type {
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

func (m *ConnectorResourceModel) ToAPIModel(ctx context.Context) (*sgsdkgo.Integration, diag.Diagnostics) {
	apiModel := sgsdkgo.Integration{
		ResourceName: m.ResourceName.ValueStringPointer(),
		Description:  m.Description.ValueStringPointer(),
	}

	// is active
	if !m.IsActive.IsNull() && !m.IsActive.IsUnknown() {
		apiModel.IsActive = (*sgsdkgo.IsArchiveEnum)(m.IsActive.ValueStringPointer())
	}

	// Set kind and config in Settings
	var settingsModelValue *ConnectorSettingsModel
	diags := m.Settings.As(context.Background(), &settingsModelValue, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
	if diags.HasError() {
		return nil, diags
	}

	var settingsConfigModel []*ConnectorSettingsConfigModel
	diags = settingsModelValue.Config.ElementsAs(context.TODO(), &settingsConfigModel, false)
	if diags.HasError() {
		return nil, diags
	}

	settingsConfigAPIValue := []*sgsdkgo.SettingsConfig{{
		InstallationId:          settingsConfigModel[0].InstallationId.ValueStringPointer(),
		GithubAppId:             settingsConfigModel[0].GithubAppId.ValueStringPointer(),
		GithubAppWebhookSecret:  settingsConfigModel[0].GithubAppWebhookSecret.ValueStringPointer(),
		GithubApiUrl:            settingsConfigModel[0].GithubApiUrl.ValueStringPointer(),
		GithubHttpUrl:           settingsConfigModel[0].GithubHttpUrl.ValueStringPointer(),
		GithubAppClientId:       settingsConfigModel[0].GithubAppClientId.ValueStringPointer(),
		GithubAppClientSecret:   settingsConfigModel[0].GithubAppClientSecret.ValueStringPointer(),
		GithubAppPemFileContent: settingsConfigModel[0].GithubAppPemFileContent.ValueStringPointer(),
		GithubAppWebhookUrl:     settingsConfigModel[0].GithubAppWebhookURL.ValueStringPointer(),
		GitlabCreds:             settingsConfigModel[0].GitlabCreds.ValueStringPointer(),
		GitlabHttpUrl:           settingsConfigModel[0].GitlabHttpUrl.ValueStringPointer(),
		GitlabApiUrl:            settingsConfigModel[0].GitlabApiUrl.ValueStringPointer(),
		AzureCreds:              settingsConfigModel[0].AzureCreds.ValueStringPointer(),
		AzureDevopsHttpUrl:      settingsConfigModel[0].AzureDevopsHttpUrl.ValueStringPointer(),
		AzureDevopsApiUrl:       settingsConfigModel[0].AzureDevopsApiUrl.ValueStringPointer(),
		BitbucketCreds:          settingsConfigModel[0].BitbucketCreds.ValueStringPointer(),
		AwsAccessKeyId:          settingsConfigModel[0].AwsAccessKeyId.ValueStringPointer(),
		AwsSecretAccessKey:      settingsConfigModel[0].AwsSecretAccessKey.ValueStringPointer(),
		AwsDefaultRegion:        settingsConfigModel[0].AwsDefaultRegion.ValueStringPointer(),
		ArmTenantId:             settingsConfigModel[0].ArmTenantId.ValueStringPointer(),
		ArmSubscriptionId:       settingsConfigModel[0].ArmSubscriptionId.ValueStringPointer(),
		ArmClientId:             settingsConfigModel[0].ArmClientId.ValueStringPointer(),
		ArmClientSecret:         settingsConfigModel[0].ArmClientSecret.ValueStringPointer(),
		GcpConfigFileContent:    settingsConfigModel[0].GcpConfigFileContent.ValueStringPointer(),
	}}

	settings := &sgsdkgo.Settings{
		Kind:   sgsdkgo.SettingsKindEnum(settingsModelValue.Kind.ValueString()),
		Config: settingsConfigAPIValue,
	}
	apiModel.Settings = settings

	// Parse discovery settings
	discoverySettingsAPIModel := &sgsdkgo.Discoverysettings{}
	var discoverySettingsModel *ConnectorDiscoverySettingsModel
	if !m.DiscoverySettings.IsNull() && !m.DiscoverySettings.IsUnknown() {
		diags := m.DiscoverySettings.As(context.Background(), &discoverySettingsModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
		if diags.HasError() {
			return nil, diags
		}

		// Parse discovery interval
		discoveryInterval := int(*discoverySettingsModel.DiscoveryInterval.ValueInt64Pointer())
		discoverySettingsAPIModel.DiscoveryInterval = &discoveryInterval

		// Parse regions
		var regionsModel []*ConnectorDiscoverySettingsRegionModel
		if !discoverySettingsModel.Regions.IsNull() {
			diags = discoverySettingsModel.Regions.ElementsAs(context.Background(), &regionsModel, false)
			if diags.HasError() {
				return nil, diags
			}
		}
		regions := []*sgsdkgo.DiscoverySettingsRegions{}
		for _, region := range regionsModel {
			regions = append(regions, &sgsdkgo.DiscoverySettingsRegions{Region: region.Region.ValueString()})
		}
		discoverySettingsAPIModel.Regions = regions

		// Parse benchmarks
		var benchmarksModel map[string]*ConnectorDiscoverySettingsBenchmarksModel
		diags = discoverySettingsModel.Benchmarks.ElementsAs(context.Background(), &benchmarksModel, false)
		if diags.HasError() {
			return nil, diags
		}

		benchmarksAPIModel := map[string]*sgsdkgo.DiscoveryBenchmark{}
		for benchmarkName, benchmark := range benchmarksModel {
			var benchmarkChecksModel []types.String
			diags = benchmark.Checks.ElementsAs(context.Background(), &benchmarkChecksModel, false)
			if diags.HasError() {
				return nil, diags
			}
			var benchmarkChecks []string
			for _, check := range benchmarkChecksModel {
				benchmarkChecks = append(benchmarkChecks, check.ValueString())
			}

			var benchmarkRegionsModel map[string]*ConnectorDiscoverySettingsBenchmarksRegionsModel
			diags = benchmark.Regions.ElementsAs(context.Background(), &benchmarkRegionsModel, false)
			if diags.HasError() {
				return nil, diags
			}

			benchmarkRegions := map[string]*sgsdkgo.DiscoveryRegion{}
			for region, regionValue := range benchmarkRegionsModel {
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

				benchmarkRegions[region] = &sgsdkgo.DiscoveryRegion{
					Emails: emailsAPIModel,
				}

			}

			// runtime resource
			benchmarkRuntimeResource := sgsdkgo.DiscoveryBenchmarkRuntimeSource{}
			if !benchmark.RuntimeSource.IsNull() && !benchmark.RuntimeSource.IsUnknown() {
				var customSourceModel ConnectorDiscoverySettingsBenchmarksRuntimeSourceCustomSourceModel
				customSourceAPIModel := &sgsdkgo.CustomSource{}

				diags = benchmark.RuntimeSource.As(context.TODO(), &customSourceModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
				if diags.HasError() {
					return nil, diags
				}

				customSourceAPIModel.SourceConfigDestKind = customSourceModel.SourceConfigDestKind.ValueStringPointer()

				if !customSourceModel.Config.IsNull() && customSourceModel.Config.IsUnknown() {
					var customSourceConfigModel ConnectorDiscoverySettingsBenchmarksRuntimeSourceCustomSourceConfigModel
					diags = customSourceModel.Config.As(context.TODO(), &customSourceConfigModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
					if diags.HasError() {
						return nil, diags
					}

					customSourceConfigAPIModel := &sgsdkgo.CustomSourceConfig{
						Auth:             customSourceConfigModel.Auth.ValueStringPointer(),
						IncludeSubModule: customSourceConfigModel.IncludeSubModule.ValueBoolPointer(),
						Ref:              customSourceConfigModel.Ref.ValueStringPointer(),
						GitCoreAutoCrlf:  customSourceConfigModel.GitCoreAutoCRLF.ValueBoolPointer(),
						WorkingDir:       customSourceConfigModel.WorkingDir.ValueStringPointer(),
						Repo:             customSourceConfigModel.Repo.ValueStringPointer(),
						IsPrivate:        customSourceConfigModel.IsPrivate.ValueBoolPointer(),
					}

					customSourceAPIModel.Config = customSourceConfigAPIModel
				}

				benchmarkRuntimeResource.CustomSource = customSourceAPIModel
			}

			benchmarkAPIModel := &sgsdkgo.DiscoveryBenchmark{
				Description:   benchmark.Description.ValueStringPointer(),
				SummaryDesc:   benchmark.SummaryDescription.ValueStringPointer(),
				SummaryTitle:  benchmark.SummaryTitle.ValueStringPointer(),
				Label:         benchmark.Label.ValueStringPointer(),
				Active:        benchmark.Active.ValueBoolPointer(),
				IsCustomCheck: benchmark.IsCustomCheck.ValueBoolPointer(),
				Checks:        benchmarkChecks,
				Regions:       benchmarkRegions,
				RuntimeSource: &benchmarkRuntimeResource,
			}

			if !benchmark.LastDiscoveryTime.IsNull() {
				intValue := int(benchmark.LastDiscoveryTime.ValueInt64())
				benchmarkAPIModel.LastDiscoveryTime = &intValue
			}

			if !benchmark.DiscoveryInterval.IsNull() {
				intValue := int(benchmark.DiscoveryInterval.ValueInt64())
				benchmarkAPIModel.DiscoveryInterval = &intValue
			}

			benchmarksAPIModel[benchmarkName] = benchmarkAPIModel
		}
		discoverySettingsAPIModel.Benchmarks = benchmarksAPIModel

		apiModel.DiscoverySettings = discoverySettingsAPIModel
	}

	// Parse Scope
	if !m.Scope.IsNull() && !m.Scope.IsUnknown() {
		var scopeModel []types.String
		diags = m.Scope.ElementsAs(context.TODO(), &scopeModel, false)
		if diags.HasError() {
			return nil, diags
		}

		var scopeAPIModel []string
		for _, scope := range scopeModel {
			scopeAPIModel = append(scopeAPIModel, scope.ValueString())
		}
		apiModel.Scope = scopeAPIModel
	}

	// Parse tags
	if !m.Tags.IsNull() {
		var tagsModel []types.String
		diags = m.Tags.ElementsAs(context.TODO(), &tagsModel, false)
		if diags.HasError() {
			return nil, diags
		}

		var tagsAPIModel []string
		for _, scope := range tagsModel {
			tagsAPIModel = append(tagsAPIModel, scope.ValueString())
		}
		apiModel.Tags = tagsAPIModel
	}

	return &apiModel, nil
}

func (m *ConnectorResourceModel) ToAPIPatchedModel(ctx context.Context) (*sgsdkgo.PatchedIntegration, diag.Diagnostics) {
	apiPatchedModel := &sgsdkgo.PatchedIntegration{
		ResourceName: m.ResourceName.ValueStringPointer(),
		Description:  m.Description.ValueStringPointer(),
	}

	// Parse Scope
	if !m.Scope.IsNull() && !m.Scope.IsUnknown() {
		var scopeModel []types.String
		diags := m.Scope.ElementsAs(context.TODO(), &scopeModel, false)
		if diags.HasError() {
			return nil, diags
		}

		var scopeAPIModel []string
		for _, scope := range scopeModel {
			scopeAPIModel = append(scopeAPIModel, scope.ValueString())
		}
		apiPatchedModel.Scope = scopeAPIModel
	}

	// Parse tags
	if !m.Tags.IsNull() {
		var tagsModel []types.String
		diags := m.Tags.ElementsAs(context.TODO(), &tagsModel, false)
		if diags.HasError() {
			return nil, diags
		}

		var tagsAPIModel []string
		for _, scope := range tagsModel {
			tagsAPIModel = append(tagsAPIModel, scope.ValueString())
		}
		apiPatchedModel.Tags = tagsAPIModel
	}

	// Parse Settings
	var settingsModelValue *ConnectorSettingsModel
	diags := m.Settings.As(context.Background(), &settingsModelValue, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
	if diags.HasError() {
		return nil, diags
	}

	var settingsConfigModel []*ConnectorSettingsConfigModel
	diags = settingsModelValue.Config.ElementsAs(context.TODO(), &settingsConfigModel, false)
	if diags.HasError() {
		return nil, diags
	}
	settingsConfigAPIValue := []*sgsdkgo.SettingsConfig{{
		InstallationId:          settingsConfigModel[0].InstallationId.ValueStringPointer(),
		GithubAppId:             settingsConfigModel[0].GithubAppId.ValueStringPointer(),
		GithubAppWebhookSecret:  settingsConfigModel[0].GithubAppWebhookSecret.ValueStringPointer(),
		GithubApiUrl:            settingsConfigModel[0].GithubApiUrl.ValueStringPointer(),
		GithubHttpUrl:           settingsConfigModel[0].GithubHttpUrl.ValueStringPointer(),
		GithubAppClientId:       settingsConfigModel[0].GithubAppClientId.ValueStringPointer(),
		GithubAppClientSecret:   settingsConfigModel[0].GithubAppClientSecret.ValueStringPointer(),
		GithubAppPemFileContent: settingsConfigModel[0].GithubAppPemFileContent.ValueStringPointer(),
		GithubAppWebhookUrl:     settingsConfigModel[0].GithubAppWebhookURL.ValueStringPointer(),
		GitlabCreds:             settingsConfigModel[0].GitlabCreds.ValueStringPointer(),
		GitlabHttpUrl:           settingsConfigModel[0].GitlabHttpUrl.ValueStringPointer(),
		GitlabApiUrl:            settingsConfigModel[0].GitlabApiUrl.ValueStringPointer(),
		AzureCreds:              settingsConfigModel[0].AzureCreds.ValueStringPointer(),
		AzureDevopsHttpUrl:      settingsConfigModel[0].AzureDevopsHttpUrl.ValueStringPointer(),
		AzureDevopsApiUrl:       settingsConfigModel[0].AzureDevopsApiUrl.ValueStringPointer(),
		BitbucketCreds:          settingsConfigModel[0].BitbucketCreds.ValueStringPointer(),
		AwsAccessKeyId:          settingsConfigModel[0].AwsAccessKeyId.ValueStringPointer(),
		AwsSecretAccessKey:      settingsConfigModel[0].AwsSecretAccessKey.ValueStringPointer(),
		AwsDefaultRegion:        settingsConfigModel[0].AwsDefaultRegion.ValueStringPointer(),
		ArmTenantId:             settingsConfigModel[0].ArmTenantId.ValueStringPointer(),
		ArmSubscriptionId:       settingsConfigModel[0].ArmSubscriptionId.ValueStringPointer(),
		ArmClientId:             settingsConfigModel[0].ArmClientId.ValueStringPointer(),
		ArmClientSecret:         settingsConfigModel[0].ArmClientSecret.ValueStringPointer(),
		GcpConfigFileContent:    settingsConfigModel[0].GcpConfigFileContent.ValueStringPointer(),
	}}

	settings := &sgsdkgo.Settings{
		Kind:   sgsdkgo.SettingsKindEnum(settingsModelValue.Kind.ValueString()),
		Config: settingsConfigAPIValue,
	}
	apiPatchedModel.Settings = settings

	// Parse discovery settings
	discoverySettingsAPIModel := &sgsdkgo.Discoverysettings{}
	var discoverySettingsModel *ConnectorDiscoverySettingsModel
	if !m.DiscoverySettings.IsNull() && !m.DiscoverySettings.IsUnknown() {
		diags := m.DiscoverySettings.As(context.Background(), &discoverySettingsModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
		if diags.HasError() {
			return nil, diags
		}

		// Parse discovery interval
		discoveryInterval := int(*discoverySettingsModel.DiscoveryInterval.ValueInt64Pointer())
		discoverySettingsAPIModel.DiscoveryInterval = &discoveryInterval

		// Parse regions
		var regionsModel []*ConnectorDiscoverySettingsRegionModel
		if !discoverySettingsModel.Regions.IsNull() {
			diags = discoverySettingsModel.Regions.ElementsAs(context.Background(), &regionsModel, false)
			if diags.HasError() {
				return nil, diags
			}
		}
		regions := []*sgsdkgo.DiscoverySettingsRegions{}
		for _, region := range regionsModel {
			regions = append(regions, &sgsdkgo.DiscoverySettingsRegions{Region: region.Region.ValueString()})
		}
		discoverySettingsAPIModel.Regions = regions

		// Parse benchmarks
		var benchmarksModel map[string]*ConnectorDiscoverySettingsBenchmarksModel
		diags = discoverySettingsModel.Benchmarks.ElementsAs(context.Background(), &benchmarksModel, false)
		if diags.HasError() {
			return nil, diags
		}

		benchmarksAPIModel := map[string]*sgsdkgo.DiscoveryBenchmark{}
		for benchmarkName, benchmark := range benchmarksModel {
			var benchmarkChecksModel []types.String
			diags = benchmark.Checks.ElementsAs(context.Background(), &benchmarkChecksModel, false)
			if diags.HasError() {
				return nil, diags
			}
			var benchmarkChecks []string
			for _, check := range benchmarkChecksModel {
				benchmarkChecks = append(benchmarkChecks, check.ValueString())
			}

			var benchmarkRegionsModel map[string]*ConnectorDiscoverySettingsBenchmarksRegionsModel
			diags = benchmark.Regions.ElementsAs(context.Background(), &benchmarkRegionsModel, false)
			if diags.HasError() {
				return nil, diags
			}

			benchmarkRegions := map[string]*sgsdkgo.DiscoveryRegion{}
			for region, regionValue := range benchmarkRegionsModel {
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

				benchmarkRegions[region] = &sgsdkgo.DiscoveryRegion{
					Emails: emailsAPIModel,
				}
			}

			// runtime resource
			benchmarkRuntimeResource := sgsdkgo.DiscoveryBenchmarkRuntimeSource{}
			if !benchmark.RuntimeSource.IsNull() && !benchmark.RuntimeSource.IsUnknown() {
				var customSourceModel ConnectorDiscoverySettingsBenchmarksRuntimeSourceCustomSourceModel
				customSourceAPIModel := &sgsdkgo.CustomSource{}

				diags = benchmark.RuntimeSource.As(context.TODO(), &customSourceModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
				if diags.HasError() {
					return nil, diags
				}

				customSourceAPIModel.SourceConfigDestKind = customSourceModel.SourceConfigDestKind.ValueStringPointer()

				if !customSourceModel.Config.IsNull() && !customSourceModel.Config.IsUnknown() {
					var customSourceConfigModel ConnectorDiscoverySettingsBenchmarksRuntimeSourceCustomSourceConfigModel
					diags = customSourceModel.Config.As(context.TODO(), &customSourceConfigModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
					if diags.HasError() {
						return nil, diags
					}

					customSourceConfigAPIModel := &sgsdkgo.CustomSourceConfig{
						Auth:             customSourceConfigModel.Auth.ValueStringPointer(),
						IncludeSubModule: customSourceConfigModel.IncludeSubModule.ValueBoolPointer(),
						Ref:              customSourceConfigModel.Ref.ValueStringPointer(),
						GitCoreAutoCrlf:  customSourceConfigModel.GitCoreAutoCRLF.ValueBoolPointer(),
						WorkingDir:       customSourceConfigModel.WorkingDir.ValueStringPointer(),
						Repo:             customSourceConfigModel.Repo.ValueStringPointer(),
						IsPrivate:        customSourceConfigModel.IsPrivate.ValueBoolPointer(),
					}

					customSourceAPIModel.Config = customSourceConfigAPIModel
				}

				benchmarkRuntimeResource.CustomSource = customSourceAPIModel
			}

			benchmarksModel := &sgsdkgo.DiscoveryBenchmark{
				Description:   benchmark.Description.ValueStringPointer(),
				SummaryDesc:   benchmark.SummaryDescription.ValueStringPointer(),
				SummaryTitle:  benchmark.SummaryTitle.ValueStringPointer(),
				Label:         benchmark.Label.ValueStringPointer(),
				Active:        benchmark.Active.ValueBoolPointer(),
				IsCustomCheck: benchmark.IsCustomCheck.ValueBoolPointer(),
				Checks:        benchmarkChecks,
				Regions:       benchmarkRegions,
				RuntimeSource: &benchmarkRuntimeResource,
			}

			if !benchmark.LastDiscoveryTime.IsNull() {
				intValue := int(benchmark.LastDiscoveryTime.ValueInt64())
				benchmarksModel.LastDiscoveryTime = &intValue
			}

			if !benchmark.DiscoveryInterval.IsNull() {
				intValue := int(benchmark.DiscoveryInterval.ValueInt64())
				benchmarksModel.DiscoveryInterval = &intValue
			}

			benchmarksAPIModel[benchmarkName] = benchmarksModel
		}
		discoverySettingsAPIModel.Benchmarks = benchmarksAPIModel

		apiPatchedModel.DiscoverySettings = discoverySettingsAPIModel
	}

	// IsActive
	if !m.IsActive.IsNull() {
		apiPatchedModel.IsActive = (*sgsdkgo.IsArchiveEnum)(m.IsActive.ValueStringPointer())
	}

	return apiPatchedModel, nil
}

func buildAPIModelToConnectorModel(apiResponse *sgsdkgo.GeneratedConnectorReadResponseMsg) (*ConnectorResourceModel, diag.Diagnostics) {
	connectorModel := &ConnectorResourceModel{
		ResourceName: flatteners.String(apiResponse.ResourceName),
		Description:  flatteners.String(apiResponse.Description),
		IsActive:     flatteners.String(apiResponse.IsActive),
	}

	settingsConfigModel := []*ConnectorSettingsConfigModel{
		{
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
		// discovery interval
		DiscoverySettingsModel.DiscoveryInterval = flatteners.Int64Ptr(apiResponse.DiscoverySettings.DiscoveryInterval)

		// benchmarks
		if apiResponse.DiscoverySettings.Benchmarks == nil || len(apiResponse.DiscoverySettings.Benchmarks) == 0 {
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
				benchmarksModel.LastDiscoveryTime = flatteners.Int64(int64(*benchmark.LastDiscoveryTime))
				benchmarksModel.IsCustomCheck = types.BoolPointerValue(benchmark.IsCustomCheck)
				benchmarksModel.Active = types.BoolValue(benchmark.Active)

				// TODO: runtime resource
				if benchmark.RuntimeSource != nil {
					runtimeSourceModel := ConnectorDiscoverySettingsBenchmarksRuntimeSourceModel{}
					if benchmark.RuntimeSource.CustomSource != nil {
						customSourceModel := ConnectorDiscoverySettingsBenchmarksRuntimeSourceCustomSourceModel{
							SourceConfigDestKind: flatteners.StringPtr(benchmark.RuntimeSource.CustomSource.SourceConfigDestKind),
						}
						if benchmark.RuntimeSource.CustomSource.Config != nil {
							configModel := ConnectorDiscoverySettingsBenchmarksRuntimeSourceCustomSourceConfigModel{
								IncludeSubModule: types.BoolValue(*benchmark.RuntimeSource.CustomSource.Config.IncludeSubModule),
								Ref:              flatteners.StringPtr(benchmark.RuntimeSource.CustomSource.Config.Ref),
								GitCoreAutoCRLF:  types.BoolValue(*benchmark.RuntimeSource.CustomSource.Config.GitCoreAutoCrlf),
								Auth:             flatteners.StringPtr(benchmark.RuntimeSource.CustomSource.Config.Auth),
								WorkingDir:       flatteners.StringPtr(benchmark.RuntimeSource.CustomSource.Config.WorkingDir),
								Repo:             flatteners.StringPtr(benchmark.RuntimeSource.CustomSource.Config.Repo),
								IsPrivate:        types.BoolValue(*benchmark.RuntimeSource.CustomSource.Config.IsPrivate),
							}
							customSourceModel.Config, diags = types.ObjectValueFrom(context.TODO(), configModel.AttributeTypes(), &configModel)
							if diags.HasError() {
								return nil, diags
							}
						} else {
							customSourceModel.Config = types.ObjectNull(ConnectorDiscoverySettingsBenchmarksRuntimeSourceCustomSourceConfigModel{}.AttributeTypes())
						}
						runtimeSourceModel.CustomSource, diags = types.ObjectValueFrom(context.TODO(), customSourceModel.AttributeTypes(), &customSourceModel)
						if diags.HasError() {
							return nil, diags
						}
					} else {
						runtimeSourceModel.CustomSource = types.ObjectNull(ConnectorDiscoverySettingsBenchmarksRuntimeSourceCustomSourceModel{}.AttributeTypes())
					}
					benchmarksModel.RuntimeSource, diags = types.ObjectValueFrom(context.TODO(), runtimeSourceModel.AttributeTypes(), runtimeSourceModel)
					if diags.HasError() {
						return nil, diags
					}
				} else {
					benchmarksModel.RuntimeSource = types.ObjectNull(ConnectorDiscoverySettingsBenchmarksRuntimeSourceModel{}.AttributeTypes())
				}

				// regions
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

		// regions
		var regionModel []*ConnectorDiscoverySettingsRegionModel
		for _, regionAPIModel := range apiResponse.DiscoverySettings.Regions {
			regionModel = append(regionModel, &ConnectorDiscoverySettingsRegionModel{
				Region: flatteners.String(regionAPIModel.Region),
			})
		}
		regionTerraType, diags := types.ListValueFrom(context.TODO(), ConnectorDiscoverySettingsRegionModel{}.AttributeTypes(), &regionModel)
		if diags.HasError() {
			return nil, diags
		}
		DiscoverySettingsModel.Regions = regionTerraType

		connectorModel.DiscoverySettings, diags = types.ObjectValueFrom(context.TODO(), ConnectorDiscoverySettingsModel{}.AttributeTypes(), DiscoverySettingsModel)
		if diags.HasError() {
			return nil, diags
		}
	}

	if apiResponse.Scope == nil || len(apiResponse.Scope) == 0 {
		connectorModel.Scope = types.ListNull(types.StringType)
	} else {
		scopeModel := []types.String{}
		for _, scope := range apiResponse.Scope {
			scopeModel = append(scopeModel, flatteners.String(scope))
		}
		scopeTerraType, diags := types.ListValueFrom(context.TODO(), types.StringType, &scopeModel)
		if diags.HasError() {
			return nil, diags
		}
		connectorModel.Scope = scopeTerraType
	}

	if apiResponse.Tags == nil || len(apiResponse.Tags) == 0 {
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

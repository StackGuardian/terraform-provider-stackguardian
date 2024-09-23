package connector

import (
	"context"
	"encoding/json"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	flatteners "github.com/StackGuardian/terraform-provider-stackguardian/internal/flattners"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ConnectorResourceModel struct {
	Organization      types.String `tfsdk:"organization"`
	ResourceName      types.String `tfsdk:"resource_name"`
	Description       types.String `tfsdk:"description"`
	Settings          types.Object `tfsdk:"settings"`
	DiscoverySettings types.Object `tfsdk:"discovery_settings"`
	IsActive          types.String `tfsdk:"is_active"`
	Scope             types.List   `tfsdk:"scope"`
}

type ConnectorSettingsModel struct {
	Kind   types.String `tfsdk:"kind"`
	Config types.String `tfsdk:"config"`
}

func (ConnectorSettingsModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"kind":   types.StringType,
		"config": types.StringType,
	}
}

type ConnectorDiscoverySettingsModel struct {
	DiscoveryInterval types.Float64 `tfsdk:"discovery_interval"`

	// Convert to []Region
	Regions types.List `tfsdk:"regions"`

	// Convert to map[string]interface{}
	Benchmarks types.Map `tfsdk:"benchmarks"`
}

type ConnectorDiscoverySettingsRegionModel struct {
	Region types.String `tfsdk:"region"`
}

type ConnectorDiscoverySettingsBenchmarksModel struct {
	Description        types.String  `tfsdk:"description"`
	Label              types.String  `tfsdk:"label"`
	RuntimeSource      types.String  `tfsdk:"runtime_source"`
	SummaryDescription types.String  `tfsdk:"summary_description"`
	SummaryTitle       types.String  `tfsdk:"summary_title"`
	DiscoveryInterval  types.Float64 `tfsdk:"discovery_interval"`
	LastDiscoveryTime  types.Float64 `tfsdk:"last_discovery_time"`
	IsCustomCheck      types.Bool    `tfsdk:"is_custom_check"`
	Active             types.Bool    `tfsdk:"active"`
	Checks             types.List    `tfsdk:"checks"`
	Regions            types.Map     `tfsdk:"regions"`
}

type ConnectorDiscoverySettingsBenchmarksRegionsModel struct {
	Emails types.List `tfsdk:"emails"`
}

func (m *ConnectorResourceModel) ToAPIModel(ctx context.Context) (*sgsdkgo.Integration, diag.Diagnostics) {
	apiModel := sgsdkgo.Integration{
		ResourceName: m.ResourceName.ValueStringPointer(),
		Description:  m.Description.ValueStringPointer(),
	}

	// Set kind and config in Settings
	var settingsModelValue *ConnectorSettingsModel
	diags := m.Settings.As(context.Background(), &settingsModelValue, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
	if diags.HasError() {
		return nil, diags
	}

	settings := &sgsdkgo.Settings{
		Kind: sgsdkgo.SettingsKindEnum(settingsModelValue.Kind.ValueString()),
	}

	var settingsConfig []map[string]interface{}
	err := json.Unmarshal([]byte(settingsModelValue.Config.ValueString()), &settingsConfig)
	if err != nil {
		tflog.Debug(ctx, err.Error())
		return nil, []diag.Diagnostic{diag.NewErrorDiagnostic("Invalid attribute", "Settings.Config is invalid")}
	}
	settings.Config = settingsConfig
	apiModel.Settings = settings

	// Parse discovery settings
	discoverySettingsAPIModel := &sgsdkgo.Discoverysettings{}
	var discoverySettingsModel *ConnectorDiscoverySettingsModel
	if !m.DiscoverySettings.IsNull() {
		diags := m.DiscoverySettings.As(context.Background(), &discoverySettingsModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
		if diags.HasError() {
			return nil, diags
		}

		// Parse discovery interval
		discoverySettingsAPIModel.DiscoveryInterval = discoverySettingsModel.DiscoveryInterval.ValueFloat64Pointer()

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
			benchmarkRegions := map[string]*sgsdkgo.DiscoveryRegion{}
			diags = benchmark.Regions.ElementsAs(context.Background(), &benchmarkRegionsModel, false)
			if diags.HasError() {
				return nil, diags
			}
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

			benchmarksAPIModel[benchmarkName] = &sgsdkgo.DiscoveryBenchmark{
				RuntimeSource:     benchmark.RuntimeSource.ValueStringPointer(),
				Description:       benchmark.Description.ValueStringPointer(),
				SummaryDesc:       benchmark.SummaryDescription.ValueStringPointer(),
				SummaryTitle:      benchmark.SummaryTitle.ValueStringPointer(),
				Label:             benchmark.Label.ValueStringPointer(),
				LastDiscoveryTime: benchmark.LastDiscoveryTime.ValueFloat64Pointer(),
				DiscoveryInterval: benchmark.DiscoveryInterval.ValueFloat64Pointer(),
				Active:            benchmark.Active.ValueBoolPointer(),
				IsCustomCheck:     benchmark.IsCustomCheck.ValueBoolPointer(),
				Checks:            benchmarkChecks,
				Regions:           benchmarkRegions,
			}
		}
		discoverySettingsAPIModel.Benchmarks = benchmarksAPIModel

		apiModel.DiscoverySettings = discoverySettingsAPIModel
	}

	return &apiModel, nil
}

func buildAPIModelToConnectorModel(apiResponse *sgsdkgo.GeneratedConnectorReadResponseMsg) (*ConnectorResourceModel, diag.Diagnostics) {
	connectorModel := &ConnectorResourceModel{
		Organization: flatteners.String(apiResponse.OrgId),
		ResourceName: flatteners.String(apiResponse.ResourceName),
		Description:  flatteners.String(apiResponse.Description),
		IsActive:     flatteners.String(apiResponse.IsActive),
	}

	settingsConfig, err := json.Marshal(apiResponse.Settings.Config)
	if err != nil {
		return nil, []diag.Diagnostic{diag.NewErrorDiagnostic("Unmarshal error", "Cannot unmarhsal Connector.Settings.Config object in response from sdk")}
	}
	connectorSettingsModel := ConnectorSettingsModel{
		Kind:   flatteners.String(apiResponse.Settings.Kind),
		Config: flatteners.String(string(settingsConfig)),
	}
	var settings, diags = types.ObjectValueFrom(context.Background(), connectorSettingsModel.AttributeTypes(), connectorSettingsModel)
	if diags.HasError() {
		return nil, diags
	}
	connectorModel.Settings = settings

	// Discovery Settings

	//var regions []Region
	//if apiResponse.DiscoverySettings.Regions != nil {
	//	for _, r := range apiResponse.DiscoverySettings.Regions {
	//		regions = append(regions, Region{region: flatteners.String(r.Region)})
	//	}
	//	connectorModel.DiscoverySettings.Regions = regions
	//}

	//benchmarks, err := json.Marshal(apiResponse.DiscoverySettings.Benchmarks)
	//if err != nil {
	//	return nil, []diag.Diagnostic{diag.NewErrorDiagnostic("Unmarshal error", "Cannot unmarhsal Connector.DiscoverySettings.Benchmarks object in response from sdk")}
	//}
	//connectorModel.DiscoverySettings.Benchmarks = flatteners.String(string(benchmarks))

	//TODO: process scope
	return connectorModel, nil
}

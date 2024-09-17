package connector

import (
	"context"
	"encoding/json"
	"strings"

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
	RuntimeSource      types.String `tfsdk:"runtime_source"`
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
			"runtime_source":      types.StringType,
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

			benchmarksModel := &sgsdkgo.DiscoveryBenchmark{
				RuntimeSource: benchmark.RuntimeSource.ValueStringPointer(),
				Description:   benchmark.Description.ValueStringPointer(),
				SummaryDesc:   benchmark.SummaryDescription.ValueStringPointer(),
				SummaryTitle:  benchmark.SummaryTitle.ValueStringPointer(),
				Label:         benchmark.Label.ValueStringPointer(),
				Active:        benchmark.Active.ValueBoolPointer(),
				IsCustomCheck: benchmark.IsCustomCheck.ValueBoolPointer(),
				Checks:        benchmarkChecks,
				Regions:       benchmarkRegions,
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

		apiModel.DiscoverySettings = discoverySettingsAPIModel
	}

	// Parse Scope
	if !m.Scope.IsNull() {
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

	return &apiModel, nil
}

func buildAPIModelToConnectorModel(apiResponse *sgsdkgo.GeneratedConnectorReadResponseMsg) (*ConnectorResourceModel, diag.Diagnostics) {
	connectorModel := &ConnectorResourceModel{
		ResourceName: flatteners.String(apiResponse.ResourceName),
		Description:  flatteners.String(apiResponse.Description),
		IsActive:     flatteners.String(apiResponse.IsActive),
	}
	org := strings.Split(apiResponse.OrgId, "/")[2]
	connectorModel.Organization = flatteners.String(org)

	settings := map[string]*string{}
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
	if apiResponse.DiscoverySettings == nil {
		connectorModel.DiscoverySettings = types.ObjectNull(ConnectorDiscoverySettingsModel{}.AttributeTypes())
	} else {
		DiscoverySettingsModel := &ConnectorDiscoverySettingsModel{}
		// discovery interval
		DiscoverySettingsModel.DiscoveryInterval = flatteners.Int64(int64(*apiResponse.DiscoverySettings.DiscoveryInterval))

		// benchmarks
		if apiResponse.DiscoverySettings.Benchmarks == nil || len(apiResponse.DiscoverySettings.Benchmarks) == 0 {
			DiscoverySettingsModel.Benchmarks = types.MapNull(ConnectorDiscoverySettingsBenchmarksModel{}.AttributeTypes())
		} else {
			// if benchmarks is not nil
			benchmarks := make(map[string]*ConnectorDiscoverySettingsBenchmarksModel, len(apiResponse.DiscoverySettings.Benchmarks))
			for benchmarkKey, benchmark := range apiResponse.DiscoverySettings.Benchmarks {
				benchmarksModel := &ConnectorDiscoverySettingsBenchmarksModel{}
				benchmarksModel.Description = types.StringPointerValue(benchmark.Description)
				benchmarksModel.Label = types.StringPointerValue(benchmark.Label)
				benchmarksModel.RuntimeSource = types.StringPointerValue(benchmark.RuntimeSource)
				benchmarksModel.SummaryDescription = types.StringPointerValue(benchmark.SummaryDesc)
				benchmarksModel.SummaryTitle = types.StringPointerValue(benchmark.SummaryTitle)
				benchmarksModel.DiscoveryInterval = flatteners.Int64(int64(*benchmark.DiscoveryInterval))
				benchmarksModel.LastDiscoveryTime = flatteners.Int64(int64(*benchmark.LastDiscoveryTime))
				benchmarksModel.IsCustomCheck = types.BoolPointerValue(benchmark.IsCustomCheck)
				benchmarksModel.Active = types.BoolValue(benchmark.Active)

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

	connectorModel.Scope = types.ListNull(types.StringType)

	//TODO: process scope
	return connectorModel, nil
}

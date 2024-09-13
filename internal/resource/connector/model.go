package connector

import (
	"context"
	"encoding/json"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	flatteners "github.com/StackGuardian/terraform-provider-stackguardian/internal/flattners"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ConnectorResourceModel struct {
	Organization      types.String      `tfsdk:"organization"`
	Name              types.String      `tfsdk:"name"`
	Description       types.String      `tfsdk:"description"`
	Settings          Settings          `tfsdk:"settings"`
	DiscoverySettings DiscoverySettings `tfsdk:"discovery_settings"`
	IsActive          types.String      `tfsdk:"is_active"`
	Scope             types.List        `tfsdk:"scope"`
}

type Settings struct {
	// Convert to []map[string]interface{}
	Config types.String `tfsdk:"config"`
	Kind   types.String `tfsdk:"kind"`
}

type DiscoverySettings struct {
	DiscoveryInterval types.Float64 `tfsdk:"discovery_interval"`

	// Convert to []Region
	Regions []Region `tfsdk:"regions"`

	// Convert to map[string]interface{}
	Benchmarks types.String `tfsdk:"benchmarks"`
}
type Region struct {
	region types.String `tfsdk:"region"`
}

func (m *ConnectorResourceModel) ToAPIModel(ctx context.Context) (*sgsdkgo.Integration, diag.Diagnostics) {
	apiModel := sgsdkgo.Integration{
		ResourceName: m.Name.ValueStringPointer(),
		Description:  m.Description.ValueStringPointer(),
		Settings: &sgsdkgo.Settings{
			Kind: sgsdkgo.SettingsKindEnum(m.Settings.Config.ValueString()),
		},
		IsActive: (*sgsdkgo.IsArchiveEnum)(m.IsActive.ValueStringPointer()),
		DiscoverySettings: &sgsdkgo.Discoverysettings{
			DiscoveryInterval: m.DiscoverySettings.DiscoveryInterval.ValueFloat64(),
		},
	}

	var settingsConfig []map[string]interface{}
	err := json.Unmarshal([]byte(m.Settings.Config.ValueString()), &settingsConfig)
	if err != nil {
		tflog.Debug(ctx, err.Error())
		return nil, []diag.Diagnostic{diag.NewErrorDiagnostic("Invalid attribute", "Settings.Config is invalid")}
	}
	apiModel.Settings.Config = settingsConfig

	var regions []*sgsdkgo.DiscoverySettingsRegions
	for _, region := range m.DiscoverySettings.Regions {
		regions = append(regions, &sgsdkgo.DiscoverySettingsRegions{Region: region.region.ValueString()})
	}
	apiModel.DiscoverySettings.Regions = regions

	var benchmarks map[string]interface{}
	err = json.Unmarshal([]byte(m.DiscoverySettings.Benchmarks.ValueString()), &benchmarks)
	if err != nil {
		tflog.Debug(ctx, err.Error())
		diag.NewErrorDiagnostic("Invalid DiscoverySettings", "Error decoding json for benchmarks")
	}

	return &apiModel, nil
}

func buildAPIModelToConnectorModel(apiResponse *sgsdkgo.GeneratedConnectorReadResponseMsg) (*ConnectorResourceModel, diag.Diagnostics) {
	connectorModel := &ConnectorResourceModel{
		Organization: flatteners.String(apiResponse.OrgId),
		Name:         flatteners.String(apiResponse.ResourceName),
		Description:  flatteners.String(apiResponse.Description),
		Settings: Settings{
			Kind: flatteners.String(apiResponse.Settings.Kind),
		},
		DiscoverySettings: DiscoverySettings{
			DiscoveryInterval: flatteners.Float64(apiResponse.DiscoverySettings.DiscoveryInterval),
		},
		IsActive: flatteners.String(apiResponse.IsActive),
	}

	settingsConfig, err := json.Marshal(apiResponse.Settings.Config)
	if err != nil {
		return nil, []diag.Diagnostic{diag.NewErrorDiagnostic("Unmarshal error", "Cannot unmarhsal Connector.Settings.Config object in response from sdk")}
	}
	connectorModel.Settings.Config = flatteners.String(string(settingsConfig))

	var regions []Region
	if apiResponse.DiscoverySettings.Regions != nil {
		for _, r := range apiResponse.DiscoverySettings.Regions {
			regions = append(regions, Region{region: flatteners.String(r.Region)})
		}
	}
	connectorModel.DiscoverySettings.Regions = regions

	benchmarks, err := json.Marshal(apiResponse.DiscoverySettings.Benchmarks)
	if err != nil {
		return nil, []diag.Diagnostic{diag.NewErrorDiagnostic("Unmarshal error", "Cannot unmarhsal Connector.DiscoverySettings.Benchmarks object in response from sdk")}
	}
	connectorModel.DiscoverySettings.Benchmarks = flatteners.String(string(benchmarks))

	//TODO: process scope
	return connectorModel, nil
}

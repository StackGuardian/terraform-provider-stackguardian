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
	Organization      types.String       `tfsdk:"organization"`
	ResourceName      types.String       `tfsdk:"resource_name"`
	Description       types.String       `tfsdk:"description"`
	Settings          types.Map          `tfsdk:"settings"`
	DiscoverySettings *DiscoverySettings `tfsdk:"discovery_settings"`
	IsActive          types.String       `tfsdk:"is_active"`
	Scope             types.List         `tfsdk:"scope"`
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
		ResourceName: m.ResourceName.ValueStringPointer(),
		Description:  m.Description.ValueStringPointer(),
	}

	// Set kind and config in Settings
	apiSettigns := sgsdkgo.Settings{}

	settings := make(map[string]types.String, len(m.Settings.Elements()))
	diags := m.Settings.ElementsAs(ctx, &settings, false)
	if diags.HasError() {
		tflog.Debug(ctx, "Connector kind not found")
		return nil, diags
	}
	kindValue := settings["kind"]
	if !kindValue.IsNull() {
		kind, err := sgsdkgo.NewSettingsKindEnumFromString(kindValue.ValueString())
		if err != nil {
			return nil, []diag.Diagnostic{diag.NewErrorDiagnostic("Cannot parse api response", "Error while parsing settings kind: "+err.Error())}
		}
		apiSettigns.Kind = kind
	}

	var settingsConfig []map[string]interface{}
	err := json.Unmarshal([]byte(*settings["config"].ValueStringPointer()), &settingsConfig)
	if err != nil {
		tflog.Debug(ctx, err.Error())
		return nil, []diag.Diagnostic{diag.NewErrorDiagnostic("Invalid attribute", "Settings.Config is invalid")}
	}
	apiSettigns.Config = settingsConfig

	apiModel.Settings = &apiSettigns

	// Convert discovery settings
	if m.DiscoverySettings != nil {
		if !m.DiscoverySettings.DiscoveryInterval.IsNull() {
			apiModel.DiscoverySettings = &sgsdkgo.Discoverysettings{
				DiscoveryInterval: m.DiscoverySettings.DiscoveryInterval.ValueFloat64(),
			}
		}

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
	}

	// Set isActive
	if !m.IsActive.IsNull() {
		apiModel.IsActive = (*sgsdkgo.IsArchiveEnum)(m.IsActive.ValueStringPointer())
	}

	return &apiModel, nil
}

func buildAPIModelToConnectorModel(apiResponse *sgsdkgo.GeneratedConnectorReadResponseMsg) (*ConnectorResourceModel, diag.Diagnostics) {
	connectorModel := &ConnectorResourceModel{
		Organization: flatteners.String(apiResponse.OrgId),
		ResourceName: flatteners.String(apiResponse.ResourceName),
		Description:  flatteners.String(apiResponse.Description),
		DiscoverySettings: &DiscoverySettings{
			DiscoveryInterval: flatteners.Float64(apiResponse.DiscoverySettings.DiscoveryInterval),
		},
		IsActive: flatteners.String(apiResponse.IsActive),
	}

	settings := map[string]*string{}
	settingsConfig, err := json.Marshal(apiResponse.Settings.Config)
	if err != nil {
		return nil, []diag.Diagnostic{diag.NewErrorDiagnostic("Unmarshal error", "Cannot unmarhsal Connector.Settings.Config object in response from sdk")}
	}
	config := string(settingsConfig)
	settings["config"] = &config
	settings["kind"] = &apiResponse.Settings.Kind
	settingsModel, diags := types.MapValueFrom(context.Background(), types.StringType, &settings)
	if diags.HasError() {
		return nil, diags
	}
	connectorModel.Settings = settingsModel

	// Discovery Settings
	var regions []Region
	if apiResponse.DiscoverySettings.Regions != nil {
		for _, r := range apiResponse.DiscoverySettings.Regions {
			regions = append(regions, Region{region: flatteners.String(r.Region)})
		}
		connectorModel.DiscoverySettings.Regions = regions
	}

	benchmarks, err := json.Marshal(apiResponse.DiscoverySettings.Benchmarks)
	if err != nil {
		return nil, []diag.Diagnostic{diag.NewErrorDiagnostic("Unmarshal error", "Cannot unmarhsal Connector.DiscoverySettings.Benchmarks object in response from sdk")}
	}
	connectorModel.DiscoverySettings.Benchmarks = flatteners.String(string(benchmarks))

	//TODO: process scope
	return connectorModel, nil
}

package connector

import (
	"context"
	"fmt"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	core "github.com/StackGuardian/sg-sdk-go/core"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &connectorResource{}

type connectorResource struct {
	client   *sgclient.Client
	org_name string
}

// NewConnectorResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &connectorResource{}
}

// Metadata returns the resource type name.
func (r *connectorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connector"
}

// Configure adds the provider configured client to the resource.
func (r *connectorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(*customTypes.ProviderInfo)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hashicups.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = provider.Client
	r.org_name = provider.Org_name
}

func (r *connectorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("resource_name"), req.ID)...)
}

// Create creates the resource and sets the initial Terraform state.
func (r *connectorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ConnectorResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := plan.ToAPIModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	reqResp, err := r.client.Connectors.CreateConnector(ctx, r.org_name, payload)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error creating connector", "Error in creating connector API call: "+err.Error())
		return
	}

	reqResp.Data.Settings.Config = payload.Settings.Config

	connectorModel, diags := buildAPIModelToConnectorModel(reqResp.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, &connectorModel)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *connectorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state ConnectorResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed state from client
	reqResp, err := r.client.Connectors.ReadConnector(ctx, state.ResourceName.ValueString(), r.org_name)
	if err != nil {
		apiErr := err.(*core.APIError)
		if apiErr.StatusCode == 404 {
			resp.State.RemoveResource(ctx)
			return
		}
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error reading connector", "Could not read connector "+state.ResourceName.ValueString()+": "+err.Error())
		return
	}

	if !state.Settings.IsNull() {
		var settingsModelValue *ConnectorSettingsModel
		diags = state.Settings.As(context.Background(), &settingsModelValue, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		var settingsConfigValue []*ConnectorSettingsConfigModel
		diags = settingsModelValue.Config.ElementsAs(ctx, &settingsConfigValue, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		reqResp.Msg.Settings.Config[0] = &sgsdkgo.SettingsConfig{
			InstallationId:          settingsConfigValue[0].InstallationId.ValueStringPointer(),
			GithubAppId:             settingsConfigValue[0].GithubAppId.ValueStringPointer(),
			GithubAppWebhookSecret:  settingsConfigValue[0].GithubAppWebhookSecret.ValueStringPointer(),
			GithubApiUrl:            settingsConfigValue[0].GithubApiUrl.ValueStringPointer(),
			GithubHttpUrl:           settingsConfigValue[0].GithubHttpUrl.ValueStringPointer(),
			GithubAppClientId:       settingsConfigValue[0].GithubAppClientId.ValueStringPointer(),
			GithubAppClientSecret:   settingsConfigValue[0].GithubAppClientSecret.ValueStringPointer(),
			GithubAppPemFileContent: settingsConfigValue[0].GithubAppPemFileContent.ValueStringPointer(),
			GithubAppWebhookUrl:     settingsConfigValue[0].GithubAppWebhookURL.ValueStringPointer(),
			GitlabCreds:             settingsConfigValue[0].GitlabCreds.ValueStringPointer(),
			GitlabHttpUrl:           settingsConfigValue[0].GitlabHttpUrl.ValueStringPointer(),
			GitlabApiUrl:            settingsConfigValue[0].GitlabApiUrl.ValueStringPointer(),
			AzureCreds:              settingsConfigValue[0].AzureCreds.ValueStringPointer(),
			AzureDevopsHttpUrl:      settingsConfigValue[0].AzureDevopsHttpUrl.ValueStringPointer(),
			AzureDevopsApiUrl:       settingsConfigValue[0].AzureDevopsApiUrl.ValueStringPointer(),
			BitbucketCreds:          settingsConfigValue[0].BitbucketCreds.ValueStringPointer(),
			AwsAccessKeyId:          settingsConfigValue[0].AwsAccessKeyId.ValueStringPointer(),
			AwsSecretAccessKey:      settingsConfigValue[0].AwsSecretAccessKey.ValueStringPointer(),
			AwsDefaultRegion:        settingsConfigValue[0].AwsDefaultRegion.ValueStringPointer(),
			ArmTenantId:             settingsConfigValue[0].ArmTenantId.ValueStringPointer(),
			ArmSubscriptionId:       settingsConfigValue[0].ArmSubscriptionId.ValueStringPointer(),
			ArmClientId:             settingsConfigValue[0].ArmClientId.ValueStringPointer(),
			ArmClientSecret:         settingsConfigValue[0].ArmClientSecret.ValueStringPointer(),
			GcpConfigFileContent:    settingsConfigValue[0].GcpConfigFileContent.ValueStringPointer(),
		}
	}

	connectorResourceModel, diags := buildAPIModelToConnectorModel(reqResp.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, connectorResourceModel)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *connectorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ConnectorResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := plan.ToAPIPatchedModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	reqResp, err := r.client.Connectors.UpdateConnector(ctx, plan.ResourceName.ValueString(), r.org_name, payload)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error updating connector", "Error in updating connector API call: "+err.Error())
		return
	}

	reqResp.Data.Settings.Config = payload.Settings.Config

	reqResp.Data.ResourceName = plan.ResourceName.ValueString()

	connectorModel, diags := buildAPIModelToConnectorModel(reqResp.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, &connectorModel)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *connectorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ConnectorResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.Connectors.DeleteConnector(ctx, state.ResourceName.ValueString(), r.org_name)
	if err != nil {
		apiErr := err.(*core.APIError)
		if apiErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting Connector",
			"Could not delete connector, unexpected error: "+err.Error(),
		)
		return
	}
}

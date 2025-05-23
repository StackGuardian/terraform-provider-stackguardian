package sgprovider

import (
	"context"
	"fmt"
	"os"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	sgoption "github.com/StackGuardian/sg-sdk-go/option"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	connectordatasource "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/connector"
	policydatasource "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/policy"
	roledatasource "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/role"
	roleassignmentdatasource "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/role_assignment"
	runnergroupdatasource "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/runner_group"
	runnergrouptoken "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/runner_group_token"
	stackoutputs "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/stack_outputs"
	stackworkflowoutputs "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/stack_workflow_outputs"
	workflowgroupdatasource "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/workflow_group"
	workflowoutputs "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/workflow_outputs"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/connector"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/policy"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/role"
	roleassignment "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/role_assignment"
	runnergroup "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/runner_group"
	workflowgroup "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/workflow_group"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &stackguardianProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &stackguardianProvider{
			version: version,
		}
	}
}

// stackguardianProvider is the provider implementation.
type stackguardianProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type stackguardianProviderModel struct {
	Api_key  types.String `tfsdk:"api_key"`
	Api_uri  types.String `tfsdk:"api_uri"`
	Org_name types.String `tfsdk:"org_name"`
}

// Metadata returns the provider type name.
func (p *stackguardianProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "stackguardian"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *stackguardianProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"org_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Stackguardian Organization name. **Required** if not using environment variable STACKGUARDIAN_ORG_NAME",
			},
			"api_key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "API key to authenticate on StackGuardian API. **Required** if not using environment variable STACKGUARDIAN_API_KEY",
			},
			"api_uri": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "API URI to set as prefix URL for StackGuardian API. Can also be configured using environment variable STACKGUARDIAN_API_URI",
			},
		},
	}
}

// Configure prepares a Stackguardian API client for data sources and resources.
func (p *stackguardianProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Stackguardian client")

	var config stackguardianProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Org_name.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("org_name"),
			"Unknown Stackguardian Organization Name",
			"The provider cannot create the Stackguardian API client as there is an unknown configuration value for the Stackguardian organization name. "+
				"Either set the value statically in the configuration, or use the STACKGUARDIAN_ORG_NAME environment variable.",
		)
	}

	if config.Api_key.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown Stackguardian API Key",
			"The provider cannot create the Stackguardian API client as there is an unknown configuration value for the Stackguardian API Key. "+
				"Either set the value statically in the configuration, or use the STACKGUARDIAN_API_URI environment variable.",
		)
	}

	if diags.HasError() {
		return
	}

	api_uri := "https://api.app.stackguardian.io"

	if os.Getenv("STACKGUARDIAN_API_URI") != "" {
		api_uri = os.Getenv("STACKGUARDIAN_API_URI")
	}

	org_name := os.Getenv("STACKGUARDIAN_ORG_NAME")
	api_key := os.Getenv("STACKGUARDIAN_API_KEY")

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	if !config.Org_name.IsNull() {
		org_name = config.Org_name.ValueString()
	}

	if !config.Api_key.IsNull() {
		api_key = config.Api_key.ValueString()
	}

	if !config.Api_uri.IsNull() {
		api_uri = config.Api_uri.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.
	if org_name == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("org_name"),
			"Missing Organization Name",
			"The provider cannot create the Stackguardian API client as there is an unknown configuration value for the Stackguardian organization name. "+
				"Either set the value statically in the configuration, or use the STACKGUARDIAN_ORG_NAME environment variable.",
		)
	}
	if api_key == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Organization Name",
			"The provider cannot create the Stackguardian API client as there is an unknown configuration value for the Stackguardian API Key. "+
				"Either set the value statically in the configuration, or use the STACKGUARDIAN_API_URI environment variable.",
		)
	}
	if api_uri == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_uri"),
			"Missing Organization Name",
			"The provider cannot create the Stackguardian API client as there is an unknown configuration value for the Stackguardian API URI. "+
				"Either set the value statically in the configuration, or use the STACKGUARDIAN_API_URI environment variable.",
		)
	}

	api_key = "apikey " + api_key
	client := sgclient.NewClient(
		sgoption.WithApiKey(api_key),
		sgoption.WithBaseURL(api_uri),
	)
	//Set the values in our struct
	provInfo := customTypes.ProviderInfo{
		ApiBaseURL: api_uri,
		ApiKey:     api_key,
		Org_name:   org_name,
		Client:     client,
	}
	// Make the HashiCups client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = &provInfo
	resp.ResourceData = &provInfo

	// Create a new client using the API key and base URL
	tflog.Debug(ctx, fmt.Sprintf("Organization: %s", org_name))
	tflog.Debug(ctx, fmt.Sprintf("API Key: %s", api_key))
	tflog.Debug(ctx, fmt.Sprintf("API URI: %s", api_uri))

	tflog.Debug(ctx, "Creating Stackguardian client")

	tflog.Info(ctx, "Configured Stackguardian client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *stackguardianProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		stackoutputs.NewDataSource,
		stackworkflowoutputs.NewDataSource,
		workflowoutputs.NewDataSource,
		connectordatasource.NewDataSource,
		roleassignmentdatasource.NewDataSource,
		workflowgroupdatasource.NewDataSource,
		roledatasource.NewDataSource,
		policydatasource.NewDataSource,
		runnergroupdatasource.NewDataSource,
		runnergrouptoken.NewDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *stackguardianProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		connector.NewResource,
		workflowgroup.NewResource,
		role.NewResource,
		roleassignment.NewResource,
		policy.NewResource,
		runnergroup.NewResource,
	}
}

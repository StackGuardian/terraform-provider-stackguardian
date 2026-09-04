package sgprovider

import (
	"context"
	"fmt"
	"net/http"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	sgoption "github.com/StackGuardian/sg-sdk-go/option"
	sgconfig "github.com/StackGuardian/terraform-provider-stackguardian/internal/config"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	connectordatasource "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/connector"
	policydatasource "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/policy"
	roledatasource "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/role"
	roleassignmentdatasource "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/role_assignment"
	runnergroupdatasource "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/runner_group"
	runnergrouptoken "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/runner_group_token"
	stackoutputs "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/stack_outputs"
	stacktemplatedatasource "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/stack_template"
	stacktemplaterevisiondatasource "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/stack_template_revision"
	stackworkflowoutputs "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/stack_workflow_outputs"
	workflowgitdatasource "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/workflow_git"
	workflowgroupdatasource "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/workflow_group"
	workflowoutputs "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/workflow_outputs"
	workflowsteptemplatedatasource "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/workflow_step_template"
	workflowsteptemplaterevisiondatasource "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/workflow_step_template_revision"
	workflowtemplatedatasource "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/workflow_template"
	workflowtemplaterevisiondatasource "github.com/StackGuardian/terraform-provider-stackguardian/internal/datasources/workflow_template_revision"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/connector"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/policy"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/role"
	roleassignment "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/role_assignment"
	rolev4 "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/role_v4"
	runnergroup "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/runner_group"
	stacktemplate "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/stack_template"
	stacktemplaterevision "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/stack_template_revision"
	workflowfromtemplate "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/workflow_from_template"
	workflowgit "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/workflow_git"
	workflowgroup "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/workflow_group"
	workflowsteptemplate "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/workflow_step_template"
	workflowsteptemplaterevision "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/workflow_step_template_revision"
	workflowtemplate "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/workflow_template"
	workflowtemplaterevision "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/workflow_template_revision"
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
func New(version string, customHeader http.Header) func() provider.Provider {
	return func() provider.Provider {
		return &stackguardianProvider{
			version:       version,
			customHeaders: customHeader,
		}
	}
}

// stackguardianProvider is the provider implementation.
type stackguardianProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version       string
	customHeaders http.Header
}

type stackguardianProviderModel struct {
	APIKey  types.String `tfsdk:"api_key"`
	APIUri  types.String `tfsdk:"api_uri"`
	OrgName types.String `tfsdk:"org_name"`
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
				MarkdownDescription: "StackGuardian Organization name. **Required** if not using environment variable STACKGUARDIAN_ORG_NAME",
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

// Configure prepares a StackGuardian API client for data sources and resources.
func (p *stackguardianProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring StackGuardian client")

	var config stackguardianProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.OrgName.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("org_name"),
			"Unknown StackGuardian Organization Name",
			"The provider cannot create the StackGuardian API client as there is an unknown configuration value for the StackGuardian organization name. "+
				"Either set the value statically in the configuration, or use the STACKGUARDIAN_ORG_NAME environment variable.",
		)
	}

	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown StackGuardian API Key",
			"The provider cannot create the StackGuardian API client as there is an unknown configuration value for the StackGuardian API Key. "+
				"Either set the value statically in the configuration, or use the STACKGUARDIAN_API_URI environment variable.",
		)
	}

	if diags.HasError() {
		return
	}

	cgf := sgconfig.Get()

	apiURI := cgf.ApiUri
	if !config.APIUri.IsNull() {
		apiURI = config.APIUri.ValueString()
	}

	orgName := cgf.OrgName
	if !config.OrgName.IsNull() {
		orgName = config.OrgName.ValueString()
	}

	apiKey := cgf.ApiKey
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.
	if orgName == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("org_name"),
			"Missing Organization Name",
			"The provider cannot create the StackGuardian API client as there is an unknown configuration value for the StackGuardian organization name. "+
				"Either set the value statically in the configuration, or use the STACKGUARDIAN_ORG_NAME environment variable.",
		)
	}
	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Organization Name",
			"The provider cannot create the StackGuardian API client as there is an unknown configuration value for the StackGuardian API Key. "+
				"Either set the value statically in the configuration, or use the STACKGUARDIAN_API_URI environment variable.",
		)
	}
	if apiURI == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_uri"),
			"Missing Organization Name",
			"The provider cannot create the StackGuardian API client as there is an unknown configuration value for the StackGuardian API URI. "+
				"Either set the value statically in the configuration, or use the STACKGUARDIAN_API_URI environment variable.",
		)
	}

	apiKey = "apikey " + apiKey
	client := sgclient.NewClient(
		sgoption.WithApiKey(apiKey),
		sgoption.WithBaseURL(apiURI),
		sgoption.WithHTTPHeader(p.customHeaders),
	)
	//Set the values in our struct
	provInfo := customTypes.ProviderInfo{
		ApiBaseURL: apiURI,
		ApiKey:     apiKey,
		OrgName:    orgName,
		Client:     client,
	}
	// Make the HashiCups client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = &provInfo
	resp.ResourceData = &provInfo

	// Create a new client using the API key and base URL
	tflog.Debug(ctx, fmt.Sprintf("Organization: %s", orgName))
	tflog.Debug(ctx, fmt.Sprintf("API Key: %s", apiKey))
	tflog.Debug(ctx, fmt.Sprintf("API URI: %s", apiURI))

	tflog.Debug(ctx, "Creating StackGuardian client")

	tflog.Info(ctx, "Configured StackGuardian client", map[string]any{"success": true})
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
		workflowsteptemplatedatasource.NewDataSource,
		workflowsteptemplaterevisiondatasource.NewDataSource,
		workflowtemplatedatasource.NewDataSource,
		workflowtemplaterevisiondatasource.NewDataSource,
		stacktemplatedatasource.NewDataSource,
		stacktemplaterevisiondatasource.NewDataSource,
		workflowgitdatasource.NewDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *stackguardianProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		connector.NewResource,
		workflowgroup.NewResource,
		workflowsteptemplate.NewResource,
		workflowsteptemplaterevision.NewResource,
		role.NewResource,
		roleassignment.NewResource,
		policy.NewResource,
		runnergroup.NewResource,
		rolev4.NewResource,
		workflowtemplate.NewResource,
		workflowtemplaterevision.NewResource,
		stacktemplate.NewResource,
		stacktemplaterevision.NewResource,
		workflowgit.NewResource,
		workflowfromtemplate.NewResource,
	}
}

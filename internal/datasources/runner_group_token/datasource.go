package runnergrouptoken

import (
	"context"
	"fmt"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

var (
	_ datasource.DataSource              = &runnerGroupTokenDataSource{}
	_ datasource.DataSourceWithConfigure = &runnerGroupTokenDataSource{}
)

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &runnerGroupTokenDataSource{}
}

type runnerGroupTokenDataSource struct {
	client     *sgclient.Client
	orgName    string
	apiBaseURL string
	apiKey     string
}

func (d *runnerGroupTokenDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_runner_group_token"
}

func (d *runnerGroupTokenDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	provInfo, ok := req.ProviderData.(*customTypes.ProviderInfo)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hashicups.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.apiBaseURL = provInfo.ApiBaseURL
	d.apiKey = provInfo.ApiKey
	d.client = provInfo.Client
	d.orgName = provInfo.Org_name
}

func (d *runnerGroupTokenDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config runnerGroupTokenModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResponse, err := getAPIToken(config.RunnerGroupID.ValueString(), d.apiBaseURL, d.apiKey)
	if err != nil {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic("failed to call api token url", err.Error()))
		return
	}

	runnerGroupTokenModel, diags := buildAPIModelToRunnerGroupTokenModel(apiResponse)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	runnerGroupTokenModel.RunnerGroupID = config.RunnerGroupID

	resp.Diagnostics.Append(resp.State.Set(ctx, runnerGroupTokenModel)...)
}

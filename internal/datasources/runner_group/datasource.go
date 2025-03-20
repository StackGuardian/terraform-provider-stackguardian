package runnergroupdatasource

import (
	"context"
	"fmt"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	runnergroup "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/runner_group"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &runnerGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &runnerGroupDataSource{}
)

func NewDataSource() datasource.DataSource {
	return &runnerGroupDataSource{}
}

type runnerGroupDataSource struct {
	client   *sgclient.Client
	org_name string
}

func (d *runnerGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_runner_group"
}

func (d *runnerGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = provInfo.Client
	d.org_name = provInfo.Org_name
}

func (d *runnerGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config runnergroup.RunnerGroupResourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	readRunnerGroupReqBools := false
	reqResp, err := d.client.RunnerGroups.ReadRunnerGroup(ctx, d.org_name, config.ResourceName.ValueString(), &sgsdkgo.ReadRunnerGroupRequest{
		GetActiveWorkflows:        &readRunnerGroupReqBools,
		GetActiveWorkflowsDetails: &readRunnerGroupReqBools,
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to read runner group.", err.Error())
		return
	}

	roleModel, diags := runnergroup.BuildAPIModelToRunnerGroupModel(reqResp.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, roleModel)...)
}

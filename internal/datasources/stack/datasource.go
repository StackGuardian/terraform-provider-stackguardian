package stackdatasource

import (
	"context"
	"fmt"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	stackresource "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/stack"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &stackDataSource{}
	_ datasource.DataSourceWithConfigure = &stackDataSource{}
)

func NewDataSource() datasource.DataSource {
	return &stackDataSource{}
}

type stackDataSource struct {
	client  *sgclient.Client
	orgName string
}

func (d *stackDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack"
}

func (d *stackDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	provInfo, ok := req.ProviderData.(*customTypes.ProviderInfo)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *customTypes.ProviderInfo, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = provInfo.Client
	d.orgName = provInfo.OrgName
}

func (d *stackDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config stackresource.StackResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.Id.ValueString()
	workflowGroupId := config.WorkflowGroupId.ValueString()

	readResp, err := d.client.Stacks.ReadStack(ctx, d.orgName, id, workflowGroupId, &sgsdkgo.ReadStackQueryParams{RefreshWorkflowsConfig: sgsdkgo.Bool(true)})
	if err != nil {
		resp.Diagnostics.AddError("Unable to read stack.", err.Error())
		return
	}
	if readResp == nil || readResp.Msg == nil {
		resp.Diagnostics.AddError("Error reading stack", "API response is empty")
		return
	}

	model, diags := stackresource.BuildAPIModelToStackModel(ctx, d.orgName, readResp.Msg, config.WorkflowGroupId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

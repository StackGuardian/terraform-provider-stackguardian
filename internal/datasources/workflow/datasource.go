package workflow

import (
	"context"
	"fmt"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	workflowresource "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/workflow"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &workflowDataSource{}
	_ datasource.DataSourceWithConfigure = &workflowDataSource{}
)

func NewDataSource() datasource.DataSource {
	return &workflowDataSource{}
}

type workflowDataSource struct {
	client  *sgclient.Client
	orgName string
}

func (d *workflowDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow"
}

func (d *workflowDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.orgName = provInfo.Org_name
}

func (d *workflowDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config workflowresource.WorkflowResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readResp, err := d.client.Workflows.ReadWorkflow(ctx, d.orgName, config.ResourceName.ValueString(), config.WorkflowGroupId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read workflow.", err.Error())
		return
	}

	if readResp == nil {
		resp.Diagnostics.AddError("Error reading workflow", "API response is empty")
		return
	}

	model, diags := workflowresource.ConvertWorkflowFromAPI(ctx, readResp, config.WorkflowGroupId.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

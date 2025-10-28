package stackworkflowoutputs

import (
	"context"
	"fmt"
	"io"
	"net/http"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &stackWorkflowOutputsDataSource{}
	_ datasource.DataSourceWithConfigure = &stackWorkflowOutputsDataSource{}
)

func NewDataSource() datasource.DataSource {
	return &stackWorkflowOutputsDataSource{}
}

type stackWorkflowOutputsDataSource struct {
	client  *sgclient.Client
	orgName string
}

func (d *stackWorkflowOutputsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_workflow_outputs"
}

func (d *stackWorkflowOutputsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.orgName = provInfo.Org_name
}

func (d *stackWorkflowOutputsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config stackWorkflowOutputsDataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stackWorkflowOutputs, err := d.client.StackWorkflows.StackWorkflowOutputs(ctx, d.orgName, config.Stack.ValueString(), config.Workflow.ValueString(), config.WorkflowGroup.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Request failure to read stack workflow outputs", err.Error())
		return
	}

	if stackWorkflowOutputs.Data.OutputsSignedUrl == "" {
		resp.Diagnostics.AddError("Unable to fetch stack workflow outputs", "")
		return
	}

	reqResp, err := http.Get(stackWorkflowOutputs.Data.OutputsSignedUrl)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read stack workflow outputs", "")
		return
	}
	defer reqResp.Body.Close()

	body, err := io.ReadAll(reqResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to parse stack workflow outputs", "")
		return
	}

	stackWorkflowOutputsDataSourceModel, diags := buildAPIModelToTerraformModel(body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stackWorkflowOutputsDataSourceModel.Stack = config.Stack
	stackWorkflowOutputsDataSourceModel.Workflow = config.Workflow
	stackWorkflowOutputsDataSourceModel.WorkflowGroup = config.WorkflowGroup

	resp.Diagnostics.Append(resp.State.Set(ctx, stackWorkflowOutputsDataSourceModel)...)
}

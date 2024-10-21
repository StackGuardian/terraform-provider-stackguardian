package workflowoutputs

import (
	"context"
	"fmt"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &workflowOutputsDataSource{}
	_ datasource.DataSourceWithConfigure = &workflowOutputsDataSource{}
)

// NewCoffeesDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &workflowOutputsDataSource{}
}

// coffeesDataSource is the data source implementation.
type workflowOutputsDataSource struct {
	client  *sgclient.Client
	orgName string
}

// Metadata returns the data source type name.
func (d *workflowOutputsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow_outputs"
}

func (d *workflowOutputsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *workflowOutputsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config workflowOutputsDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	workflowOuputs, err := d.client.Workflows.Outputs(ctx, d.orgName, config.Workflow.ValueString(), config.WorkflowGroup.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read stack outputs", err.Error())
	}

	stackOutputsDataSourceModel, diags := buildAPIModelToTerraformModel(workflowOuputs.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	stackOutputsDataSourceModel.Workflow = config.Workflow
	stackOutputsDataSourceModel.WorkflowGroup = config.WorkflowGroup

	resp.Diagnostics.Append(resp.State.Set(ctx, stackOutputsDataSourceModel)...)
}

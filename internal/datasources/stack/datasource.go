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

// NewDataSource returns the stackguardian_stack data source; registered in the provider's DataSources().
func NewDataSource() datasource.DataSource {
	return &stackDataSource{}
}

// stackDataSource reads an existing stack by id and workflow_group_id. It reuses the
// stackguardian_stack resource's model and API-to-model conversion
// (stackresource.BuildAPIModelToStackModel), so the two always expose the same
// attributes with the same values.
type stackDataSource struct {
	client  *sgclient.Client
	orgName string
}

// Metadata sets the data source type name, stackguardian_stack.
func (d *stackDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack"
}

// Configure stores the API client and organization from the provider configuration.
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

// Read fetches the stack with refresh_workflows_config=true, so workflows_config
// reflects the stack's live workflows rather than the copy stored on the stack (the
// same read the resource does), and converts it with the resource's
// BuildAPIModelToStackModel. workflow_group_id isn't part of the API response, so
// it's carried over from the configuration.
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

	model, diags := stackresource.BuildAPIModelToStackModel(ctx, readResp.Msg, config.WorkflowGroupId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

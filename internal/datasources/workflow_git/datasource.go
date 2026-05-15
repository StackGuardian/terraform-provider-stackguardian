package workflowgit

import (
	"context"
	"fmt"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	workflowgit "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/workflow_git"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &workflowGitDataSource{}
	_ datasource.DataSourceWithConfigure = &workflowGitDataSource{}
)

func NewDataSource() datasource.DataSource {
	return &workflowGitDataSource{}
}

type workflowGitDataSource struct {
	client  *sgclient.Client
	orgName string
}

func (d *workflowGitDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow_git"
}

func (d *workflowGitDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *workflowGitDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config workflowgit.WorkflowGitResourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.Id.ValueString()
	if id == "" {
		resp.Diagnostics.AddError("id must be provided", "")
		return
	}

	workflowGroupId := config.WorkflowGroupId.ValueString()
	if workflowGroupId == "" {
		resp.Diagnostics.AddError("workflow_group_id must be provided", "")
		return
	}

	readResp, err := d.client.Workflows.ReadWorkflow(ctx, d.orgName, id, workflowGroupId)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read workflow_git.", err.Error())
		return
	}

	if readResp == nil {
		resp.Diagnostics.AddError("Error reading workflow_git", "API response is empty")
		return
	}

	model, diags := workflowgit.ConvertWorkflowGitFromAPI(ctx, readResp)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.WorkflowGroupId = config.WorkflowGroupId

	resp.Diagnostics.Append(resp.State.Set(ctx, model)...)
}

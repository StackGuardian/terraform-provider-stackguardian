package workflowgroupsdatasource

import (
	"context"
	"fmt"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/workflowGroups"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &workflowGroupsDataSource{}
	_ datasource.DataSourceWithConfigure = &workflowGroupsDataSource{}
)

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &workflowGroupsDataSource{}
}

type workflowGroupsDataSource struct {
	client  *sgclient.Client
	orgName string
}

func (d *workflowGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow_group"
}

func (d *workflowGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *workflowGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config workflowGroups.WorkflowGroupResourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	reqResp, err := d.client.WorkflowGroups.ReadWorkflowGroup(ctx, d.orgName, config.ResourceName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read workflow group.", err.Error())
		return
	}

	workflowGroupsDataSourceModel, diags := workflowGroups.BuildAPIModelToWorkflowGroupModel(reqResp.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	workflowGroupsDataSourceModel.ResourceName = config.ResourceName

	resp.Diagnostics.Append(resp.State.Set(ctx, workflowGroupsDataSourceModel)...)
}

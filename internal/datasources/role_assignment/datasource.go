package roleassignment

import (
	"context"
	"fmt"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	roleassignment "github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/role_assignment"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &roleAssignmentDataSource{}
	_ datasource.DataSourceWithConfigure = &roleAssignmentDataSource{}
)

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &roleAssignmentDataSource{}
}

type roleAssignmentDataSource struct {
	client  *sgclient.Client
	orgName string
}

func (d *roleAssignmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_assignment"
}

func (d *roleAssignmentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *roleAssignmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config roleassignment.RoleAssignmentResourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	getUserRequest := &sgsdkgo.GetorRemoveUserFromOrganization{
		UserId: config.UserId.ValueStringPointer(),
	}

	readRoleAssignmentResponse, err := d.client.AccessManagement.ReadUser(ctx, d.orgName, getUserRequest)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read role assignment.", err.Error())
		return
	}

	roleAssignmentDataSourceModel, diags := roleassignment.BuildAPIModelToRoleAssignmentModel(readRoleAssignmentResponse.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, roleAssignmentDataSourceModel)...)
}

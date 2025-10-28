package policy

import (
	"context"
	"fmt"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/policy"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = &policyDatasource{}
	_ datasource.DataSourceWithConfigure = &policyDatasource{}
)

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &policyDatasource{}
}

type policyDatasource struct {
	client  *sgclient.Client
	orgName string
}

func (d *policyDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

func (d *policyDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *policyDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config policy.PolicyResourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	reqResp, err := d.client.Policies.ReadPolicy(ctx, d.orgName, config.ResourceName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read role assignment.", err.Error())
		return
	}

	respPolicyGeneral := reqResp.Msg.General
	policyGeneralModel := sgsdkgo.PolicyGeneralResponse{
		ResourceName: respPolicyGeneral.ResourceName,
		Description:  respPolicyGeneral.Description,
		Approvers:    respPolicyGeneral.Approvers,

		NumberOfApprovalsRequired: respPolicyGeneral.NumberOfApprovalsRequired,
		Tags:                      respPolicyGeneral.Tags,
		ContextTags:               respPolicyGeneral.ContextTags,
		EnforcedOn:                respPolicyGeneral.EnforcedOn,
		PoliciesConfig:            respPolicyGeneral.PoliciesConfig,
	}

	policyModel, diags := policy.BuildAPIModelToPolicyModel(&policyGeneralModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, policyModel)...)
}

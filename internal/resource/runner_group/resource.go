package runnergroup

import (
	"context"
	"fmt"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/sg-sdk-go/core"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type runnerGroupResource struct {
	resource.Resource
	client   *sgclient.Client
	org_name string
}

var (
	_ resource.Resource               = &runnerGroupResource{}
	_ resource.ResourceWithConfigure  = &runnerGroupResource{}
	_ resource.ResourceWithModifyPlan = &runnerGroupResource{}
)

func NewResource() resource.Resource {
	return &runnerGroupResource{}
}

// Metadata returns the resource type name.
func (r *runnerGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_runner_group"
}

func (r *runnerGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	providerInfo, ok := req.ProviderData.(*customTypes.ProviderInfo)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hashicups.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = providerInfo.Client
	r.org_name = providerInfo.Org_name
}

func (r *runnerGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("resource_name"), req.ID)...)
}

func (r *runnerGroupResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if !req.State.Raw.IsNull() && !req.Plan.Raw.IsNull() {
		var state RunnerGroupResourceModel
		var plan RunnerGroupResourceModel

		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if !plan.ResourceName.Equal(state.ResourceName) {
			resp.RequiresReplace = append(resp.RequiresReplace, path.Root("resource_name"))
		}
	}
}

func (r *runnerGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RunnerGroupResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := plan.ToAPIModel()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	runnerGroup, err := r.client.RunnerGroups.CreateNewRunnerGroup(context.TODO(), r.org_name, payload)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error create runner group", "Error in creating runner group API call: "+err.Error())
		return
	}

	runnerGroupResourceModel, diags := BuildAPIModelToRunnerGroupModel(runnerGroup.Data)
	runnerGroupResourceModel.RunnerToken = types.StringNull()

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, runnerGroupResourceModel)...)
}

func (r *runnerGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RunnerGroupResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	readRunnerGroupReqBools := false
	runnerGroup, err := r.client.RunnerGroups.ReadRunnerGroup(ctx, r.org_name, state.ResourceName.ValueString(), &sgsdkgo.ReadRunnerGroupRequest{
		GetActiveWorkflows:        &readRunnerGroupReqBools,
		GetActiveWorkflowsDetails: &readRunnerGroupReqBools,
	})
	if err != nil {
		// if a managed resource is no longer found then remove it from state
		if apiErr, ok := err.(*core.APIError); ok {
			if apiErr.StatusCode == 404 {
				resp.State.RemoveResource(ctx)
			}
		}
		resp.Diagnostics.AddError("Error reading policy", err.Error())
		return
	}

	runnerGroupResourceModel, diags := BuildAPIModelToRunnerGroupModel(runnerGroup.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, runnerGroupResourceModel)...)

}

func (r *runnerGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RunnerGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	patchedAPIModel, diags := plan.ToPatchedAPIModel()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updatedRunnerGroup, err := r.client.RunnerGroups.UpdateRunnerGroup(ctx, r.org_name, plan.ResourceName.ValueString(), patchedAPIModel)
	if err != nil {
		resp.Diagnostics.AddError("Error updating policy", err.Error())
		return
	}

	runnerGroupResourceModel, diags := BuildAPIModelToRunnerGroupModel(updatedRunnerGroup.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	runnerGroupResourceModel.ResourceName = plan.ResourceName

	resp.Diagnostics.Append(resp.State.Set(ctx, runnerGroupResourceModel)...)
}

func (r *runnerGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RunnerGroupResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.RunnerGroups.DeleteRunnerGroup(ctx, r.org_name, state.ResourceName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting policy", err.Error())
		return
	}
}

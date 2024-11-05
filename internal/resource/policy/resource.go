package policy

import (
	"context"
	"fmt"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	core "github.com/StackGuardian/sg-sdk-go/core"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource = &policyResrouce{}
)

func NewResource() resource.Resource {
	return &policyResrouce{}
}

type policyResrouce struct {
	orgName string
	client  *sgclient.Client
}

func (r *policyResrouce) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

func (r *policyResrouce) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(*customTypes.ProviderInfo)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hashicups.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = provider.Client
	r.orgName = provider.Org_name
}

func (r *policyResrouce) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("resource_name"), req.ID)...)
}

func (r *policyResrouce) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if !req.State.Raw.IsNull() && !req.Plan.Raw.IsNull() {
		var state PolicyResourceModel
		var plan PolicyResourceModel

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

func (r *policyResrouce) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PolicyResourceModel
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

	policy, err := r.client.Policies.CreatePolicy(ctx, r.orgName, payload)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error creating policy", "Error in creating policy API call: "+err.Error())
		return
	}

	policyResourceModel, diags := BuildAPIModelToPolicyModel(policy.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, policyResourceModel)...)

}

func (r *policyResrouce) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PolicyResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.Policies.ReadPolicy(ctx, r.orgName, state.ResourceName.ValueString())
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

	policyResourceModel, diags := BuildAPIModelToPolicyModel(policy.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, policyResourceModel)...)

}

func (r *policyResrouce) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PolicyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := plan.ToPatchedAPIModel()
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updatedPolicy, err := r.client.Policies.UpdatePolicy(ctx, r.orgName, plan.ResourceName.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError("Error updating policy", err.Error())
		return
	}

	policyResourceModel, diags := BuildAPIModelToPolicyModel(updatedPolicy.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	policyResourceModel.ResourceName = plan.ResourceName

	resp.Diagnostics.Append(resp.State.Set(ctx, policyResourceModel)...)
}

func (r *policyResrouce) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PolicyResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Policies.DeletePolicy(ctx, r.orgName, state.ResourceName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting policy", err.Error())
		return
	}
}

package stack

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
	_ resource.Resource                = &stackResource{}
	_ resource.ResourceWithConfigure   = &stackResource{}
	_ resource.ResourceWithImportState = &stackResource{}
)

type stackResource struct {
	client   *sgclient.Client
	org_name string
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &stackResource{}
}

// Metadata returns the resource type name.
func (r *stackResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack"
}

// Configure adds the provider configured client to the resource.
func (r *stackResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Add a nil check when handling ProviderData because Terraform
	// sets that data after it calls the ConfigureProvider RPC.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(*customTypes.ProviderInfo)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *customTypes.ProviderInfo, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = provider.Client
	r.org_name = provider.Org_name
}

// ImportState imports a stack using its ID.
func (r *stackResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// Create creates the resource and sets the initial Terraform state.
func (r *stackResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StackResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := plan.ToAPIModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	wfGrpId := plan.WorkflowGroupId.ValueString()

	createResp, err := r.client.Stacks.CreateStack(ctx, r.org_name, wfGrpId, payload)
	if err != nil {
		resp.Diagnostics.AddError("failed to create stack", err.Error())
		return
	}

	// Get the stack ID from create response (use SubResourceId as the primary identifier)
	stackID := createResp.Data.Stack.SubResourceId

	// Call read to get the full state since create response may not return all values
	readResp, err := r.client.Stacks.ReadStack(ctx, r.org_name, stackID, wfGrpId)
	if err != nil {
		resp.Diagnostics.AddError("Error reading created stack", "Could not read the created stack: "+err.Error())
		return
	}

	stackModel, diags := BuildAPIModelToStackModel(ctx, readResp.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, &stackModel)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *stackResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state StackResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stackId := state.Id.ValueString()
	wfGrpId := state.WorkflowGroupId.ValueString()

	// Get refreshed state from client
	readResp, err := r.client.Stacks.ReadStack(ctx, r.org_name, stackId, wfGrpId)
	if err != nil {
		// If a managed resource is no longer found then remove it from the state
		if apiErr, ok := err.(*core.APIError); ok {
			if apiErr.StatusCode == 404 {
				resp.State.RemoveResource(ctx)
				return
			}
		}
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error reading stack", "Could not read stack "+state.ResourceName.ValueString()+": "+err.Error())
		return
	}

	stackModel, diags := BuildAPIModelToStackModel(ctx, readResp.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, stackModel)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *stackResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan StackResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := plan.ToUpdateAPIModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stackId := plan.Id.ValueString()
	wfGrpId := plan.WorkflowGroupId.ValueString()

	_, err := r.client.Stacks.UpdateStack(ctx, r.org_name, stackId, wfGrpId, payload)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error updating stack", "Error in updating stack "+
			plan.ResourceName.ValueString()+": "+err.Error())
		return
	}

	// Call read to get the updated stack resource to set the state
	updatedStack, err := r.client.Stacks.ReadStack(ctx, r.org_name, stackId, wfGrpId)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error reading the updated state of stack",
			"Could not read the updated state of stack "+plan.ResourceName.ValueString()+": "+err.Error())
		return
	}

	stackModel, diags := BuildAPIModelToStackModel(ctx, updatedStack.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, stackModel)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *stackResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state StackResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stackId := state.Id.ValueString()
	wfGrpId := state.WorkflowGroupId.ValueString()

	_, err := r.client.Stacks.DeleteStack(ctx, r.org_name, stackId, wfGrpId)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error deleting stack", "Error in deleting stack "+state.ResourceName.ValueString()+": "+err.Error())
		return
	}
}

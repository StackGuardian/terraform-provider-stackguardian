package user

import (
	"context"
	"fmt"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	core "github.com/StackGuardian/sg-sdk-go/core"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &userResource{}

type userResource struct {
	resource.Resource
	client   *sgclient.Client
	org_name string
}

// NewUserResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &userResource{}
}

// Metadata returns the resource type name.
func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// Configure adds the provider configured client to the resource.
func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates the resource and sets the initial Terraform state.
func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan UserResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, planDiags := plan.ToCreateAPIModel(ctx)
	resp.Diagnostics.Append(planDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	reqResp, err := r.client.UsersRoles.AddUser(ctx, r.org_name, payload)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error creating User", "Error in creating User via API call: "+err.Error())
		return
	}

	userModel, diags := buildAPIModelToUserModel(reqResp.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, &userModel)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state UserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, paylodDiags := state.ToGetAPIModel(ctx)
	resp.Diagnostics.Append(paylodDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed state from client
	user, err := r.client.UsersRoles.GetUser(ctx, r.org_name, payload)
	if err != nil {
		// If a managed resource is no longer found then remove it from the state
		if apiErr, ok := err.(*core.APIError); ok {
			if apiErr.StatusCode == 404 {
				resp.State.RemoveResource(ctx)
				return
			}
		}
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error reading user", "Could not read user "+state.UserId.ValueString()+": "+err.Error())
		return
	}

	roleResourceModel, diags := buildAPIModelToUserModel(user.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, roleResourceModel)...)

}

// Update updates the resource and sets the updated Terraform state on success.
func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan UserResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := plan.ToCreateAPIModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.UsersRoles.UpdateUser(ctx, r.org_name, payload)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error updating user", "Error in updating user "+
			plan.UserId.ValueString()+": "+err.Error())
		return
	}

	getPayload, paylodDiags := plan.ToGetAPIModel(ctx)
	resp.Diagnostics.Append(paylodDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call read to get the updated WFG resource to set the state
	updatedUser, err := r.client.UsersRoles.GetUser(ctx, r.org_name, getPayload)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error reading the updated state of user",
			"Could not read the updated state of user "+plan.UserId.ValueString()+": "+err.Error())
		return
	}

	roleResourceModel, diags := buildAPIModelToUserModel(updatedUser.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, roleResourceModel)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state UserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	removePayload, paylodDiags := state.ToGetAPIModel(ctx)
	resp.Diagnostics.Append(paylodDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.UsersRoles.RemoveUser(ctx, r.org_name, removePayload)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error deleting user", "Error in deleting user "+state.UserId.ValueString()+": "+err.Error())
		return
	}
	//TODO: check if we need to update the state
}

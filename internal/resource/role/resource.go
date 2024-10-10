package role

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

var _ resource.Resource = &roleResource{}

type roleResource struct {
	client   *sgclient.Client
	org_name string
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &roleResource{}
}

// Metadata returns the resource type name.
func (r *roleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

// Configure adds the provider configured client to the resource.
func (r *roleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.org_name = provider.Org_name
}

func (r *roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("resource_name"), req.ID)...)
}

// Create creates the resource and sets the initial Terraform state.
func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RoleResourceModel
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

	reqResp, err := r.client.UsersRoles.CreateRole(ctx, r.org_name, payload)
	if err != nil {
		if apiErr, ok := err.(*core.APIError); ok {
			// Check if resource already exists
			if apiErr.StatusCode == 400 {
				role, readErr := r.client.UsersRoles.ReadRole(ctx, r.org_name, plan.ResourceName.ValueString())
				if readErr != nil {
					tflog.Error(ctx, readErr.Error())
					//Return the original error if read also fails
					resp.Diagnostics.AddError("Error creating Role", "Could not create Role "+plan.ResourceName.ValueString()+": "+err.Error())
					return
				}

				roleResourceModel, diags := buildAPIModelToRoleModel(role.Msg)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, roleResourceModel)...)
				return
			}
		}
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error creating Role", "Error in creating Role API call: "+err.Error())
		return
	}

	roleModel, diags := buildAPIModelToRoleModel(reqResp.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, &roleModel)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state RoleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed state from client
	role, err := r.client.UsersRoles.ReadRole(ctx, r.org_name, state.ResourceName.ValueString())
	if err != nil {
		// If a managed resource is no longer found then remove it from the state
		if apiErr, ok := err.(*core.APIError); ok {
			if apiErr.StatusCode == 404 {
				resp.State.RemoveResource(ctx)
				return
			}
		}
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error reading Role", "Could not read Role "+state.ResourceName.ValueString()+": "+err.Error())
		return
	}

	roleResourceModel, diags := buildAPIModelToRoleModel(role.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, roleResourceModel)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RoleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := plan.ToPatchedAPIModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.UsersRoles.UpdateRole(ctx, r.org_name, plan.ResourceName.ValueString(), payload)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error updating role", "Error in updating role "+
			plan.ResourceName.ValueString()+": "+err.Error())
		return
	}

	// Call read to get the updated WFG resource to set the state
	updatedRole, err := r.client.UsersRoles.ReadRole(ctx, r.org_name, plan.ResourceName.ValueString())
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error reading the updated state of role",
			"Could not read the updated state of role "+plan.ResourceName.ValueString()+": "+err.Error())
		return
	}

	roleResourceModel, diags := buildAPIModelToRoleModel(updatedRole.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, roleResourceModel)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RoleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UsersRoles.DeleteRole(ctx, r.org_name, state.ResourceName.ValueString())
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error deleting role", "Error in deleting role "+state.ResourceName.ValueString()+": "+err.Error())
		return
	}
	//TODO: check if we need to update the state
}

package rolev4

import (
	"context"
	"fmt"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	core "github.com/StackGuardian/sg-sdk-go/core"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/role"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource               = &RoleV4Resource{}
	_ resource.ResourceWithConfigure  = &RoleV4Resource{}
	_ resource.ResourceWithModifyPlan = &RoleV4Resource{}
)

type RoleV4Resource struct {
	role.RoleResource
	client   *sgclient.Client
	org_name string
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &RoleV4Resource{}
}

// Metadata returns the resource type name.
func (r *RoleV4Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rolev4"
}

// Configure adds the provider configured client to the resource.
func (r *RoleV4Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RoleV4Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("resource_name"), req.ID)...)
}

func (r *RoleV4Resource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if !req.State.Raw.IsNull() && !req.Plan.Raw.IsNull() {
		var state RoleV4ResourceModel
		var plan RoleV4ResourceModel

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

// Create creates the resource and sets the initial Terraform state.
func (r *RoleV4Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RoleV4ResourceModel

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

				roleResourceModel, diags := BuildAPIModelToRoleModel(role.Msg)
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

	roleModel, diags := BuildAPIModelToRoleModel(reqResp.Data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, &roleModel)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *RoleV4Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state RoleV4ResourceModel

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

	roleResourceModel, diags := BuildAPIModelToRoleModel(role.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, roleResourceModel)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *RoleV4Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RoleV4ResourceModel
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

	roleResourceModel, diags := BuildAPIModelToRoleModel(updatedRole.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, roleResourceModel)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *RoleV4Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RoleV4ResourceModel
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
}

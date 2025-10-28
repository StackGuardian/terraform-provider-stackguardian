package workflowgroup

import (
	"context"
	"fmt"
	"strings"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	core "github.com/StackGuardian/sg-sdk-go/core"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource               = &workflowGroupResource{}
	_ resource.ResourceWithConfigure  = &workflowGroupResource{}
	_ resource.ResourceWithModifyPlan = &workflowGroupResource{}
)

type workflowGroupResource struct {
	client   *sgclient.Client
	org_name string
}

// NewWorkflowGroupResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &workflowGroupResource{}
}

// Metadata returns the resource type name.
func (r *workflowGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow_group"
}

// Configure adds the provider configured client to the resource.
func (r *workflowGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *workflowGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("resource_name"), req.ID)...)
}

func (r *workflowGroupResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if !req.State.Raw.IsNull() && !req.Plan.Raw.IsNull() {
		var state WorkflowGroupResourceModel
		var plan WorkflowGroupResourceModel

		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if !plan.Id.Equal(state.Id) {
			resp.RequiresReplace = append(resp.RequiresReplace, path.Root("resource_name"))
		}
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *workflowGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WorkflowGroupResourceModel
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

	var responseData *sgsdkgo.WorkflowGroupDataResponse

	// Check if resource name is nested. If so, create a child workflow group
	var reqResp *sgsdkgo.WorkflowGroupCreateResponse
	var err error
	if workflows := strings.Split(*payload.ResourceName, "/"); len(workflows) > 1 {
		resourceName := workflows[len(workflows)-1]
		payload.ResourceName = &resourceName
		ParentWfg := strings.Join(workflows[:len(workflows)-1], "/")

		reqResp, err = r.client.WorkflowGroups.CreateChildWorkflowGroup(ctx, r.org_name,
			ParentWfg, payload)
		if err != nil {
			// Check if resource already exists
			if apiErr, ok := err.(*core.APIError); ok {
				if apiErr.StatusCode == 400 {
					workflowGroup, readErr := r.client.WorkflowGroups.ReadWorkflowGroup(ctx, r.org_name, plan.ResourceName.ValueString())
					if readErr != nil {
						// If read also fails return the original error
						tflog.Error(ctx, readErr.Error())
						resp.Diagnostics.AddError("Error creating child workflowGroup", "Could not create child workflowGroup "+plan.ResourceName.ValueString()+": "+err.Error())
						return
					}

					//For cases where WFG is a nested one, the resource name is returned as the full path to match the resource_name in resource definition
					if workflowGroup.Msg.SubResourceId != nil {
						fullResourceName := strings.Replace(*workflowGroup.Msg.SubResourceId, "/wfgrps/", "", 1)
						workflowGroup.Msg.ResourceName = &fullResourceName
					}

					workflowGroupResourceModel, diags := BuildAPIModelToWorkflowGroupModel(workflowGroup.Msg)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}

					resp.Diagnostics.Append(resp.State.Set(ctx, workflowGroupResourceModel)...)
					return
				}
			}
			tflog.Error(ctx, err.Error())
			resp.Diagnostics.AddError("Error creating child workflowGroup", "Error in creating child workflowGroup API call: "+err.Error())
			return
		}
		//For cases where WFG is a nested one, the resource name is returned as the full path to match the resource_name in resource definition
		if reqResp.Data.SubResourceId != nil {
			fullResourceName := strings.Replace(*reqResp.Data.SubResourceId, "/wfgrps/", "", 1)
			reqResp.Data.ResourceName = &fullResourceName
		}
		responseData = reqResp.Data
	} else { //If not, create a normal workflow group
		reqResp, err = r.client.WorkflowGroups.CreateWorkflowGroup(ctx, r.org_name, payload)
		if err != nil {
			// Check if resource already exists
			if apiErr, ok := err.(*core.APIError); ok {
				if apiErr.StatusCode == 400 {
					workflowGroup, readErr := r.client.WorkflowGroups.ReadWorkflowGroup(ctx, r.org_name, plan.ResourceName.ValueString())
					if readErr != nil {
						// If read also fails return the original error
						tflog.Error(ctx, readErr.Error())
						resp.Diagnostics.AddError("Error creating workflowGroup", "Could not create workflowGroup "+plan.ResourceName.ValueString()+": "+err.Error())
						return
					}

					//For cases where WFG is a nested one, the resource name is returned as the full path to match the resource_name in resource definition
					if workflowGroup.Msg.SubResourceId != nil {
						fullResourceName := strings.Replace(*workflowGroup.Msg.SubResourceId, "/wfgrps/", "", 1)
						workflowGroup.Msg.ResourceName = &fullResourceName
					}

					workflowGroupResourceModel, diags := BuildAPIModelToWorkflowGroupModel(workflowGroup.Msg)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}

					resp.Diagnostics.Append(resp.State.Set(ctx, workflowGroupResourceModel)...)
					return
				}
			}
			tflog.Error(ctx, err.Error())
			resp.Diagnostics.AddError("Error creating workflowGroup", "Error in creating workflowGroup API call: "+err.Error())
			return
		}
		responseData = reqResp.Data
	}

	workflowGroupModel, diags := BuildAPIModelToWorkflowGroupModel(responseData)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, &workflowGroupModel)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *workflowGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state WorkflowGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed state from client
	workflowGroup, err := r.client.WorkflowGroups.ReadWorkflowGroup(ctx, r.org_name, state.Id.ValueString())
	if err != nil {
		// If a managed resource is no longer found then remove it from the state
		if apiErr, ok := err.(*core.APIError); ok {
			if apiErr.StatusCode == 404 {
				resp.State.RemoveResource(ctx)
				return
			}
		}
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error reading workflowGroup", "Could not read workflowGroup "+state.ResourceName.ValueString()+": "+err.Error())
		return
	}

	//For cases where WFG is a nested one, the resource name is returned as the full path to match the resource_name in resource definition
	fullResourceId := strings.Replace(*workflowGroup.Msg.SubResourceId, "/wfgrps/", "", 1)
	workflowGroup.Msg.Id = fullResourceId

	workflowGroupResourceModel, diags := BuildAPIModelToWorkflowGroupModel(workflowGroup.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, workflowGroupResourceModel)...)

}

// Update updates the resource and sets the updated Terraform state on success.
func (r *workflowGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WorkflowGroupResourceModel
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

	_, err := r.client.WorkflowGroups.UpdateWorkflowGroup(ctx, r.org_name, plan.Id.ValueString(), payload)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error updating workflowGroup", "Error in updating workflowGroup "+
			plan.ResourceName.ValueString()+": "+err.Error())
		return
	}

	// Call read to get the updated WFG resource to set the state
	updatedWorkflowGroup, err := r.client.WorkflowGroups.ReadWorkflowGroup(ctx, r.org_name, plan.Id.ValueString())
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error reading the updated state of workflowGroup",
			"Could not read the updated state of workflowGroup "+plan.ResourceName.ValueString()+": "+err.Error())
		return
	}
	//For cases where WFG is a nested one, the resource name is returned as the full path to match the resource_name in resource definition
	fullResourceId := strings.Replace(*updatedWorkflowGroup.Msg.SubResourceId, "/wfgrps/", "", 1)
	updatedWorkflowGroup.Msg.Id = fullResourceId

	workflowGroupResourceModel, diags := BuildAPIModelToWorkflowGroupModel(updatedWorkflowGroup.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, workflowGroupResourceModel)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *workflowGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WorkflowGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.WorkflowGroups.DeleteWorkflowGroup(ctx, r.org_name, state.Id.ValueString())
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error deleting workflowGroup", "Error in deleting workflowGroup "+state.ResourceName.ValueString()+": "+err.Error())
		return
	}
}

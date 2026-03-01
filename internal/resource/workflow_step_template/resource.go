package workflowsteptemplate

import (
	"context"
	"fmt"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	core "github.com/StackGuardian/sg-sdk-go/core"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &workflowStepTemplateResource{}
	_ resource.ResourceWithConfigure   = &workflowStepTemplateResource{}
	_ resource.ResourceWithImportState = &workflowStepTemplateResource{}
)

type workflowStepTemplateResource struct {
	client   *sgclient.Client
	org_name string
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &workflowStepTemplateResource{}
}

// Metadata returns the resource type name.
func (r *workflowStepTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow_step_template"
}

// Configure adds the provider configured client to the resource.
func (r *workflowStepTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ImportState imports a workflow step template using its ID.
func (r *workflowStepTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *workflowStepTemplateResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if !req.State.Raw.IsNull() && !req.Plan.Raw.IsNull() {
		var state WorkflowStepTemplateResourceModel
		var plan WorkflowStepTemplateResourceModel

		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if !plan.Id.Equal(state.Id) {
			resp.RequiresReplace = append(resp.RequiresReplace, path.Root("template_name"))
		}
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *workflowStepTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WorkflowStepTemplateResourceModel

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

	payload.OwnerOrg = fmt.Sprintf("/orgs/%v", r.org_name)

	createResp, err := r.client.WorkflowStepTemplate.CreateWorkflowStepTemplate(ctx, r.org_name, false, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating workflow step template", "Error in creating workflow step template API call: "+err.Error())
		return
	}

	readResp, err := r.client.WorkflowStepTemplate.ReadWorkflowStepTemplate(ctx, r.org_name, createResp.Data.Parent.Id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading workflow step template after create", "Error in reading workflow step template API call: "+err.Error())
		return
	}

	templateModel, diags := BuildAPIModelToWorkflowStepTemplateModel(&readResp.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, &templateModel)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *workflowStepTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WorkflowStepTemplateResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	templateId := state.Id.ValueString()
	if templateId == "" {
		resp.Diagnostics.AddError("Error reading workflow step template", "Template ID is empty")
		return
	}

	readResp, err := r.client.WorkflowStepTemplate.ReadWorkflowStepTemplate(ctx, r.org_name, templateId)
	if err != nil {
		if apiErr, ok := err.(*core.APIError); ok {
			if apiErr.StatusCode == 404 {
				tflog.Warn(ctx, "Workflow step template not found, removing from state")
				resp.State.RemoveResource(ctx)
				return
			}
		}

		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error reading workflow step template", "Error in reading workflow step template API call: "+err.Error())
		return
	}

	if readResp == nil {
		resp.Diagnostics.AddError("Error reading workflow step template", "API response is empty")
		return
	}

	templateModel, diags := BuildAPIModelToWorkflowStepTemplateModel(&readResp.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &templateModel)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *workflowStepTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WorkflowStepTemplateResourceModel
	var state WorkflowStepTemplateResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	templateId := state.Id.ValueString()
	if templateId == "" {
		resp.Diagnostics.AddError("Error updating workflow step template", "Template ID is empty")
		return
	}

	payload, diags := plan.ToPatchedAPIModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload.OwnerOrg = sgsdkgo.Optional(fmt.Sprintf("/orgs/%v", r.org_name))

	_, err := r.client.WorkflowStepTemplate.UpdateWorkflowStepTemplate(ctx, r.org_name, templateId, payload)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error updating workflow step template", "Error in updating workflow step template API call: "+err.Error())
		return
	}

	// Read back the template to get all attributes
	readResp, err := r.client.WorkflowStepTemplate.ReadWorkflowStepTemplate(ctx, r.org_name, templateId)
	if err != nil {
		resp.Diagnostics.AddError("Error reading workflow step template after update", "Error in reading workflow step template API call: "+err.Error())
		return
	}

	templateModel, diags := BuildAPIModelToWorkflowStepTemplateModel(&readResp.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &templateModel)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *workflowStepTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WorkflowStepTemplateResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	templateId := state.Id.ValueString()
	if templateId == "" {
		resp.Diagnostics.AddError("Error deleting workflow step template", "Template ID is empty")
		return
	}

	err := r.client.WorkflowStepTemplate.DeleteWorkflowStepTemplate(ctx, r.org_name, templateId)
	if err != nil {
		if apiErr, ok := err.(*core.APIError); ok {
			if apiErr.StatusCode == 404 {
				// Resource already deleted
				return
			}
		}

		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error deleting workflow step template", "Error in deleting workflow step template API call: "+err.Error())
		return
	}
}

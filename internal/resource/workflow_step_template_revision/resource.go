package workflowsteptemplaterevision

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
	_ resource.Resource                = &workflowStepTemplateRevisionResource{}
	_ resource.ResourceWithConfigure   = &workflowStepTemplateRevisionResource{}
	_ resource.ResourceWithImportState = &workflowStepTemplateRevisionResource{}
)

type workflowStepTemplateRevisionResource struct {
	client   *sgclient.Client
	org_name string
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &workflowStepTemplateRevisionResource{}
}

// Metadata returns the resource type name.
func (r *workflowStepTemplateRevisionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow_step_template_revision"
}

// Configure adds the provider configured client to the resource.
func (r *workflowStepTemplateRevisionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ImportState imports a workflow step template revision using its ID (format: templateId:revisionNumber).
func (r *workflowStepTemplateRevisionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *workflowStepTemplateRevisionResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if !req.State.Raw.IsNull() && !req.Plan.Raw.IsNull() {
		var plan WorkflowStepTemplateRevisionResourceModel
		var state WorkflowStepTemplateRevisionResourceModel

		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if !plan.Id.Equal(state.Id) {
			resp.RequiresReplace = append(resp.RequiresReplace, path.Root("id"))
			return
		}
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *workflowStepTemplateRevisionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WorkflowStepTemplateRevisionResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	templateId := plan.TemplateId.ValueString()
	if templateId == "" {
		resp.Diagnostics.AddError("Error creating workflow step template revision", "Template ID is required")
		return
	}

	payload, diags := plan.ToAPIModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload.OwnerOrg = fmt.Sprintf("/orgs/%v", r.org_name)

	createResp, err := r.client.WorkflowStepTemplateRevision.CreateWorkflowStepTemplateRevision(ctx, r.org_name, templateId, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating workflow step template revision", "Error in creating workflow step template revision API call: "+err.Error())
		return
	}

	// Construct the revision ID from the create response
	revisionId := createResp.Data.Revision.Id

	// Read back the revision to get all attributes
	readResp, err := r.client.WorkflowStepTemplateRevision.ReadWorkflowStepTemplateRevision(ctx, r.org_name, revisionId)
	if err != nil {
		resp.Diagnostics.AddError("Error reading workflow step template revision after create", "Error in reading workflow step template revision API call: "+err.Error())
		return
	}

	revisionModel, diags := BuildAPIModelToRevisionModel(readResp.Msg, revisionId, templateId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &revisionModel)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *workflowStepTemplateRevisionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WorkflowStepTemplateRevisionResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	revisionId := state.Id.ValueString()
	if revisionId == "" {
		resp.Diagnostics.AddError("Error reading workflow step template revision", "Revision ID is empty")
		return
	}

	templateId := state.TemplateId.ValueString()

	readResp, err := r.client.WorkflowStepTemplateRevision.ReadWorkflowStepTemplateRevision(ctx, r.org_name, revisionId)
	if err != nil {
		resp.Diagnostics.AddError("Error reading workflow step template revision", "Error in reading workflow step template revision API call: "+err.Error())
		return
	}

	revisionModel, diags := BuildAPIModelToRevisionModel(readResp.Msg, revisionId, templateId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &revisionModel)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *workflowStepTemplateRevisionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WorkflowStepTemplateRevisionResourceModel
	var state WorkflowStepTemplateRevisionResourceModel

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

	revisionId := state.Id.ValueString()
	if revisionId == "" {
		resp.Diagnostics.AddError("Error updating workflow step template revision", "Revision ID is empty")
		return
	}

	templateId := state.TemplateId.ValueString()

	payload, diags := plan.ToPatchedAPIModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload.OwnerOrg = fmt.Sprintf("/orgs/%v", r.org_name)

	_, err := r.client.WorkflowStepTemplateRevision.UpdateWorkflowStepTemplateRevision(ctx, r.org_name, revisionId, payload)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error updating workflow step template revision", "Error in updating workflow step template revision API call: "+err.Error())
		return
	}

	// Read back the revision to get all attributes
	readResp, err := r.client.WorkflowStepTemplateRevision.ReadWorkflowStepTemplateRevision(ctx, r.org_name, revisionId)
	if err != nil {
		resp.Diagnostics.AddError("Error reading workflow step template revision after update", "Error in reading workflow step template revision API call: "+err.Error())
		return
	}

	if readResp == nil || readResp.Msg == nil {
		resp.Diagnostics.AddError("Error reading workflow step template revision after update", "API response is empty")
		return
	}

	revisionModel, diags := BuildAPIModelToRevisionModel(readResp.Msg, revisionId, templateId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &revisionModel)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *workflowStepTemplateRevisionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WorkflowStepTemplateRevisionResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	revisionId := state.Id.ValueString()
	if revisionId == "" {
		resp.Diagnostics.AddError("Error deleting workflow step template revision", "Revision ID is empty")
		return
	}

	err := r.client.WorkflowStepTemplateRevision.DeleteWorkflowStepTemplateRevision(ctx, r.org_name, revisionId, true)
	if err != nil {
		if apiErr, ok := err.(*core.APIError); ok {
			if apiErr.StatusCode == 404 {
				// Resource already deleted
				return
			}
		}

		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error deleting workflow step template revision", "Error in deleting workflow step template revision API call: "+err.Error())
		return
	}
}

package workflowtemplaterevision

import (
	"context"
	"fmt"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &workflowTemplateRevisionResource{}
	_ resource.ResourceWithConfigure   = &workflowTemplateRevisionResource{}
	_ resource.ResourceWithImportState = &workflowTemplateRevisionResource{}
)

type workflowTemplateRevisionResource struct {
	client   *sgclient.Client
	org_name string
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &workflowTemplateRevisionResource{}
}

// Metadata returns the resource type name.
func (r *workflowTemplateRevisionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow_template_revision"
}

// Configure adds the provider configured client to the resource.
func (r *workflowTemplateRevisionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ImportState imports a workflow template revision using its ID.
func (r *workflowTemplateRevisionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// Create creates the resource and sets the initial Terraform state.
func (r *workflowTemplateRevisionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WorkflowTemplateRevisionResourceModel

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

	templateID := plan.TemplateId.ValueString()

	payload.OwnerOrg = fmt.Sprintf("/orgs/%s", r.org_name)

	createResp, err := r.client.WorkflowTemplatesRevisions.CreateWorkflowTemplateRevision(ctx, r.org_name, templateID, payload)
	if err != nil {
		resp.Diagnostics.AddError("failed to create template revision", err.Error())
		return
	}

	// Set the ID from the create response
	revisionID := createResp.Data.Revision.Id

	// Call read to get the full state since create response doesn't return all values
	readResp, err := r.client.WorkflowTemplatesRevisions.ReadWorkflowTemplateRevision(ctx, r.org_name, revisionID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading created workflow template revision", "Could not read the created workflow template revision: "+err.Error())
		return
	}

	revisionModel, diags := BuildAPIModelToWorkflowTemplateRevisionModel(ctx, &readResp.Msg)
	revisionModel.TemplateId = plan.TemplateId
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, &revisionModel)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *workflowTemplateRevisionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WorkflowTemplateRevisionResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	revisionID := state.Id.ValueString()
	if revisionID == "" {
		resp.Diagnostics.AddError("Error reading workflow template revision", "Revision ID is empty")
		return
	}

	readResp, err := r.client.WorkflowTemplatesRevisions.ReadWorkflowTemplateRevision(ctx, r.org_name, revisionID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading workflow template revision", "Error in reading workflow template revision API call: "+err.Error())
		return
	}

	if readResp == nil {
		resp.Diagnostics.AddError("Error reading workflow template revision", "API response is empty")
		return
	}

	revisionModel, diags := BuildAPIModelToWorkflowTemplateRevisionModel(ctx, &readResp.Msg)
	revisionModel.TemplateId = state.TemplateId
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &revisionModel)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *workflowTemplateRevisionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WorkflowTemplateRevisionResourceModel
	var state WorkflowTemplateRevisionResourceModel

	diags := req.Plan.Get(ctx, &plan)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = req.State.Get(ctx, &state)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	revisionID := state.Id.ValueString()

	payload, diags := plan.ToUpdateAPIModel(ctx)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	ownerOrg := fmt.Sprintf("/orgs/%v", r.org_name)
	payload.OwnerOrg = sgsdkgo.Optional(ownerOrg)

	_, err := r.client.WorkflowTemplatesRevisions.UpdateWorkflowTemplateRevision(ctx, r.org_name, revisionID, payload)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error updating workflow template revision", "Error in updating workflow template revision API call: "+err.Error())
		return
	}

	// Call read to get the updated state since update response doesn't return all values
	readResp, err := r.client.WorkflowTemplatesRevisions.ReadWorkflowTemplateRevision(ctx, r.org_name, revisionID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading updated workflow template revision", "Could not read the updated workflow template revision: "+err.Error())
		return
	}

	revisionModel, diags := BuildAPIModelToWorkflowTemplateRevisionModel(ctx, &readResp.Msg)
	revisionModel.TemplateId = state.TemplateId
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &revisionModel)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *workflowTemplateRevisionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WorkflowTemplateRevisionResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	revisionID := state.Id.ValueString()

	if revisionID == "" {
		resp.Diagnostics.AddError("Error deleting workflow template revision", "Revision ID is empty")
		return
	}

	err := r.client.WorkflowTemplatesRevisions.DeleteWorkflowTemplateRevision(ctx, r.org_name, revisionID, true)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting workflow template revision", "Error in deleting workflow template revision API call: "+err.Error())
		return
	}
}

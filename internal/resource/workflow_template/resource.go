package workflowtemplate

import (
	"context"
	"fmt"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = &workflowTemplateResource{}
	_ resource.ResourceWithConfigure   = &workflowTemplateResource{}
	_ resource.ResourceWithImportState = &workflowTemplateResource{}
)

type workflowTemplateResource struct {
	client   *sgclient.Client
	org_name string
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &workflowTemplateResource{}
}

// Metadata returns the resource type name.
func (r *workflowTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow_template"
}

// Configure adds the provider configured client to the resource.
func (r *workflowTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ImportState imports a workflow template using its ID.
func (r *workflowTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// Create creates the resource and sets the initial Terraform state.
func (r *workflowTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WorkflowTemplateResourceModel

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
	payload.TemplateType = sgsdkgo.TemplateTypeEnum("IAC")

	createResp, err := r.client.WorkflowTemplates.CreateWorkflowTemplate(ctx, r.org_name, false, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating workflow template", "Error in creating workflow template API call: "+err.Error())
		return
	}

	// Set the ID from the create response
	templateID := *createResp.Data.Parent.Id

	// Call read to get the full state since create response doesn't return all values
	readResp, err := r.client.WorkflowTemplates.ReadWorkflowTemplate(ctx, r.org_name, templateID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading created workflow template", "Could not read the created workflow template: "+err.Error())
		return
	}

	if readResp == nil {
		resp.Diagnostics.AddError("Error reading workflow template", "API response is empty")
		return
	}

	templateModel, diags := BuildAPIModelToWorkflowTemplateModel(&readResp.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, &templateModel)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *workflowTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WorkflowTemplateResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	templateID := state.Id.ValueString()

	readResp, err := r.client.WorkflowTemplates.ReadWorkflowTemplate(ctx, r.org_name, templateID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading workflow template", "Error in reading workflow template API call: "+err.Error())
		return
	}

	templateModel, diags := BuildAPIModelToWorkflowTemplateModel(&readResp.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &templateModel)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *workflowTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WorkflowTemplateResourceModel
	var state WorkflowTemplateResourceModel

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

	templateID := state.Id.ValueString()

	payload, diags := plan.ToUpdateAPIModel(ctx)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	payload.OwnerOrg = sgsdkgo.Optional(r.org_name)

	_, err := r.client.WorkflowTemplates.UpdateWorkflowTemplate(ctx, r.org_name, templateID, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error updating workflow template", "Error in updating workflow template API call: "+err.Error())
		return
	}

	// Call read to get the updated state since update response doesn't return all values
	readResp, err := r.client.WorkflowTemplates.ReadWorkflowTemplate(ctx, r.org_name, templateID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading updated workflow template", "Could not read the updated workflow template: "+err.Error())
		return
	}

	templateModel, diags := BuildAPIModelToWorkflowTemplateModel(&readResp.Msg)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &templateModel)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *workflowTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WorkflowTemplateResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.WorkflowTemplates.DeleteWorkflowTemplate(ctx, r.org_name, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting workflow template", "Error in deleting workflow template API call: "+err.Error())
		return
	}
}

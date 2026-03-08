package stacktemplaterevision

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
	_ resource.Resource                = &stackTemplateRevisionResource{}
	_ resource.ResourceWithConfigure   = &stackTemplateRevisionResource{}
	_ resource.ResourceWithImportState = &stackTemplateRevisionResource{}
)

type stackTemplateRevisionResource struct {
	client   *sgclient.Client
	org_name string
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &stackTemplateRevisionResource{}
}

// Metadata returns the resource type name.
func (r *stackTemplateRevisionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_template_revision"
}

// Configure adds the provider configured client to the resource.
func (r *stackTemplateRevisionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ImportState imports a stack template revision using its ID.
func (r *stackTemplateRevisionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// Create creates the resource and sets the initial Terraform state.
func (r *stackTemplateRevisionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StackTemplateRevisionResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ownerOrg := fmt.Sprintf("/orgs/%s", r.org_name)

	payload, diags := plan.ToAPIModel(ctx, r.org_name)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	templateID := plan.ParentTemplateId.ValueString()
	payload.OwnerOrg = ownerOrg

	createResp, err := r.client.StackTemplateRevisions.CreateStackTemplateRevision(ctx, r.org_name, templateID, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating stack template revision", "Error in creating stack template revision API call: "+err.Error())
		return
	}

	revisionID := createResp.Data.Revision.Id

	// Call read to get the full state since create response doesn't return all values
	readResp, err := r.client.StackTemplateRevisions.ReadStackTemplateRevision(ctx, r.org_name, revisionID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading created stack template revision", "Could not read the created stack template revision: "+err.Error())
		return
	}

	revisionModel, diags := BuildAPIModelToStackTemplateRevisionModel(ctx, &readResp.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve the user-provided parent_template_id from the plan
	revisionModel.ParentTemplateId = plan.ParentTemplateId

	resp.Diagnostics.Append(resp.State.Set(ctx, &revisionModel)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *stackTemplateRevisionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state StackTemplateRevisionResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	revisionID := state.Id.ValueString()
	if revisionID == "" {
		resp.Diagnostics.AddError("Error reading stack template revision", "Revision ID is empty")
		return
	}

	readResp, err := r.client.StackTemplateRevisions.ReadStackTemplateRevision(ctx, r.org_name, revisionID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading stack template revision", "Error in reading stack template revision API call: "+err.Error())
		return
	}

	if readResp == nil {
		resp.Diagnostics.AddError("Error reading stack template revision", "API response is empty")
		return
	}

	revisionModel, diags := BuildAPIModelToStackTemplateRevisionModel(ctx, &readResp.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve parent_template_id from prior state
	revisionModel.ParentTemplateId = state.ParentTemplateId

	resp.Diagnostics.Append(resp.State.Set(ctx, &revisionModel)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *stackTemplateRevisionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan StackTemplateRevisionResourceModel
	var state StackTemplateRevisionResourceModel

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
	ownerOrg := fmt.Sprintf("/orgs/%s", r.org_name)

	payload, diags := plan.ToUpdateAPIModel(ctx, r.org_name)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	payload.OwnerOrg = sgsdkgo.Optional(ownerOrg)

	_, err := r.client.StackTemplateRevisions.UpdateStackTemplateRevision(ctx, r.org_name, revisionID, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error updating stack template revision", "Error in updating stack template revision API call: "+err.Error())
		return
	}

	// Call read to get the updated state since update response doesn't return all values
	readResp, err := r.client.StackTemplateRevisions.ReadStackTemplateRevision(ctx, r.org_name, revisionID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading updated stack template revision", "Could not read the updated stack template revision: "+err.Error())
		return
	}

	revisionModel, diags := BuildAPIModelToStackTemplateRevisionModel(ctx, &readResp.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve parent_template_id from plan
	revisionModel.ParentTemplateId = plan.ParentTemplateId

	resp.Diagnostics.Append(resp.State.Set(ctx, &revisionModel)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *stackTemplateRevisionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state StackTemplateRevisionResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	revisionID := state.Id.ValueString()
	if revisionID == "" {
		resp.Diagnostics.AddError("Error deleting stack template revision", "Revision ID is empty")
		return
	}

	err := r.client.StackTemplateRevisions.DeleteStackTemplateRevision(ctx, r.org_name, revisionID, true)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting stack template revision", "Error in deleting stack template revision API call: "+err.Error())
		return
	}
}

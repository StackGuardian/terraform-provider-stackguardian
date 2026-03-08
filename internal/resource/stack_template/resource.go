package stacktemplate

import (
	"context"
	"fmt"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                = &stackTemplateResource{}
	_ resource.ResourceWithConfigure   = &stackTemplateResource{}
	_ resource.ResourceWithImportState = &stackTemplateResource{}
)

type stackTemplateResource struct {
	client   *sgclient.Client
	org_name string
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &stackTemplateResource{}
}

// Metadata returns the resource type name.
func (r *stackTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack_template"
}

// Configure adds the provider configured client to the resource.
func (r *stackTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ImportState imports a stack template using its ID.
func (r *stackTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// Create creates the resource and sets the initial Terraform state.
func (r *stackTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StackTemplateResourceModel

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

	createResp, err := r.client.StackTemplates.CreateStackTemplate(ctx, r.org_name, false, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating stack template", "Error in creating stack template API call: "+err.Error())
		return
	}

	templateID := *createResp.Data.Parent.Id

	// Call read to get the full state since create response doesn't return all values
	readResp, err := r.client.StackTemplates.ReadStackTemplate(ctx, r.org_name, templateID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading created stack template", "Could not read the created stack template: "+err.Error())
		return
	}

	if readResp == nil {
		resp.Diagnostics.AddError("Error reading stack template", "API response is empty")
		return
	}

	templateModel, diags := BuildAPIModelToStackTemplateModel(&readResp.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &templateModel)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *stackTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state StackTemplateResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	templateID := state.Id.ValueString()

	readResp, err := r.client.StackTemplates.ReadStackTemplate(ctx, r.org_name, templateID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading stack template", "Error in reading stack template API call: "+err.Error())
		return
	}

	templateModel, diags := BuildAPIModelToStackTemplateModel(&readResp.Msg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &templateModel)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *stackTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan StackTemplateResourceModel
	var state StackTemplateResourceModel

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

	_, err := r.client.StackTemplates.UpdateStackTemplate(ctx, r.org_name, templateID, payload)
	if err != nil {
		resp.Diagnostics.AddError("Error updating stack template", "Error in updating stack template API call: "+err.Error())
		return
	}

	// Call read to get the updated state since update response doesn't return all values
	readResp, err := r.client.StackTemplates.ReadStackTemplate(ctx, r.org_name, templateID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading updated stack template", "Could not read the updated stack template: "+err.Error())
		return
	}

	templateModel, diags := BuildAPIModelToStackTemplateModel(&readResp.Msg)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &templateModel)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *stackTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state StackTemplateResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.StackTemplates.DeleteStackTemplate(ctx, r.org_name, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting stack template", "Error in deleting stack template API call: "+err.Error())
		return
	}
}

package workflow

import (
	"context"
	"fmt"
	"strings"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &workflowResource{}
	_ resource.ResourceWithConfigure   = &workflowResource{}
	_ resource.ResourceWithImportState = &workflowResource{}
)

type workflowResource struct {
	client   *sgclient.Client
	org_name string
}

func NewResource() resource.Resource {
	return &workflowResource{}
}

func (r *workflowResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow"
}

func (r *workflowResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(*customTypes.ProviderInfo)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *customTypes.ProviderInfo, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = provider.Client
	r.org_name = provider.Org_name
}

func (r *workflowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workflowResourceModel

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

	createResp, err := r.client.Workflows.CreateWorkflow(context.TODO(), r.org_name, plan.WorkflowGroupId.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating workflow", "Error in creating workflow API call: "+err.Error())
		return
	}

	id := createResp.Data.Id

	workflow, err := r.client.Workflows.ReadWorkflow(ctx, r.org_name, id, plan.WorkflowGroupId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading create workflow", "Could not read the created workflow"+err.Error())
		return
	}

	workflowTerraModel, diags := convertWorkflowFromAPI(ctx, workflow, plan.WorkflowGroupId.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	workflowTerraModel.WorkflowGroupId = plan.WorkflowGroupId

	resp.Diagnostics.Append(resp.State.Set(ctx, &workflowTerraModel)...)
}

func (r *workflowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workflowResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	readResp, err := r.client.Workflows.ReadWorkflow(ctx, r.org_name, state.Id.ValueString(), state.WorkflowGroupId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading workflow", "Error in reading workflow API call: "+err.Error())
		return
	}

	workflowTerraModel, diags := convertWorkflowFromAPI(ctx, readResp, state.WorkflowGroupId.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	workflowTerraModel.WorkflowGroupId = state.WorkflowGroupId

	resp.Diagnostics.Append(resp.State.Set(ctx, &workflowTerraModel)...)
}

func (r *workflowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan workflowResourceModel
	var state workflowResourceModel

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

	id := state.Id.ValueString()
	workflowGroupId := state.WorkflowGroupId.ValueString()

	payload, diags := plan.ToUpdateAPIModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Workflows.UpdateWorkflow(ctx, r.org_name, id, workflowGroupId, payload)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error updating workflow", "Error in updating workflow API call: "+err.Error())
		return
	}

	readResp, err := r.client.Workflows.ReadWorkflow(ctx, r.org_name, id, workflowGroupId)
	if err != nil {
		resp.Diagnostics.AddError("Error reading updated workflow", "Could not read the updated workflow: "+err.Error())
		return
	}

	workflowTerraModel, diags := convertWorkflowFromAPI(ctx, readResp, workflowGroupId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	workflowTerraModel.WorkflowGroupId = state.WorkflowGroupId

	resp.Diagnostics.Append(resp.State.Set(ctx, &workflowTerraModel)...)
}

func (r *workflowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state workflowResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Workflows.DeleteWorkflow(ctx, r.org_name, state.Id.ValueString(), state.WorkflowGroupId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting workflow", "Error in deleting workflow API call: "+err.Error())
		return
	}
}

func (r *workflowResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Expected import ID format: workflow_group_id/workflow_id
	parts := strings.Split(req.ID, "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected format: workflow_group_id/workflow_id, got: %q", req.ID),
		)
		return
	}

	workflowId := parts[len(parts)-1]
	workflowGroupId := strings.Join(parts[0:len(parts)-1], "/")

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workflow_group_id"), workflowGroupId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), workflowId)...)
}

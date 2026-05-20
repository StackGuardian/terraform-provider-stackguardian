package workflowgit

import (
	"context"
	"fmt"
	"strings"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	sgworkflows "github.com/StackGuardian/sg-sdk-go/workflows"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/expanders"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &workflowGitResource{}
	_ resource.ResourceWithConfigure   = &workflowGitResource{}
	_ resource.ResourceWithImportState = &workflowGitResource{}
)

type workflowGitResource struct {
	client   *sgclient.Client
	org_name string
}

func NewResource() resource.Resource {
	return &workflowGitResource{}
}

func (r *workflowGitResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow_git"
}

func (r *workflowGitResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *workflowGitResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

func (r *workflowGitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WorkflowGitResourceModel

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
	payload.VcsConfig.IacVcsConfig.UseMarketplaceTemplate = expanders.BoolPtr(false)

	createResp, err := r.client.Workflows.CreateWorkflow(context.TODO(), r.org_name, plan.WorkflowGroupId.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating workflow_git", "Error in creating workflow_git API call: "+err.Error())
		return
	}

	id := createResp.Data.Id

	if !plan.VcsTriggers.IsNull() && !plan.VcsTriggers.IsUnknown() && payload.VcsTriggers != nil {
		vcsTriggerReq := &sgworkflows.CreateVcsTriggersRequest{
			VcsTriggers: payload.VcsTriggers,
		}
		if payload.VcsConfig != nil {
			vcsTriggerReq.VcsConfig = payload.VcsConfig
		}

		_, err := r.client.Workflows.CreateVcsTriggers(ctx, r.org_name, plan.WorkflowGroupId.ValueString(), id, vcsTriggerReq)
		if err != nil {
			r.client.Workflows.DeleteWorkflow(ctx, r.org_name, id, plan.WorkflowGroupId.ValueString())
			resp.Diagnostics.AddError(
				"Error creating vcs_triggers for workflow_git",
				"VCS trigger registration failed: "+err.Error()+". The workflow was deleted to avoid leaving orphaned resources.",
			)
			return
		}
	}

	readResp, err := r.client.Workflows.ReadWorkflow(ctx, r.org_name, id, plan.WorkflowGroupId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading created workflow_git", "Could not read the created workflow_git: "+err.Error())
		return
	}

	model, diags := ConvertWorkflowGitFromAPI(ctx, readResp)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.WorkflowGroupId = plan.WorkflowGroupId

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *workflowGitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WorkflowGitResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	readResp, err := r.client.Workflows.ReadWorkflow(ctx, r.org_name, state.Id.ValueString(), state.WorkflowGroupId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading workflow_git", "Error in reading workflow_git API call: "+err.Error())
		return
	}

	model, diags := ConvertWorkflowGitFromAPI(ctx, readResp)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.WorkflowGroupId = state.WorkflowGroupId

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *workflowGitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WorkflowGitResourceModel
	var state WorkflowGitResourceModel

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
	payload.VcsConfig.Value.IacVcsConfig.UseMarketplaceTemplate = expanders.BoolPtr(false)

	_, err := r.client.Workflows.UpdateWorkflow(ctx, r.org_name, id, workflowGroupId, sgworkflows.UpgradeModeEnumPreserveSettings.Ptr(), payload)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error updating workflow_git", "Error in updating workflow_git API call: "+err.Error())
		return
	}

	if !plan.VcsTriggers.IsNull() && !plan.VcsTriggers.IsUnknown() {
		createPayload, diags := plan.ToAPIModel(ctx)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		createPayload.VcsConfig.IacVcsConfig.UseMarketplaceTemplate = expanders.BoolPtr(false)
		if createPayload.VcsTriggers != nil {
			vcsTriggerReq := &sgworkflows.CreateVcsTriggersRequest{
				VcsTriggers: createPayload.VcsTriggers,
			}
			if createPayload.VcsConfig != nil {
				vcsTriggerReq.VcsConfig = createPayload.VcsConfig
			}
			_, err := r.client.Workflows.CreateVcsTriggers(ctx, r.org_name, workflowGroupId, id, vcsTriggerReq)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error updating vcs_triggers for workflow_git",
					"VCS trigger update failed: "+err.Error(),
				)
			}
		}
	}

	readResp, err := r.client.Workflows.ReadWorkflow(ctx, r.org_name, id, workflowGroupId)
	if err != nil {
		resp.Diagnostics.AddError("Error reading updated workflow_git", "Could not read the updated workflow_git: "+err.Error())
		return
	}

	model, diags := ConvertWorkflowGitFromAPI(ctx, readResp)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.WorkflowGroupId = state.WorkflowGroupId

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *workflowGitResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WorkflowGitResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Workflows.DeleteWorkflow(ctx, r.org_name, state.Id.ValueString(), state.WorkflowGroupId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting workflow_git", "Error in deleting workflow_git API call: "+err.Error())
		return
	}
}

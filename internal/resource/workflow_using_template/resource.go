package workflowusingtemplate

import (
	"context"
	"fmt"
	"strings"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	sgworkflows "github.com/StackGuardian/sg-sdk-go/workflows"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &workflowUsingTemplateResource{}
	_ resource.ResourceWithConfigure   = &workflowUsingTemplateResource{}
	_ resource.ResourceWithImportState = &workflowUsingTemplateResource{}
)

type workflowUsingTemplateResource struct {
	client   *sgclient.Client
	org_name string
}

func NewResource() resource.Resource {
	return &workflowUsingTemplateResource{}
}

func (r *workflowUsingTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workflow_using_template"
}

func (r *workflowUsingTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *workflowUsingTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

func (r *workflowUsingTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WorkflowUsingTemplateResourceModel

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

	createResp, err := r.client.Workflows.CreateWorkflow(ctx, r.org_name, plan.WorkflowGroupId.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating workflow_using_template", "Error in creating workflow_using_template API call: "+err.Error())
		return
	}

	id := createResp.Data.Id

	readResp, err := r.client.Workflows.ReadWorkflow(ctx, r.org_name, id, plan.WorkflowGroupId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading created workflow_using_template", "Could not read the created workflow_using_template: "+err.Error())
		return
	}

	model, diags := ConvertWorkflowUsingTemplateFromAPI(ctx, readResp)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.WorkflowGroupId = plan.WorkflowGroupId
	model.ResourceName = plan.ResourceName
	model.Description = plan.Description
	model.EnvironmentVariables = plan.EnvironmentVariables
	model.MiniSteps = plan.MiniSteps
	model.RunnerConstraints = plan.RunnerConstraints
	model.Tags = plan.Tags
	model.UserSchedules = plan.UserSchedules
	model.ContextTags = plan.ContextTags
	model.Approvers = plan.Approvers
	model.NumberOfApprovalsRequired = plan.NumberOfApprovalsRequired
	model.UserJobCpu = plan.UserJobCpu
	model.UserJobMemory = plan.UserJobMemory
	model.TerraformConfig = plan.TerraformConfig
	model.DeploymentPlatformConfig = plan.DeploymentPlatformConfig
	model.WfStepsConfig = plan.WfStepsConfig

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *workflowUsingTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WorkflowUsingTemplateResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	readResp, err := r.client.Workflows.ReadWorkflow(ctx, r.org_name, state.Id.ValueString(), state.WorkflowGroupId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading workflow_using_template", "Error in reading workflow_using_template API call: "+err.Error())
		return
	}

	model, diags := ConvertWorkflowUsingTemplateFromAPI(ctx, readResp)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.WorkflowGroupId = state.WorkflowGroupId
	model.ResourceName = state.ResourceName
	model.Description = state.Description
	model.EnvironmentVariables = state.EnvironmentVariables
	model.MiniSteps = state.MiniSteps
	model.RunnerConstraints = state.RunnerConstraints
	model.Tags = state.Tags
	model.UserSchedules = state.UserSchedules
	model.ContextTags = state.ContextTags
	model.Approvers = state.Approvers
	model.NumberOfApprovalsRequired = state.NumberOfApprovalsRequired
	model.UserJobCpu = state.UserJobCpu
	model.UserJobMemory = state.UserJobMemory
	model.TerraformConfig = state.TerraformConfig
	model.DeploymentPlatformConfig = state.DeploymentPlatformConfig
	model.WfStepsConfig = state.WfStepsConfig

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *workflowUsingTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WorkflowUsingTemplateResourceModel
	var state WorkflowUsingTemplateResourceModel

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

	originalPlan := plan

	if !state.ResolvedSchema.IsNull() && !state.ResolvedSchema.IsUnknown() {
		var resolvedSchema ResolvedSchemaModel
		diags = state.ResolvedSchema.As(ctx, &resolvedSchema, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if plan.ResourceName.IsNull() || plan.ResourceName.IsUnknown() {
			originalPlan.ResourceName = resolvedSchema.ResourceName
		}
		if plan.RunnerConstraints.IsNull() || plan.RunnerConstraints.IsUnknown() {
			originalPlan.RunnerConstraints = resolvedSchema.RunnerConstraints
		}
		if plan.UserJobCpu.IsNull() || plan.UserJobCpu.IsUnknown() {
			originalPlan.UserJobCpu = resolvedSchema.UserJobCpu
		}
		if plan.UserJobMemory.IsNull() || plan.UserJobMemory.IsUnknown() {
			originalPlan.UserJobMemory = resolvedSchema.UserJobMemory
		}
	}

	payload, diags := originalPlan.ToUpdateAPIModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Workflows.UpdateWorkflow(ctx, r.org_name, id, workflowGroupId, sgworkflows.UpgradeModeEnumPreserveSettings.Ptr(), payload)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error updating workflow_using_template", "Error in updating workflow_using_template API call: "+err.Error())
		return
	}

	readResp, err := r.client.Workflows.ReadWorkflow(ctx, r.org_name, id, workflowGroupId)
	if err != nil {
		resp.Diagnostics.AddError("Error reading updated workflow_using_template", "Could not read the updated workflow_using_template: "+err.Error())
		return
	}

	model, diags := ConvertWorkflowUsingTemplateFromAPI(ctx, readResp)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.WorkflowGroupId = state.WorkflowGroupId
	model.ResourceName = plan.ResourceName
	model.Description = plan.Description
	model.EnvironmentVariables = plan.EnvironmentVariables
	model.MiniSteps = plan.MiniSteps
	model.RunnerConstraints = plan.RunnerConstraints
	model.Tags = plan.Tags
	model.UserSchedules = plan.UserSchedules
	model.ContextTags = plan.ContextTags
	model.Approvers = plan.Approvers
	model.NumberOfApprovalsRequired = plan.NumberOfApprovalsRequired
	model.UserJobCpu = plan.UserJobCpu
	model.UserJobMemory = plan.UserJobMemory
	model.TerraformConfig = plan.TerraformConfig
	model.DeploymentPlatformConfig = plan.DeploymentPlatformConfig
	model.WfStepsConfig = plan.WfStepsConfig

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *workflowUsingTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WorkflowUsingTemplateResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Workflows.DeleteWorkflow(ctx, r.org_name, state.Id.ValueString(), state.WorkflowGroupId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting workflow_using_template", "Error in deleting workflow_using_template API call: "+err.Error())
		return
	}
}

package workflowfromtemplate

import (
	"context"
	"fmt"
	"strings"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/sg-sdk-go/workflowtemplaterevisions"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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
	resp.TypeName = req.ProviderTypeName + "_workflow_from_template"
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

// ModifyPlan handles a template revision change. When iac_template_id changes, the
// Optional+Computed fields the user did NOT declare must be re-resolved against the new
// revision. Without this, UseStateForUnknown carries the OLD revision's values forward
// (they were never unknown), so the merge never runs for them. Setting those fields to
// unknown here lets Create/Update's merge fill them from the new revision at apply.
// Fields the user declared in config are left untouched (their value always wins).
func (r *workflowUsingTemplateResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return // create or destroy — no revision transition to handle
	}

	var plan, state, config WorkflowUsingTemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planTpl, d := plan.IacTemplateId(ctx)
	resp.Diagnostics.Append(d...)
	stateTpl, d := state.IacTemplateId(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() || planTpl == stateTpl || planTpl == "" {
		return // no revision change
	}

	// Revision changed: for each template-resolved field the user left null in config,
	// reset the planned value to unknown so it re-resolves against the new revision.
	resp.Diagnostics.Append(resetUnsetComputedToUnknown(ctx, &plan, config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

// fetchTemplateRevision reads the workflow template revision referenced by the
// model's vcs_config.iac_vcs_config.iac_template_id. It returns nil (without an
// error) when no template id is set, so merging is a no-op in that case.
func (r *workflowUsingTemplateResource) fetchTemplateRevision(ctx context.Context, m WorkflowUsingTemplateResourceModel) (*workflowtemplaterevisions.ReadWorkflowTemplateRevisionModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	templateId, d := m.IacTemplateId(ctx)
	diags.Append(d...)
	if diags.HasError() || templateId == "" {
		return nil, diags
	}

	// iac_template_id is stored fully-qualified ("/<org>/<name>:<rev>"), but
	// ReadWorkflowTemplateRevision builds its URL as
	// /templatetypes/IAC/<org>/<revisionId>/ and supplies the org separately. Pass
	// only the bare "<name>:<rev>" so the org is not duplicated in the path.
	revisionId := strings.TrimPrefix(templateId, fmt.Sprintf("/%s/", r.org_name))

	readResp, err := r.client.WorkflowTemplatesRevisions.ReadWorkflowTemplateRevision(ctx, r.org_name, revisionId)
	if err != nil {
		diags.AddError("Error reading workflow template revision",
			fmt.Sprintf("Could not read template revision %q to resolve workflow defaults: %s", templateId, err.Error()))
		return nil, diags
	}
	if readResp == nil {
		diags.AddError("Error reading workflow template revision",
			fmt.Sprintf("API returned an empty response for template revision %q", templateId))
		return nil, diags
	}
	return &readResp.Msg, diags
}

func (r *workflowUsingTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WorkflowUsingTemplateResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tpl, diags := r.fetchTemplateRevision(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := plan.ToAPIModel(ctx, tpl)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createResp, err := r.client.Workflows.CreateWorkflow(ctx, r.org_name, plan.WorkflowGroupId.ValueString(), payload)
	if err != nil {
		resp.Diagnostics.AddError("Error creating workflow_from_template", "Error in creating workflow_from_template API call: "+err.Error())
		return
	}

	id := createResp.Data.Id

	readResp, err := r.client.Workflows.ReadWorkflow(ctx, r.org_name, id, plan.WorkflowGroupId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading created workflow_from_template", "Could not read the created workflow_from_template: "+err.Error())
		return
	}

	model, diags := ConvertWorkflowUsingTemplateFromAPI(ctx, readResp, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

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
		resp.Diagnostics.AddError("Error reading workflow_from_template", "Error in reading workflow_from_template API call: "+err.Error())
		return
	}

	model, diags := ConvertWorkflowUsingTemplateFromAPI(ctx, readResp, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

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

	// Provider-side resolution: fetch the template revision and let ToUpdateAPIModel
	// fill any field the user did not declare with the template's value. This
	// replaces the former resolved_schema back-fill, which re-sent stale state and
	// masked drift.
	tpl, diags := r.fetchTemplateRevision(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := plan.ToUpdateAPIModel(ctx, tpl)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// No upgrade mode: the provider already resolved the full config (user + template
	// merge, including re-resolution on revision change via ModifyPlan), so the payload
	// is authoritative. Passing nil makes the API apply it directly instead of running
	// its own preserve/reset logic, which would otherwise keep stale fields (e.g. old
	// env vars) on a revision change.
	_, err := r.client.Workflows.UpdateWorkflow(ctx, r.org_name, id, workflowGroupId, nil, payload)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error updating workflow_from_template", "Error in updating workflow_from_template API call: "+err.Error())
		return
	}

	readResp, err := r.client.Workflows.ReadWorkflow(ctx, r.org_name, id, workflowGroupId)
	if err != nil {
		resp.Diagnostics.AddError("Error reading updated workflow_from_template", "Could not read the updated workflow_from_template: "+err.Error())
		return
	}

	model, diags := ConvertWorkflowUsingTemplateFromAPI(ctx, readResp, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.WorkflowGroupId = state.WorkflowGroupId

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
		resp.Diagnostics.AddError("Error deleting workflow_from_template", "Error in deleting workflow_from_template API call: "+err.Error())
		return
	}
}

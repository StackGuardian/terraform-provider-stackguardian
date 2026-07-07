package workflowfromtemplate

import (
	"context"
	"errors"
	"fmt"
	"strings"

	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	"github.com/StackGuardian/sg-sdk-go/core"
	"github.com/StackGuardian/sg-sdk-go/workflowtemplaterevisions"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

// isWorkflowNotFound reports whether err from ReadWorkflow indicates the workflow no longer
// exists (deleted out-of-band). The API returns either a 404, or a 400 whose body is
// {"msg":"Workflow does not exist"}; for the 400 case the message is required so a genuine
// bad request is not mistaken for a missing resource.
func isWorkflowNotFound(err error) bool {
	var apiErr *core.APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	if apiErr.StatusCode == 404 {
		return true
	}
	return apiErr.StatusCode == 400 && strings.Contains(err.Error(), "does not exist")
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

	// Enforce the drift_check/drift_cron coupling in the plan so plan == apply. A cron is
	// only meaningful when drift checking is on; the apply-time Read clears drift_cron to ""
	// when drift_check is false (see convertTerraformConfigFromAPI). When the user flips
	// drift_check to false but leaves drift_cron unset, UseStateForUnknown would otherwise
	// carry the old cron forward in the plan, which then mismatches the cleared apply value
	// ("inconsistent result after apply"). Predict the clear here by writing drift_cron = ""
	// directly onto the response plan. Runs on every update, independent of revision change.
	driftCheckPath := path.Root("terraform_config").AtName("drift_check")
	driftCronPath := path.Root("terraform_config").AtName("drift_cron")
	var plannedDriftCheck types.Bool
	resp.Diagnostics.Append(resp.Plan.GetAttribute(ctx, driftCheckPath, &plannedDriftCheck)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Only predict when drift_check is known-false; unknown/true can't be cleared here.
	if !plannedDriftCheck.IsNull() && !plannedDriftCheck.IsUnknown() && !plannedDriftCheck.ValueBool() {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, driftCronPath, types.StringValue(""))...)
		if resp.Diagnostics.HasError() {
			return
		}
		// Re-read the plan so the revision-change logic below sees the updated value.
		resp.Diagnostics.Append(resp.Plan.Get(ctx, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	planTpl, d := plan.IacTemplateId(ctx)
	resp.Diagnostics.Append(d...)
	stateTpl, d := state.IacTemplateId(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() || planTpl == stateTpl || planTpl == "" {
		return // no revision change
	}

	// Revision changed. Fetch the NEW revision and re-resolve the template-derived fields
	// the user did not declare, writing the concrete merged values into the plan. We must
	// set known values (not unknown) for terraform_config's nested fields: marking the
	// object unknown doesn't stop the nested UseStateForUnknown modifiers from carrying the
	// OLD revision's value (incl. a stale known-empty "") forward, which then mismatches the
	// new revision at apply ("inconsistent result"). Computing the value here makes
	// plan == apply.
	tpl, d := r.fetchTemplateRevision(ctx, plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() || tpl == nil {
		return
	}

	resp.Diagnostics.Append(reResolveOnRevisionChange(ctx, &plan, config, tpl)...)
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
	// ReadWorkflowTemplateRevision builds its URL as /templatetypes/IAC/<org>/<name>:<rev>/
	// (org also goes in the x-sg-orgid header). Parse the OWNING org out of the id prefix
	// rather than assuming the caller's org — a shared template belongs to another org, so
	// stripping only the caller's org would leave the foreign "/<owner>/" in the revisionId
	// and target the wrong org. Falls back to the caller's org if the id has no prefix.
	tplOrg, revisionId := splitTemplateOrg(templateId, r.org_name)

	readResp, err := r.client.WorkflowTemplatesRevisions.ReadWorkflowTemplateRevision(ctx, tplOrg, revisionId)
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

// splitTemplateOrg parses a fully-qualified template id "/<org>/<name>:<rev>" into the
// owning org and the bare "<name>:<rev>" revision id. The owning org may differ from the
// caller's org when the template is shared from another org, so it must drive the read.
// If id has no leading "/<org>/" prefix (already bare), the caller's org is used.
func splitTemplateOrg(id, callerOrg string) (org, revisionId string) {
	if rest, ok := strings.CutPrefix(id, "/"); ok {
		if owner, revID, ok := strings.Cut(rest, "/"); ok {
			return owner, revID
		}
	}
	return callerOrg, id
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
		// If the workflow was deleted out-of-band (e.g. from the platform UI), drop it
		// from state so Terraform plans a recreate instead of erroring forever. The API
		// signals a missing workflow with a 404, or a 400 whose body is
		// {"msg":"Workflow does not exist"} — match both, but for 400 require the message
		// so a genuine bad request is not silently swallowed.
		if isWorkflowNotFound(err) {
			tflog.Warn(ctx, "workflow_from_template not found; removing from state")
			resp.State.RemoveResource(ctx)
			return
		}
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
		// Already gone (deleted out-of-band) — treat as a successful delete so the resource
		// is removed from state instead of erroring.
		if isWorkflowNotFound(err) {
			tflog.Warn(ctx, "workflow_from_template already deleted; treating delete as success")
			return
		}
		resp.Diagnostics.AddError("Error deleting workflow_from_template", "Error in deleting workflow_from_template API call: "+err.Error())
		return
	}
}

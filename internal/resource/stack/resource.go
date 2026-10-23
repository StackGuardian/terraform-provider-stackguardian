package stack

import (
	"context"
	"fmt"
	"strings"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	core "github.com/StackGuardian/sg-sdk-go/core"
	"github.com/StackGuardian/sg-sdk-go/stacktemplaterevisions"
	"github.com/StackGuardian/sg-sdk-go/workflowtemplaterevisions"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/customTypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &stackResource{}
	_ resource.ResourceWithConfigure   = &stackResource{}
	_ resource.ResourceWithImportState = &stackResource{}
	_ resource.ResourceWithModifyPlan  = &stackResource{}
)

type stackResource struct {
	client   *sgclient.Client
	org_name string
}

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &stackResource{}
}

// Metadata returns the resource type name.
func (r *stackResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stack"
}

// Configure adds the provider configured client to the resource.
func (r *stackResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.org_name = provider.OrgName
}

// ImportState imports a stack using "workflow_group_id/id".
func (r *stackResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected format: workflow_group_id/id, got: %q", req.ID),
		)
		return
	}
	stackId := parts[len(parts)-1]
	workflowGroupId := strings.Join(parts[0:len(parts)-1], "/")
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workflow_group_id"), workflowGroupId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), stackId)...)
}

// isStackNotFound reports whether err from a Stack API call indicates the
// stack no longer exists (deleted out-of-band).
func isStackNotFound(err error) bool {
	apiErr, ok := err.(*core.APIError)
	return ok && apiErr.StatusCode == 404
}

// fetchTemplateRevision reads the stack template revision referenced by the
// model's template_group_id. It returns nil (without an error) when no
// template group id is set, so merging is a no-op in that case.
//
// template_group_id is the fully qualified "/<owner-org>/<template-id>:<revision>".
// A stack template can be owned by another organization (shared with this
// one, or public), so the revision is read from the org named in the id, not
// the stack's own org — the same way fetchWorkflowTemplateRevision reads
// workflow templates.
func (r *stackResource) fetchTemplateRevision(ctx context.Context, m StackResourceModel) (*stacktemplaterevisions.ReadStackTemplateRevisionModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	templateGroupId := m.TemplateGroupId.ValueString()
	if templateGroupId == "" {
		return nil, diags
	}

	tplOrg, revisionId := splitTemplateOrg(templateGroupId, r.org_name)
	readResp, err := r.client.StackTemplateRevisions.ReadStackTemplateRevision(ctx, tplOrg, revisionId)
	if err != nil {
		diags.AddError("Error reading stack template revision",
			fmt.Sprintf("Could not read template revision %q to resolve stack defaults: %s", templateGroupId, err.Error()))
		return nil, diags
	}
	return &readResp.Msg, diags
}

// fetchWorkflowTemplateRevision reads the workflow template revision referenced by
// templateId — a stack template revision workflow entry's
// vcs_config.iac_vcs_config.iac_template_id (never the stack's own, since that
// attribute is Computed-only and always inherited from the matching workflow
// entry on the stack template revision). Returns
// nil (without an error) when templateId is empty. Mirrors
// workflow_from_template's fetchTemplateRevision.
func (r *stackResource) fetchWorkflowTemplateRevision(ctx context.Context, templateId string) (*workflowtemplaterevisions.ReadWorkflowTemplateRevisionModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if templateId == "" {
		return nil, diags
	}

	// templateId is stored fully-qualified ("/<org>/<template-id>:<rev>"), but
	// ReadWorkflowTemplateRevision builds its URL as /templatetypes/IAC/<org>/<template-id>:<rev>/
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

// splitTemplateOrg parses a fully-qualified template id "/<org>/<template-id>:<rev>" into the
// owning org and the bare "<template-id>:<rev>" revision id. The owning org may differ from the
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

// resolveWorkflowTemplates fetches, for every workflow entry declared in
// plan.WorkflowsConfig, the workflow template revision that workflow resolves
// against — read from the MATCHING workflow entry on stackTpl (the stack
// template revision), never from the stack's own entry, since
// vcs_config.iac_vcs_config is Computed-only on the stack resource and is
// always inherited from the stack template. Entries are matched by id (the
// workflowUUID the stack template revision declares for the workflow). A
// workflow with no match on stackTpl, or whose matching entry has no
// iac_template_id, is simply omitted from the result — expandWorkflowsConfig
// then merges only the layers that exist for that workflow. The returned map
// is keyed by workflowUUID.
func (r *stackResource) resolveWorkflowTemplates(ctx context.Context, plan StackResourceModel, stackTpl *stacktemplaterevisions.ReadStackTemplateRevisionModel) (map[string]*workflowtemplaterevisions.ReadWorkflowTemplateRevisionModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	if plan.WorkflowsConfig.IsNull() || plan.WorkflowsConfig.IsUnknown() {
		return nil, diags
	}
	var wfc WorkflowsConfigModel
	if d := plan.WorkflowsConfig.As(ctx, &wfc, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}); d.HasError() {
		diags.Append(d...)
		return nil, diags
	}
	if wfc.Workflows.IsNull() || wfc.Workflows.IsUnknown() {
		return nil, diags
	}
	var wfModels []WorkflowInStackModel
	if d := wfc.Workflows.ElementsAs(ctx, &wfModels, false); d.HasError() {
		diags.Append(d...)
		return nil, diags
	}
	if len(wfModels) == 0 {
		return nil, diags
	}

	stackTplWorkflowsById := make(map[string]*stacktemplaterevisions.StackTemplateRevisionWorkflow)
	if stackTpl != nil && stackTpl.WorkflowsConfig != nil {
		for _, w := range stackTpl.WorkflowsConfig.Workflows {
			if w != nil && w.Id != nil {
				stackTplWorkflowsById[*w.Id] = w
			}
		}
	}

	result := make(map[string]*workflowtemplaterevisions.ReadWorkflowTemplateRevisionModel, len(wfModels))
	for _, wm := range wfModels {
		slotId := wm.Id.ValueString()
		matched := stackTplWorkflowsById[slotId]
		if matched == nil || matched.VcsConfig == nil || matched.VcsConfig.IacVcsConfig == nil {
			continue
		}
		templateId := matched.VcsConfig.IacVcsConfig.IacTemplateId
		if templateId == nil || *templateId == "" {
			continue
		}
		tpl, d := r.fetchWorkflowTemplateRevision(ctx, *templateId)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		if tpl != nil {
			result[slotId] = tpl
		}
	}

	return result, diags
}

// ModifyPlan does two things.
//
// First, on every create and update whose template_group_id is known, it
// checks workflows_config against the stack template revision
// (validateWorkflowsConfigMatchesRevision), so a missing or reordered workflow
// fails `terraform plan` instead of only `terraform apply`. A plan that
// succeeds would otherwise promise a change the apply then refuses, which
// matters for pipelines that review a plan before applying it. This costs one
// read of the revision per plan (Create()/Update() read it again to build the
// request); the check is the most likely mistake when writing a stack, so the
// earlier feedback is worth the extra read. Create() and Update() still run
// the same check, which is what catches it when template_group_id is unknown
// at plan time.
//
// Second, on a template_group_id change, the Optional+Computed fields the
// user did NOT declare must be re-resolved against the new revision. Without
// that re-resolution, UseStateForUnknown carries the OLD revision's values
// forward (they were never unknown), so the merge never runs for them.
// Setting those fields to concrete resolved values here lets plan == apply.
// Fields the user declared in config are left untouched. It covers
// description/tags/context_tags/actions (stack-level) and workflows_config's
// per-workflow template-derived fields. actions the user declared in config
// has no re-resolution to do, but is validated: a reference to a workflow the
// new revision dropped is an error, not a silent carry-forward.
func (r *stackResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return // destroy — nothing to validate or re-resolve
	}
	isCreate := req.State.Raw.IsNull()

	var plan, state, config StackResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if !isCreate {
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	}
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.TemplateGroupId.IsUnknown() {
		// The revision isn't known until apply; Create()/Update() run the
		// workflows_config check then, against the revision they fetch.
		return
	}

	tpl, d := r.fetchTemplateRevision(ctx, plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() || tpl == nil {
		return
	}

	resp.Diagnostics.Append(validateWorkflowsConfigMatchesRevision(ctx, plan.WorkflowsConfig, tpl)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if isCreate || plan.TemplateGroupId.ValueString() == state.TemplateGroupId.ValueString() {
		return // no revision change: nothing to re-resolve
	}

	resp.Diagnostics.Append(validateActionsAgainstRevision(ctx, config.Actions, plan.WorkflowsConfig, tpl)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workflowTemplates, d := r.resolveWorkflowTemplates(ctx, plan, tpl)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(reResolveOnRevisionChange(ctx, &plan, config, tpl)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(reResolveWorkflowsConfigOnRevisionChange(ctx, &plan, config, tpl, workflowTemplates)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

// Create creates the resource and sets the initial Terraform state.
func (r *stackResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StackResourceModel

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

	// ModifyPlan already ran this check at plan time whenever template_group_id
	// was known. It runs again here for when it wasn't (e.g. it references a
	// revision created in the same apply), using the tpl this function already
	// fetched to build the request.
	resp.Diagnostics.Append(validateWorkflowsConfigMatchesRevision(ctx, plan.WorkflowsConfig, tpl)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workflowTemplates, diags := r.resolveWorkflowTemplates(ctx, plan, tpl)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := plan.ToAPIModel(ctx, tpl, workflowTemplates)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	wfGrpId := plan.WorkflowGroupId.ValueString()

	createResp, err := r.client.Stacks.CreateStack(ctx, r.org_name, wfGrpId, payload)
	if err != nil {
		resp.Diagnostics.AddError("failed to create stack", err.Error())
		return
	}

	// Get the stack ID from create response (use SubResourceId as the primary identifier)
	stackID := createResp.Data.Stack.Id

	// Call read to get the full state since create response may not return all values
	readResp, err := r.client.Stacks.ReadStack(ctx, r.org_name, stackID, wfGrpId, &sgsdkgo.ReadStackQueryParams{RefreshWorkflowsConfig: sgsdkgo.Bool(true)})
	if err != nil {
		resp.Diagnostics.AddError("Error reading created stack", "Could not read the created stack: "+err.Error())
		return
	}

	stackModel, diags := BuildAPIModelToStackModel(ctx, readResp.Msg, plan.WorkflowGroupId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Creating the workflows rewrites ${reference::} values to ${workflow::}; keep
	// the planned form where that is the only difference (see references.go).
	resp.Diagnostics.Append(preserveReferenceForms(ctx, stackModel, plan.WorkflowsConfig)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, &stackModel)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *stackResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state StackResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stackId := state.Id.ValueString()
	wfGrpId := state.WorkflowGroupId.ValueString()

	// Get refreshed state from client
	readResp, err := r.client.Stacks.ReadStack(ctx, r.org_name, stackId, wfGrpId, &sgsdkgo.ReadStackQueryParams{RefreshWorkflowsConfig: sgsdkgo.Bool(true)})
	if err != nil {
		// If a managed resource is no longer found then remove it from the state
		if isStackNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error reading stack", "Could not read stack "+state.ResourceName.ValueString()+": "+err.Error())
		return
	}

	stackModel, diags := BuildAPIModelToStackModel(ctx, readResp.Msg, state.WorkflowGroupId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Keep the stored ${reference::} form where the platform holds core's
	// ${workflow::} rewrite of it, so a refresh doesn't report it as drift (see
	// references.go).
	resp.Diagnostics.Append(preserveReferenceForms(ctx, stackModel, state.WorkflowsConfig)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, stackModel)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *stackResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan StackResourceModel
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

	// ModifyPlan already ran this check at plan time whenever template_group_id
	// was known. It runs again here for when it wasn't (e.g. it references a
	// revision created in the same apply), using the tpl this function already
	// fetched to build the request.
	resp.Diagnostics.Append(validateWorkflowsConfigMatchesRevision(ctx, plan.WorkflowsConfig, tpl)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workflowTemplates, diags := r.resolveWorkflowTemplates(ctx, plan, tpl)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := plan.ToUpdateAPIModel(ctx, tpl, workflowTemplates)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stackId := plan.Id.ValueString()
	wfGrpId := plan.WorkflowGroupId.ValueString()

	_, err := r.client.Stacks.UpdateStack(ctx, r.org_name, stackId, wfGrpId, payload)
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error updating stack", "Error in updating stack "+
			plan.ResourceName.ValueString()+": "+err.Error())
		return
	}

	// Call read to get the updated stack resource to set the state
	updatedStack, err := r.client.Stacks.ReadStack(ctx, r.org_name, stackId, wfGrpId, &sgsdkgo.ReadStackQueryParams{RefreshWorkflowsConfig: sgsdkgo.Bool(true)})
	if err != nil {
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error reading the updated state of stack",
			"Could not read the updated state of stack "+plan.ResourceName.ValueString()+": "+err.Error())
		return
	}

	stackModel, diags := BuildAPIModelToStackModel(ctx, updatedStack.Msg, plan.WorkflowGroupId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Updates store ${reference::} values as sent, so this normally changes nothing;
	// it runs so the state never depends on whether the platform rewrote a value
	// (see references.go).
	resp.Diagnostics.Append(preserveReferenceForms(ctx, stackModel, plan.WorkflowsConfig)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, stackModel)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *stackResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state StackResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stackId := state.Id.ValueString()
	wfGrpId := state.WorkflowGroupId.ValueString()

	_, err := r.client.Stacks.DeleteStack(ctx, r.org_name, stackId, wfGrpId, &sgsdkgo.DeleteStackRequest{
		ForceDelete: sgsdkgo.Bool(true),
	})
	if err != nil {
		// Already gone (deleted out-of-band) — treat as a successful delete so
		// the resource is removed from state instead of erroring.
		if isStackNotFound(err) {
			tflog.Warn(ctx, "stack already deleted; treating delete as success")
			return
		}
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error deleting stack", "Error in deleting stack "+state.ResourceName.ValueString()+": "+err.Error())
		return
	}
}

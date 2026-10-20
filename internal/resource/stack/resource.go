package stack

import (
	"context"
	"errors"
	"fmt"
	"slices"
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
	"github.com/hashicorp/terraform-plugin-framework/types"
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

// isNotFound reports whether err from an API call is a 404: the stack was
// deleted out-of-band, or the stack template revision does not exist (yet).
func isNotFound(err error) bool {
	var apiErr *core.APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == 404
}

// fetchTemplateRevision reads the stack template revision referenced by the
// model's template_group_id. It returns nil (without an error) when no
// template group id is set, so merging is a no-op in that case.
//
// The Terraform-facing template_group_id is stored bare ("<name>:<revision>",
// no org prefix) — ToAPIModel/ToUpdateAPIModel add the "/<org>/" prefix the
// wire format documents (see the schema description) only when building the
// request, and BuildAPIModelToStackModel strips it back off on read. Unlike
// workflow_from_template's iac_template_id, stack templates aren't
// documented as cross-org shareable, so this is always read against the
// stack's own org rather than an org parsed out of the id.
func (r *stackResource) fetchTemplateRevision(ctx context.Context, m StackResourceModel) (*stacktemplaterevisions.ReadStackTemplateRevisionModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	templateGroupId := m.TemplateGroupId.ValueString()
	if templateGroupId == "" {
		return nil, diags
	}

	tpl, err := r.readTemplateRevision(ctx, templateGroupId)
	if err != nil {
		diags.AddError("Error reading stack template revision",
			fmt.Sprintf("Could not read template revision %q to resolve stack defaults: %s", templateGroupId, err.Error()))
		return nil, diags
	}
	return tpl, diags
}

// readTemplateRevision reads a stack template revision by its bare id, keeping
// the API error so the caller can tell a missing revision (404) from a failure.
func (r *stackResource) readTemplateRevision(ctx context.Context, templateGroupId string) (*stacktemplaterevisions.ReadStackTemplateRevisionModel, error) {
	readResp, err := r.client.StackTemplateRevisions.ReadStackTemplateRevision(ctx, r.org_name, templateGroupId)
	if err != nil {
		return nil, err
	}
	if readResp == nil {
		return nil, fmt.Errorf("API returned an empty response")
	}
	return &readResp.Msg, nil
}

// workflowIdsChanged reports whether the workflow ids declared in plan differ
// from those in state, in order. An unknown or null value on either side
// counts as a change, so the caller re-validates rather than assuming.
func workflowIdsChanged(ctx context.Context, plan, state types.Object) bool {
	planIds, ok := declaredWorkflowIds(ctx, plan)
	if !ok {
		return true
	}
	stateIds, ok := declaredWorkflowIds(ctx, state)
	if !ok {
		return true
	}
	return !slices.Equal(planIds, stateIds)
}

// declaredWorkflowIds returns workflows_config.workflows[*].id in order, and
// false when the value or any id is null or unknown.
func declaredWorkflowIds(ctx context.Context, wfc types.Object) ([]string, bool) {
	if wfc.IsNull() || wfc.IsUnknown() {
		return nil, false
	}
	var m WorkflowsConfigModel
	if d := wfc.As(ctx, &m, basetypes.ObjectAsOptions{}); d.HasError() || m.Workflows.IsNull() || m.Workflows.IsUnknown() {
		return nil, false
	}
	var wfs []WorkflowInStackModel
	if d := m.Workflows.ElementsAs(ctx, &wfs, false); d.HasError() {
		return nil, false
	}
	ids := make([]string, 0, len(wfs))
	for _, wf := range wfs {
		if wf.Id.IsNull() || wf.Id.IsUnknown() {
			return nil, false
		}
		ids = append(ids, wf.Id.ValueString())
	}
	return ids, true
}

// fetchWorkflowTemplateRevision reads the workflow template revision referenced by
// templateId — a stack-template-revision workflow slot's
// vcs_config.iac_vcs_config.iac_template_id (never the stack's own, since that
// attribute is Computed-only and always inherited from the matched slot). Returns
// nil (without an error) when templateId is empty. Mirrors
// workflow_from_template's fetchTemplateRevision.
func (r *stackResource) fetchWorkflowTemplateRevision(ctx context.Context, templateId string) (*workflowtemplaterevisions.ReadWorkflowTemplateRevisionModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	if templateId == "" {
		return nil, diags
	}

	// templateId is stored fully-qualified ("/<org>/<name>:<rev>"), but
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

// resolveWorkflowTemplates fetches, for every workflow slot declared in
// plan.WorkflowsConfig, the workflow template revision that slot resolves
// against — read from the MATCHING slot on stackTpl (the stack template
// revision), never from the stack's own entry, since
// vcs_config.iac_vcs_config is Computed-only on the stack resource and is
// always inherited from the stack template. Slots are matched by id (the
// stack-template-defined workflow slot UUID). A slot with no match on
// stackTpl, or whose matched slot has no iac_template_id, is simply omitted
// from the result — expandWorkflowsConfig then merges only the layers that
// exist for that slot. The returned map is keyed by workflow slot id.
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
// It rejects, at plan time, a workflows_config.workflows that does not declare
// exactly the workflows of the stack template revision, in its order
// (validateWorkflowsConfigMatchesRevision). That runs on create, on a
// template_group_id change, and on an update that changes the declared
// workflow ids; a plan that changes none of these was already checked when
// the current state was applied, so it skips the extra template read. The
// check is skipped when template_group_id is unknown at plan time, or on
// create when the revision is not found yet — both happen when the revision
// is created in the same apply. Create and Update repeat the check against
// the revision they read, so those cases are still caught at apply.
//
// On a template_group_id change it also re-resolves the Optional+Computed
// fields the user did NOT declare against the new revision. Without that,
// UseStateForUnknown carries the OLD revision's values forward (they were
// never unknown), so the merge never runs for them. Setting those fields to
// concrete resolved values here lets plan == apply. Fields the user declared
// in config are left untouched. This covers description/tags/context_tags/
// actions (stack-level) and workflows_config's per-workflow template-derived
// fields. actions the user declared in config has no re-resolution to do, but
// is validated: a reference to a workflow the new revision dropped is an
// error, not a silent carry-forward.
func (r *stackResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return // destroy
	}
	creating := req.State.Raw.IsNull()

	var plan, state, config StackResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if !creating {
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.TemplateGroupId.IsUnknown() {
		return // template not resolved yet — Create/Update validate at apply
	}

	revisionChanged := !creating && plan.TemplateGroupId.ValueString() != state.TemplateGroupId.ValueString()
	checkWorkflows := creating || revisionChanged || workflowIdsChanged(ctx, plan.WorkflowsConfig, state.WorkflowsConfig)

	// Overrides the config declared at the last apply but no longer does. A
	// revision change re-resolves every undeclared field anyway, so this is
	// only needed for an update that keeps the revision.
	var removed declaredFields
	if !creating && !revisionChanged {
		prev, d := loadDeclaredFields(ctx, req.Private)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		removed = prev.without(declaredFieldsFromConfig(config))
	}

	if !checkWorkflows && removed.isEmpty() {
		return // the current state already matched this revision when it was applied
	}

	tpl, err := r.readTemplateRevision(ctx, plan.TemplateGroupId.ValueString())
	if err != nil {
		if creating && isNotFound(err) {
			return // the revision may be created by this apply — Create validates it
		}
		resp.Diagnostics.AddError("Error reading stack template revision",
			fmt.Sprintf("Could not read template revision %q to resolve stack defaults: %s", plan.TemplateGroupId.ValueString(), err.Error()))
		return
	}

	if checkWorkflows {
		resp.Diagnostics.Append(validateWorkflowsConfigMatchesRevision(ctx, plan.WorkflowsConfig, tpl)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if !revisionChanged {
		if !removed.isEmpty() {
			resp.Diagnostics.Append(r.revertRemovedOverrides(ctx, &plan, config, state, tpl, removed)...)
			if resp.Diagnostics.HasError() {
				return
			}
			resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
		}
		return
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

	resp.Diagnostics.Append(reResolveWorkflowsConfigOnRevisionChange(ctx, &plan, config, state.WorkflowsConfig, tpl, workflowTemplates)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

// revertRemovedOverrides sets each field in removed — an override the config no
// longer declares — back to the value it resolves to without the override. It
// re-resolves the whole stack against tpl the same way a revision change does
// (template value, else the attribute's default, else the prior value for the
// fields the platform cannot clear; see resolveUnsentWorkflowFields), then
// copies only the removed fields into plan. Every other field keeps its planned
// value, so an in-place change to a template does not show up as a diff.
func (r *stackResource) revertRemovedOverrides(ctx context.Context, plan *StackResourceModel, config, state StackResourceModel, tpl *stacktemplaterevisions.ReadStackTemplateRevisionModel, removed declaredFields) diag.Diagnostics {
	var diags diag.Diagnostics

	workflowTemplates, d := r.resolveWorkflowTemplates(ctx, *plan, tpl)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	resolved := *plan
	diags.Append(reResolveOnRevisionChange(ctx, &resolved, config, tpl)...)
	if diags.HasError() {
		return diags
	}
	diags.Append(reResolveWorkflowsConfigOnRevisionChange(ctx, &resolved, config, state.WorkflowsConfig, tpl, workflowTemplates)...)
	if diags.HasError() {
		return diags
	}

	diags.Append(applyRemovedFields(plan, resolved, removed)...)
	return diags
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

	// ModifyPlan already ran this check at plan time unless template_group_id
	// was unknown then or the revision did not exist yet (both happen when the
	// revision is created in the same apply). Repeat it against the tpl this
	// function read to build the request, so those cases fail here.
	resp.Diagnostics.Append(validateWorkflowsConfigMatchesRevision(ctx, plan.WorkflowsConfig, tpl)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workflowTemplates, diags := r.resolveWorkflowTemplates(ctx, plan, tpl)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := plan.ToAPIModel(ctx, r.org_name, tpl, workflowTemplates)
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

	stackModel, diags := BuildAPIModelToStackModel(ctx, r.org_name, readResp.Msg, plan.WorkflowGroupId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Keep the configured form of an input or variable the platform stored with
	// its references resolved; see preserveConfiguredReferences.
	stackModel.WorkflowsConfig, diags = preserveConfiguredReferences(stackModel, plan.WorkflowsConfig)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	resp.Diagnostics.Append(resp.State.Set(ctx, &stackModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(saveDeclaredFieldsFromConfig(ctx, req.Config, resp.Private)...)
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
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error reading stack", "Could not read stack "+state.ResourceName.ValueString()+": "+err.Error())
		return
	}

	stackModel, diags := BuildAPIModelToStackModel(ctx, r.org_name, readResp.Msg, state.WorkflowGroupId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Keep the configured form of an input or variable the platform stored with
	// its references resolved; see preserveConfiguredReferences.
	stackModel.WorkflowsConfig, diags = preserveConfiguredReferences(stackModel, state.WorkflowsConfig)
	resp.Diagnostics.Append(diags...)
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

	// ModifyPlan already ran this check at plan time unless template_group_id
	// was unknown then. Repeat it against the tpl this function read to build
	// the request: it costs nothing, and covers that case.
	resp.Diagnostics.Append(validateWorkflowsConfigMatchesRevision(ctx, plan.WorkflowsConfig, tpl)...)
	if resp.Diagnostics.HasError() {
		return
	}

	workflowTemplates, diags := r.resolveWorkflowTemplates(ctx, plan, tpl)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, diags := plan.ToUpdateAPIModel(ctx, r.org_name, tpl, workflowTemplates)
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

	stackModel, diags := BuildAPIModelToStackModel(ctx, r.org_name, updatedStack.Msg, plan.WorkflowGroupId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Keep the configured form of an input or variable the platform stored with
	// its references resolved; see preserveConfiguredReferences.
	stackModel.WorkflowsConfig, diags = preserveConfiguredReferences(stackModel, plan.WorkflowsConfig)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, stackModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(saveDeclaredFieldsFromConfig(ctx, req.Config, resp.Private)...)
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
		if isNotFound(err) {
			tflog.Warn(ctx, "stack already deleted; treating delete as success")
			return
		}
		tflog.Error(ctx, err.Error())
		resp.Diagnostics.AddError("Error deleting stack", "Error in deleting stack "+state.ResourceName.ValueString()+": "+err.Error())
		return
	}
}

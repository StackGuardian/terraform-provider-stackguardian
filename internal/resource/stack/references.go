package stack

import (
	"context"
	"encoding/json"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ${reference::template-group.<index>...} values in a stack's workflows refer to the
// outputs of a sibling workflow by its position in the stack. When core CREATES a
// stack's workflows it rewrites them to ${workflow::<group>.<stack>.<workflow id>...}
// (__process_string_reference in core's stacks resource), in
// vcs_config.iac_input_data.data and in PLAIN_TEXT environment_variables text_value.
// Updates store the value as sent, without rewriting.
//
// So a create that sends a ${reference::} value (declared in config, or inherited
// from the stack template revision and known at plan time) reads back the rewritten
// form, and Terraform fails the apply with "Provider produced inconsistent result
// after apply"; a later refresh would show it as a diff. preserveReferenceForms maps
// those values back: after every read, a value whose prior form (the plan on
// Create/Update, the state on Read) is a ${reference::} that core's rewrite turns
// into exactly what the API returned keeps the prior form. Anything else is left as
// the API returned it, so genuine drift still shows.

// referencePattern matches a value core rewrites: it must START with the reference
// (core only checks the prefix), naming the sibling workflow by its 0-based index.
var referencePattern = regexp.MustCompile(`^\$\{reference::template-group\.(\d+)\.(.*)$`)

// referenceRewrittenAs reports whether core's rewrite of the reference prior yields
// actual. workflowIds lists the stack's workflow ids (workflow_id) in the order of
// workflows_config.workflows, which is the stack template revision's order (enforced
// by validateWorkflowsConfigMatchesRevision) and the order core indexes into.
//
// Core replaces "${reference::" with "${workflow::" and "template-group" with
// "<group>.<stack>", splits on "." and replaces the third segment (the index) with
// the workflow id. The group and stack segments are not predicted (core uses the
// workflow group name and a slug of the stack name); they only have to be present.
func referenceRewrittenAs(prior, actual string, workflowIds []string) bool {
	m := referencePattern.FindStringSubmatch(prior)
	if m == nil {
		return false
	}
	index, err := strconv.Atoi(m[1])
	if err != nil || index < 0 || index >= len(workflowIds) || workflowIds[index] == "" {
		return false
	}
	rest := m[2]
	if strings.Contains(rest, "template-group") {
		// Core would rewrite that occurrence too; not worth predicting.
		return false
	}

	afterPrefix, ok := strings.CutPrefix(actual, "${workflow::")
	if !ok {
		return false
	}
	parts := strings.SplitN(afterPrefix, ".", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" {
		return false
	}
	return parts[2] == workflowIds[index]+"."+rest
}

// jsonEquivalentModuloReferences reports whether two decoded JSON values are equal,
// treating a string in prior as equal to the string in actual when core's reference
// rewrite turns one into the other. Object keys must match exactly.
func jsonEquivalentModuloReferences(prior, actual interface{}, workflowIds []string) bool {
	switch p := prior.(type) {
	case map[string]interface{}:
		a, ok := actual.(map[string]interface{})
		if !ok || len(a) != len(p) {
			return false
		}
		for k, pv := range p {
			av, ok := a[k]
			if !ok || !jsonEquivalentModuloReferences(pv, av, workflowIds) {
				return false
			}
		}
		return true
	case []interface{}:
		a, ok := actual.([]interface{})
		if !ok || len(a) != len(p) {
			return false
		}
		for i := range p {
			if !jsonEquivalentModuloReferences(p[i], a[i], workflowIds) {
				return false
			}
		}
		return true
	case string:
		a, ok := actual.(string)
		return ok && (a == p || referenceRewrittenAs(p, a, workflowIds))
	default:
		return reflect.DeepEqual(prior, actual)
	}
}

// preserveReferenceForms keeps the prior ${reference::} form of
// vcs_config.iac_input_data.data and PLAIN_TEXT environment_variables text_value in
// result.WorkflowsConfig wherever the API returned core's rewrite of it (see the
// comment at the top of this file). Workflows are matched by id, environment
// variables by position and var_name.
func preserveReferenceForms(ctx context.Context, result *StackResourceModel, priorWorkflowsConfig types.Object) diag.Diagnostics {
	var diags diag.Diagnostics

	resultWorkflows, ok := listAttr(result.WorkflowsConfig, "workflows")
	if !ok {
		return diags
	}
	priorWorkflows, ok := listAttr(priorWorkflowsConfig, "workflows")
	if !ok {
		return diags
	}

	priorById := make(map[string]types.Object)
	for _, v := range priorWorkflows.Elements() {
		wf, ok := v.(types.Object)
		if !ok {
			continue
		}
		if id, ok := stringAttr(wf, "id"); ok {
			priorById[id] = wf
		}
	}
	if len(priorById) == 0 {
		return diags
	}

	elems := resultWorkflows.Elements()
	workflowIds := make([]string, len(elems))
	for i, v := range elems {
		if wf, ok := v.(types.Object); ok {
			workflowIds[i], _ = stringAttr(wf, "workflow_id")
		}
	}

	changed := false
	for i, v := range elems {
		wf, ok := v.(types.Object)
		if !ok {
			continue
		}
		id, _ := stringAttr(wf, "id")
		prior, ok := priorById[id]
		if !ok {
			continue
		}

		newWf, wfChanged, d := preserveWorkflowReferenceForms(ctx, wf, prior, workflowIds)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		if wfChanged {
			elems[i] = newWf
			changed = true
		}
	}
	if !changed {
		return diags
	}

	newList, d := types.ListValue(resultWorkflows.ElementType(ctx), elems)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	newWfc, d := withAttr(ctx, result.WorkflowsConfig, "workflows", newList)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	result.WorkflowsConfig = newWfc
	return diags
}

// preserveWorkflowReferenceForms applies preserveReferenceForms to one workflow.
func preserveWorkflowReferenceForms(ctx context.Context, wf, prior types.Object, workflowIds []string) (types.Object, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	changed := false

	// vcs_config.iac_input_data.data: a JSON string; compare decoded, keep the prior
	// string (its exact text is what config and plan hold).
	if vcs, ok := objectAttr(wf, "vcs_config"); ok {
		if iac, ok := objectAttr(vcs, "iac_input_data"); ok {
			data, dataOk := stringAttr(iac, "data")
			priorData := ""
			if pv, ok := objectAttr(prior, "vcs_config"); ok {
				if pi, ok := objectAttr(pv, "iac_input_data"); ok {
					priorData, _ = stringAttr(pi, "data")
				}
			}
			if dataOk && priorData != "" && priorData != data && strings.Contains(priorData, "${reference::") {
				var p, a interface{}
				if json.Unmarshal([]byte(priorData), &p) == nil && json.Unmarshal([]byte(data), &a) == nil &&
					jsonEquivalentModuloReferences(p, a, workflowIds) {
					newIac, d := withAttr(ctx, iac, "data", types.StringValue(priorData))
					diags.Append(d...)
					newVcs, d2 := withAttr(ctx, vcs, "iac_input_data", newIac)
					diags.Append(d2...)
					newWf, d3 := withAttr(ctx, wf, "vcs_config", newVcs)
					diags.Append(d3...)
					if diags.HasError() {
						return wf, false, diags
					}
					wf, changed = newWf, true
				}
			}
		}
	}

	// environment_variables[*].config.text_value: core rewrites PLAIN_TEXT values.
	envs, ok := listAttr(wf, "environment_variables")
	priorEnvs, priorOk := listAttr(prior, "environment_variables")
	if ok && priorOk {
		elems, priorElems := envs.Elements(), priorEnvs.Elements()
		envChanged := false
		for j := 0; j < len(elems) && j < len(priorElems); j++ {
			env, ok1 := elems[j].(types.Object)
			priorEnv, ok2 := priorElems[j].(types.Object)
			if !ok1 || !ok2 {
				continue
			}
			cfg, ok1 := objectAttr(env, "config")
			priorCfg, ok2 := objectAttr(priorEnv, "config")
			if !ok1 || !ok2 {
				continue
			}
			name, _ := stringAttr(cfg, "var_name")
			priorName, _ := stringAttr(priorCfg, "var_name")
			text, ok1 := stringAttr(cfg, "text_value")
			priorText, ok2 := stringAttr(priorCfg, "text_value")
			if !ok1 || !ok2 || name != priorName || !referenceRewrittenAs(priorText, text, workflowIds) {
				continue
			}
			newCfg, d := withAttr(ctx, cfg, "text_value", types.StringValue(priorText))
			diags.Append(d...)
			newEnv, d2 := withAttr(ctx, env, "config", newCfg)
			diags.Append(d2...)
			if diags.HasError() {
				return wf, false, diags
			}
			elems[j] = newEnv
			envChanged = true
		}
		if envChanged {
			newEnvs, d := types.ListValue(envs.ElementType(ctx), elems)
			diags.Append(d...)
			newWf, d2 := withAttr(ctx, wf, "environment_variables", newEnvs)
			diags.Append(d2...)
			if diags.HasError() {
				return wf, false, diags
			}
			wf, changed = newWf, true
		}
	}

	return wf, changed, diags
}

// objectAttr returns the known, non-null object attribute name of o.
func objectAttr(o types.Object, name string) (types.Object, bool) {
	if o.IsNull() || o.IsUnknown() {
		return types.Object{}, false
	}
	v, ok := o.Attributes()[name].(types.Object)
	if !ok || v.IsNull() || v.IsUnknown() {
		return types.Object{}, false
	}
	return v, true
}

// listAttr returns the known, non-null list attribute name of o.
func listAttr(o types.Object, name string) (types.List, bool) {
	if o.IsNull() || o.IsUnknown() {
		return types.List{}, false
	}
	v, ok := o.Attributes()[name].(types.List)
	if !ok || v.IsNull() || v.IsUnknown() {
		return types.List{}, false
	}
	return v, true
}

// stringAttr returns the known, non-null string attribute name of o.
func stringAttr(o types.Object, name string) (string, bool) {
	if o.IsNull() || o.IsUnknown() {
		return "", false
	}
	v, ok := o.Attributes()[name].(types.String)
	if !ok || v.IsNull() || v.IsUnknown() {
		return "", false
	}
	return v.ValueString(), true
}

// withAttr returns a copy of o with attribute name set to v.
func withAttr(ctx context.Context, o types.Object, name string, v attr.Value) (types.Object, diag.Diagnostics) {
	attrs := o.Attributes()
	attrs[name] = v
	return types.ObjectValue(o.AttributeTypes(ctx), attrs)
}

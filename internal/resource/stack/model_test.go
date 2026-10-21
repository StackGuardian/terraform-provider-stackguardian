package stack_test

import (
	"context"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/resource/stack"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestOverlayConfigUnknowns covers the value shapes overlayConfigUnknowns walks:
// a value unknown in config must stay unknown in the result, a value null in
// config (not set by the user) keeps the re-resolved planned value, and lists
// are paired by position only when both sides have the same length.
func TestOverlayConfigUnknowns(t *testing.T) {
	ctx := context.Background()

	envType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"var_name":   types.StringType,
		"text_value": types.StringType,
	}}
	wfType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"id":                    types.StringType,
		"description":           types.StringType,
		"resource_name":         types.StringType,
		"environment_variables": types.ListType{ElemType: envType},
		"context_tags":          types.MapType{ElemType: types.StringType},
	}}

	env := func(name string, value types.String) attr.Value {
		return types.ObjectValueMust(envType.AttrTypes, map[string]attr.Value{
			"var_name":   types.StringValue(name),
			"text_value": value,
		})
	}
	wf := func(id, description, resourceName types.String, envVars types.List, contextTags types.Map) attr.Value {
		return types.ObjectValueMust(wfType.AttrTypes, map[string]attr.Value{
			"id":                    id,
			"description":           description,
			"resource_name":         resourceName,
			"environment_variables": envVars,
			"context_tags":          contextTags,
		})
	}
	list := func(elemType attr.Type, elems ...attr.Value) types.List {
		return types.ListValueMust(elemType, elems)
	}
	strMap := func(m map[string]attr.Value) types.Map {
		return types.MapValueMust(types.StringType, m)
	}

	cfg := list(wfType, wf(
		types.StringUnknown(), // id from another resource
		types.StringUnknown(), // declared, not known until apply
		types.StringNull(),    // not set: re-resolved from the template
		list(envType, env("BUCKET", types.StringUnknown())),
		strMap(map[string]attr.Value{"env": types.StringUnknown(), "team": types.StringValue("core")}),
	))
	planned := list(wfType, wf(
		types.StringValue(""),
		types.StringValue("from workflow template"),
		types.StringValue("wf-1"),
		list(envType, env("BUCKET", types.StringValue(""))),
		strMap(map[string]attr.Value{"env": types.StringValue(""), "team": types.StringValue("core")}),
	))
	want := list(wfType, wf(
		types.StringUnknown(),
		types.StringUnknown(),
		types.StringValue("wf-1"),
		list(envType, env("BUCKET", types.StringUnknown())),
		strMap(map[string]attr.Value{"env": types.StringUnknown(), "team": types.StringValue("core")}),
	))

	got, diags := stack.OverlayConfigUnknowns(ctx, cfg, planned)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !got.Equal(want) {
		t.Errorf("unknown values not carried over from config\ngot:  %s\nwant: %s", got, want)
	}

	t.Run("length mismatch keeps planned", func(t *testing.T) {
		shortCfg := list(envType, env("A", types.StringUnknown()))
		longPlanned := list(envType, env("A", types.StringValue("")), env("B", types.StringValue("")))
		got, diags := stack.OverlayConfigUnknowns(ctx, shortCfg, longPlanned)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if !got.Equal(longPlanned) {
			t.Errorf("got %s, want planned unchanged %s", got, longPlanned)
		}
	})

	t.Run("null config keeps planned", func(t *testing.T) {
		nullCfg := types.ListNull(envType)
		got, diags := stack.OverlayConfigUnknowns(ctx, nullCfg, list(envType, env("A", types.StringValue("x"))))
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if got.IsNull() {
			t.Errorf("got null, want planned value")
		}
	})
}

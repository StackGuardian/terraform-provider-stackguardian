package stack_test

import (
	"context"
	"strings"
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

// TestWorkflowsMismatchDetail checks the workflows_config mismatch error says
// what is wrong (missing, not in the revision, duplicated, or only the order),
// names the revision, and labels ids with their resource_name.
func TestWorkflowsMismatchDetail(t *testing.T) {
	const (
		a = "d8dfaf15-2ad9-da29-8af0-c6b288b12089"
		b = "3f7c9e2a-5b1d-4e6f-8a2c-9d4b6e1f0a3c"
		x = "11111111-2222-3333-4444-555555555555"
	)
	required := []string{a, b}
	names := map[string]string{a: "wf-1", b: "wf-2"}
	revision := "/my-org/my-stack-template:2"
	expectedOrder := "Expected order:\n  1. " + a + " (wf-1)\n  2. " + b + " (wf-2)"

	cases := []struct {
		name     string
		declared []string
		want     []string // substrings the detail must contain
	}{
		{"missing", []string{a}, []string{"Missing:", b + " (wf-2)"}},
		{"wrong order", []string{b, a}, []string{"The workflows are all declared, but not in the revision's order."}},
		{"not in revision", []string{a, b, x}, []string{"Not in the revision:", x}},
		{"duplicated", []string{a, a, b}, []string{"Declared more than once:", a + " (wf-1)"}},
		{"none declared", nil, []string{"No workflows are declared."}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			detail, mismatch := stack.WorkflowsMismatchDetail(required, c.declared, names, revision)
			if !mismatch {
				t.Fatalf("mismatch = false, want true")
			}
			for _, want := range append(c.want, revision, expectedOrder) {
				if !strings.Contains(detail, want) {
					t.Errorf("detail missing %q:\n%s", want, detail)
				}
			}
		})
	}

	t.Run("match", func(t *testing.T) {
		if detail, mismatch := stack.WorkflowsMismatchDetail(required, []string{a, b}, names, revision); mismatch {
			t.Errorf("mismatch = true for an exact match, detail:\n%s", detail)
		}
	})

	t.Run("no resource name", func(t *testing.T) {
		detail, _ := stack.WorkflowsMismatchDetail(required, []string{a}, map[string]string{}, "")
		if strings.Contains(detail, "()") || !strings.Contains(detail, "the stack template revision defines") {
			t.Errorf("unexpected detail without names or revision id:\n%s", detail)
		}
	})
}

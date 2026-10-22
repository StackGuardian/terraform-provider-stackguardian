package stack

import (
	"reflect"
	"testing"
)

func TestStackReferencesResolveString(t *testing.T) {
	sr := stackReferences{
		wfGrpId:           "platform",
		stackId:           "web-stack",
		stackName:         "Web Stack",
		workflowIds:       []string{"vpc-aaaa1111", "app-bbbb2222"},
		workflowIdsByName: map[string]string{"vpc": "vpc-aaaa1111", "app": "app-bbbb2222"},
	}
	for _, tc := range []struct {
		name, in, want string
	}{
		{"reference by index", "${reference::template-group.0.outputs.vpc_id}", "${workflow::platform.web-stack.vpc-aaaa1111.outputs.vpc_id}"},
		{"second workflow", "${reference::template-group.1.outputs.url}", "${workflow::platform.web-stack.app-bbbb2222.outputs.url}"},
		{"index out of range", "${reference::template-group.5.outputs.vpc_id}", "${reference::template-group.5.outputs.vpc_id}"},
		{"index not a number", "${reference::template-group.vpc.outputs.vpc_id}", "${reference::template-group.vpc.outputs.vpc_id}"},
		{"workflow reference by names", "${workflow::platform.Web Stack.vpc.outputs.vpc_id}", "${workflow::platform.web-stack.vpc-aaaa1111.outputs.vpc_id}"},
		{"workflow reference by ids", "${workflow::platform.web-stack.vpc-aaaa1111.outputs.vpc_id}", "${workflow::platform.web-stack.vpc-aaaa1111.outputs.vpc_id}"},
		{"not at the start", "id=${reference::template-group.0.outputs.vpc_id}", "id=${reference::template-group.0.outputs.vpc_id}"},
		{"plain value", "eu-central-1", "eu-central-1"},
		{"secret reference", "${secret::db-password}", "${secret::db-password}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := sr.resolveString(tc.in); got != tc.want {
				t.Errorf("resolveString(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}

	t.Run("no workflow ids yet", func(t *testing.T) {
		in := "${reference::template-group.0.outputs.vpc_id}"
		if got := (stackReferences{}).resolveString(in); got != in {
			t.Errorf("resolveString = %q, want it unchanged", got)
		}
	})
}

func TestStackReferencesResolveNested(t *testing.T) {
	sr := stackReferences{wfGrpId: "g", stackId: "s", workflowIds: []string{"vpc-1"}}
	in := map[string]any{
		"subnets": []any{"${reference::template-group.0.outputs.subnet}", 3.0},
		"tags":    map[string]any{"${reference::template-group.0.outputs.key}": "x"},
	}
	want := map[string]any{
		"subnets": []any{"${workflow::g.s.vpc-1.outputs.subnet}", 3.0},
		"tags":    map[string]any{"${workflow::g.s.vpc-1.outputs.key}": "x"},
	}
	if got := sr.resolve(in); !reflect.DeepEqual(got, want) {
		t.Errorf("resolve = %v, want %v", got, want)
	}
}

func TestEquivalentInputData(t *testing.T) {
	sr := stackReferences{wfGrpId: "g", stackId: "s", workflowIds: []string{"vpc-1"}}
	for _, tc := range []struct {
		name, configured, stored string
		want                     bool
	}{
		{"resolved reference", `{"vpc":"${reference::template-group.0.outputs.id}"}`, `{"vpc":"${workflow::g.s.vpc-1.outputs.id}"}`, true},
		{"formatting and key order", `{ "b": 1,  "a": "x" }`, `{"a":"x","b":1}`, true},
		{"different value", `{"a":"x"}`, `{"a":"y"}`, false},
		{"invalid configured JSON", `{"a":`, `{"a":"x"}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := equivalentInputData(sr, tc.configured, tc.stored); got != tc.want {
				t.Errorf("equivalentInputData = %v, want %v", got, tc.want)
			}
		})
	}
}

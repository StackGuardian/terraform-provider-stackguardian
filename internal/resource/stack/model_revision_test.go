package stack

import (
	"context"
	"testing"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	existingSlot = "d8dfaf15-2ad9-da29-8af0-c6b288b12089"
	newSlot      = "3f7c9e2a-5b1d-4e6f-8a2c-9d4b6e1f0a3c"
)

func mustFlattenWorkflows(t *testing.T, workflows ...*sgsdkgo.StackWorkflowsConfigWorkflow) types.Object {
	t.Helper()
	obj, diags := flattenWorkflowsConfig(context.Background(), &sgsdkgo.StackWorkflowsConfig{Workflows: workflows})
	if diags.HasError() {
		t.Fatalf("flattenWorkflowsConfig: %v", diags)
	}
	return obj
}

func workflowAttrs(t *testing.T, wfc types.Object, slot string) map[string]attr.Value {
	t.Helper()
	obj, ok := workflowObjectsById(wfc)[slot]
	if !ok {
		t.Fatalf("workflow %s not found", slot)
	}
	return obj.Attributes()
}

func iacInputData(t *testing.T, wf map[string]attr.Value) attr.Value {
	t.Helper()
	vcs, ok := wf["vcs_config"].(types.Object)
	if !ok || vcs.IsNull() || vcs.IsUnknown() {
		t.Fatalf("vcs_config = %v, want a known object", wf["vcs_config"])
	}
	return vcs.Attributes()["iac_input_data"]
}

func TestResolveUnsentWorkflowFields(t *testing.T) {
	rawJSON := sgsdkgo.IacInputDataSchemaTypeEnum("RAW_JSON")
	priorData := map[string]interface{}{"region": "eu"}
	terraform := sgsdkgo.WfTypeEnum("TERRAFORM")
	opentofu := sgsdkgo.WfTypeEnum("OPENTOFU")

	// State before the revision change: the workflow as the platform created it.
	prior := mustFlattenWorkflows(t, &sgsdkgo.StackWorkflowsConfigWorkflow{
		Id:              sgsdkgo.String(existingSlot),
		WorkflowId:      sgsdkgo.String("old-tpl-d8dfaf15"),
		ResourceName:    sgsdkgo.String("old-tpl-0"),
		Description:     sgsdkgo.String("from the platform"),
		WfType:          &terraform,
		UserJobCpu:      sgsdkgo.Int(2),
		TerraformConfig: &sgsdkgo.TerraformConfig{TerraformVersion: sgsdkgo.String("1.5.7")},
		VcsConfig: &sgsdkgo.VcsConfig{
			IacVcsConfig: &sgsdkgo.IacvcsConfig{IacTemplateId: sgsdkgo.String("/org/old-tpl:1")},
			IacInputData: &sgsdkgo.IacInputData{SchemaType: &rawJSON, Data: &priorData},
		},
	})

	// The merge against the new revision: a different workflow template, a
	// resource_name the stack template now sets, and nothing for the other fields.
	resolved := mustFlattenWorkflows(t,
		&sgsdkgo.StackWorkflowsConfigWorkflow{
			Id:           sgsdkgo.String(existingSlot),
			WorkflowId:   sgsdkgo.String("new-tpl-d8dfaf15"),
			ResourceName: sgsdkgo.String("renamed-by-template"),
			WfType:       &opentofu,
			VcsConfig: &sgsdkgo.VcsConfig{
				IacVcsConfig: &sgsdkgo.IacvcsConfig{IacTemplateId: sgsdkgo.String("/org/new-tpl:1")},
			},
		},
		&sgsdkgo.StackWorkflowsConfigWorkflow{
			Id:         sgsdkgo.String(newSlot),
			WorkflowId: sgsdkgo.String("new-tpl-3f7c9e2a"),
			VcsConfig: &sgsdkgo.VcsConfig{
				IacVcsConfig: &sgsdkgo.IacvcsConfig{IacTemplateId: sgsdkgo.String("/org/new-tpl:1")},
			},
		},
	)

	got, diags := resolveUnsentWorkflowFields(resolved, prior)
	if diags.HasError() {
		t.Fatalf("resolveUnsentWorkflowFields: %v", diags)
	}

	t.Run("existing workflow", func(t *testing.T) {
		wf := workflowAttrs(t, got, existingSlot)
		priorWf := workflowAttrs(t, prior, existingSlot)

		// The platform ignores WorkflowId and strips WfType when updating a workflow.
		assertString(t, wf, "workflow_id", "old-tpl-d8dfaf15")
		assertString(t, wf, "wf_type", "TERRAFORM")
		// A value a template layer resolves is sent, so it wins over the prior one.
		assertString(t, wf, "resource_name", "renamed-by-template")
		// Nothing resolves these, so the platform keeps the current value.
		assertString(t, wf, "description", "from the platform")
		if !wf["user_job_cpu"].Equal(types.Int64Value(2)) {
			t.Errorf("user_job_cpu = %v, want 2", wf["user_job_cpu"])
		}
		if !wf["terraform_config"].Equal(priorWf["terraform_config"]) {
			t.Errorf("terraform_config = %v, want the prior value", wf["terraform_config"])
		}
		if !iacInputData(t, wf).Equal(iacInputData(t, priorWf)) {
			t.Errorf("iac_input_data = %v, want the prior value", iacInputData(t, wf))
		}
		// iac_vcs_config always comes from the new stack template revision.
		vcs := wf["vcs_config"].(types.Object).Attributes()["iac_vcs_config"].(types.Object)
		assertString(t, vcs.Attributes(), "iac_template_id", "/org/new-tpl:1")
		// A field with a sendable default keeps it: the revision change resets it.
		if !wf["number_of_approvals_required"].Equal(types.Int64Value(0)) {
			t.Errorf("number_of_approvals_required = %v, want 0", wf["number_of_approvals_required"])
		}
	})

	t.Run("new workflow", func(t *testing.T) {
		wf := workflowAttrs(t, got, newSlot)

		// The platform creates the workflow under the id the provider sends.
		assertString(t, wf, "workflow_id", "new-tpl-3f7c9e2a")
		for _, name := range append([]string{"wf_type"}, unsentWhenNullWorkflowFields...) {
			if !wf[name].IsUnknown() {
				t.Errorf("%s = %v, want unknown", name, wf[name])
			}
		}
		if !iacInputData(t, wf).IsUnknown() {
			t.Errorf("iac_input_data = %v, want unknown", iacInputData(t, wf))
		}
	})
}

func TestResolveUnsentWorkflowFieldsNullPrior(t *testing.T) {
	resolved := mustFlattenWorkflows(t, &sgsdkgo.StackWorkflowsConfigWorkflow{Id: sgsdkgo.String(newSlot)})
	got, diags := resolveUnsentWorkflowFields(resolved, types.ObjectNull(WorkflowsConfigModel{}.AttributeTypes()))
	if diags.HasError() {
		t.Fatalf("resolveUnsentWorkflowFields: %v", diags)
	}
	wf := workflowAttrs(t, got, newSlot)
	if !wf["resource_name"].IsUnknown() {
		t.Errorf("resource_name = %v, want unknown", wf["resource_name"])
	}
	// No VcsConfig resolved at all: the whole object is left to the platform.
	if !wf["vcs_config"].IsUnknown() {
		t.Errorf("vcs_config = %v, want unknown", wf["vcs_config"])
	}
}

func assertString(t *testing.T, attrs map[string]attr.Value, name, want string) {
	t.Helper()
	if !attrs[name].Equal(types.StringValue(want)) {
		t.Errorf("%s = %v, want %q", name, attrs[name], want)
	}
}

func TestWorkflowIdsChanged(t *testing.T) {
	ctx := context.Background()
	wfc := func(ids ...string) types.Object {
		workflows := make([]*sgsdkgo.StackWorkflowsConfigWorkflow, 0, len(ids))
		for _, id := range ids {
			workflows = append(workflows, &sgsdkgo.StackWorkflowsConfigWorkflow{Id: sgsdkgo.String(id)})
		}
		return mustFlattenWorkflows(t, workflows...)
	}
	null := types.ObjectNull(WorkflowsConfigModel{}.AttributeTypes())
	unknown := types.ObjectUnknown(WorkflowsConfigModel{}.AttributeTypes())

	for _, tc := range []struct {
		name        string
		plan, state types.Object
		want        bool
	}{
		{"same ids, same order", wfc("a", "b"), wfc("a", "b"), false},
		{"reordered", wfc("b", "a"), wfc("a", "b"), true},
		{"added", wfc("a", "b"), wfc("a"), true},
		{"removed", wfc("a"), wfc("a", "b"), true},
		{"unknown plan", unknown, wfc("a"), true},
		{"null state", wfc("a"), null, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := workflowIdsChanged(ctx, tc.plan, tc.state); got != tc.want {
				t.Errorf("workflowIdsChanged = %v, want %v", got, tc.want)
			}
		})
	}
}

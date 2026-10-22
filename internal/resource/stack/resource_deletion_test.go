package stack

import (
	"context"
	"strings"
	"testing"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// stackState returns a schema-backed value holding a two-workflow stack.
func stackState(t *testing.T, id string) (resource.SchemaResponse, StackResourceModel) {
	t.Helper()
	ctx := context.Background()
	var schemaResp resource.SchemaResponse
	(&stackResource{}).Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	model, diags := BuildAPIModelToStackModel(ctx, "org", &sgsdkgo.StackData{
		Id:              id,
		TemplateGroupId: sgsdkgo.String("/org/tpl:1"),
		WorkflowsConfig: &sgsdkgo.StackWorkflowsConfig{Workflows: []*sgsdkgo.StackWorkflowsConfigWorkflow{
			{Id: sgsdkgo.String("wf-a")}, {Id: sgsdkgo.String("wf-b")},
		}},
	}, types.StringValue("grp"))
	if diags.HasError() {
		t.Fatalf("BuildAPIModelToStackModel: %v", diags)
	}
	return schemaResp, *model
}

func modifyPlan(t *testing.T, schemaResp resource.SchemaResponse, state StackResourceModel, plan *StackResourceModel) resource.ModifyPlanResponse {
	t.Helper()
	ctx := context.Background()
	nullValue := tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil)

	stateValue := tfsdk.State{Schema: schemaResp.Schema, Raw: nullValue}
	if diags := stateValue.Set(ctx, &state); diags.HasError() {
		t.Fatalf("state: %v", diags)
	}
	planValue := tfsdk.Plan{Schema: schemaResp.Schema, Raw: nullValue}
	configValue := tfsdk.Config{Schema: schemaResp.Schema, Raw: nullValue}
	if plan != nil {
		if diags := planValue.Set(ctx, plan); diags.HasError() {
			t.Fatalf("plan: %v", diags)
		}
		configValue.Raw = planValue.Raw
	}

	resp := resource.ModifyPlanResponse{Plan: planValue}
	(&stackResource{}).ModifyPlan(ctx, resource.ModifyPlanRequest{State: stateValue, Plan: planValue, Config: configValue}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("ModifyPlan: %v", resp.Diagnostics)
	}
	return resp
}

func deletionWarning(resp resource.ModifyPlanResponse) string {
	for _, d := range resp.Diagnostics.Warnings() {
		if d.Summary() == stackDeletionSummary {
			return d.Detail()
		}
	}
	return ""
}

func TestModifyPlanWarnsOnDestroy(t *testing.T) {
	schemaResp, state := stackState(t, "web-stack")
	detail := deletionWarning(modifyPlan(t, schemaResp, state, nil))
	for _, want := range []string{`"web-stack"`, "its 2 workflow(s)", "does not run their destroy action"} {
		if !strings.Contains(detail, want) {
			t.Errorf("warning %q does not mention %q", detail, want)
		}
	}
}

func TestModifyPlanWarnsOnReplacement(t *testing.T) {
	schemaResp, state := stackState(t, "web-stack")
	plan := state
	plan.Id = types.StringValue("web-stack-v2")
	if deletionWarning(modifyPlan(t, schemaResp, state, &plan)) == "" {
		t.Error("changing id replaces the stack, but no deletion warning was given")
	}
}

func TestModifyPlanNoWarningOnUpdate(t *testing.T) {
	schemaResp, state := stackState(t, "web-stack")
	plan := state
	plan.Description = types.StringValue("changed")
	if detail := deletionWarning(modifyPlan(t, schemaResp, state, &plan)); detail != "" {
		t.Errorf("an in-place update gave a deletion warning: %s", detail)
	}
}

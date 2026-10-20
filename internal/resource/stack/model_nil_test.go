package stack

import (
	"context"
	"testing"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// The API can return input data without a schema type. Converting it must not panic.

func TestFlattenVcsConfigWithoutSchemaType(t *testing.T) {
	data := map[string]interface{}{"region": "eu"}
	obj, diags := flattenVcsConfig(context.Background(), &sgsdkgo.VcsConfig{
		IacInputData: &sgsdkgo.IacInputData{Data: &data},
	})
	if diags.HasError() {
		t.Fatalf("flattenVcsConfig: %v", diags)
	}
	input := obj.Attributes()["iac_input_data"].(types.Object)
	if !input.Attributes()["schema_type"].IsNull() {
		t.Errorf("schema_type = %v, want null", input.Attributes()["schema_type"])
	}
	if !input.Attributes()["data"].Equal(types.StringValue(`{"region":"eu"}`)) {
		t.Errorf("data = %v, want the input data", input.Attributes()["data"])
	}
}

func TestFlattenWfStepsConfigWithoutSchemaType(t *testing.T) {
	list, diags := flattenWfStepsConfig(context.Background(), []*sgsdkgo.WfStepsConfig{{
		Name:            sgsdkgo.String("step"),
		WfStepInputData: &sgsdkgo.WfStepInputData{Data: map[string]interface{}{"k": "v"}},
	}})
	if diags.HasError() {
		t.Fatalf("flattenWfStepsConfig: %v", diags)
	}
	step := list.Elements()[0].(types.Object)
	input := step.Attributes()["wf_step_input_data"].(types.Object)
	if !input.Attributes()["schema_type"].IsNull() {
		t.Errorf("schema_type = %v, want null", input.Attributes()["schema_type"])
	}
}

func TestBuildAPIModelToStackModelWithoutStack(t *testing.T) {
	model, diags := BuildAPIModelToStackModel(context.Background(), "org", nil, types.StringValue("grp"))
	if !diags.HasError() || model != nil {
		t.Fatalf("got model %v, diags %v; want an error and no model", model, diags)
	}
}

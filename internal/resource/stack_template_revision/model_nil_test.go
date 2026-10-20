package stacktemplaterevision

import (
	"testing"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// The API can return input data without a schema type. Converting it must not panic.

func TestVcsConfigFromAPIWithoutSchemaType(t *testing.T) {
	data := map[string]interface{}{"region": "eu"}
	obj, diags := vcsConfigFromAPI(&sgsdkgo.VcsConfig{IacInputData: &sgsdkgo.IacInputData{Data: &data}})
	if diags.HasError() {
		t.Fatalf("vcsConfigFromAPI: %v", diags)
	}
	input := obj.Attributes()["iac_input_data"].(types.Object)
	if !input.Attributes()["schema_type"].IsNull() {
		t.Errorf("schema_type = %v, want null", input.Attributes()["schema_type"])
	}
}

func TestWfStepsConfigFromAPIWithoutSchemaType(t *testing.T) {
	list, diags := wfStepsConfigFromAPI([]sgsdkgo.WfStepsConfig{{
		Name:            sgsdkgo.String("step"),
		WfStepInputData: &sgsdkgo.WfStepInputData{Data: map[string]interface{}{"k": "v"}},
	}})
	if diags.HasError() {
		t.Fatalf("wfStepsConfigFromAPI: %v", diags)
	}
	step := list.Elements()[0].(types.Object)
	input := step.Attributes()["wf_step_input_data"].(types.Object)
	if !input.Attributes()["schema_type"].IsNull() {
		t.Errorf("schema_type = %v, want null", input.Attributes()["schema_type"])
	}
}

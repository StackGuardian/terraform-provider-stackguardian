package workflowgit

import (
	"context"
	"testing"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
)

// The platform accepts and stores `iacInputData: {}` ("no inputs"); reading it must not panic,
// and it reads as no iac_input_data.
func TestConvertVcsConfigFromAPI_EmptyIacInputData(t *testing.T) {
	obj, diags := convertVcsConfigFromAPI(context.Background(), &sgsdkgo.VcsConfig{IacInputData: &sgsdkgo.IacInputData{}})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if got := obj.Attributes()["iac_input_data"]; !got.IsNull() {
		t.Fatalf("iac_input_data = %v, want null", got)
	}
}

// Data without a schema type is not something the platform accepts, but reading it must not panic.
func TestConvertVcsConfigFromAPI_IacInputDataWithoutSchemaType(t *testing.T) {
	data := map[string]interface{}{"a": "b"}
	obj, diags := convertVcsConfigFromAPI(context.Background(), &sgsdkgo.VcsConfig{IacInputData: &sgsdkgo.IacInputData{Data: &data}})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if got := obj.Attributes()["iac_input_data"]; got.IsNull() {
		t.Fatal("iac_input_data is null, want the data kept")
	}
}

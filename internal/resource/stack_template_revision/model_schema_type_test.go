package stacktemplaterevision

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NO_CODE_JSON is a schema type the platform accepts and returns, though the SDK enum lacks it:
// a revision read with it must be sendable again, unchanged.
func TestConvertVcsConfigToAPI_PassesSchemaTypeThrough(t *testing.T) {
	ctx := context.Background()
	iacVcs, diags := types.ObjectValueFrom(ctx, IacVcsConfigModel{}.AttributeTypes(), IacVcsConfigModel{
		IacTemplateId: types.StringValue("my-workflow-template:1"),
	})
	if diags.HasError() {
		t.Fatal(diags)
	}
	input, diags := types.ObjectValueFrom(ctx, IacInputDataModel{}.AttributeTypes(), IacInputDataModel{
		SchemaId:   types.StringNull(),
		SchemaType: types.StringValue("NO_CODE_JSON"),
		Data:       types.StringValue(`{"region":"eu-central-1"}`),
	})
	if diags.HasError() {
		t.Fatal(diags)
	}
	vcs, diags := types.ObjectValueFrom(ctx, VcsConfigModel{}.AttributeTypes(), VcsConfigModel{IacVcsConfig: iacVcs, IacInputData: input})
	if diags.HasError() {
		t.Fatal(diags)
	}

	out, diags := convertVcsConfigToAPI(ctx, vcs, "my-org")
	if diags.HasError() {
		t.Fatalf("convertVcsConfigToAPI: %v", diags)
	}
	if out.IacInputData == nil || out.IacInputData.SchemaType == nil || *out.IacInputData.SchemaType != "NO_CODE_JSON" {
		t.Fatalf("iacInputData = %+v, want schemaType NO_CODE_JSON", out.IacInputData)
	}
}

// iac_input_data = {} is StackGuardian's "no inputs" (the workflow template's input defaults are
// then not applied): it is sent as {}, and {} reads back as the same empty object.
func TestConvertVcsConfigToAPI_EmptyIacInputDataRoundTrips(t *testing.T) {
	ctx := context.Background()
	iacVcs, diags := types.ObjectValueFrom(ctx, IacVcsConfigModel{}.AttributeTypes(), IacVcsConfigModel{
		IacTemplateId: types.StringValue("my-workflow-template:1"),
	})
	if diags.HasError() {
		t.Fatal(diags)
	}
	empty, diags := types.ObjectValueFrom(ctx, IacInputDataModel{}.AttributeTypes(), IacInputDataModel{
		SchemaId: types.StringNull(), SchemaType: types.StringNull(), Data: types.StringNull(),
	})
	if diags.HasError() {
		t.Fatal(diags)
	}
	vcs, diags := types.ObjectValueFrom(ctx, VcsConfigModel{}.AttributeTypes(), VcsConfigModel{IacVcsConfig: iacVcs, IacInputData: empty})
	if diags.HasError() {
		t.Fatal(diags)
	}

	out, diags := convertVcsConfigToAPI(ctx, vcs, "my-org")
	if diags.HasError() {
		t.Fatalf("convertVcsConfigToAPI: %v", diags)
	}
	sent, err := json.Marshal(out.IacInputData)
	if err != nil || string(sent) != "{}" {
		t.Fatalf("iacInputData sent as %s (%v), want {}", sent, err)
	}

	read, diags := vcsConfigFromAPI(out)
	if diags.HasError() {
		t.Fatalf("vcsConfigFromAPI: %v", diags)
	}
	if got := read.Attributes()["iac_input_data"]; !got.Equal(empty) {
		t.Fatalf("iac_input_data read back as %v, want %v", got, empty)
	}
}

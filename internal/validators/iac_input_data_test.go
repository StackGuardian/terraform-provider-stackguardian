package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestIacInputData(t *testing.T) {
	attrTypes := map[string]attr.Type{"schema_id": types.StringType, "schema_type": types.StringType, "data": types.StringType}
	obj := func(schemaID, schemaType, data types.String) types.Object {
		return types.ObjectValueMust(attrTypes, map[string]attr.Value{"schema_id": schemaID, "schema_type": schemaType, "data": data})
	}
	null, unknown := types.StringNull(), types.StringUnknown()
	str := types.StringValue

	for _, tc := range []struct {
		name    string
		value   types.Object
		wantErr int
	}{
		{"null object", types.ObjectNull(attrTypes), 0},
		{"empty object means no inputs", obj(null, null, null), 0},
		{"complete", obj(null, str("RAW_JSON"), str(`{"a":1}`)), 0},
		{"unknown data is not checked", obj(null, str("RAW_JSON"), unknown), 0},
		{"schema_type without data", obj(null, str("RAW_JSON"), null), 1},
		{"data without schema_type", obj(null, null, str(`{"a":1}`)), 1},
		{"only schema_id", obj(str("x"), null, null), 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := &validator.ObjectResponse{}
			IacInputData().ValidateObject(context.Background(), validator.ObjectRequest{Path: path.Root("iac_input_data"), ConfigValue: tc.value}, resp)
			if got := resp.Diagnostics.ErrorsCount(); got != tc.wantErr {
				t.Fatalf("errors = %d, want %d: %v", got, tc.wantErr, resp.Diagnostics)
			}
		})
	}
}

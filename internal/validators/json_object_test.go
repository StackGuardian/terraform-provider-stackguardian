package validators_test

import (
	"context"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestJSONObject(t *testing.T) {
	for _, tc := range []struct {
		name    string
		value   types.String
		wantErr bool
	}{
		{"object", types.StringValue(`{"region":"eu","count":2}`), false},
		{"empty object", types.StringValue(`{}`), false},
		{"empty string", types.StringValue(""), false},
		{"null", types.StringNull(), false},
		{"unknown", types.StringUnknown(), false},
		{"malformed", types.StringValue(`{"region":`), true},
		{"array", types.StringValue(`["a"]`), true},
		{"scalar", types.StringValue(`"eu"`), true},
		{"hcl-like", types.StringValue(`{ region = "eu" }`), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var resp validator.StringResponse
			validators.JSONObject().ValidateString(context.Background(), validator.StringRequest{
				Path:        path.Root("data"),
				ConfigValue: tc.value,
			}, &resp)
			if got := resp.Diagnostics.HasError(); got != tc.wantErr {
				t.Errorf("error = %v, want %v (%v)", got, tc.wantErr, resp.Diagnostics)
			}
		})
	}
}

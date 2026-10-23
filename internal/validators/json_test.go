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
	cases := []struct {
		value types.String
		valid bool
	}{
		{types.StringValue(`{"bucket_region":"eu-central-1"}`), true},
		{types.StringValue(`{}`), true},
		{types.StringValue(""), true},
		{types.StringNull(), true},
		{types.StringUnknown(), true},
		{types.StringValue(`{"a":`), false},
		{types.StringValue(`[1,2]`), false},
		{types.StringValue(`null`), false},
	}
	for _, c := range cases {
		resp := &validator.StringResponse{}
		validators.JSONObject().ValidateString(context.Background(), validator.StringRequest{
			Path:        path.Root("data"),
			ConfigValue: c.value,
		}, resp)
		if got := !resp.Diagnostics.HasError(); got != c.valid {
			t.Errorf("JSONObject(%s) valid = %v, want %v", c.value, got, c.valid)
		}
	}
}

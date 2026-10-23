package validators

import (
	"context"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/expanders"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// JSONObject validates that a string attribute holds a JSON object, such as
// iac_input_data.data. Invalid input used to be replaced by {} when building the
// request, so the user's inputs were silently dropped and the apply failed with
// "inconsistent result after apply". Checking it here fails `terraform plan`
// with a message that names the attribute instead. An empty string means no data
// and is accepted; null and unknown values are left to the attribute's other
// rules (an unknown value is checked again when the request is built).
func JSONObject() validator.String {
	return jsonObjectValidator{}
}

type jsonObjectValidator struct{}

func (jsonObjectValidator) Description(_ context.Context) string {
	return "value must be a JSON object"
}

func (v jsonObjectValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (jsonObjectValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if _, err := expanders.JSONStringToObject(req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid JSON object", "The value "+err.Error()+".")
	}
}

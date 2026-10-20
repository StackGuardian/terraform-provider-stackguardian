// Package validators holds plan-time attribute validators shared by resources.
package validators

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = jsonObjectValidator{}

type jsonObjectValidator struct{}

// JSONObject validates that a string attribute holds a JSON object, such as the
// output of jsonencode({ ... }). An empty string is accepted and means "no
// value"; null and unknown values are not checked.
func JSONObject() validator.String {
	return jsonObjectValidator{}
}

func (v jsonObjectValidator) Description(_ context.Context) string {
	return "value must be a JSON object, for example jsonencode({ key = \"value\" })"
}

func (v jsonObjectValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v jsonObjectValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || req.ConfigValue.ValueString() == "" {
		return
	}
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(req.ConfigValue.ValueString()), &obj); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid JSON object",
			fmt.Sprintf("The value must be a JSON object, for example jsonencode({ key = \"value\" }): %s", err))
	}
}

package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Object = iacInputDataValidator{}

type iacInputDataValidator struct{}

// IacInputData validates an iac_input_data object the way StackGuardian does: either empty
// (`{}`, "no inputs"), or with both schema_type and data set. Null objects and objects with an
// unknown attribute are not checked.
func IacInputData() validator.Object {
	return iacInputDataValidator{}
}

func (v iacInputDataValidator) Description(_ context.Context) string {
	return "must be {} for no inputs, or set both schema_type and data"
}

func (v iacInputDataValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v iacInputDataValidator) ValidateObject(_ context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	attrs := req.ConfigValue.Attributes()
	empty := true
	for _, v := range attrs {
		if v.IsUnknown() {
			return
		}
		if !v.IsNull() {
			empty = false
		}
	}
	if empty {
		return
	}
	for _, name := range []string{"schema_type", "data"} {
		if v, ok := attrs[name]; ok && v.IsNull() {
			resp.Diagnostics.AddAttributeError(req.Path.AtName(name), "Incomplete iac_input_data",
				"Set both schema_type and data, or set iac_input_data to {} for no inputs.")
		}
	}
}

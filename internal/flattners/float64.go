package flatteners

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Int64 accepts a `int64` and returns a `basetypes.Int64Value`. The
// return type automatically handles `Int64Null` should the integer be 0.
//
// Removes the need for the following code when saving to state.
//
//	if response.MyField == "" {
//	    state.MyField = types.Int64Value(response.MyField)
//	} else {
//	    state.MyField = types.Int64Null()
//	}
//
// Not recommended if you care about returning an empty string for the state.
//
// nolint: contextcheck
func Float64(in float64) basetypes.Float64Value {
	if in == 0 {
		return types.Float64Null()
	}
	return types.Float64Value(in)
}

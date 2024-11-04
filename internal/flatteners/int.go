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
func Int64(in int64) basetypes.Int64Value {
	if in == 0 {
		return types.Int64Null()
	}
	return types.Int64Value(in)
}

func Int64Ptr(in *int) basetypes.Int64Value {
	if in == nil {
		return types.Int64Null()
	}
	return Int64(int64(*in))
}

func Int32(in int) basetypes.Int32Value {
	//if in == 0 {
	//	return types.Int32Null()
	//}
	return types.Int32Value(int32(in))
}

func Int32Ptr(in *int) basetypes.Int32Value {
	if in == nil {
		return types.Int32Null()
	}
	return Int32(*in)
}

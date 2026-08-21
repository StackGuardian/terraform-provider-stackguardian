package flatteners

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func BoolPtr(in *bool) basetypes.BoolValue {
	if in == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*in)
}

// BoolPtrDefault returns false (not null) when in is nil, otherwise *in. Use
// for Computed bool attributes that must hold a known value in state — a null
// value is skipped by UseStateForUnknown and re-plans as "known after apply"
// forever.
func BoolPtrDefault(in *bool) basetypes.BoolValue {
	if in == nil {
		return types.BoolValue(false)
	}
	return types.BoolValue(*in)
}

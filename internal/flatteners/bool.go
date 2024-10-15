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

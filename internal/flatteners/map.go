package flatteners

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MapStringString converts a map[string]string to types.Map.
// Returns null if the map is nil or empty.
func MapStringString(ctx context.Context, m map[string]string) (types.Map, diag.Diagnostics) {
	if len(m) == 0 {
		return types.MapNull(types.StringType), nil
	}
	return types.MapValueFrom(ctx, types.StringType, m)
}

// MapStringStringOrEmpty converts a map[string]string to types.Map.
// Returns an empty (non-null) map when nil or empty, so UseStateForUnknown
// has a concrete value to carry forward on subsequent plans.
func MapStringStringOrEmpty(ctx context.Context, m map[string]string) (types.Map, diag.Diagnostics) {
	if len(m) == 0 {
		return types.MapValue(types.StringType, map[string]attr.Value{})
	}
	return types.MapValueFrom(ctx, types.StringType, m)
}

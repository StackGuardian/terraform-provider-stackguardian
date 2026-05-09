package flatteners

import (
	"context"

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

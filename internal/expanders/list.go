package expanders

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StringList accepts a `types.List` and returns a slice of strings.
func StringList(ctx context.Context, in types.List) ([]string, diag.Diagnostics) {
	if in.IsNull() || in.IsUnknown() {
		return nil, nil
	}
	results := []string{}
	diags := in.ElementsAs(ctx, &results, false)
	if diags.HasError() {
		return nil, diags
	}
	return results, nil
}

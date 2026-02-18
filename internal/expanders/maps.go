package expanders

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func MapStringString(ctx context.Context, input types.Map) (map[string]string, diag.Diagnostics) {
	if input.IsNull() || input.IsUnknown() {
		return nil, nil
	}

	result := make(map[string]string)
	diags := input.ElementsAs(ctx, &result, false)
	if diags.HasError() {
		return nil, diags
	}

	return result, nil
}

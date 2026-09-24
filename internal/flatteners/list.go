package flatteners

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ListOfStringToTerraformList converts l to a Terraform list: a nil slice (the field
// genuinely absent/null) becomes null, while any non-nil slice — including an empty one —
// becomes a known list ([] for zero-length). Preserving that distinction matters for
// Optional (and Optional+Computed) attributes: an explicitly-configured empty list plans
// as a known, empty value, so collapsing the read-back value to null would disagree with
// that known plan and fail Terraform's post-apply consistency check ("Provider produced
// inconsistent result after apply").
func ListOfStringToTerraformList(l []string) (types.List, diag.Diagnostics) {
	if l == nil {
		return types.ListNull(types.StringType), nil
	}

	terraType, diags := types.ListValueFrom(context.TODO(), types.StringType, l)
	if diags.HasError() {
		return types.ListNull(types.StringType), diags
	}

	return terraType, diags
}

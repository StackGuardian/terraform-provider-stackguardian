package flatteners

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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

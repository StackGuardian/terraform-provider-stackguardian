package sgprovider

import (
	"fmt"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// checkTerraformVersion rejects Terraform CLI releases older than the provider's
// support floor.
//
// The registry has no way to express a minimum Terraform version for a provider,
// so without this a user on an unsupported CLI installs the provider happily and
// then hits whatever obscure failure the old CLI produces. Failing here turns
// "we don't test below MinTerraformVersion" into "it refuses below
// MinTerraformVersion", with a message that says what to do about it.
//
// Prereleases are compared on their core version, so 1.6.0-rc1 counts as 1.6.0.
// This matches how terraform-plugin-testing's tfversion checks treat them, which
// keeps the acceptance tests and the runtime check in agreement.
func checkTerraformVersion(terraformVersion string) diag.Diagnostics {
	var diags diag.Diagnostics

	// Something other than the CLI is driving the provider (the plugin protocol
	// leaves this empty for some callers). There is no version to check.
	if terraformVersion == "" {
		return diags
	}

	current, err := version.NewVersion(terraformVersion)
	if err != nil {
		// An unparseable version is not a reason to refuse to run.
		return diags
	}

	if current.Core().LessThan(constants.MinTerraformVersion) {
		diags.AddError(
			"Unsupported Terraform version",
			fmt.Sprintf(
				"The StackGuardian provider requires Terraform %s or later, but this configuration "+
					"is being run by Terraform %s.\n\n"+
					"Upgrade the Terraform CLI, or pin a provider version released before %s became "+
					"the minimum supported version.",
				constants.MinTerraformVersion, terraformVersion, constants.MinTerraformVersion,
			),
		)
	}

	return diags
}

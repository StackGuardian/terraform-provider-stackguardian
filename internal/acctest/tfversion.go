package acctest

import (
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// VersionChecks returns the Terraform CLI version checks every acceptance test
// case must declare, pinned to the provider's support floor
// (constants.MinTerraformVersion).
//
// This deliberately uses RequireAbove rather than SkipBelow. SkipBelow *passes*
// a test when the CLI is older than the floor, so a CI job that resolves the
// wrong Terraform version would skip all of the tests and report green — a
// false guarantee. RequireAbove fails loudly instead, which is what we want for
// the one version we promise to support.
func VersionChecks() []tfversion.TerraformVersionCheck {
	return []tfversion.TerraformVersionCheck{
		tfversion.RequireAbove(constants.MinTerraformVersion),
	}
}

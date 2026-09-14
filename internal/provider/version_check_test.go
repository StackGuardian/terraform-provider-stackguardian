package sgprovider

import (
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/constants"
)

// TestCheckTerraformVersion guards the support floor without needing TF_ACC or
// API credentials, so it runs on every pull request.
func TestCheckTerraformVersion(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		terraformVersion string
		expectError      bool
	}{
		"below floor":         {terraformVersion: "1.4.6", expectError: true},
		"below floor patch":   {terraformVersion: "1.5.6", expectError: true},
		"far below floor":     {terraformVersion: "1.0.0", expectError: true},
		"at floor":            {terraformVersion: "1.5.7", expectError: false},
		"above floor patch":   {terraformVersion: "1.5.8", expectError: false},
		"above floor minor":   {terraformVersion: "1.14.0", expectError: false},
		"prerelease at floor": {terraformVersion: "1.5.7-rc1", expectError: false},
		"prerelease above":    {terraformVersion: "1.6.0-alpha1", expectError: false},
		"prerelease below":    {terraformVersion: "1.5.6-rc1", expectError: true},
		"empty version":       {terraformVersion: "", expectError: false},
		"unparseable version": {terraformVersion: "not-a-version", expectError: false},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := checkTerraformVersion(testCase.terraformVersion)

			if testCase.expectError && !diags.HasError() {
				t.Fatalf("expected Terraform %s to be rejected (floor is %s), got no error",
					testCase.terraformVersion, constants.MinTerraformVersion)
			}
			if !testCase.expectError && diags.HasError() {
				t.Fatalf("expected Terraform %s to be accepted (floor is %s), got: %v",
					testCase.terraformVersion, constants.MinTerraformVersion, diags.Errors())
			}
		})
	}
}

// TestMinTerraformVersionMatchesDocs fails if the floor is bumped without the
// README being updated to match, so the documented promise and the enforced one
// cannot drift apart.
func TestMinTerraformVersionMatchesDocs(t *testing.T) {
	t.Parallel()

	if got, want := constants.MinTerraformVersion.String(), "1.5.7"; got != want {
		t.Fatalf("MinTerraformVersion is %s, not %s. If this bump is intentional, update "+
			"README.md, the docs-guides-assets required_version constraints, the CI "+
			"terraform matrices in .github/workflows/, and this test.", got, want)
	}
}

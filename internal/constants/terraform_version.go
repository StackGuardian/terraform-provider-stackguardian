package constants

import "github.com/hashicorp/go-version"

// MinTerraformVersion is the oldest Terraform CLI release this provider supports.
//
// The provider refuses to configure below it (internal/provider/version_check.go)
// and the acceptance tests fail — rather than silently skip — below it
// (acctest.VersionChecks).
//
// Raising it is a breaking change for users: bump the provider's major/minor
// version alongside it and note it in CHANGELOG.md.
var MinTerraformVersion = version.Must(version.NewVersion("1.5.7"))

// MaxTerraformVersion is the newest Terraform CLI release CI exercises.
//
// This is not an upper bound on support — the provider imposes no ceiling — it
// is the newest version we have actually proven the provider against. Bump it
// when a new Terraform minor ships.
var MaxTerraformVersion = version.Must(version.NewVersion("1.16.2"))

// SupportedTerraformVersions is the exact Terraform matrix CI runs: the support
// floor, the newest tested release, and the latest patch of every minor series
// in between. Testing the newest patch of each minor is what gives coverage of
// the whole supported range without running every patch release ever published.
//
// The same list is hardcoded into the matrix of both
// .github/workflows/compat.yaml and .github/workflows/test.yaml.
// TestTerraformMatrixMatchesWorkflows fails if they drift apart, so this stays
// the one place to edit when the range changes.
var SupportedTerraformVersions = []string{
	"1.5.7", // MinTerraformVersion
	"1.6.6",
	"1.7.5",
	"1.8.5",
	"1.9.8",
	"1.10.5",
	"1.11.4",
	"1.12.2",
	"1.13.5",
	"1.14.9",
	"1.15.9",
	"1.16.2", // MaxTerraformVersion
}

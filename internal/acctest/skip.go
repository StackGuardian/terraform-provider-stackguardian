package acctest

import (
	"os"
	"os/exec"
	"testing"
)

// SkipUnlessAcceptance skips the calling test unless TF_ACC is set.
//
// resource.Test performs this check itself, but only once it is reached. The
// acceptance tests in this repository create their API fixtures before that
// point, so without this guard a run without credentials fails on a 401 rather
// than skipping. That in turn makes `make test` unusable as a credential-free
// pull request gate.
//
// Call it as the first statement of every TestAcc function.
func SkipUnlessAcceptance(t *testing.T) {
	t.Helper()

	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance test skipped unless env 'TF_ACC' set")
	}
}

// SkipUnlessTerraformCLI skips the calling test unless a Terraform CLI is
// available without a download: on PATH, or named by TF_ACC_TERRAFORM_PATH.
//
// Tests that run terraform against a fake API need no credentials or TF_ACC,
// so they run under `make test`. Without a CLI, terraform-plugin-testing would
// download Terraform for every test, which is slow enough to exceed the unit
// test timeout. Setting TF_ACC_TERRAFORM_VERSION asks for that download
// explicitly, so it is allowed too.
func SkipUnlessTerraformCLI(t *testing.T) {
	t.Helper()

	if os.Getenv("TF_ACC_TERRAFORM_PATH") != "" || os.Getenv("TF_ACC_TERRAFORM_VERSION") != "" {
		return
	}
	if _, err := exec.LookPath("terraform"); err == nil {
		return
	}
	t.Skip("Terraform CLI not found: install it or set TF_ACC_TERRAFORM_PATH to run this test")
}

package acctest

import (
	"os"
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

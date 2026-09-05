package acctest

import (
	"context"
	"os"
	"testing"
)

// TestSweepOrphans deletes resources left behind by an earlier acceptance run.
//
// It is a normal test rather than a terraform-plugin-testing sweeper because the
// sweeper API needs resource.TestMain in every package that registers one, and
// the identifiers these tests create span twenty packages. One entry point is
// easier to reason about, and easier to be careful with.
//
// It does nothing unless SG_ACC_SWEEP is set, so `go test ./...` never sweeps by
// accident, and it only reports unless SG_ACC_SWEEP_APPLY is also set.
//
//	# see what would go
//	SG_ACC_SWEEP=1 go test ./internal/acctest -run TestSweepOrphans -v
//
//	# actually delete it
//	SG_ACC_SWEEP=1 SG_ACC_SWEEP_APPLY=1 go test ./internal/acctest -run TestSweepOrphans -v
func TestSweepOrphans(t *testing.T) {
	if os.Getenv("SG_ACC_SWEEP") == "" {
		t.Skip("set SG_ACC_SWEEP=1 to sweep resources left over from an earlier run")
	}

	for _, v := range []string{"STACKGUARDIAN_API_KEY", "STACKGUARDIAN_API_URI", "STACKGUARDIAN_ORG_NAME"} {
		if os.Getenv(v) == "" {
			t.Fatalf("%s must be set to sweep", v)
		}
	}

	org := os.Getenv("STACKGUARDIAN_ORG_NAME")
	report := Sweep(context.Background(), SGClient(), org)

	t.Logf("org %s\n%s", org, report)

	if report.DryRun && len(report.Swept) > 0 {
		t.Logf("nothing was deleted: set SG_ACC_SWEEP_APPLY=1 to remove the %d resource(s) above",
			len(report.Swept))
	}
	// A listing that failed means the sweep cannot claim the org is clean.
	if len(report.Errors) > 0 {
		t.Fatalf("sweep finished with %d error(s); see the report above", len(report.Errors))
	}
}

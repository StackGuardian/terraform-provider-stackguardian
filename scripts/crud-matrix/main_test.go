package main

import (
	"strings"
	"testing"
)

// info builds a testInfo the way parse() would have collected it: one output
// element per go test -json Output event, each carrying its own newline.
func info(action string, output ...string) *testInfo {
	return &testInfo{pkg: "resource", action: action, output: output}
}

func wantOps(t *testing.T, ops map[string]string, want map[string]string) {
	t.Helper()
	for op, status := range want {
		if got := ops[op]; got != status {
			t.Errorf("%s = %q, want %q", op, got, status)
		}
	}
}

// A gone step whose out-of-band delete fails kills the test in PreConfig,
// without a "Step 4/4 error" line. The steps before it passed, and the
// deferred teardown still exercises delete.
func TestClassifyGonePreConfigFailure(t *testing.T) {
	c := info("fail",
		"=== RUN   TestAccPolicy_CRUD\n",
		"    crud.go:142: crud: resource=stackguardian_policy steps=4\n",
		"    crud.go:129: deleting stackguardian_policy.this outside Terraform: 500 {\"detail\": \"upstream error\"}\n",
		"--- FAIL: TestAccPolicy_CRUD (12.3s)\n",
	)
	resource, ops, reasons := classify(c)
	if resource != "stackguardian_policy" {
		t.Errorf("resource = %q", resource)
	}
	wantOps(t, ops, map[string]string{
		"create": pass, "update": pass, "import": pass, "gone": fail, "delete": pass,
	})
	if got := reasons["gone"]; !strings.Contains(got, "outside Terraform") {
		t.Errorf("gone reason = %q, want the out-of-band delete message", got)
	}
}

// The same failure with a teardown that also failed: both gone and delete
// are reported, with their own reasons.
func TestClassifyGonePreConfigAndDestroyFailure(t *testing.T) {
	c := info("fail",
		"    crud.go:142: crud: resource=stackguardian_policy steps=4\n",
		"    crud.go:129: deleting stackguardian_policy.this outside Terraform: 500 {\"detail\": \"boom\"}\n",
		"    resource.go:77: Error running post-test destroy, err: could not destroy\n",
	)
	_, ops, reasons := classify(c)
	wantOps(t, ops, map[string]string{
		"gone": fail, "delete": fail, "create": pass, "update": pass, "import": pass,
	})
	if reasons["gone"] == "" || reasons["delete"] == "" {
		t.Errorf("reasons = %v, want both gone and delete explained", reasons)
	}
}

// A panic kills the test at a step the output does not identify; create
// carries the failure, and the reason must be the panic line, not the crud:
// layout log that happens to be the first output line.
func TestClassifyPanic(t *testing.T) {
	c := info("fail",
		"=== RUN   TestAccWorkflowGit_CRUD\n",
		"    crud.go:142: crud: resource=stackguardian_workflow_git steps=4\n",
		"panic: runtime error: invalid memory address or nil pointer dereference\n",
		"[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x0]\n",
		"goroutine 42 [running]:\n",
	)
	_, ops, reasons := classify(c)
	wantOps(t, ops, map[string]string{
		"create": fail, "update": notRun, "import": notRun, "gone": notRun, "delete": notRun,
	})
	if got := reasons["create"]; !strings.Contains(got, "panic: runtime error") {
		t.Errorf("create reason = %q, want the panic line", got)
	}
}

// A runtime fatal, same shape as a panic for attribution purposes.
func TestClassifyRuntimeFatal(t *testing.T) {
	c := info("fail",
		"    crud.go:142: crud: resource=stackguardian_stack steps=4\n",
		"fatal error: all goroutines are asleep - deadlock!\n",
	)
	_, ops, reasons := classify(c)
	wantOps(t, ops, map[string]string{"create": fail})
	if got := reasons["create"]; !strings.Contains(got, "fatal error:") {
		t.Errorf("create reason = %q, want the fatal error line", got)
	}
}

// A failure with no recognizable signature at all (here: PreCheck's t.Fatal):
// the reason is the first line that is not framework noise, and never the
// crud: layout log, which says nothing about the failure.
func TestClassifyCatchAllSkipsLayoutLine(t *testing.T) {
	c := info("fail",
		"=== RUN   TestAccRole_CRUD\n",
		"    crud.go:142: crud: resource=stackguardian_role steps=4\n",
		"    acctest.go:24: STACKGUARDIAN_API_KEY must be set for acceptance tests\n",
		"--- FAIL: TestAccRole_CRUD (0.1s)\n",
	)
	_, ops, reasons := classify(c)
	wantOps(t, ops, map[string]string{"create": fail})
	const want = "acctest.go:24: STACKGUARDIAN_API_KEY must be set for acceptance tests"
	if got := reasons["create"]; got != want {
		t.Errorf("create reason = %q, want %q", got, want)
	}
}

// The step-error path is unchanged: step N's operation fails, the ones after
// it did not run, and the reason comes from the step error line.
func TestClassifyStepError(t *testing.T) {
	c := info("fail",
		"    crud.go:142: crud: resource=stackguardian_connector steps=4\n",
		"    resource.go:99: Step 2/4 error: After applying this step, the plan was not empty: [attribute diff]\n",
	)
	_, ops, reasons := classify(c)
	wantOps(t, ops, map[string]string{
		"create": pass, "update": fail, "import": notRun, "gone": notRun, "delete": pass,
	})
	if got := reasons["update"]; !strings.Contains(got, "Step 2/4 error") {
		t.Errorf("update reason = %q, want the step error line", got)
	}
}

// A post-test destroy failure marks delete and nothing else.
func TestClassifyPostTestDestroyFailure(t *testing.T) {
	c := info("fail",
		"    crud.go:142: crud: resource=stackguardian_runner_group steps=4\n",
		"    resource.go:77: Error running post-test destroy, err: could not delete\n",
	)
	_, ops, reasons := classify(c)
	wantOps(t, ops, map[string]string{
		"create": pass, "update": pass, "import": pass, "gone": pass, "delete": fail,
	})
	if got := reasons["delete"]; !strings.Contains(got, "post-test destroy") {
		t.Errorf("delete reason = %q, want the destroy error line", got)
	}
}

// A three-step resource has no gone step; a step-3 failure leaves gone n/a.
func TestClassifyNoGoneStep(t *testing.T) {
	c := info("fail",
		"    crud.go:142: crud: resource=stackguardian_workflow_group steps=3\n",
		"    resource.go:99: Step 3/3 error: ImportState error: not found\n",
	)
	_, ops, _ := classify(c)
	wantOps(t, ops, map[string]string{
		"create": pass, "update": pass, "import": fail, "gone": na, "delete": pass,
	})
}

// A skipped test marks every operation skipped, including gone.
func TestClassifySkip(t *testing.T) {
	c := info("skip",
		"=== RUN   TestAccStack_CRUD\n",
		"--- SKIP: TestAccStack_CRUD (0.0s) acceptance tests disabled\n",
	)
	_, ops, reasons := classify(c)
	wantOps(t, ops, map[string]string{
		"create": skip, "update": skip, "import": skip, "gone": skip, "delete": skip,
	})
	if len(reasons) != 0 {
		t.Errorf("reasons = %v, want none", reasons)
	}
}

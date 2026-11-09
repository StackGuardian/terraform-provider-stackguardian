package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// info builds a testInfo the way parse() would have collected it: one output
// element per go test -json Output event, each carrying its own newline, and
// the test package's last path segment as pkg.
func info(pkg, action string, output ...string) *testInfo {
	return &testInfo{pkg: pkg, action: action, output: output}
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
	c := info("policy", "fail",
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
	c := info("policy", "fail",
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
	c := info("workflow_git", "fail",
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
	c := info("stack", "fail",
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
	c := info("role", "fail",
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
	c := info("connector", "fail",
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
	c := info("runner_group", "fail",
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
	c := info("workflow_group", "fail",
		"    crud.go:142: crud: resource=stackguardian_workflow_group steps=3\n",
		"    resource.go:99: Step 3/3 error: ImportState error: not found\n",
	)
	_, ops, _ := classify(c)
	wantOps(t, ops, map[string]string{
		"create": pass, "update": pass, "import": fail, "gone": na, "delete": pass,
	})
}

// A test that skips inside RunCRUD emits the crud: layout line first, so its
// row carries the resource name from the line like any other row.
func TestClassifySkipWithLayoutLine(t *testing.T) {
	c := info("policy", "skip",
		"=== RUN   TestAccPolicy_CRUD\n",
		"    crud.go:142: crud: resource=stackguardian_policy steps=4\n",
		"--- SKIP: TestAccPolicy_CRUD (0.0s) Acceptance test skipped unless env 'TF_ACC' set\n",
	)
	resource, ops, reasons := classify(c)
	if resource != "stackguardian_policy" {
		t.Errorf("resource = %q, want stackguardian_policy", resource)
	}
	wantOps(t, ops, map[string]string{
		"create": skip, "update": skip, "import": skip, "gone": skip, "delete": skip,
	})
	if len(reasons) != 0 {
		t.Errorf("reasons = %v, want none", reasons)
	}
}

// A test that skips before RunCRUD (the fixture-creating CRUD tests) has no
// layout line; the row falls back to the Terraform type derived from the
// package directory.
func TestClassifySkipWithoutLayoutLine(t *testing.T) {
	c := info("connector", "skip",
		"=== RUN   TestAccConnector_CRUD\n",
		"--- SKIP: TestAccConnector_CRUD (0.0s) Acceptance test skipped unless env 'TF_ACC' set\n",
	)
	resource, ops, _ := classify(c)
	if resource != "stackguardian_connector" {
		t.Errorf("resource = %q, want stackguardian_connector", resource)
	}
	wantOps(t, ops, map[string]string{
		"create": skip, "update": skip, "import": skip, "gone": skip, "delete": skip,
	})
}

// role_v4 is the one package directory whose Terraform type is not
// "stackguardian_" + directory.
func TestClassifyFallbackRoleV4(t *testing.T) {
	c := info("role_v4", "skip",
		"--- SKIP: TestAccRoleV4_CRUD (0.0s) Acceptance test skipped unless env 'TF_ACC' set\n",
	)
	resource, _, _ := classify(c)
	if resource != "stackguardian_rolev4" {
		t.Errorf("resource = %q, want stackguardian_rolev4", resource)
	}
}

// A test that started but never reported a terminal result. A timeout panics
// the test binary, and go test reports that at the package level, so the
// running test gets no fail event of its own. Nothing is known to have passed;
// create carries the failure and the panic line is the reason.
func TestClassifyInterruptedRun(t *testing.T) {
	c := info("workflow_template_revision", "",
		"=== RUN   TestAccWorkflowTemplateRevision_CRUD\n",
		"    crud.go:150: crud: resource=stackguardian_workflow_template_revision steps=4\n",
		"panic: test timed out after 20m0s\n",
		"\trunning tests:\n",
		"\t\tTestAccWorkflowTemplateRevision_CRUD (20m0s)\n",
	)
	_, ops, reasons := classify(c)
	wantOps(t, ops, map[string]string{
		"create": fail, "update": notRun, "import": notRun, "gone": notRun, "delete": notRun,
	})
	if got := reasons["create"]; !strings.Contains(got, "panic: test timed out") {
		t.Errorf("create reason = %q, want the timeout panic line", got)
	}
}

// The same interrupted run for a resource without a gone step.
func TestClassifyInterruptedRunNoGoneStep(t *testing.T) {
	c := info("workflow_group", "",
		"    crud.go:150: crud: resource=stackguardian_workflow_group steps=3\n",
		"panic: test timed out after 20m0s\n",
	)
	_, ops, _ := classify(c)
	wantOps(t, ops, map[string]string{
		"create": fail, "update": notRun, "import": notRun, "gone": na, "delete": notRun,
	})
}

// writeEvents writes one go test -json event per argument to a file and returns
// its path, so parse() can be exercised with real event shapes.
func writeEvents(t *testing.T, events ...string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "go-test.json")
	if err := os.WriteFile(path, []byte(strings.Join(events, "")), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

// One passing CRUD test plus another package's build failure: the build
// failure emits build-output events (keyed by ImportPath, not Package) and a
// package fail, but no test events at all. It must surface as a package
// failure with the compile error as its reason, or the matrix would exit 0.
func TestParsePackageBuildFailure(t *testing.T) {
	path := writeEvents(t,
		`{"ImportPath":"example.com/p/internal/resource/bad [example.com/p/internal/resource/bad.test]","Action":"build-output","Output":"# example.com/p/internal/resource/bad [example.com/p/internal/resource/bad.test]\n"}`,
		"\n",
		`{"ImportPath":"example.com/p/internal/resource/bad [example.com/p/internal/resource/bad.test]","Action":"build-output","Output":"bad/resource_test.go:42:5: undefined: deleteBadFixture\n"}`,
		"\n",
		`{"ImportPath":"example.com/p/internal/resource/bad [example.com/p/internal/resource/bad.test]","Action":"build-fail"}`,
		"\n",
		`{"Action":"output","Package":"example.com/p/internal/resource/bad","Output":"FAIL\texample.com/p/internal/resource/bad [build failed]\n"}`,
		"\n",
		`{"Action":"fail","Package":"example.com/p/internal/resource/bad","Elapsed":0,"FailedBuild":"example.com/p/internal/resource/bad [example.com/p/internal/resource/bad.test]"}`,
		"\n",
		`{"Action":"start","Package":"example.com/p/internal/resource/policy"}`,
		"\n",
		`{"Action":"run","Package":"example.com/p/internal/resource/policy","Test":"TestAccPolicy_CRUD"}`,
		"\n",
		`{"Action":"output","Package":"example.com/p/internal/resource/policy","Test":"TestAccPolicy_CRUD","Output":"    crud.go:150: crud: resource=stackguardian_policy steps=4\n"}`,
		"\n",
		`{"Action":"pass","Package":"example.com/p/internal/resource/policy","Test":"TestAccPolicy_CRUD","Elapsed":12.3}`,
		"\n",
		`{"Action":"pass","Package":"example.com/p/internal/resource/policy","Elapsed":12.9}`,
		"\n",
	)
	tests, pkgFails := parse(path)
	if len(tests) != 1 {
		t.Fatalf("parse found %d tests, want 1 (the passing CRUD test)", len(tests))
	}
	if got := tests["TestAccPolicy_CRUD"]; got == nil || got.action != pass {
		t.Errorf("TestAccPolicy_CRUD action = %v, want pass", got)
	}
	if len(pkgFails) != 1 {
		t.Fatalf("parse found %d package failures, want 1", len(pkgFails))
	}
	pf := pkgFails[0]
	if pf.pkg != "example.com/p/internal/resource/bad" {
		t.Errorf("package failure pkg = %q", pf.pkg)
	}
	if !strings.Contains(pf.reason, "undefined: deleteBadFixture") {
		t.Errorf("package failure reason = %q, want the compile error", pf.reason)
	}
}

// A timed-out CRUD test: output attributed to the running test, including the
// panic line, but no test-level terminal event -- only the package fails. The
// row keeps its (empty) action for classify() to fail, and the package
// failure is reported too.
func TestParseTimedOutTest(t *testing.T) {
	path := writeEvents(t,
		`{"Action":"run","Package":"example.com/p/internal/resource/slow","Test":"TestAccSlow_CRUD"}`,
		"\n",
		`{"Action":"output","Package":"example.com/p/internal/resource/slow","Test":"TestAccSlow_CRUD","Output":"    crud.go:150: crud: resource=stackguardian_slow steps=4\n"}`,
		"\n",
		`{"Action":"output","Package":"example.com/p/internal/resource/slow","Test":"TestAccSlow_CRUD","Output":"panic: test timed out after 2s\n"}`,
		"\n",
		`{"Action":"output","Package":"example.com/p/internal/resource/slow","Output":"FAIL\texample.com/p/internal/resource/slow\t2.8s\n"}`,
		"\n",
		`{"Action":"fail","Package":"example.com/p/internal/resource/slow","Elapsed":2.8}`,
		"\n",
	)
	tests, pkgFails := parse(path)
	ti, ok := tests["TestAccSlow_CRUD"]
	if !ok {
		t.Fatal("parse did not collect TestAccSlow_CRUD")
	}
	if ti.action != "" {
		t.Errorf("action = %q, want empty (no terminal event arrived)", ti.action)
	}
	if len(pkgFails) != 1 || pkgFails[0].pkg != "example.com/p/internal/resource/slow" {
		t.Fatalf("package failures = %v, want the slow package", pkgFails)
	}
	if pkgFails[0].reason != "no test reported a result" {
		t.Errorf("reason = %q, want the fallback wording", pkgFails[0].reason)
	}
}

// An ordinary test failure also produces a package fail event, but the
// package failure must not be reported: the test's own row already covers it.
func TestParseTestFailureIsNotPackageFailure(t *testing.T) {
	path := writeEvents(t,
		`{"Action":"run","Package":"example.com/p/internal/resource/policy","Test":"TestAccPolicy_CRUD"}`,
		"\n",
		`{"Action":"output","Package":"example.com/p/internal/resource/policy","Test":"TestAccPolicy_CRUD","Output":"    resource.go:99: Step 1/4 error: Error: 409 conflict\n"}`,
		"\n",
		`{"Action":"fail","Package":"example.com/p/internal/resource/policy","Test":"TestAccPolicy_CRUD","Elapsed":3.2}`,
		"\n",
		`{"Action":"output","Package":"example.com/p/internal/resource/policy","Output":"FAIL\texample.com/p/internal/resource/policy\t3.2s\n"}`,
		"\n",
		`{"Action":"fail","Package":"example.com/p/internal/resource/policy","Elapsed":3.4}`,
		"\n",
	)
	tests, pkgFails := parse(path)
	if got := tests["TestAccPolicy_CRUD"]; got == nil || got.action != fail {
		t.Errorf("TestAccPolicy_CRUD action = %v, want fail", got)
	}
	if len(pkgFails) != 0 {
		t.Errorf("package failures = %v, want none (the test row reports it)", pkgFails)
	}
}

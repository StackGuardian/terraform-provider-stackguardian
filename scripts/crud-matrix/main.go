// crud-matrix turns `go test -json` output of the TestAcc*_CRUD tests into a
// resource x operation matrix.
//
// The step layout is defined in internal/acctest/crud.go (CRUDStep*): a failing
// "Step N/M" maps to the operation at position N; the steps before it passed
// and the ones after it did not run. A "post-test destroy" error marks delete
// as failed. Exits 1 if any CRUD test failed, if a package failed without any
// of its tests reporting a result (a build failure), or if a test ended
// without a terminal result (an interrupted run, e.g. a timeout: go test
// reports the timeout at the package level, so the running test never gets a
// fail event).
//
// Two failure shapes bypass the "Step N/M" machinery and are attributed from
// their own signature lines instead: the gone step's PreConfig t.Fatalf (the
// out-of-band delete failed, so the step never ran -- create, update and
// import passed by then, and the deferred teardown still exercises delete),
// and a panic or runtime fatal, which kills the test at a step the output
// does not identify.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

var ops = []string{"create", "update", "import", "gone", "delete"}

var stepOps = []string{"create", "update", "import", "gone"}

var stepErrRe = regexp.MustCompile(`Step (\d+)/(\d+) error`)
var layoutRe = regexp.MustCompile(`crud: resource=(\S+) steps=(\d+)`)

// The gone step's PreConfig t.Fatalf, when the out-of-band delete fails
// (internal/acctest/crud.go). The step never runs, so no "Step 4/4 error"
// line exists; this message is the only place "outside Terraform" appears in
// test output.
var gonePreConfigRe = regexp.MustCompile(`deleting \S+ outside Terraform:`)

// A panic or runtime fatal (a test timeout reports as "panic: test timed
// out") also kills the test without a step error line. Matched against
// trimmed lines, so t.Log output with its "file:line:" prefix cannot match.
var panicRe = regexp.MustCompile(`^(panic:|fatal error:)`)
var apiErrRe = regexp.MustCompile(`(\d{3}): (\{.*)`)

// go-cmp, which prints the ImportStateVerify diff, randomly uses U+00A0 in
// place of spaces to discourage parsing its output; RE2's \s does not match it.
var diffLineRe = regexp.MustCompile(`^[\s\x{00a0}]*([+-])[\s\x{00a0}]+"([^"]+)":[\s\x{00a0}]*(".*?"),?$`)

const destroyErr = "Error running post-test destroy"

const (
	pass   = "pass"
	fail   = "fail"
	notRun = "not run"
	na     = "n/a"
	skip   = "skip"
)

var symbols = map[string]string{
	pass:   "\u2705",
	fail:   "\u274c",
	notRun: "\u23ed\ufe0f",
	na:     "\u2014",
	skip:   "\u26aa",
}

type testEvent struct {
	Test    string `json:"Test"`
	Package string `json:"Package"`
	// ImportPath is set instead of Package on build-output events: go test
	// attributes a compile failure to the test binary it could not build,
	// not to the package ("pkg [pkg.test]").
	ImportPath string  `json:"ImportPath"`
	Action     string  `json:"Action"`
	Output     string  `json:"Output"`
	Elapsed    float64 `json:"Elapsed"`
	// FailedBuild marks a package fail event caused by a build failure.
	FailedBuild string `json:"FailedBuild"`
}

type testInfo struct {
	pkg     string
	output  []string
	action  string
	elapsed float64
}

// pkgFailure is a package that failed without any of its tests reporting a
// terminal result: a build failure emits build-output events and a package
// fail but no test events at all, and a run that dies mid-test (a timeout
// panics the test binary) emits the running test's output but no test-level
// pass or fail -- only the package gets the fail event.
type pkgFailure struct {
	pkg    string
	reason string
}

// pkgState accumulates a package's Test-free events, to attribute a
// package-level failure none of its tests reported.
type pkgState struct {
	output []string
	// terminal records that some test in the package reported a result, so
	// a package fail event is already covered by that test's row.
	terminal bool
	failed   bool
}

type row struct {
	Resource string            `json:"resource"`
	Test     string            `json:"test"`
	Result   string            `json:"result"`
	Seconds  float64           `json:"seconds"`
	Ops      map[string]string `json:"ops"`
	Reasons  map[string]string `json:"reasons"`
}

// eventPkg returns the package an event belongs to: the Package field for test
// and package events, or the ImportPath of build-output events, which carry it
// instead, in the "pkg [pkg.test]" form -- the bracketed part names the test
// binary and is stripped.
func eventPkg(ev testEvent) string {
	if ev.Package != "" {
		return ev.Package
	}
	if i := strings.Index(ev.ImportPath, " ["); i >= 0 {
		return ev.ImportPath[:i]
	}
	return ev.ImportPath
}

func parse(path string) (map[string]*testInfo, []pkgFailure) {
	f, err := os.Open(path)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error opening %s: %v\n", path, err)
		os.Exit(2)
	}
	defer f.Close()

	tests := make(map[string]*testInfo)
	pkgs := make(map[string]*pkgState)
	pkgStateFor := func(ev testEvent) *pkgState {
		key := eventPkg(ev)
		if key == "" {
			return nil
		}
		st, ok := pkgs[key]
		if !ok {
			st = &pkgState{}
			pkgs[key] = st
		}
		return st
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)
	for scanner.Scan() {
		var ev testEvent
		if err = json.Unmarshal(scanner.Bytes(), &ev); err != nil {
			continue
		}
		if ev.Test != "" {
			name := ev.Test
			if strings.Contains(name, "/") || !strings.HasSuffix(name, "_CRUD") {
				continue
			}
			t, ok := tests[name]
			if !ok {
				pkg := ev.Package
				if idx := strings.LastIndex(pkg, "/"); idx >= 0 {
					pkg = pkg[idx+1:]
				}
				t = &testInfo{pkg: pkg}
				tests[name] = t
			}
			switch ev.Action {
			case "output":
				t.output = append(t.output, ev.Output)
			case "pass", "fail", "skip":
				t.action = ev.Action
				t.elapsed = ev.Elapsed
				if st := pkgStateFor(ev); st != nil {
					st.terminal = true
				}
			}
			continue
		}
		// A package-level event: a terminal result, build output, or the
		// FAIL summary line. Dropped by the test branch above, but a
		// package can fail with no test to report it.
		st := pkgStateFor(ev)
		if st == nil {
			continue
		}
		switch ev.Action {
		case "output", "build-output":
			st.output = append(st.output, ev.Output)
		case "fail", "build-fail":
			st.failed = true
		}
	}

	var pkgFails []pkgFailure
	for key, st := range pkgs {
		if st.failed && !st.terminal {
			pkgFails = append(pkgFails, pkgFailure{pkg: key, reason: packageReason(st.output)})
		}
	}
	sort.Slice(pkgFails, func(i, j int) bool { return pkgFails[i].pkg < pkgFails[j].pkg })
	return tests, pkgFails
}

// packageReason extracts the diagnostic from a failed package's own output: a
// build failure's compile error. The FAIL summary line carries no reason, and
// a timeout's package output holds nothing but that summary.
func packageReason(lines []string) string {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if t == "" || strings.HasPrefix(t, "#") || strings.HasPrefix(t, "FAIL") ||
			strings.HasPrefix(t, "ok  ") || strings.HasPrefix(t, "PASS") ||
			strings.HasPrefix(t, "no tests") {
			continue
		}
		return truncate(t, 200)
	}
	return "no test reported a result"
}

// matchIndex returns the index of the first line matching, or -1.
func matchIndex(lines []string, match func(string) bool) int {
	for i, l := range lines {
		if match(l) {
			return i
		}
	}
	return -1
}

func firstError(lines []string, start int) string {
	head := strings.TrimSpace(lines[start])
	end := min(start+40, len(lines))
	block := lines[start:end]

	if strings.Contains(head, "not equivalent") {
		var diff []string
		for _, l := range block {
			if m := diffLineRe.FindStringSubmatch(l); m != nil {
				diff = append(diff, m[1]+m[2]+"="+m[3])
			}
		}
		return "import state differs (- state, + imported): " + strings.Join(diff, ", ")
	}

	if strings.Contains(head, "still exists") {
		idx := strings.Index(head, "stackguardian_")
		if idx >= 0 {
			s := head[idx:]
			if len(s) > 200 {
				s = s[:200]
			}
			return s
		}
	}

	joined := strings.Join(func() []string {
		out := make([]string, len(block))
		for i, l := range block {
			out[i] = strings.TrimSpace(l)
		}
		return out
	}(), " ")

	if m := apiErrRe.FindString(joined); m != "" {
		sm := apiErrRe.FindStringSubmatch(joined)
		body := sm[2]
		if idx := strings.Index(body, "}"); idx >= 0 {
			body = body[:idx+1]
		}
		prefix := "API error"
		if strings.Contains(head, "pre-apply plan") || strings.Contains(head, "refresh") {
			prefix = "refresh failed"
		} else if strings.Contains(head, "post-test destroy") {
			prefix = "destroy failed"
		}
		result := fmt.Sprintf("%s: %s %s", prefix, sm[1], body)
		if len(result) > 220 {
			result = result[:220]
		}
		return result
	}

	if strings.Contains(joined, "ExpectResourceAction") || strings.Contains(joined, "expected") {
		re := regexp.MustCompile(`'?[\w.\-]+'? - expected \w+, got action\(s\): \[[^\]]*\]`)
		if m := re.FindString(joined); m != "" {
			return m
		}
	}

	if len(head) > 200 {
		return head[:200]
	}
	return head
}

func classify(t *testInfo) (string, map[string]string, map[string]string) {
	out := strings.Join(t.output, "")
	lines := strings.Split(out, "\n")

	resource := resourceFromPkg(t.pkg)
	nSteps := 3
	if m := layoutRe.FindStringSubmatch(out); m != nil {
		resource = m[1]
		nSteps = atoiSafe(m[2])
	}

	opsMap := make(map[string]string)
	for _, op := range ops {
		opsMap[op] = pass
	}
	if nSteps < 4 {
		opsMap["gone"] = na
	}
	reasons := make(map[string]string)

	if t.action == "skip" {
		for _, op := range ops {
			opsMap[op] = skip
		}
		return resource, opsMap, reasons
	}

	if t.action == "" {
		// The test started but never reported a terminal result: an
		// interrupted run. A timeout, for one, panics the test binary and
		// go test reports it at the package level, so the running test
		// gets no fail event of its own. Where it died is unknown, so
		// nothing is marked passed; create carries the failure and the
		// reason, like the panic branch below.
		for _, op := range ops {
			opsMap[op] = notRun
		}
		if nSteps < 4 {
			opsMap["gone"] = na
		}
		reasons["create"] = catchAllReason(lines)
		opsMap["create"] = fail
		return resource, opsMap, reasons
	}

	if m := stepErrRe.FindStringSubmatch(out); m != nil {
		failedStep := atoiSafe(m[1])
		opsMap[stepOps[failedStep-1]] = fail
		for i := failedStep; i < nSteps && i < len(stepOps); i++ {
			opsMap[stepOps[i]] = notRun
		}
		if idx := matchIndex(lines, stepErrRe.MatchString); idx >= 0 {
			reasons[stepOps[failedStep-1]] = firstError(lines, idx)
		}
	}

	// The gone step's PreConfig deletes the object behind Terraform's back;
	// when that fails, its t.Fatalf kills the test without a "Step 4/4 error"
	// line, because the step never runs. Create, update and import passed by
	// then, and the deferred teardown still exercises delete, which the
	// post-test destroy branch below reports on.
	if idx := matchIndex(lines, gonePreConfigRe.MatchString); idx >= 0 {
		opsMap["gone"] = fail
		reasons["gone"] = truncate(strings.TrimSpace(lines[idx]), 200)
	}

	if strings.Contains(out, destroyErr) {
		opsMap["delete"] = fail
		if idx := matchIndex(lines, func(l string) bool { return strings.Contains(l, destroyErr) }); idx >= 0 {
			reasons["delete"] = firstError(lines, idx)
		}
	}

	if t.action == "fail" && len(reasons) == 0 {
		// No step, PreConfig or destroy line identified the failure: a panic
		// or runtime fatal killed the test at a step the output does not
		// reveal. Where it happened is unknown, so every operation is "not
		// run" except create, which carries the failure and its reason.
		for _, op := range ops {
			opsMap[op] = notRun
		}
		reasons["create"] = catchAllReason(lines)
		opsMap["create"] = fail
	}

	return resource, opsMap, reasons
}

// catchAllReason diagnoses a failure that produced no identifiable step,
// PreConfig or destroy line. The first output line of every test is the
// "crud: resource=… steps=…" layout log, which says nothing about the
// failure, so prefer the panic or fatal line; otherwise take the first line
// that is not test-framework noise.
func catchAllReason(lines []string) string {
	if idx := matchIndex(lines, panicRe.MatchString); idx >= 0 {
		return truncate(strings.TrimSpace(lines[idx]), 200)
	}
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if t == "" || strings.HasPrefix(t, "===") || strings.HasPrefix(t, "---") || layoutRe.MatchString(t) {
			continue
		}
		return truncate(t, 200)
	}
	return "failed"
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

func atoiSafe(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return n
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// resourceFromPkg derives the Terraform type from the test package directory,
// for rows with no "crud: resource=…" layout line -- a test that skipped
// before RunCRUD ran. SkipUnlessAcceptance is the first statement of every
// TestAcc*_CRUD test (before any API call or cleanup registration), so a run
// without TF_ACC reaches this path for every row. The directory names the
// resource in every package; role_v4 is the one exception, whose Terraform
// type is stackguardian_rolev4 (see CLAUDE.md).
func resourceFromPkg(pkg string) string {
	if pkg == "role_v4" {
		return "stackguardian_rolev4"
	}
	return "stackguardian_" + pkg
}

func main() {
	var input, markdownOut, jsonOut string
	flag.StringVar(&markdownOut, "markdown", "", "write markdown table to this file")
	flag.StringVar(&jsonOut, "json", "", "write JSON results to this file")
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: %s <go-test.json> [--markdown out.md] [--json out.json]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(2)
	}
	input = args[0]
	// flag stops at the first positional argument; parse the flags after the
	// input file too, so the documented "<input> --markdown x" order works.
	if err := flag.CommandLine.Parse(args[1:]); err != nil {
		os.Exit(2)
	}

	tests, pkgFails := parse(input)
	if len(tests) == 0 && len(pkgFails) == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "no TestAcc*_CRUD results found in %s\n", input)
		os.Exit(1)
	}

	type sortedEntry struct {
		name string
		t    *testInfo
	}
	var entries []sortedEntry
	for name, t := range tests {
		entries = append(entries, sortedEntry{name, t})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].t.pkg < entries[j].t.pkg
	})

	var rows []row
	for _, e := range entries {
		resource, opsMap, reasons := classify(e.t)
		result := e.t.action
		if result == "" {
			// classify() reports an interrupted test (no terminal
			// result) as failed; the row must agree, and the failed
			// count below keys off this field.
			result = fail
		}
		sec := e.t.elapsed
		sec = float64(int(sec*10)) / 10
		rows = append(rows, row{
			Resource: resource,
			Test:     e.name,
			Result:   result,
			Seconds:  sec,
			Ops:      opsMap,
			Reasons:  reasons,
		})
	}

	var md []string
	header := "| Resource | " + strings.Join(ops, " | ") + " | Time |"
	md = append(md, header)
	var sep strings.Builder
	sep.WriteString("|---|")
	for range ops {
		sep.WriteString("---|")
	}
	sep.WriteString("---|")
	md = append(md, sep.String())

	for _, r := range rows {
		cells := make([]string, len(ops))
		for i, op := range ops {
			cells[i] = symbols[r.Ops[op]]
		}
		md = append(md, fmt.Sprintf("| `%s` | %s | %.1fs |", r.Resource, strings.Join(cells, " | "), r.Seconds))
	}

	var failed []row
	passed := 0
	for _, r := range rows {
		if r.Result == "fail" {
			failed = append(failed, r)
		}
		if r.Result == "pass" {
			passed++
		}
	}

	md = append(md, "")
	md = append(md, fmt.Sprintf("%d/%d resources passed. \u2705 pass \u00b7 \u274c fail \u00b7 \u23ed\ufe0f not run (an earlier step failed) \u00b7 \u2014 not applicable \u00b7 \u26aa skipped", passed, len(rows)))
	md = append(md, "`gone` = object deleted outside Terraform; refresh must drop it and plan a recreate.")
	if len(pkgFails) > 0 {
		md = append(md, fmt.Sprintf("%d package(s) failed with no test result (build failure or interrupted run).", len(pkgFails)))
	}
	if len(failed) > 0 || len(pkgFails) > 0 {
		md = append(md, "", "### Failures", "")
		for _, r := range failed {
			for _, op := range ops {
				if why, ok := r.Reasons[op]; ok {
					md = append(md, fmt.Sprintf("- `%s` **%s** (`%s`): %s", r.Resource, op, r.Test, why))
				}
			}
		}
		for _, pf := range pkgFails {
			md = append(md, fmt.Sprintf("- package `%s` failed with no test result: %s", pf.pkg, pf.reason))
		}
	}

	text := strings.Join(md, "\n")
	fmt.Println(text)

	if markdownOut != "" {
		content := "# Provider CRUD matrix\n\n" + text + "\n"
		if err := os.WriteFile(markdownOut, []byte(content), 0644); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error writing markdown: %v\n", err)
			os.Exit(2)
		}
	}

	if jsonOut != "" {
		data, err := json.MarshalIndent(rows, "", "  ")
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error marshalling JSON: %v\n", err)
			os.Exit(2)
		}
		if err = os.WriteFile(jsonOut, data, 0644); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error writing JSON: %v\n", err)
			os.Exit(2)
		}
	}

	if len(failed) > 0 || len(pkgFails) > 0 {
		os.Exit(1)
	}
}

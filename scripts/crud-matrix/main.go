// crud-matrix turns `go test -json` output of the TestAcc*_CRUD tests into a
// resource x operation matrix.
//
// The step layout is defined in internal/acctest/crud.go (CRUDStep*): a failing
// "Step N/M" maps to the operation at position N; the steps before it passed
// and the ones after it did not run. A "post-test destroy" error marks delete
// as failed. Exits 1 if any CRUD test failed.
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
	Test    string  `json:"Test"`
	Package string  `json:"Package"`
	Action  string  `json:"Action"`
	Output  string  `json:"Output"`
	Elapsed float64 `json:"Elapsed"`
}

type testInfo struct {
	pkg     string
	output  []string
	action  string
	elapsed float64
}

type row struct {
	Resource string            `json:"resource"`
	Test     string            `json:"test"`
	Result   string            `json:"result"`
	Seconds  float64           `json:"seconds"`
	Ops      map[string]string `json:"ops"`
	Reasons  map[string]string `json:"reasons"`
}

func parse(path string) map[string]*testInfo {
	f, err := os.Open(path)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error opening %s: %v\n", path, err)
		os.Exit(2)
	}
	defer f.Close()

	tests := make(map[string]*testInfo)
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)
	for scanner.Scan() {
		var ev testEvent
		if err = json.Unmarshal(scanner.Bytes(), &ev); err != nil {
			continue
		}
		name := ev.Test
		if name == "" || strings.Contains(name, "/") || !strings.HasSuffix(name, "_CRUD") {
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
		}
	}
	return tests
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

	resource := t.pkg
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

	tests := parse(input)
	if len(tests) == 0 {
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
		sec := e.t.elapsed
		sec = float64(int(sec*10)) / 10
		rows = append(rows, row{
			Resource: resource,
			Test:     e.name,
			Result:   e.t.action,
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
	if len(failed) > 0 {
		md = append(md, "", "### Failures", "")
		for _, r := range failed {
			for _, op := range ops {
				if why, ok := r.Reasons[op]; ok {
					md = append(md, fmt.Sprintf("- `%s` **%s** (`%s`): %s", r.Resource, op, r.Test, why))
				}
			}
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

	if len(failed) > 0 {
		os.Exit(1)
	}
}

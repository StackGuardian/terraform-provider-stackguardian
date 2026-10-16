// extract-doc-blocks extracts every ```terraform block from the documentation
// prose and classifies it.
//
// The examples under docs-examples/ are whole files and are already type-checked
// by validate-examples.sh. The blocks *inside* prose are not, and they are not
// uniform, so each kind has to be checked the way it can actually be checked:
//
//	standalone  declares at least one top-level block (resource, data, locals, ...).
//	            Type-checked by `terraform validate` once a provider block is
//	            injected, exactly like a docs-examples fragment.
//
//	mixed       a bare attribute assignment *and* a block in the same fence. The
//	            attribute is looked up in the schema; the block is validated.
//
//	import      contains only `import {}` blocks. Terraform rejects these unless the
//	            resource named in `to` is also in the configuration, so validating one
//	            alone reports a failure that says nothing about the example. The `to`
//	            address is checked against the provider schema instead, which is the
//	            part that actually goes stale.
//
//	fragment    a bare attribute assignment shown out of context, e.g.
//	            `iac_template_id = "my-template:1"`. Not valid at the top level of a
//	            .tf file, so its attribute names are looked up in the schema.
//
// Blocks in a guide are cumulative: a later fence refers to a resource an earlier one
// declared. Validating each fence alone would report "Reference to undeclared
// resource" for examples that are correct as written, so every block that can be
// validated is given the earlier blocks of the same page as context.tf. Where a page
// redeclares an address, the latest definition wins, which is what a reader following
// along would end up with.
//
// Blocks are written to --out as per-block directories and described in a
// tab-separated manifest, so the shell harness can drive `terraform validate` without
// parsing JSON. Blocks that cannot be validated as configuration are adjudicated here
// against the schema in --schema and carry their verdict in the manifest.
//
// Manifest columns:
//
//	index  kind  source  line  dir  verdict  detail
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var configBlocks = map[string]bool{
	"resource": true, "data": true, "module": true, "variable": true,
	"output": true, "locals": true, "provider": true, "terraform": true,
}

var addressed = map[string]bool{
	"resource": true, "data": true, "module": true, "variable": true, "output": true,
}

const (
	fenceOpen  = "```terraform"
	fenceClose = "```"
)

var identRe = regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_-]*)`)
var blockHeaderRe = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_-]*)\s+"`)
var blockBraceRe = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_-]*)\s*\{`)
var attrRe = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_-]*)\s*=`)
var identAssignOrBraceRe = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_-]*)\s*(=|\{)`)
var elidedRe = regexp.MustCompile(`^(#|//)\s*\.{3}\.*$`)
var quotedRe = regexp.MustCompile(`"([^"]+)"`)
var importToRe = regexp.MustCompile(`(?m)^\s*to\s*=\s*([A-Za-z_][A-Za-z0-9_]*)\.`)

type chunk struct {
	kind string
	name string
	text string
}

type block struct {
	source string
	line   int
	body   string
}

func braceDelta(line string) int {
	var kept []rune
	inString := false
	runes := []rune(line)
	i := 0
	for i < len(runes) {
		c := runes[i]
		if inString {
			if c == '\\' {
				i += 2
				continue
			}
			if c == '"' {
				inString = false
			}
			i++
			continue
		}
		if c == '"' {
			inString = true
			i++
			continue
		}
		if c == '#' {
			break
		}
		if c == '/' && i+1 < len(runes) && runes[i+1] == '/' {
			break
		}
		kept = append(kept, c)
		i++
	}
	return strings.Count(string(kept), "{") - strings.Count(string(kept), "}")
}

func splitChunks(body string) []chunk {
	lines := strings.Split(body, "\n")
	var chunks []chunk
	var pending []string
	i := 0
	for i < len(lines) {
		stripped := strings.TrimSpace(lines[i])
		if stripped == "" || strings.HasPrefix(stripped, "#") || strings.HasPrefix(stripped, "//") {
			pending = append(pending, lines[i])
			i++
			continue
		}

		var blockMatch []string
		if m := blockHeaderRe.FindStringSubmatch(stripped); m != nil {
			blockMatch = m
		} else if m := blockBraceRe.FindStringSubmatch(stripped); m != nil {
			blockMatch = m
		}
		attrMatch := attrRe.FindStringSubmatch(stripped)

		start := i
		depth := 0
		for i < len(lines) {
			depth += braceDelta(lines[i])
			i++
			if depth <= 0 {
				break
			}
		}
		text := strings.Join(append(pending, lines[start:i]...), "\n")
		pending = nil

		var kind, name string
		if blockMatch != nil {
			kind = "block"
			name = blockMatch[1]
		} else if attrMatch != nil {
			kind = "attr"
			name = attrMatch[1]
		} else {
			kind = "other"
			if len(stripped) > 24 {
				name = stripped[:24]
			} else {
				name = stripped
			}
		}
		chunks = append(chunks, chunk{kind: kind, name: name, text: text})
	}
	return chunks
}

func addressOf(c chunk) []string {
	if c.kind != "block" {
		return nil
	}
	btype := c.name
	if !addressed[btype] {
		return []string{btype}
	}
	header := ""
	for _, l := range strings.Split(c.text, "\n") {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "//") {
			header = l
			break
		}
	}
	matches := quotedRe.FindAllStringSubmatch(header, 2)
	result := []string{btype}
	for _, m := range matches {
		result = append(result, m[1])
	}
	return result
}

func isElided(body string) bool {
	for _, line := range strings.Split(body, "\n") {
		if elidedRe.MatchString(strings.TrimSpace(line)) {
			return true
		}
	}
	return false
}

func classify(chunks []chunk, body string) string {
	blockTypes := make(map[string]bool)
	hasAttr := false
	for _, c := range chunks {
		if c.kind == "block" {
			blockTypes[c.name] = true
		}
		if c.kind == "attr" {
			hasAttr = true
		}
	}

	hasConfigBlock := false
	for bt := range blockTypes {
		if configBlocks[bt] {
			hasConfigBlock = true
			break
		}
	}

	if hasConfigBlock {
		if isElided(body) {
			return "partial"
		}
		if hasAttr {
			return "mixed"
		}
		return "standalone"
	}
	if blockTypes["import"] {
		return "import"
	}
	if hasAttr {
		return "fragment"
	}
	return "unknown"
}

func blockSurface(text string) (string, []string) {
	lines := strings.Split(text, "\n")
	headerIdx := -1
	for i, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "//") {
			headerIdx = i
			break
		}
	}
	if headerIdx == -1 {
		return "", nil
	}
	header := lines[headerIdx]
	labels := quotedRe.FindAllStringSubmatch(header, -1)
	var btype string
	if len(labels) > 0 {
		btype = labels[0][1]
	}

	var names []string
	depth := braceDelta(header)
	for _, line := range lines[headerIdx+1:] {
		stripped := strings.TrimSpace(line)
		if depth == 1 && stripped != "" && !strings.HasPrefix(stripped, "#") && !strings.HasPrefix(stripped, "//") {
			if m := identAssignOrBraceRe.FindStringSubmatch(stripped); m != nil {
				names = append(names, m[1])
			}
		}
		depth += braceDelta(line)
	}
	return btype, names
}

func extract(docsDirs []string) []block {
	var blocks []block
	var files []string
	for _, d := range docsDirs {
		filepath.WalkDir(d, func(path string, _ os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if path != d && (strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".md.tmpl")) {
				files = append(files, path)
			}
			return nil
		})
	}
	sort.Strings(files)

	seen := make(map[string]bool)
	for _, path := range files {
		if seen[path] {
			continue
		}
		seen[path] = true
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		lines := strings.Split(string(content), "\n")
		inBlock := false
		var buf []string
		start := 0
		for lineno, line := range lines {
			if !inBlock && strings.TrimSpace(line) == fenceOpen {
				inBlock = true
				buf = nil
				start = lineno + 1
			} else if inBlock && strings.TrimSpace(line) == fenceClose {
				inBlock = false
				blocks = append(blocks, block{
					source: path,
					line:   start,
					body:   strings.Join(buf, "\n"),
				})
			} else if inBlock {
				buf = append(buf, line)
			}
		}
		if inBlock {
			fmt.Fprintf(os.Stderr, "WARN unclosed ```terraform fence in %s at line %d\n", path, start)
		}
	}
	return blocks
}

func loadSchema(path string) (map[string]bool, map[string]bool, map[string]bool) {
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR reading schema: %v\n", err)
		os.Exit(2)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(content, &raw); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR parsing schema: %v\n", err)
		os.Exit(2)
	}

	resources := make(map[string]bool)
	datasources := make(map[string]bool)
	attributes := make(map[string]bool)

	var walkBlock func(b map[string]interface{})
	walkBlock = func(b map[string]interface{}) {
		if attrs, ok := b["attributes"].(map[string]interface{}); ok {
			for name, attr := range attrs {
				attributes[name] = true
				if am, ok := attr.(map[string]interface{}); ok {
					if nested, ok := am["nested_type"]; ok {
						if nb, ok := nested.(map[string]interface{}); ok {
							walkBlock(nb)
						}
					}
				}
			}
		}
		if bts, ok := b["block_types"].(map[string]interface{}); ok {
			for name, nested := range bts {
				attributes[name] = true
				if nm, ok := nested.(map[string]interface{}); ok {
					if inner, ok := nm["block"].(map[string]interface{}); ok {
						walkBlock(inner)
					}
				}
			}
		}
	}

	providerSchemas, _ := raw["provider_schemas"].(map[string]interface{})
	for _, provider := range providerSchemas {
		pm, ok := provider.(map[string]interface{})
		if !ok {
			continue
		}
		if rs, ok := pm["resource_schemas"].(map[string]interface{}); ok {
			for name, body := range rs {
				resources[name] = true
				if bm, ok := body.(map[string]interface{}); ok {
					if block, ok := bm["block"].(map[string]interface{}); ok {
						walkBlock(block)
					}
				}
			}
		}
		if ds, ok := pm["data_source_schemas"].(map[string]interface{}); ok {
			for name, body := range ds {
				datasources[name] = true
				if bm, ok := body.(map[string]interface{}); ok {
					if block, ok := bm["block"].(map[string]interface{}); ok {
						walkBlock(block)
					}
				}
			}
		}
	}

	return resources, datasources, attributes
}

func checkImport(body string, resources, datasources map[string]bool) (string, string) {
	targets := importToRe.FindAllStringSubmatch(body, -1)
	if len(targets) == 0 {
		return "FAIL", "import block has no `to` address"
	}
	var bad []string
	for _, m := range targets {
		if !resources[m[1]] {
			bad = append(bad, m[1])
		}
	}
	if len(bad) > 0 {
		unique := make(map[string]bool)
		for _, b := range bad {
			unique[b] = true
		}
		var sorted []string
		for b := range unique {
			sorted = append(sorted, b)
		}
		sort.Strings(sorted)
		hint := ""
		for _, b := range bad {
			if datasources[b] {
				hint = " (that is a data source, not a resource)"
				break
			}
		}
		return "FAIL", fmt.Sprintf("unknown resource type: %s%s", strings.Join(sorted, ", "), hint)
	}
	unique := make(map[string]bool)
	for _, m := range targets {
		unique[m[1]] = true
	}
	var sorted []string
	for t := range unique {
		sorted = append(sorted, t)
	}
	sort.Strings(sorted)
	return "PASS", strings.Join(sorted, ",")
}

func checkAttrs(names []string, attributes map[string]bool) (string, string) {
	if len(names) == 0 {
		return "FAIL", "no attributes found"
	}
	var bad []string
	for _, n := range names {
		if !attributes[n] {
			bad = append(bad, n)
		}
	}
	if len(bad) > 0 {
		unique := make(map[string]bool)
		for _, b := range bad {
			unique[b] = true
		}
		var sorted []string
		for b := range unique {
			sorted = append(sorted, b)
		}
		sort.Strings(sorted)
		return "FAIL", fmt.Sprintf("unknown attribute: %s", strings.Join(sorted, ", "))
	}
	unique := make(map[string]bool)
	for _, n := range names {
		unique[n] = true
	}
	var sorted []string
	for n := range unique {
		sorted = append(sorted, n)
	}
	sort.Strings(sorted)
	return "PASS", strings.Join(sorted, ",")
}

func main() {
	var docsDirs []string
	var outDir, schemaPath, manifestPath string
	flag.Func("docs-dir", "directory of markdown sources (repeatable)", func(s string) error {
		docsDirs = append(docsDirs, s)
		return nil
	})
	flag.StringVar(&outDir, "out", "", "directory to write block dirs into")
	flag.StringVar(&schemaPath, "schema", "", "terraform providers schema -json")
	flag.StringVar(&manifestPath, "manifest", "", "TSV manifest to write")
	flag.Parse()

	if len(docsDirs) == 0 || outDir == "" || schemaPath == "" || manifestPath == "" {
		fmt.Fprintln(os.Stderr, "Usage: extract-doc-blocks --docs-dir <dir> [--docs-dir <dir>] --out <dir> --schema <file> --manifest <file>")
		os.Exit(2)
	}

	resources, datasources, attributes := loadSchema(schemaPath)
	if len(resources) == 0 {
		fmt.Fprintln(os.Stderr, "ERROR: provider schema contained no resource types")
		os.Exit(2)
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR creating out dir: %v\n", err)
		os.Exit(2)
	}

	blocks := extract(docsDirs)

	type manifestRow struct {
		index   string
		kind    string
		source  string
		line    string
		bdir    string
		verdict string
		detail  string
	}
	var rows []manifestRow

	context := make(map[string]map[string]string)

	for idx, block := range blocks {
		blockNum := idx + 1
		chunks := splitChunks(block.body)
		kind := classify(chunks, block.body)
		source := block.source
		seen, ok := context[source]
		if !ok {
			seen = make(map[string]string)
			context[source] = seen
		}

		var blockChunks []chunk
		var attrNames []string
		for _, c := range chunks {
			if c.kind == "block" {
				blockChunks = append(blockChunks, c)
			}
			if c.kind == "attr" {
				attrNames = append(attrNames, c.name)
			}
		}

		bdir := filepath.Join(outDir, fmt.Sprintf("%03d", blockNum))

		if kind == "partial" {
			var problems []string
			var shown []string
			for _, c := range blockChunks {
				btype, names := blockSurface(c.text)
				if btype != "" && !resources[btype] && !datasources[btype] {
					problems = append(problems, fmt.Sprintf("unknown resource type: %s", btype))
				}
				var bad []string
				for _, n := range names {
					if !attributes[n] {
						bad = append(bad, n)
					}
				}
				if len(bad) > 0 {
					unique := make(map[string]bool)
					for _, b := range bad {
						unique[b] = true
					}
					var sorted []string
					for b := range unique {
						sorted = append(sorted, b)
					}
					sort.Strings(sorted)
					problems = append(problems, fmt.Sprintf("unknown attribute: %s", strings.Join(sorted, ", ")))
				}
				shown = append(shown, names...)
			}
			var verdict, detail string
			if len(problems) > 0 {
				verdict = "FAIL"
				detail = strings.Join(problems, "; ")
			} else {
				verdict = "PASS"
				unique := make(map[string]bool)
				for _, s := range shown {
					unique[s] = true
				}
				var sorted []string
				for s := range unique {
					sorted = append(sorted, s)
				}
				sort.Strings(sorted)
				detail = strings.Join(sorted, ",")
				if detail == "" {
					detail = "elided"
				}
			}
			if err := os.MkdirAll(bdir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR creating dir: %v\n", err)
				os.Exit(2)
			}
			if err := os.WriteFile(filepath.Join(bdir, "main.tf"), []byte(block.body+"\n"), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR writing main.tf: %v\n", err)
				os.Exit(2)
			}
			rows = append(rows, manifestRow{
				index:   fmt.Sprintf("%d", blockNum),
				kind:    kind,
				source:  source,
				line:    fmt.Sprintf("%d", block.line),
				bdir:    bdir,
				verdict: verdict,
				detail:  detail,
			})
		} else if kind == "standalone" || kind == "mixed" {
			verdict := "DEFER"
			detail := "terraform validate"
			if kind == "mixed" {
				verdictAttr, detailAttr := checkAttrs(attrNames, attributes)
				if verdictAttr == "FAIL" {
					verdict = "FAIL"
					detail = detailAttr
				}
			}

			own := make(map[string]bool)
			for _, c := range blockChunks {
				if addr := addressOf(c); addr != nil {
					own[strings.Join(addr, "\x00")] = true
				}
			}

			if err := os.MkdirAll(bdir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR creating dir: %v\n", err)
				os.Exit(2)
			}

			var blockTexts []string
			for _, c := range blockChunks {
				blockTexts = append(blockTexts, c.text)
			}
			if err := os.WriteFile(filepath.Join(bdir, "main.tf"), []byte(strings.Join(blockTexts, "\n\n")+"\n"), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR writing main.tf: %v\n", err)
				os.Exit(2)
			}

			var prior []string
			for addrKey, text := range seen {
				if !own[addrKey] {
					prior = append(prior, text)
				}
			}
			sort.Strings(prior)
			ctxPath := filepath.Join(bdir, "context.tf")
			if len(prior) > 0 {
				if err := os.WriteFile(ctxPath, []byte(strings.Join(prior, "\n\n")+"\n"), 0644); err != nil {
					fmt.Fprintf(os.Stderr, "ERROR writing context.tf: %v\n", err)
					os.Exit(2)
				}
			} else {
				os.Remove(ctxPath)
			}

			for _, c := range blockChunks {
				if addr := addressOf(c); addr != nil {
					seen[strings.Join(addr, "\x00")] = c.text
				}
			}

			rows = append(rows, manifestRow{
				index:   fmt.Sprintf("%d", blockNum),
				kind:    kind,
				source:  source,
				line:    fmt.Sprintf("%d", block.line),
				bdir:    bdir,
				verdict: verdict,
				detail:  detail,
			})
		} else {
			var verdict, detail string
			if kind == "import" {
				verdict, detail = checkImport(block.body, resources, datasources)
			} else if kind == "fragment" {
				verdict, detail = checkAttrs(attrNames, attributes)
			} else {
				verdict = "FAIL"
				detail = "no recognisable terraform content"
			}

			if err := os.MkdirAll(bdir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR creating dir: %v\n", err)
				os.Exit(2)
			}
			if err := os.WriteFile(filepath.Join(bdir, "main.tf"), []byte(block.body+"\n"), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR writing main.tf: %v\n", err)
				os.Exit(2)
			}
			rows = append(rows, manifestRow{
				index:   fmt.Sprintf("%d", blockNum),
				kind:    kind,
				source:  source,
				line:    fmt.Sprintf("%d", block.line),
				bdir:    bdir,
				verdict: verdict,
				detail:  detail,
			})
		}
	}

	var manifestLines []string
	for _, r := range rows {
		manifestLines = append(manifestLines, strings.Join([]string{
			r.index, r.kind, r.source, r.line, r.bdir, r.verdict, r.detail,
		}, "\t"))
	}
	if err := os.WriteFile(manifestPath, []byte(strings.Join(manifestLines, "\n")+"\n"), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR writing manifest: %v\n", err)
		os.Exit(2)
	}

	counts := make(map[string]int)
	for _, r := range rows {
		counts[r.kind]++
	}
	var parts []string
	var keys []string
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s %d", k, counts[k]))
	}
	fmt.Printf("==> extracted %d terraform blocks (%s)\n", len(rows), strings.Join(parts, ", "))
}

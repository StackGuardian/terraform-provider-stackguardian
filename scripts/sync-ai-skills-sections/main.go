// sync-ai-skills-sections writes the "Building this with AI" section into every
// documentation source page.
//
// The section names the skill an agent should load for that page and states the
// mistakes that page's resource actually invites -- the ones that pass
// `terraform validate` and fail later, or worse, succeed while meaning something
// other than intended. Those are worth stating on the page itself, because they
// are as useful to a person reading it as to an agent generating from it.
//
// Content lives here rather than in 43 files so the wording stays consistent and
// a correction lands everywhere at once. The script replaces whatever sits
// between the AI-SKILLS markers, so it is idempotent and safe to re-run after
// editing this file.
//
// Pages with no specific entry below fall back to a generic note; that is
// deliberate, so a newly added page is covered rather than silently skipped.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const heading = "## Building this with AI"
const startMarker = "<!-- AI-SKILLS:START -->"
const endMarker = "<!-- AI-SKILLS:END -->"

const skillsURL = "https://github.com/StackGuardian/terraform-provider-stackguardian/tree/main/.claude/skills"
const agentsURL = "https://github.com/StackGuardian/terraform-provider-stackguardian/blob/main/AGENTS.md"

type pageEntry struct {
	skill   string
	warning string
}

var pages = map[string]pageEntry{
	"policy": {"stackguardian-policies",
		"An inline policy sets **both** `policy_input_data` and `policy_vcs_config` — they are " +
			"not alternatives — and `schema_type` is `TIRITH_JSON`. `RAW_JSON` is accepted but the " +
			"platform stores `TIRITH_JSON`, leaving a permanent diff. `enforced_on = [\"*\"]` is " +
			"organization-wide."},
	"role": {"stackguardian-access",
		"This resource is **deprecated** in favour of `stackguardian_rolev4`, and moving is not a " +
			"rename: this one expands permissions as a cartesian product of path values, `rolev4` maps " +
			"them one to one."},
	"rolev4": {"stackguardian-access",
		"Path values map **one to one**, unlike `stackguardian_role`, which takes their cartesian " +
			"product. Converting means combining values into a single alternation string " +
			"(`\"a|b|c\"`); listing them separately silently narrows the role instead of failing."},
	"role_assignment": {"stackguardian-access",
		"`user_id` is an email address for a user, or an SSO group identifier to grant the " +
			"role to everyone in that group."},
	"workflow_git": {"stackguardian-workflows",
		"Three nesting modes are easy to get backwards: `deployment_platform_config`, " +
			"`environment_variables` and `user_schedules` are **lists**; " +
			"`mini_steps.notifications.email.<event>` is a **list** of `{recipients}`; and " +
			"`vcs_triggers.push` is a **map** keyed `createWfRun`. `terraform_version` takes the " +
			"bare form — no `TERRAFORM-` prefix."},
	"workflow_from_template": {"stackguardian-workflows",
		"Attributes you leave unset adopt the template revision's values, so they can " +
			"change when the revision does. Declare anything that must not move, and pin " +
			"the revision rather than tracking `:latest`."},
	"workflow_group": {"stackguardian-workflows",
		"Nesting is literal: `platform/networking` does **not** create `platform`, which must " +
			"already exist. If a group of the same name exists, the provider adopts it into state " +
			"rather than failing — so a later `destroy` will delete a group Terraform did not " +
			"create."},
	"workflow_template": {"stackguardian-templates",
		"The container holds no content; the revision does. Create this first, then a " +
			"`stackguardian_workflow_template_revision` against it."},
	"workflow_template_revision": {"stackguardian-templates",
		"Revision numbers are assigned by the platform, not chosen. The matching " +
			"**data source** accepts only the bare `<name>:<revision>` form — the " +
			"organization-qualified path returns `Unauthorized`, not a not-found."},
	"workflow_step_template": {"stackguardian-templates",
		"Step templates are container images and belong in `wf_step_template_id`. Using " +
			"one as an `iac_template_id` is a category error that validates and then fails " +
			"at apply."},
	"workflow_step_template_revision": {"stackguardian-templates",
		"Each step runs in its **own** container, so files and environment " +
			"changes do not carry between steps except through the shared workspace " +
			"mount."},
	"stack_template": {"stackguardian-templates",
		"A stack differs from a workflow group in that its workflows are ordered: the revision's " +
			"`actions.order` expresses the dependencies between them."},
	"stack_template_revision": {"stackguardian-templates",
		"Objects nested inside `workflows_config.workflows[]` belong to the revision — " +
			"they are created and destroyed with it and do not collide with org-level " +
			"names."},
	"connector": {"stackguardian-provider",
		"Reference a connector rather than typing its ID: `stackguardian_connector.aws.id` gives " +
			"Terraform the dependency edge. A workflow's `deployment_platform_config[].kind` must match " +
			"the connector's own kind."},
	"runner_group": {"stackguardian-workflows",
		"Pin a workflow to a runner group with `runner_constraints = {type = \"private\", names = " +
			"[\"<group-name>\"]}` — `names` takes the group's **name**, not an ID or a path."},
	"runner_group_token": {"stackguardian-workflows",
		"Tokens are read at registration time; treat the value as a credential and keep it " +
			"out of configuration and state."},
	"workflow_outputs": {"stackguardian-workflows",
		"Outputs are published by a completed run, so a plan against a workflow that has never " +
			"run returns nothing."},
	"stack_outputs": {"stackguardian-templates",
		"Stack outputs aggregate across every workflow in the stack."},
	"stack_workflow_outputs": {"stackguardian-templates",
		"This reads the outputs of one workflow inside a stack; use " +
			"`stackguardian_stack_outputs` for the whole stack."},
}

var guides = map[string]pageEntry{
	"GettingStarted": {"stackguardian-provider",
		"Order matters: the workflow group and connector must exist before the workflow that " +
			"references them."},
	"Installation": {"stackguardian-provider",
		"Every provider argument falls back to an environment variable, so credentials need not " +
			"appear in configuration."},
	"ObjectModel": {"stackguardian-provider",
		"A workflow group is both a folder and the unit access is scoped to, so the group layout " +
			"decides the permission model."},
	"ResourceIDs": {"stackguardian-resource-ids",
		"IDs are position-sensitive: the same value can be valid in one attribute and rejected in " +
			"another, and the rejection reads as `Unauthorized` rather than a not-found."},
	"Templates": {"stackguardian-templates",
		"An attribute never written down can change when a revision changes, because it was always " +
			"coming from the template."},
	"Policies": {"stackguardian-policies",
		"`approval_pre_apply` decides whether a run stops; `approvers` decides who may release it. A " +
			"gate with no approvers can be released by anyone."},
	"AccessManagement": {"stackguardian-access",
		"Permission map keys are API routes, and `paths` fills their placeholders with bare " +
			"names — not `/wfgrps/...` resource paths."},
	"TeamOnboarding": {"stackguardian-access",
		"Which permissions belong at which level is a decision about your organization, not a " +
			"fact about the provider."},
	"ImportingResources": {"stackguardian-import",
		"Run a plan straight after importing. A diff is expected, and it is closed by " +
			"declaring what the platform reports, not by removing it from state."},
	"Troubleshooting": {"stackguardian-provider",
		"A permanent diff usually means an attribute the platform computes is being written, or " +
			"one he returns is being omitted."},
}

var generic = pageEntry{"stackguardian-provider",
	"Check the schema below rather than inferring an attribute or enum value from its name."}

func body(kind, subject, skill, warning string) string {
	var lead string
	if kind == "guide" {
		lead = fmt.Sprintf("Generating configuration from this guide? Load the **`%s`** skill, which turns "+
			"the guidance here into rules an agent can follow.", skill)
	} else {
		lead = fmt.Sprintf("Generating `%s` configuration with an AI assistant? Load the **`%s`** "+
			"skill, which covers this resource's arguments and the mistakes it invites.", subject, skill)
	}
	return fmt.Sprintf("%s\n\n"+
		"**Worth knowing either way:** %s\n\n"+
		"The skills live in [the provider repository](%s) and work with Claude Code, "+
		"Cursor, Copilot, Windsurf and any agent that reads [`AGENTS.md`](%s).",
		lead, warning, skillsURL, agentsURL)
}

func section(text string) string {
	return fmt.Sprintf("\n%s\n\n%s\n%s\n%s\n", heading, startMarker, text, endMarker)
}

var blockRe = regexp.MustCompile(`(?s)\n## Building this with AI\n\n` + regexp.QuoteMeta(startMarker) + `\n.*?\n` + regexp.QuoteMeta(endMarker) + `\n\z`)

func apply(path, text string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("error reading file: %v", err)
	}
	s := string(content)
	newSection := section(text)

	if strings.Contains(s, startMarker) {
		if !blockRe.MatchString(s) {
			return "MARKERS NOT AT END OF FILE"
		}
		s = blockRe.ReplaceAllString(s, "")
	}
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}

	if err := os.WriteFile(path, []byte(s+newSection), 0644); err != nil {
		return fmt.Sprintf("error writing file: %v", err)
	}
	return "ok"
}

var pageHeadingRe = regexp.MustCompile(`(?m)^# (stackguardian_\S+) \((Resource|Data Source)\)\s*$`)

func pageHeading(text string) (string, string) {
	m := pageHeadingRe.FindStringSubmatch(text)
	if m == nil {
		return "", ""
	}
	return m[1], m[2]
}

func trimPrefix(s, prefix string) string {
	if strings.HasPrefix(s, prefix) {
		return s[len(prefix):]
	}
	return s
}

func trimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		return s[:len(s)-len(suffix)]
	}
	return s
}

func main() {
	var problems []string
	done := 0

	resourceFiles, _ := filepath.Glob("docs-templates/resources/*.md.tmpl")
	dataSourceFiles, _ := filepath.Glob("docs-templates/data-sources/*.md.tmpl")
	allFiles := append(resourceFiles, dataSourceFiles...)
	sort.Strings(allFiles)

	for _, path := range allFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", path, err))
			continue
		}
		name, _ := pageHeading(string(content))
		if name == "" {
			problems = append(problems, fmt.Sprintf("%s: no 'stackguardian_x (Resource|Data Source)' H1", path))
			continue
		}
		key := trimPrefix(name, "stackguardian_")
		entry, ok := pages[key]
		if !ok {
			entry = generic
		}
		r := apply(path, body("resource", name, entry.skill, entry.warning))
		if r != "ok" {
			problems = append(problems, fmt.Sprintf("%s: %s", path, r))
		} else {
			done++
		}
		fmt.Printf("  %-48s %s\n", filepath.Base(path), entry.skill)
	}

	guideFiles, _ := filepath.Glob("docs-templates/guides/*.md")
	sort.Strings(guideFiles)
	for _, path := range guideFiles {
		key := trimSuffix(filepath.Base(path), ".md")
		entry, ok := guides[key]
		if !ok {
			entry = generic
		}
		r := apply(path, body("guide", key, entry.skill, entry.warning))
		if r != "ok" {
			problems = append(problems, fmt.Sprintf("%s: %s", path, r))
		} else {
			done++
		}
		fmt.Printf("  %-48s %s\n", filepath.Base(path), entry.skill)
	}

	r := apply("docs-templates/index.md.tmpl",
		body("guide", "the provider", "stackguardian-provider",
			"Start with `stackguardian-provider`; it routes to the skill for the area in play."))
	if r != "ok" {
		problems = append(problems, fmt.Sprintf("index.md.tmpl: %s", r))
	} else {
		done++
	}
	fmt.Printf("  %-48s stackguardian-provider\n", "index.md.tmpl")

	fmt.Printf("\n==> %d pages synced\n", done)
	for _, p := range problems {
		fmt.Fprintf(os.Stderr, "PROBLEM %s\n", p)
	}
	if len(problems) > 0 {
		os.Exit(1)
	}
}

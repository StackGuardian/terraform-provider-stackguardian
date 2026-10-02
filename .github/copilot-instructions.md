# GitHub Copilot instructions

This repository is the **StackGuardian Terraform Provider**.

Full guidance is in [`AGENTS.md`](../AGENTS.md); per-area detail is in
[`.claude/skills/`](../.claude/skills/); provider internals are in [`CLAUDE.md`](../CLAUDE.md).
Read those before making non-trivial changes. What follows is the short version.

## Two jobs, different rules

**Writing StackGuardian Terraform configuration** — consult the skill for the area:
`stackguardian-resource-ids`, `-workflows`, `-templates`, `-policies`, `-access`, `-import`.

**Developing the provider** — Go, terraform-plugin-framework v1. See `CLAUDE.md` for the
per-resource file pattern and the expander/flattener conventions.

## Rules that always apply

1. **Never invent a resource type, attribute, or enum value.** The authoritative list is
   `internal/provider/provider.go` — 15 resources, 17 data sources.
2. **IDs are path-form**: `/integrations/<name>`, `/secrets/<name>`, `<name>:<rev>` in your own org,
   `/<org>/<name>:<rev>` elsewhere. The `stackguardian_workflow_template_revision` data source
   accepts **only** the bare form; the qualified path returns `Unauthorized`.
3. **`stackguardian_role` is deprecated — use `stackguardian_rolev4`.** Not a rename: V3 takes the
   cartesian product of path values, V4 maps one to one. Combine values into one alternation string
   (`"a|b|c"`) when converting.
4. **An inline policy sets both** `policy_input_data` **and** `policy_vcs_config`, with
   `schema_type = "TIRITH_JSON"`.
5. **Secret references use a double colon**: `${secret::<name>}`, written `$${secret::<name>}` in
   Terraform. Never put credentials in `text_value`.
6. **Nested workflow groups do not create their parents.**

## Repository conventions

- **`docs/` is generated — never edit it directly.** Edit `docs-templates/`, `docs-examples/`, or
  the `MarkdownDescription` values in `internal/constants/` and each `schema.go`, then run
  `make docs-generate`.
- Every documentation example is type-checked in CI by `make docs-validate-examples`, including
  snippets embedded in guide prose. Adding an example means it must validate.
- Acceptance tests create real infrastructure and need credentials. Identifiers are randomised with
  a `tfacc-` prefix via `acctest.ResourceName`; do not hardcode resource names in new tests.
- Go 1.21.4. Run `make test` for unit tests; `go build ./...` is the source of truth for compilation.

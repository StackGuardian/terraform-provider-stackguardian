---
trigger: glob
globs: **/*.tf,internal/**/*.go,docs-templates/**,docs-examples/**
description: StackGuardian Terraform provider conventions — resource IDs, roles, policies, secrets, and generated docs.
---

# StackGuardian Terraform Provider

Full guidance: `AGENTS.md`. Per-area skills: `.claude/skills/`. Provider internals: `CLAUDE.md`.

## Writing StackGuardian configuration

Consult the skill for the area — `stackguardian-resource-ids`, `-workflows`, `-templates`,
`-policies`, `-access`, `-import`.

Rules where being wrong is silent rather than loud:

- **Never invent a resource type, attribute, or enum value.** `internal/provider/provider.go` is
  authoritative: 15 resources, 17 data sources.
- **IDs are path-form**: `/integrations/<name>`, `/secrets/<name>`, `<name>:<rev>` in your own org,
  `/<org>/<name>:<rev>` elsewhere. The `stackguardian_workflow_template_revision` data source takes
  only the bare form; the qualified path returns `Unauthorized`.
- **`stackguardian_role` is deprecated — use `stackguardian_rolev4`.** V3 takes the cartesian
  product of path values, V4 maps one to one. Convert by combining values into a single alternation
  string (`"a|b|c"`).
- **Inline policies set both `policy_input_data` and `policy_vcs_config`**, `schema_type` is
  `TIRITH_JSON`.
- **Secret references use a double colon**: `${secret::<name>}` → `$${secret::<name>}` in Terraform.
  Credentials never go in `text_value`.
- **Nested workflow groups do not create their parents.**
- Prefer resource references over literal IDs so Terraform orders the graph.

## Repository conventions

- `docs/` is **generated** — edit `docs-templates/`, `docs-examples/`, or the schema
  `MarkdownDescription` values, then `make docs-generate`.
- Examples are type-checked in CI via `make docs-validate-examples`.
- Acceptance tests create real infrastructure; identifiers are randomised with a `tfacc-` prefix.
- Go 1.21.4; `go build ./...` is the source of truth for compilation.

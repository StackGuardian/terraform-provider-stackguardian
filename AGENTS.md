# AGENTS.md

Guidance for AI coding agents working in this repository or writing StackGuardian Terraform
configuration. Read by agents that follow the `AGENTS.md` convention — OpenAI Codex, Gemini CLI,
Cursor, Zed, Aider, Jules and others. Claude Code additionally loads the skills below directly.

There are two distinct jobs here. Pick the right one.

## 1. Writing StackGuardian Terraform configuration

Skills live in [`.claude/skills/`](.claude/skills/), one per area. Each is a `SKILL.md` whose
frontmatter says when it applies. Load the one that matches the task:

| Skill | Load when |
| --- | --- |
| [`stackguardian-provider`](.claude/skills/stackguardian-provider/SKILL.md) | Any `stackguardian_*` config; start here to route |
| [`stackguardian-resource-ids`](.claude/skills/stackguardian-resource-ids/SKILL.md) | A reference returns 404 or `Unauthorized`; any attribute naming another resource |
| [`stackguardian-workflows`](.claude/skills/stackguardian-workflows/SKILL.md) | Workflows, workflow groups, drift, triggers, schedules |
| [`stackguardian-templates`](.claude/skills/stackguardian-templates/SKILL.md) | Templates, revisions, pinning, upgrades |
| [`stackguardian-policies`](.claude/skills/stackguardian-policies/SKILL.md) | Guardrails, approval gates, policy bodies |
| [`stackguardian-access`](.claude/skills/stackguardian-access/SKILL.md) | Roles, permissions, assignments |
| [`stackguardian-import`](.claude/skills/stackguardian-import/SKILL.md) | Adopting existing resources; `409 already exists` |

Reference material the skills point at lives in [`docs/guides/`](docs/guides/) and on the
[Terraform Registry](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs).

### Rules that always apply

These are the ones where being wrong is **silent** — the config validates and then misbehaves:

1. **Never invent a resource type, attribute, or enum value.** `internal/provider/provider.go` is
   the authoritative list: 15 resources, 17 data sources. Check the schema; do not infer from a
   plausible name.
2. **IDs are path-form and position-sensitive.** `/integrations/<name>`, `/secrets/<name>`,
   `<name>:<rev>` in your own org, `/<org>/<name>:<rev>` elsewhere. The
   `stackguardian_workflow_template_revision` **data source** takes only the bare form — the
   qualified path returns `Unauthorized`, not "not found".
3. **`stackguardian_role` is deprecated. Use `stackguardian_rolev4`** — and it is not a rename. V3
   takes the cartesian product of path values; V4 maps one to one. Converting means combining values
   into one alternation string (`"a|b|c"`). A naive swap silently narrows the role.
4. **An inline policy sets both `policy_input_data` and `policy_vcs_config`.** They are not
   alternatives. `schema_type` is `TIRITH_JSON`.
5. **Secret references use a double colon** — `${secret::<name>}`, written `$${secret::<name>}` in
   Terraform. Credentials never go in `text_value`.
6. **Nested workflow groups do not create their parents.** `platform/networking` requires
   `platform` to exist.
7. **State what cannot be verified statically.** `terraform validate` checks the schema; it cannot
   tell you whether a referenced connector, secret, or template exists in the organization.

## 2. Developing the provider itself

Read [`CLAUDE.md`](CLAUDE.md) — architecture, the per-resource file pattern, expander/flattener
conventions, and the gotchas that cause runtime panics.

Two things that catch agents out:

- **`docs/` is generated. Never edit it by hand.** Attribute text comes from `MarkdownDescription`
  in each `schema.go`, mostly via shared constants in `internal/constants/`; page prose lives in
  `docs-templates/`. Edit those, then `make docs-generate`.
- **Acceptance tests create real infrastructure.** They need credentials and an organization, and
  every identifier they create is randomised with a `tfacc-` prefix. Run `make test` for unit tests;
  `make test-acc` needs `TF_ACC=1` plus `STACKGUARDIAN_API_KEY`, `_API_URI` and `_ORG_NAME`.

Verify with:

```bash
make test                    # unit tests, no credentials
make docs-generate           # rebuild docs/ after touching schemas or templates
make docs-validate-examples  # type-check every example, no credentials, no network
make docs-check              # fail if docs/ is stale — the CI gate
```

## Output expectations

When proposing configuration, state the assumptions, what must already exist in the organization,
and how to validate. Prefer resource references over hardcoded IDs so Terraform builds the
dependency graph. Say plainly when something is a guess.

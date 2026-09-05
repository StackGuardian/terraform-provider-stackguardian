#!/usr/bin/env python3
"""Write the "Building this with AI" section into every documentation source page.

The section names the skill an agent should load for that page and states the
mistakes that page's resource actually invites -- the ones that pass
`terraform validate` and fail later, or worse, succeed while meaning something
other than intended. Those are worth stating on the page itself, because they are
as useful to a person reading it as to an agent generating from it.

Content lives here rather than in 43 files so the wording stays consistent and a
correction lands everywhere at once. The script replaces whatever sits between the
AI-SKILLS markers, so it is idempotent and safe to re-run after editing this file.

Pages with no specific entry below fall back to a generic note; that is deliberate,
so a newly added page is covered rather than silently skipped.
"""
from __future__ import annotations

import glob
import os
import re
import sys

HEADING = "## Building this with AI"
START = "<!-- AI-SKILLS:START -->"
END = "<!-- AI-SKILLS:END -->"

SKILLS_URL = "https://github.com/StackGuardian/terraform-provider-stackguardian/tree/main/.claude/skills"
AGENTS_URL = "https://github.com/StackGuardian/terraform-provider-stackguardian/blob/main/AGENTS.md"

# page basename -> (skill, page-specific warning). The warning is the part that
# cannot be inferred from the schema.
PAGES: dict[str, tuple[str, str]] = {
    # Policies
    "policy": ("stackguardian-policies",
               "An inline policy sets **both** `policy_input_data` and `policy_vcs_config` — they are "
               "not alternatives — and `schema_type` is `TIRITH_JSON`. `RAW_JSON` is accepted but the "
               "platform stores `TIRITH_JSON`, leaving a permanent diff. `enforced_on = [\"*\"]` is "
               "organization-wide."),
    # Access
    "role": ("stackguardian-access",
             "This resource is **deprecated** in favour of `stackguardian_rolev4`, and moving is not a "
             "rename: this one expands permissions as a cartesian product of path values, `rolev4` maps "
             "them one to one."),
    "rolev4": ("stackguardian-access",
               "Path values map **one to one**, unlike `stackguardian_role`, which takes their cartesian "
               "product. Converting means combining values into a single alternation string "
               "(`\"a|b|c\"`); listing them separately silently narrows the role instead of failing."),
    "role_assignment": ("stackguardian-access",
                        "`user_id` is an email address for a user, or an SSO group identifier to grant the "
                        "role to everyone in that group."),
    # Workflows
    "workflow_git": ("stackguardian-workflows",
                     "Three nesting modes are easy to get backwards: `deployment_platform_config`, "
                     "`environment_variables` and `user_schedules` are **lists**; "
                     "`mini_steps.notifications.email.<event>` is a **list** of `{recipients}`; and "
                     "`vcs_triggers.push` is a **map** keyed `createWfRun`. `terraform_version` takes the "
                     "bare form — no `TERRAFORM-` prefix."),
    "workflow_from_template": ("stackguardian-workflows",
                               "Attributes you leave unset adopt the template revision's values, so they can "
                               "change when the revision does. Declare anything that must not move, and pin "
                               "the revision rather than tracking `:latest`."),
    "workflow_group": ("stackguardian-workflows",
                       "Nesting is literal: `platform/networking` does **not** create `platform`, which must "
                       "already exist. If a group of the same name exists, the provider adopts it into state "
                       "rather than failing — so a later `destroy` will delete a group Terraform did not "
                       "create."),
    # Templates
    "workflow_template": ("stackguardian-templates",
                          "The container holds no content; the revision does. Create this first, then a "
                          "`stackguardian_workflow_template_revision` against it."),
    "workflow_template_revision": ("stackguardian-templates",
                                   "Revision numbers are assigned by the platform, not chosen. The matching "
                                   "**data source** accepts only the bare `<name>:<revision>` form — the "
                                   "organization-qualified path returns `Unauthorized`, not a not-found."),
    "workflow_step_template": ("stackguardian-templates",
                               "Step templates are container images and belong in `wf_step_template_id`. Using "
                               "one as an `iac_template_id` is a category error that validates and then fails "
                               "at apply."),
    "workflow_step_template_revision": ("stackguardian-templates",
                                        "Each step runs in its **own** container, so files and environment "
                                        "changes do not carry between steps except through the shared workspace "
                                        "mount."),
    "stack_template": ("stackguardian-templates",
                       "A stack differs from a workflow group in that its workflows are ordered: the revision's "
                       "`actions.order` expresses the dependencies between them."),
    "stack_template_revision": ("stackguardian-templates",
                                "Objects nested inside `workflows_config.workflows[]` belong to the revision — "
                                "they are created and destroyed with it and do not collide with org-level "
                                "names."),
    # Connectors and runners
    "connector": ("stackguardian-provider",
                  "Reference a connector rather than typing its ID: `stackguardian_connector.aws.id` gives "
                  "Terraform the dependency edge. A workflow's `deployment_platform_config[].kind` must match "
                  "the connector's own kind."),
    "runner_group": ("stackguardian-workflows",
                     "Pin a workflow to a runner group with `runner_constraints = {type = \"private\", names = "
                     "[\"<group-name>\"]}` — `names` takes the group's **name**, not an ID or a path."),
    "runner_group_token": ("stackguardian-workflows",
                           "Tokens are read at registration time; treat the value as a credential and keep it "
                           "out of configuration and state."),
    # Outputs
    "workflow_outputs": ("stackguardian-workflows",
                         "Outputs are published by a completed run, so a plan against a workflow that has never "
                         "run returns nothing."),
    "stack_outputs": ("stackguardian-templates",
                      "Stack outputs aggregate across every workflow in the stack."),
    "stack_workflow_outputs": ("stackguardian-templates",
                               "This reads the outputs of one workflow inside a stack; use "
                               "`stackguardian_stack_outputs` for the whole stack."),
}

GUIDES: dict[str, tuple[str, str]] = {
    "GettingStarted": ("stackguardian-provider",
                       "Order matters: the workflow group and connector must exist before the workflow that "
                       "references them."),
    "Installation": ("stackguardian-provider",
                     "Every provider argument falls back to an environment variable, so credentials need not "
                     "appear in configuration."),
    "ObjectModel": ("stackguardian-provider",
                    "A workflow group is both a folder and the unit access is scoped to, so the group layout "
                    "decides the permission model."),
    "ResourceIDs": ("stackguardian-resource-ids",
                    "IDs are position-sensitive: the same value can be valid in one attribute and rejected in "
                    "another, and the rejection reads as `Unauthorized` rather than a not-found."),
    "Templates": ("stackguardian-templates",
                  "An attribute never written down can change when a revision changes, because it was always "
                  "coming from the template."),
    "Policies": ("stackguardian-policies",
                 "`approval_pre_apply` decides whether a run stops; `approvers` decides who may release it. A "
                 "gate with no approvers can be released by anyone."),
    "AccessManagement": ("stackguardian-access",
                         "Permission map keys are API routes, and `paths` fills their placeholders with bare "
                         "names — not `/wfgrps/...` resource paths."),
    "TeamOnboarding": ("stackguardian-access",
                       "Which permissions belong at which level is a decision about your organization, not a "
                       "fact about the provider."),
    "ImportingResources": ("stackguardian-import",
                           "Run a plan straight after importing. A diff is expected, and it is closed by "
                           "declaring what the platform reports, not by removing it from state."),
    "Troubleshooting": ("stackguardian-provider",
                        "A permanent diff usually means an attribute the platform computes is being written, or "
                        "one it returns is being omitted."),
}

GENERIC = ("stackguardian-provider",
           "Check the schema below rather than inferring an attribute or enum value from its name.")


def body(kind: str, subject: str, skill: str, warning: str) -> str:
    if kind == "guide":
        lead = (f"Generating configuration from this guide? Load the **`{skill}`** skill, which turns "
                f"the guidance here into rules an agent can follow.")
    else:
        lead = (f"Generating `{subject}` configuration with an AI assistant? Load the **`{skill}`** "
                f"skill, which covers this resource's arguments and the mistakes it invites.")
    return (
        f"{lead}\n\n"
        f"**Worth knowing either way:** {warning}\n\n"
        f"The skills live in [the provider repository]({SKILLS_URL}) and work with Claude Code, "
        f"Cursor, Copilot, Windsurf and any agent that reads [`AGENTS.md`]({AGENTS_URL})."
    )


def section(text: str) -> str:
    return f"\n{HEADING}\n\n{START}\n{text}\n{END}\n"


BLOCK = re.compile(
    r"\n## Building this with AI\n\n" + re.escape(START) + r"\n.*?\n" + re.escape(END) + r"\n\Z",
    re.S,
)


def apply(path: str, text: str) -> str:
    s = open(path, encoding="utf-8").read()
    new = section(text)
    if START in s:
        if not BLOCK.search(s):
            return "MARKERS NOT AT END OF FILE"
        s = BLOCK.sub("", s)
    if not s.endswith("\n"):
        s += "\n"
    open(path, "w", encoding="utf-8").write(s + new)
    return "ok"


def page_heading(text: str):
    m = re.search(r"^# (stackguardian_\S+) \((Resource|Data Source)\)\s*$", text, re.M)
    return (m.group(1), m.group(2)) if m else (None, None)


def main() -> int:
    problems, done = [], 0

    for path in sorted(glob.glob("docs-templates/resources/*.md.tmpl")
                       + glob.glob("docs-templates/data-sources/*.md.tmpl")):
        name, kind = page_heading(open(path, encoding="utf-8").read())
        if not name:
            problems.append(f"{path}: no 'stackguardian_x (Resource|Data Source)' H1")
            continue
        key = name.removeprefix("stackguardian_")
        skill, warning = PAGES.get(key, GENERIC)
        r = apply(path, body("resource", name, skill, warning))
        (problems.append(f"{path}: {r}") if r != "ok" else None)
        done += r == "ok"
        print(f"  {os.path.basename(path):<48} {skill}")

    for path in sorted(glob.glob("docs-templates/guides/*.md")):
        key = os.path.basename(path).removesuffix(".md")
        skill, warning = GUIDES.get(key, GENERIC)
        r = apply(path, body("guide", key, skill, warning))
        (problems.append(f"{path}: {r}") if r != "ok" else None)
        done += r == "ok"
        print(f"  {os.path.basename(path):<48} {skill}")

    r = apply("docs-templates/index.md.tmpl",
              body("guide", "the provider", "stackguardian-provider",
                   "Start with `stackguardian-provider`; it routes to the skill for the area in play."))
    (problems.append(f"index.md.tmpl: {r}") if r != "ok" else None)
    done += r == "ok"
    print(f"  {'index.md.tmpl':<48} stackguardian-provider")

    print(f"\n==> {done} pages synced")
    for p in problems:
        print(f"PROBLEM {p}", file=sys.stderr)
    return 1 if problems else 0


if __name__ == "__main__":
    sys.exit(main())

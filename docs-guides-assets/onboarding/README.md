# Onboarding examples

Two worked configurations showing how a business organization can be laid out in StackGuardian:
one team per project, with workflow groups, roles, role assignments and connectors.

They back the [Team Onboarding](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/TeamOnboarding)
guide, which explains the reasoning; these are the configurations themselves.

## project-01 — a flat team

A single team with no management layer. Everyone has a full-stack profile and works across
frontend, backend and devops, so one role covers everybody.

- Three workflow groups: `Frontend`, `Backend`, `DevOps`
- One `stackguardian_rolev4` — `Developer` — scoped to all three, with the full
  workflow lifecycle including delete
- One role assignment
- A cloud connector (`AWS_RBAC`) and a VCS connector
- A workflow, so the groups are not empty and the cloud connector is actually used
- [`import.sh`](project-01/import.sh) to bring an existing organization under Terraform

Releases are frequent, through staging and production, and the whole team is responsible for
tearing down cloud resources that are no longer needed.

## project-02 — a hierarchical team

Two teams, frontend and backend, each with a manager and developers, plus a cross-functional
devops role contributing to both.

- Three workflow groups: `Frontend`, `Backend`, `DevOps`
- Five `stackguardian_rolev4` roles: a manager and a developer for each of frontend and
  backend, plus a cross-functional `Developer-DevOps`
- Five role assignments — every role declared is granted to someone
- A cloud connector (`AWS_RBAC`) and a VCS connector
- A workflow, wired to the cloud connector

The access model, which is what this example exists to show:

| Role | Can do |
| ---- | ------ |
| Developer | Read its group; create, update and run workflows in it; read logs and outputs. |
| Manager | Everything a developer can, plus delete a workflow and resume a run held for approval. |
| DevOps | The manager set, across all three groups — this role owns teardown. |

The permission sets are built once as `locals` and combined with `merge`, so the difference
between a manager and a developer is visible in one place rather than repeated five times.

Releases go out every two weeks, with the devops role responsible for tearing down cloud
resources that are no longer needed.

`project-02` has no import script; use `project-01`'s as a model.

## Before you run these

Both configurations use placeholders that must be replaced:

- `<YOUR-API-KEY>` and `<YOUR-ORG-NAME>` in the provider block — or better, set
  `STACKGUARDIAN_API_KEY` and `STACKGUARDIAN_ORG_NAME` and delete those lines.
- `REPLACE-ME` values in the connector credentials.
- `<org>` in the role permission keys — it is part of the key, not a comment.

See the [top-level README](../README.md) for how to run and validate them.

# Onboarding examples

Two worked configurations showing how a business organization can be laid out in StackGuardian:
one team per project, with workflow groups, roles, role assignments and connectors.

They back the [Team Onboarding](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/TeamOnboarding)
guide, which explains the reasoning; these are the configurations themselves.

## project-01 — a flat team

A single team with no management layer. Everyone has a full-stack profile and works across
frontend, backend and devops, so one role covers everybody.

- Three workflow groups: `Frontend`, `Backend`, `DevOps`
- One role — `Developer` — scoped to all three
- One role assignment
- A cloud connector and a VCS connector
- [`import.sh`](project-01/import.sh) to bring an existing organization under Terraform

Releases are frequent, through staging and production, and the whole team is responsible for
tearing down cloud resources that are no longer needed.

## project-02 — a hierarchical team

Two teams, frontend and backend, each with a manager and developers, plus a cross-functional
devops role contributing to both.

- Three workflow groups: `Frontend`, `Backend`, `DevOps`
- Five roles: a manager and a developer role for each of frontend and backend, plus a
  cross-functional `Developer-DevOps`
- Three role assignments, one per role that is actually granted
- A cloud connector and a VCS connector

Releases go out every two weeks, with the devops role responsible for tearing down cloud
resources that are no longer needed.

`project-02` has no import script; use `project-01`'s as a model.

## Before you run these

Both configurations use placeholders that must be replaced:

- `<YOUR-API-KEY>` and `<YOUR-ORG-NAME>` in the provider block — or better, set
  `STACKGUARDIAN_API_KEY` and `STACKGUARDIAN_ORG_NAME` and delete those lines.
- `REPLACE-ME-*` values in the connector credentials.
- The organization name embedded in role permission paths.

See the [top-level README](../README.md) for how to run and validate them.

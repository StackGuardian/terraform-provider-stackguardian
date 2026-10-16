<a href="https://www.stackguardian.io/">
    <img src=".github/stackguardian_logo.svg" alt="StackGuardian logo" title="StackGuardian" align="right" height="40" />
</a>

# StackGuardian Terraform Provider

[![Terraform Registry](https://img.shields.io/badge/terraform-registry-623CE4?logo=terraform)](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest)
[![Release](https://img.shields.io/github/v/release/StackGuardian/terraform-provider-stackguardian)](https://github.com/StackGuardian/terraform-provider-stackguardian/releases)
[![Go](https://img.shields.io/github/go-mod/go-version/StackGuardian/terraform-provider-stackguardian)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MPL--2.0-blue)](/LICENSE)

The [StackGuardian Terraform Provider](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest) allows [Terraform](https://www.terraform.io/) to programmatically interact with the [StackGuardian API](https://docs.stackguardian.io/docs/api/overview) to help you manage resources on the StackGuardian platform, ultimately enabling organizations to manage cloud infrastructure in a cost-efficient, secure, and compliant way.

It covers **15 resources** and **17 data sources** — workflows, workflow groups, stacks, templates and their revisions, connectors, policies, roles, role assignments and runner groups.

> [!TIP]
> Looking for ready-to-use examples? All StackGuardian Terraform modules and examples are maintained in the **[terraform-stackguardian-modules](https://github.com/StackGuardian/terraform-stackguardian-modules)** repository.

## Quick start

The provider is published on the [Terraform Registry](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest):

```terraform
terraform {
  required_providers {
    stackguardian = {
      source  = "StackGuardian/stackguardian"
      version = "~> 1.12"
    }
  }
}

provider "stackguardian" {
  api_key  = "<YOUR-API-KEY>"                   # or env var STACKGUARDIAN_API_KEY
  org_name = "<YOUR-ORG-NAME>"                  # or env var STACKGUARDIAN_ORG_NAME
  api_uri  = "https://api.app.stackguardian.io" # or env var STACKGUARDIAN_API_URI; use "https://api.us.stackguardian.io" for the US region
}

resource "stackguardian_workflow_group" "example" {
  resource_name = "Simple-Workflow-Group"
  description   = "Example of how to create a workflow group using the StackGuardian Terraform Provider"
  tags          = ["tf-provider-example", "example"]
}
```

## Documentation

Reference for every resource and data source is on the
[Terraform Registry](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs),
alongside eleven guides:

| Guide | What it covers |
| --- | --- |
| [Installation](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/Installation) | Installing and configuring the provider |
| [Getting Started](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/GettingStarted) | A working deployment, one resource at a time |
| [Object Model](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/ObjectModel) | How workflow groups, templates, connectors and roles relate |
| [Resource IDs](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/ResourceIDs) | The path-form IDs StackGuardian uses, and where they differ |
| [Runtime References](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/RuntimeReferences) | The `${...}` tokens the platform resolves at run time |
| [Templates and Revisions](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/Templates) | The template/revision split, and what a revision upgrade changes |
| [Policies](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/Policies) | Guardrails, approval gates, and where a policy body comes from |
| [Access Management](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/AccessManagement) | Roles, permissions and assignments |
| [Team Onboarding](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/TeamOnboarding) | Laying out groups and roles for a team |
| [Importing Existing Resources](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/ImportingResources) | Import ID formats per resource |
| [Troubleshooting](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/Troubleshooting) | Common errors and what they mean |

## Examples

- **[terraform-stackguardian-modules](https://github.com/StackGuardian/terraform-stackguardian-modules)** — the collection of StackGuardian Terraform modules and examples.
- [Quickstart](/docs-guides-assets/quickstart) — a minimal working configuration to get started.
- [Onboarding projects](/docs-guides-assets/onboarding) — end-to-end projects covering connectors, workflow groups, roles and role assignments.

## Development

### Prerequisites

- [Go](https://go.dev/) >= 1.26.7 (see [go.mod](/go.mod))
- [Terraform](https://developer.hashicorp.com/terraform/downloads) — CI tests against 1.14.0; the
  acceptance suite skips versions below 1.1.0, and the `import {}` block examples need 1.5.0 or newer
- [`tfplugindocs`](https://github.com/hashicorp/terraform-plugin-docs) for documentation work, via `make tools-install`
- [`act`](https://github.com/nektos/act) to run the GitHub workflows locally (optional)

### Common tasks

| Command | What it does |
| --- | --- |
| `make build` | Compile the provider binary |
| `make install` | Build and install into `~/.terraform.d/plugins` for local use |
| `make test` | Unit tests |
| `make test-acc` | Acceptance tests against a real organization |
| `make test-crud` | Create/update/import/delete check of every resource, as a pass/fail matrix |
| `make docs-generate` | Regenerate `docs/` from the schemas and templates |
| `make docs-validate` | Structure and frontmatter checks |
| `make docs-check` | Fail if `docs/` is out of date — the CI gate |
| `make docs-validate-examples` | Type-check every documentation example |

### Documentation workflow

`docs/` is **generated — never edit it by hand.** Content lives in two places:

- **Attribute text** comes from `MarkdownDescription` in each `schema.go`, mostly via shared
  constants in [`internal/constants/`](/internal/constants). A data source should reuse the same
  constant as its resource twin so the two cannot drift.
- **Page prose** lives in [`docs-templates/`](/docs-templates), and embedded `.tf` snippets in
  [`docs-examples/`](/docs-examples).

Edit those, then:

```bash
make docs-generate           # rebuild docs/
make docs-validate           # structure and frontmatter
make docs-validate-examples  # type-check every example
make docs-check              # fail if docs/ is stale
```

`docs-validate-examples` needs no credentials and makes no API calls. It builds the provider, serves
it from a local filesystem mirror, and runs `terraform validate` over every example — the standalone
files under `docs-examples/` and `docs-guides-assets/`, plus every ` ```terraform ` block embedded
in page and guide prose. Prose blocks are not uniform, so
[`scripts/extract-doc-blocks`](/scripts/extract-doc-blocks) classifies each one and checks it
the way it can actually be checked: self-contained snippets are validated, `import {}` blocks have
their target verified against the schema, and bare attribute fragments have their names looked up.
That catches unknown resource types, renamed attributes, bad references and type errors before a
reader copies them.

### Testing

Unit tests need nothing:

```bash
make test
```

Acceptance tests create and destroy real resources, so they need an organization:

```bash
export TF_ACC=1
export STACKGUARDIAN_API_KEY=<key>
export STACKGUARDIAN_API_URI=<uri>
export STACKGUARDIAN_ORG_NAME=<org>

make test-acc                                                   # everything
make test-acc TEST=./internal/resource/workflow_git/            # one package
make test-acc TESTARGS='-run TestAccWorkflowGit_WithVcsConfig'  # one test
```

`STACKGUARDIAN_API_URI` is the host only, e.g. `https://testapi.qa.stackguardian.io`. The SDK
appends `/api/v1` itself, and a doubled `/api/v1/api/v1` path comes back as `403 Unauthorized`
rather than 404. Use a user API key (`sgu_…`): an organization key (`sgo_…`) gets 403 on roles.

To keep the credentials out of your shell history, put the four variables in a `.env` file at the
repository root (`KEY=value` lines; `.env` is gitignored) and load it for a single command:

```bash
(set -a; source .env; set +a; make test-crud)
```

#### CRUD matrix

`make test-crud` runs one `TestAcc<Resource>_CRUD` test per resource and reports the result as a
resource × operation table. Every resource goes through the same steps, defined once in
[`internal/acctest/crud.go`](/internal/acctest/crud.go):

| Step | What is checked |
| --- | --- |
| create | Apply, then refresh: the plan must be empty |
| update | The plan must be an in-place update — not a replace, not a no-op — and apply cleanly |
| import | Import by ID; the imported state must match what Terraform created |
| gone | The object is deleted through the API behind Terraform's back; refresh must drop it and plan a recreate |
| delete | Destroy, then ask the API: the object must really be gone |

```bash
(set -a; source .env; set +a; make test-crud)                                         # all resources
(set -a; source .env; set +a; make test-crud TESTARGS='-run TestAccPolicy_CRUD')      # one resource
```

The whole run takes well under a minute. Results land in `.crud-results/`:

- `matrix.md` — the table, followed by a one-line reason for each failure (the API error, or the
  attribute diff of a failed import)
- `matrix.json` — the same, for tooling
- `go-test.json` — the raw `go test -json` output

The target exits non-zero if any resource fails. When you add a resource, add a
`resource_crud_test.go` that calls `acctest.RunCRUD`; the existing ones show the pattern.

Every identifier the suite creates is generated per run and prefixed `tfacc-`, via
`acctest.ResourceName`. Static names made the suite order-dependent: a run that failed before its
destroy step left the name taken, and every later run failed at create with `409 already exists`.
The shared prefix is what makes a leak identifiable afterwards.

When a run is interrupted between create and destroy, its resources survive. Sweep them:

```bash
make test-acc-sweep        # report what carries the prefix; deletes nothing
make test-acc-sweep-apply  # delete it
```

The sweep only ever touches names carrying the `tfacc-` prefix, and reports rather than deletes
unless you ask it to. It covers workflow groups (with their workflows), connectors, and workflow,
workflow step and stack templates (with their revisions). Policies, roles and runner groups cannot
be listed through `sg-sdk-go`, so the report names them as not swept.

### Running the workflows locally

The GitHub workflows can be run on your machine with [`act`](https://github.com/nektos/act), which
avoids pushing a branch to see a CI result:

```bash
make gh-workflow-test-provider     # the provider test workflow
make gh-workflow-test-api-stg      # API tests against staging
make gh-workflow-test-api-prd      # API tests against production
```

Release notes for each version are in the [CHANGELOG](/CHANGELOG.md) and on the
[GitHub releases page](https://github.com/StackGuardian/terraform-provider-stackguardian/releases).

### Terraform version support

The minimum supported Terraform version is **1.5.7**. The supported range is
defined in `internal/constants/terraform_version.go`:

- `MinTerraformVersion` — the support floor, **1.5.7**.
- `MaxTerraformVersion` — the newest release CI exercises. Not a ceiling on
  support; the provider imposes none.
- `SupportedTerraformVersions` — the CI matrix: the floor, the newest tested
  release, and the latest patch of every minor series in between.

Four layers keep that promise honest:

| Layer | What it does | Where |
| ----- | ------------ | ----- |
| Runtime check | The provider refuses to configure on a CLI below the floor, with a message saying so | `internal/provider/version_check.go` |
| Acceptance tests | Every test case fails — rather than silently skips — below the floor | `acctest.VersionChecks()` |
| Compatibility gate | Every pull request to `main` makes each Terraform version in the matrix load the provider, decode its full schema, and validate every example under `docs-examples/`. No credentials required. | `.github/workflows/compat.yaml` |
| Acceptance matrix | Every pull request to `main` runs the full acceptance suite against that same matrix | `.github/workflows/test.yaml` |

Both workflows hardcode the matrix in their YAML.
`TestTerraformMatrixMatchesWorkflows` fails if either drifts from
`SupportedTerraformVersions`, so the Go list stays the single place to edit when
the range changes.

Raising the floor is a breaking change for users: bump the provider version
alongside it and note it in the [CHANGELOG](/CHANGELOG.md).

### Continuous integration

All workflows gate on `main`:

```text
feature branch  ──PR──▶  main  ──tag v*──▶  release
                     │                 │
          compatibility gate   acceptance matrix
           (no credentials)     (hits the API)
```

A pull request to `main` runs the hermetic compatibility gate first, then the
expensive acceptance suite that creates real resources. Both target `main`
only — no `develop` branch is involved.

| Workflow | Runs on | Needs secrets | What it does |
| -------- | ------- | ------------- | ------------ |
| `compat.yaml` | PR and push to `main`; manual | No | Build, `go vet`, unit tests, docs validation, and the Terraform compatibility matrix |
| `test.yaml` | PR and push to `main`; manual | Yes | Full acceptance suite across the Terraform matrix, plus docs validation |
| `release.yaml` | Tag `v*`; manual | Yes | Runs both of the above, then builds and signs the release |
| `test-api-prd.yaml` | Manual only | Yes | API high-load test against production, using `main` |
| `test-api-stg.yaml` | Manual only | Yes | API high-load test against staging, using `main` |
| `test-api.yaml` | Called by the two above | Yes | Shared implementation; takes `gitref` and `testenv` inputs |
| `community.yml` | Called by `release.yaml` | Yes | GoReleaser build and GPG signing |

No workflow is scheduled. The two API load tests in particular are triggered by
hand from the Actions tab — they used to run on a cron during a fixed window,
and now need someone to start them.

Everything that calls the API shares one StackGuardian organization, and some
tests use fixed resource names, so overlapping runs collide in ways that look
like version incompatibilities rather than the race they are. `test.yaml` guards
against this twice: `max-parallel: 1` keeps its own matrix sequential, and a
`concurrency: acceptance-tests` group stops two of its runs overlapping. That
group does not cover `test-api-prd.yaml` or `test-api-stg.yaml` — check the
Actions tab before triggering a load test by hand.

## Contributing

Contributions are welcome — please see [CONTRIBUTING.md](/CONTRIBUTING.md) for guidelines and the [Code of Conduct](/CODE_OF_CONDUCT.md) for community standards. Use [GitHub issues](https://github.com/StackGuardian/terraform-provider-stackguardian/issues) to report bugs or request features.

## License

This project is licensed under the [Mozilla Public License 2.0](/LICENSE).

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
alongside ten guides:

| Guide | What it covers |
| --- | --- |
| [Installation](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/Installation) | Installing and configuring the provider |
| [Getting Started](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/GettingStarted) | A working deployment, one resource at a time |
| [Object Model](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/ObjectModel) | How workflow groups, templates, connectors and roles relate |
| [Resource IDs](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/ResourceIDs) | The path-form IDs StackGuardian uses, and where they differ |
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

- [Go](https://go.dev/) 1.21.4 or newer (see [go.mod](/go.mod))
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
[`scripts/extract-doc-blocks.py`](/scripts/extract-doc-blocks.py) classifies each one and checks it
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
unless you ask it to.

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

## Contributing

Contributions are welcome — please see [CONTRIBUTING.md](/CONTRIBUTING.md) for guidelines and the [Code of Conduct](/CODE_OF_CONDUCT.md) for community standards. Use [GitHub issues](https://github.com/StackGuardian/terraform-provider-stackguardian/issues) to report bugs or request features.

## License

This project is licensed under the [Mozilla Public License 2.0](/LICENSE).

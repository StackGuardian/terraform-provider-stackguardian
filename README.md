<a href="https://www.stackguardian.io/">
    <img src=".github/stackguardian_logo.svg" alt="StackGuardian logo" title="StackGuardian" align="right" height="40" />
</a>

# StackGuardian Terraform Provider

[![Terraform Registry](https://img.shields.io/badge/terraform-registry-623CE4?logo=terraform)](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest)
[![Release](https://img.shields.io/github/v/release/StackGuardian/terraform-provider-stackguardian)](https://github.com/StackGuardian/terraform-provider-stackguardian/releases)
[![License](https://img.shields.io/badge/license-MPL--2.0-blue)](/LICENSE)

The [StackGuardian Terraform Provider](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest) allows [Terraform](https://www.terraform.io/) to programmatically interact with the [StackGuardian API](https://docs.stackguardian.io/docs/api/overview) to help you manage resources on the StackGuardian platform, ultimately enabling organizations to manage cloud infrastructure in a cost-efficient, secure, and compliant way.

> [!TIP]
> Looking for ready-to-use examples? All StackGuardian Terraform modules and examples are maintained in the **[terraform-stackguardian-modules](https://github.com/StackGuardian/terraform-stackguardian-modules)** repository.

## Usage

The provider is available on the [Terraform Registry](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest):

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

The full list of supported resources and data sources, with documentation for each, is available on the [Terraform Registry](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs).

## Examples

- **[terraform-stackguardian-modules](https://github.com/StackGuardian/terraform-stackguardian-modules)** — the collection of StackGuardian Terraform modules and examples.
- [Quickstart guide](/docs-guides-assets/quickstart) — a minimal working configuration to get started.
- [Onboarding examples](/docs-guides-assets/onboarding) — end-to-end projects covering connectors, workflow groups, roles, and role assignments.

## Development

Requirements:

- [Go](https://go.dev/) >= 1.26.7
- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.5.7

```bash
make build                  # compile the provider binary
make install                # build and install into ~/.terraform.d/plugins for local use
make test                   # run unit tests (no credentials needed)
make tf-compat-check        # check the provider loads and examples validate on the local CLI
make test-acc               # run acceptance tests (see below)
make test-acc-min-terraform # run acceptance tests on Terraform 1.5.7
make docs-generate          # regenerate docs with tfplugindocs
make docs-validate          # validate generated docs
```

Acceptance tests run against a real StackGuardian organization and require:

```bash
export TF_ACC=1
export STACKGUARDIAN_API_KEY=<key>
export STACKGUARDIAN_API_URI=<uri>
export STACKGUARDIAN_ORG_NAME=<org>
```

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
| Compatibility gate | Every pull request to `develop` makes each Terraform version in the matrix load the provider, decode its full schema, and validate every example under `docs-examples/`. No credentials required. | `.github/workflows/compat.yaml` |
| Acceptance matrix | Every pull request to `main` runs the full acceptance suite against that same matrix | `.github/workflows/test.yaml` |

Both workflows hardcode the matrix in their YAML.
`TestTerraformMatrixMatchesWorkflows` fails if either drifts from
`SupportedTerraformVersions`, so the Go list stays the single place to edit when
the range changes.

Raising the floor is a breaking change for users: bump the provider version
alongside it and note it in the [CHANGELOG](/CHANGELOG.md).

### Continuous integration

Work flows through two branches, and each one gates on a different cost of test:

```text
feature branch  ──PR──▶  develop  ──PR──▶  main  ──tag v*──▶  release
                    │                 │                   │
         compatibility gate   acceptance matrix     both, then build
          (no credentials)     (hits the API)
```

A pull request to `develop` runs only hermetic checks, so it is fast and safe on
forks. The expensive suite that creates real resources runs later, when work is
promoted to `main`.

| Workflow | Runs on | Needs secrets | What it does |
| -------- | ------- | ------------- | ------------ |
| `compat.yaml` | PR and push to `develop`; manual | No | Build, `go vet`, unit tests, docs validation, and the Terraform compatibility matrix |
| `test.yaml` | PR and push to `main`; manual | Yes | Full acceptance suite across the Terraform matrix, plus docs validation |
| `release.yaml` | Tag `v*`; manual | Yes | Runs both of the above, then builds and signs the release |
| `test-api-prd.yaml` | Manual only | Yes | API high-load test against production, using `main` |
| `test-api-stg.yaml` | Manual only | Yes | API high-load test against staging, using `develop` |
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

Release notes for each version are in the [CHANGELOG](/CHANGELOG.md) and on the [GitHub releases page](https://github.com/StackGuardian/terraform-provider-stackguardian/releases).

## Contributing

Contributions are welcome — please see [CONTRIBUTING.md](/CONTRIBUTING.md) for guidelines and the [Code of Conduct](/CODE_OF_CONDUCT.md) for community standards. Use [GitHub issues](https://github.com/StackGuardian/terraform-provider-stackguardian/issues) to report bugs or request features.

## License

This project is licensed under the [Mozilla Public License 2.0](/LICENSE).

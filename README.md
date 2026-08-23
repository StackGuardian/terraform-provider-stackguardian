<a href="https://www.stackguardian.io/">
    <img src=".github/stackguardian_logo.svg" alt="StackGuardian logo" title="StackGuardian" align="right" height="40" />
</a>

# StackGuardian Terraform Provider

[![Terraform Registry](https://img.shields.io/badge/terraform-registry-623CE4?logo=terraform)](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest)
[![Release](https://img.shields.io/github/v/release/StackGuardian/terraform-provider-stackguardian)](https://github.com/StackGuardian/terraform-provider-stackguardian/releases)
[![License](https://img.shields.io/badge/license-MPL--2.0-blue)](/LICENSE)

The [StackGuardian Terraform Provider](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest) allows [Terraform](https://www.terraform.io/) to programmatically interact with the [StackGuardian API](https://docs.stackguardian.io/docs/api/overview) to help you manage resources on the StackGuardian platform, ultimately enabling organizations to manage cloud infrastructure in a cost-efficient, secure, and compliant way.

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

## Supported resources

| Resource                                             | Data source |
| ---------------------------------------------------- | ----------- |
| `stackguardian_connector`                            | ✓           |
| `stackguardian_workflow_group`                       | ✓           |
| `stackguardian_role`                                 | ✓           |
| `stackguardian_rolev4`                               |             |
| `stackguardian_role_assignment`                      | ✓           |
| `stackguardian_policy`                               | ✓           |
| `stackguardian_runner_group`                         | ✓           |
| `stackguardian_workflow_template`                    | ✓           |
| `stackguardian_workflow_template_revision`           | ✓           |
| `stackguardian_workflow_step_template`               | ✓           |
| `stackguardian_workflow_step_template_revision`      | ✓           |
| `stackguardian_stack_template`                       | ✓           |
| `stackguardian_stack_template_revision`              | ✓           |
| `stackguardian_workflow_git`                         | ✓           |
| `stackguardian_workflow_from_template`               |             |

Additional data sources: `stackguardian_workflow_outputs`, `stackguardian_stack_outputs`, `stackguardian_stack_workflow_outputs`, and `stackguardian_runner_group_token`.

Full documentation for every resource and data source is available on the [Terraform Registry](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs) and in the [`docs/`](/docs) directory.

## Examples

- [Quickstart guide](/docs-guides-assets/quickstart) — a minimal working configuration to get started.
- [Onboarding examples](/docs-guides-assets/onboarding) — end-to-end projects covering connectors, workflow groups, roles, and role assignments.

## Development

Requirements:

- [Go](https://go.dev/) >= 1.21
- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 0.13

```bash
make build          # compile the provider binary
make install        # build and install into ~/.terraform.d/plugins for local use
make test           # run unit tests
make test-acc       # run acceptance tests (see below)
make docs-generate  # regenerate docs with tfplugindocs
make docs-validate  # validate generated docs
```

Acceptance tests run against a real StackGuardian organization and require:

```bash
export TF_ACC=1
export STACKGUARDIAN_API_KEY=<key>
export STACKGUARDIAN_API_URI=<uri>
export STACKGUARDIAN_ORG_NAME=<org>
```

Release notes for each version are in the [CHANGELOG](/CHANGELOG.md) and on the [GitHub releases page](https://github.com/StackGuardian/terraform-provider-stackguardian/releases).

## Contributing

Contributions are welcome — please see [CONTRIBUTING.md](/CONTRIBUTING.md) for guidelines and the [Code of Conduct](/CODE_OF_CONDUCT.md) for community standards. Use [GitHub issues](https://github.com/StackGuardian/terraform-provider-stackguardian/issues) to report bugs or request features.

## License

This project is licensed under the [Mozilla Public License 2.0](/LICENSE).

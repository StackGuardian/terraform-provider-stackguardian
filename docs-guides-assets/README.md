# Provider examples

Runnable configurations that back the guides published on the Terraform Registry.

| Directory | What it is |
| --- | --- |
| [`quickstart/`](quickstart) | One working deployment: workflow group → connector → workflow. Backs the [Getting Started](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/GettingStarted) guide. |
| [`onboarding/project-01/`](onboarding/project-01) | A flat team: three workflow groups, a developer role, connectors, and an import script. |
| [`onboarding/project-02/`](onboarding/project-02) | A hierarchical team: frontend and backend groups with separate manager and developer roles. |

Per-resource snippets live in [`../docs-examples`](../docs-examples) and are embedded into the
resource and data source pages by `tfplugindocs`.

## Running an example

Each directory is a self-contained Terraform configuration. From the repository root:

```bash
# Build and install the provider locally
make install

# Then, in the example directory
cd docs-guides-assets/quickstart
export STACKGUARDIAN_API_KEY="<your-api-key>"
export STACKGUARDIAN_ORG_NAME="<your-org-name>"

terraform init
terraform plan
```

The examples use placeholder credentials, account IDs and repository URLs — replace them before
applying. They create real resources in the organization named by `STACKGUARDIAN_ORG_NAME`, so
use a test organization.

To clean up:

```bash
terraform destroy
rm -rf .terraform .terraform.lock.hcl terraform.tfstate*
```

## Validating without credentials

Every example is type-checked against the provider schema in CI. To run that locally:

```bash
make docs-validate-examples
```

This builds the provider, resolves it from a local filesystem mirror and runs
`terraform validate` on each example. It makes no API calls and needs no credentials.

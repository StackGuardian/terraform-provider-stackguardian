---
page_title: "Installation"
subcategory: "Getting Started"
description: |-
  Install the StackGuardian provider and give it credentials.
---

# Installation

## Requirements

- Terraform 0.13 or later (1.5+ to use `import` blocks)
- A StackGuardian organization

## Configure the provider

Add the provider to your configuration:

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
  api_uri = "https://api.app.stackguardian.io"
}
```

Then run `terraform init`.

## Credentials

The provider needs three values. Each can be set in the `provider` block or, preferably, as an
environment variable — keeping the API key out of your configuration and out of version control.

| Setting    | Environment variable      | Notes                                                     |
| ---------- | ------------------------- | --------------------------------------------------------- |
| `api_key`  | `STACKGUARDIAN_API_KEY`   | Required. Sensitive — prefer the environment variable.     |
| `org_name` | `STACKGUARDIAN_ORG_NAME`  | Required. The organization the provider operates against.  |
| `api_uri`  | `STACKGUARDIAN_API_URI`   | Optional. Defaults to `https://api.app.stackguardian.io`.  |

```bash
export STACKGUARDIAN_API_KEY="<your-api-key>"
export STACKGUARDIAN_ORG_NAME="<your-org-name>"
```

### Where to get an API key

Generate one in the StackGuardian UI, in your organization settings. There are two kinds, and the
difference matters for what the provider is allowed to do:

| Kind | Prefix | Permissions |
| --- | --- | --- |
| Organization API key | `sgo_` | You choose the role when creating the key, up to the role you hold yourself. |
| User API key | `sgu_` | Inherits the permissions of the user it belongs to. |

For CI, prefer an **organization key** scoped to a role that covers only what the pipeline needs —
it does not break when a person's own permissions change, or when they leave. See
[Access Control](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/AccessControl)
for how roles are scoped.

### Regions

| Region | `api_uri`                            |
| ------ | ------------------------------------ |
| EU     | `https://api.app.stackguardian.io`   |
| US     | `https://api.us.stackguardian.io`    |

Point `api_uri` at the region your organization is hosted in. Using the wrong one produces
authentication errors even with a valid key.

## Next steps

- [Getting Started](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/GettingStarted) — a working configuration, end to end.
- [Object Model](https://registry.terraform.io/providers/StackGuardian/stackguardian/latest/docs/guides/ObjectModel) — how the resources relate to each other.

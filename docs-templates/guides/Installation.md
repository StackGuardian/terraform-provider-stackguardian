---
page_title: "Installation"
description: |-
---

# Installation

To install the StackGuardian provider follow the instructions below.

- Add the following code into your Terraform configuration files:

    ```terraform
    terraform {
    required_providers {
        stackguardian = {
        source = "StackGuardian/stackguardian"
        version = "1.0.0-rc5"
        }
    }
    }

    provider "stackguardian" {
        api_key  = "<YOUR-API-KEY>"                      # Replace this with your API key
        org_name = "<YOUR-ORG-NAME>"                     # Replace this with your organization name
        api_uri  = "https://api.app.stackguardian.io"    # Use "https://api.us.stackguardian.io" for US Region
    }
    ```
- Run `terraform init` to initialize the provider.

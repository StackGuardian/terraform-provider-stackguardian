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
        version = "1.0.0-rc4"
        }
    }
    }

    provider "stackguardian" {
        api_key  = "<YOUR-API-KEY>"                      # Replace this with your API key
        org_name = "<YOUR-ORG-NAME>"                     # Replace this with your organization name
        api_uri  = "https://testapi.qa.stackguardian.io" # Use testapi instead of production for testing
    }
    ```
- Run `terraform init` to initialize the provider.
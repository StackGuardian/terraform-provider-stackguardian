terraform {
  required_providers {
    stackguardian = {
      source  = "terraform/provider/stackguardian"
      version = "0.0.0-dev"
    }
  }
}

provider "stackguardian" {}

resource "stackguardian_policy" "TPS-Example-Policy" {
  data = jsonencode({
    "ResourceName" : "TPS-Example-Policy",
    "Description" : "Example of terraform-provider-stackguardian for Policy",
    "Tags" : ["tf-provider-example", "example", "policy"]
  })
}

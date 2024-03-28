terraform {
  required_providers {
    stackguardian = {
      source  = "terraform/provider/stackguardian"
      version = "0.0.0-dev"
    }
  }
}

provider "stackguardian" {}

resource "stackguardian_secret" "TPS-Example-Secret-Name" {
  data = jsonencode({
    "ResourceName" : "TPS-Example-Secret-Name",
    "ResourceValue" : "TPS-Example-Secret-Value"
  })
}

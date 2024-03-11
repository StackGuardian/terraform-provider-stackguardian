terraform {
  required_providers {
    stackguardian = {
      source  = "terraform/provider/stackguardian"
      version = "0.0.0-dev"
    }
  }
}

provider "stackguardian" {}

resource "stackguardian_role" "TPS-Example-Role" {
  data = jsonencode({
    "ResourceName" : "TPS-Example-Role",
    "Description" : "Example of terraform-provider-stackguardian for Role",
    "Tags" : ["tf-provider-example"],
    "Actions" : [
      "Action-1"
    ],
    "AllowedPermissions" : {
      "Permission-key-1" : "Permission-val-1",
      "Permission-key-2" : "Permission-val-2"
    }
  })
}

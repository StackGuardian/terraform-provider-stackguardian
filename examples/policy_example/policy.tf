terraform {
  required_providers {
    stackguardian = {
      source = "terraform/provider/stackguardian"
      version = "0.0.0-dev"
    }
  }
}

provider "stackguardian" {
  org_name = "---" // TBD
  api_key  = "---" // TBD
}

resource "stackguardian_policy" "TestPolicy" {
  data = jsonencode(
    { "ResourceName" : "test", "Description" : "", "Tags" : ["test", "policy"] }
  )
}

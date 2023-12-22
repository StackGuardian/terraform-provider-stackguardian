terraform {
  required_providers {
    stackguardian = {
      source = "terraform/provider/stackguardian"
    }
  }
}

provider "stackguardian" {
  org_name = "---" // TBD
  api_key  = "---" // TBD
}

resource "stackguardian_tf_provider_policy" "TestPolicy" {
  data = jsonencode(
    { "ResourceName" : "test", "Description" : "", "Tags" : ["test", "policy"] }
  )
}

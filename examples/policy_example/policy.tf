terraform {
  required_providers {
    stackguardian = {
      source = "terraform/provider/stackguardian"
    }
  }
}

provider "stackguardian" {
  api_uri  = "https://api.app.stackguardian.io/api/v1/"
  org_name = "fanda"
  api_key  = "sgu_g5aHlvuHQvaarykFYHRG5"
}

resource "stackguardian_tf_provider_policy" "TestPolicy" {
  data = jsonencode(
    { "ResourceName" : "test", "Description" : "", "Tags" : ["test", "policy"] }
  )
}


terraform {
  required_providers {
    stackguardian = {
      source = "terraform/provider/stackguardian"
    }
  }
}

provider "stackguardian" {
  org_name = "fanda"
  api_key  = "sgu_g5aHlvuHQvaarykFYHRG5"
}

resource "stackguardian_tf_provider_workflow" "test_import" {
  wfgrp = "Firstworkflow"
  # stack ="example" optional
  data = jsonencode({
    "ResourceName" : "test_import",
    "wfgrpName" : "Firstworkflow",
  })
}
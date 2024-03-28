terraform {
  required_providers {
    stackguardian = {
      source  = "terraform/provider/stackguardian"
      version = "0.0.0-dev"
    }
  }
}

provider "stackguardian" {}

resource "stackguardian_workflow_group" "TPS-Example-WorkflowGroup" {
  data = jsonencode({
    "ResourceName" : "TPS-Example-WorkflowGroup",
    "Description" : "Example of terraform-provider-stackguardian for WorkflowGroup",
    "Tags" : ["tf-provider-"],
    "IsActive" : 1,
  })
}

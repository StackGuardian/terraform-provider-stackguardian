# Read the outputs of a workflow's most recent run.
#
# Use this to chain infrastructure together: one workflow produces a value, and Terraform
# feeds it into whatever comes next. Outputs arrive as a JSON string.
data "stackguardian_workflow_outputs" "vpc" {
  workflow       = "deploy-vpc"
  workflow_group = "platform"
}

locals {
  vpc_outputs = jsondecode(data.stackguardian_workflow_outputs.vpc.data_json)
}

output "subnet_ids" {
  value = try(local.vpc_outputs.subnet_ids, [])
}

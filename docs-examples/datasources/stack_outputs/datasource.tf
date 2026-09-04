# Read the outputs of a stack's most recent run, so resources outside StackGuardian can
# consume values it produced (a VPC id, a cluster endpoint, and so on).
#
# Outputs arrive as a JSON string; decode it before use.
data "stackguardian_stack_outputs" "network" {
  stack          = "network"
  workflow_group = "platform"
}

locals {
  network_outputs = jsondecode(data.stackguardian_stack_outputs.network.data_json)
}

output "vpc_id" {
  value = try(local.network_outputs.vpc_id, null)
}

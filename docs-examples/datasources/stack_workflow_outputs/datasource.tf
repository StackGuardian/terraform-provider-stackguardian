# Read the outputs of one workflow inside a stack, rather than the stack's outputs as a whole.
# Use this when a stack runs several workflows and you need a value from a specific one.
data "stackguardian_stack_workflow_outputs" "cluster" {
  stack          = "platform-stack"
  workflow       = "deploy-cluster"
  workflow_group = "platform"
}

locals {
  cluster_outputs = jsondecode(data.stackguardian_stack_workflow_outputs.cluster.data_json)
}

output "cluster_endpoint" {
  value = try(local.cluster_outputs.cluster_endpoint, null)
}

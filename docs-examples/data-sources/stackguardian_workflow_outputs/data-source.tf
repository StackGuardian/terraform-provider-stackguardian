data "stackguardian_workflow_outputs" "TPS-Example-WorkflowOutputs" {
  wfgrp           = "aws-dev-environments"
  wf              = "wf-musical-coral"
  // stack        = "test-stack-1" // optionally
}

output "website_url_from_mapstr" {
  value = data.stackguardian_workflow_outputs.TPS-Example-WorkflowOutputs.outputs_str.sample_website_url
}

output "website_url_from_json" {
  value = jsondecode(data.stackguardian_workflow_outputs.TPS-Example-WorkflowOutputs.outputs_json).sample_website_url.value
}

output "outputs_full_json" {
  value = jsondecode(data.stackguardian_workflow_outputs.TPS-Example-WorkflowOutputs.outputs_json)
}

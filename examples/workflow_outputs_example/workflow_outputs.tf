terraform {
  required_providers {
    stackguardian = {
      source = "terraform/provider/stackguardian"
      version = "0.0.0-dev"
    }
  }
}

provider "stackguardian" {
  org_name = "wicked-hop"
  // api_key must be picked up from the var env STACKGUARDIAN_API_KEY
}

data "stackguardian_wf_output" "wf-test-1" {
  # wfgrps/aws-dev-environments/wfs/wf-musical-coral?tab=outputs
  wfgrp           = "aws-dev-environments"
  wf              = "wf-musical-coral"
  // stack        = "test-stack-1" // optionally
}

output "website_url_from_mapstr" {
  value = data.stackguardian_wf_output.wf-test-1.outputs_str.sample_website_url
}


output "website_url_from_json" {
  value = jsondecode(data.stackguardian_wf_output.wf-test-1.outputs_json).sample_website_url.value
}

output "outputs_full_json" {
  value = jsondecode(data.stackguardian_wf_output.wf-test-1.outputs_json)
}

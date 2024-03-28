package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
TODO:
	- add test for output name
	- add tests for other types implemented for the outputs_str map
	- replicate tests for workflow in stack
	- fix test for `outputs_full_json`: with/without jsondecode
*/

const testAccCheckSgWorkflowOutputsConfig = `
data "stackguardian_wf_output" "TPS-Test-Outputs" {
	# wfgrps/aws-dev-environments/wfs/wf-musical-coral?tab=outputs
	wfgrp           = "aws-dev-environments"
	wf              = "wf-musical-coral"
	// stack        = "test-stack-1" // optionally
  }

  output "website_url_from_mapstr" {
	value = data.stackguardian_wf_output.TPS-Test-Outputs.outputs_str.sample_website_url
  }


  output "website_url_from_json" {
	value = jsondecode(data.stackguardian_wf_output.TPS-Test-Outputs.outputs_json).sample_website_url.value
  }

  output "outputs_full_json" {
	value = jsondecode(data.stackguardian_wf_output.TPS-Test-Outputs.outputs_json)
  }
`

func TestAcc_DatasourceSgWorkflowOutputs(t *testing.T) {
	t.Skipf("TODO: Find identical WF for PRD and STG")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckSgWorkflowOutputsConfig,
				Check: resource.ComposeTestCheckFunc(

					resource.TestCheckOutput("website_url_from_mapstr", "http://stackguardian-fast-giraffe.s3-website.eu-central-1.amazonaws.com"),
					resource.TestCheckOutput("website_url_from_json", "http://stackguardian-fast-giraffe.s3-website.eu-central-1.amazonaws.com"),
					/*
						resource.TestCheckOutput("outputs_full_json", `{
							"sample_website_url" = {
							  "sensitive" = false
							  "type" = "string"
							  "value" = "http://stackguardian-fast-giraffe.s3-website.eu-central-1.amazonaws.com"
							}
						  }`),
					*/
					/*
						resource.TestCheckOutput("outputs_full_json",
						(&terraform.OutputState{
							Sensitive: false,
							Type:      "string",
							Value:     "http://stackguardian-fast-giraffe.s3-website.eu-central-1.amazonaws.com",
							}).String(),
						),
					*/
					/*
						resource.TestCheckOutput("outputs_full_json", `map[sample_website_url:map[sensitive:false type:string value:http://stackguardian-fast-giraffe.s3-website.eu-central-1.amazonaws.com]]`),
					*/
				),
			},
		},
	})
}

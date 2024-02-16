package stackguardian_tf_provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

/*
TODO:
	- add test for output name
	- add tests for other types implemented for the outputs_str map
	- replicate tests for workflow in stack
	- fix test for `outputs_full_json`: with/without jsondecode
*/

// var testAccProviders map[string]*schema.Provider
var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"stackguardian": testAccProvider,
	}
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("STACKGUARDIAN_ORG_NAME") == "" {
		t.Fatal("STACKGUARDIAN_ORG_NAME must be set for acceptance tests")
	}

	if os.Getenv("STACKGUARDIAN_API_KEY") == "" {
		t.Fatal("STACKGUARDIAN_API_KEY must be set for acceptance tests")
	}

	// Needed ?
	// err := testAccProvider.Configure(terraform.NewResourceConfig(nil))
	// if err != nil {
	// 		t.Fatal(err)
	// }
}

const testAccCheckSgWorkflowOutputsConfig = `
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
`

func TestAccSgWorkflowOutputsDataSource_Outputs(t *testing.T) {
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

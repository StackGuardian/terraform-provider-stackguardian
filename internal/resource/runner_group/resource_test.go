package runnergroup_test

import (
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

const (
	testAccResource = `resource "stackguardian_runner_group" "example-runner-group" {
  resource_name = "example-runner-group"

  description = "RunnerGroup created using provider"
  
  tags = [
    "provider",
    "runnergroup"
  ]
	
	storage_backend_config = {
		type = "aws_s3"
		aws_region = "eu-central-1"
		s3_bucket_name = "http-proxy-private-runner"
	}

}
`
	testAccResourceUpdate = `resource "stackguardian_runner_group" "example-runner-group" {
  resource_name = "example-runner-group"

  description = "RunnerGroup created using provider"
  
  tags = [
    "provider",
    "runnergroup",
		"update"
  ]
	
	storage_backend_config = {
		type = "aws_s3"
		aws_region = "eu-central-1"
		s3_bucket_name = "http-proxy-private-runner"
	}
}`
)

func TestAccRunnerGroup(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccResource,
			},
			{
				Config: testAccResourceUpdate,
			},
		},
	})
}

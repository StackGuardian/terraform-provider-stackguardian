package runnergroupdatasource_test

import (
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccRunnerGroupDatasource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `data "stackguardian_runner_group" "example" {
						resource_name = "test-datasource-runner-group"
					}

					output "demo-runner-group-output" {
						value = data.stackguardian_runner_group.example.description
					}`,
			},
		},
	})
}

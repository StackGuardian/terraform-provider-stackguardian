package runnergroupdatasource_test

import (
	"net/http"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRunnerGroupDatasource(t *testing.T) {
	acctest.SkipUnlessAcceptance(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks:   acctest.VersionChecks(),
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
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

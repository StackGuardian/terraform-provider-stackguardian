package workflowgroup_test

import (
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

const (
	testAccResource = `resource "stackguardian_workflow_group" "ONBOARDING-Project01-Backend" {
  resource_name = "ONBOARDING-Project01-Backend"
  description   = "Onboarding example  of terraform-provider-stackguardian for WorkflowGroup"
  tags          = ["tf-provider-example", "onboarding"]
}`
	testAccResourceUpdate = `resource "stackguardian_workflow_group" "ONBOARDING-Project01-Backend" {
  resource_name = "ONBOARDING-Project01-Backend"
  description   = "Onboarding example of terraform-provider-stackguardian for WorkflowGroup"
  tags          = ["tf-provider-example", "onboarding", "update"]
}`
)

func TestAccWorkflowGroup(t *testing.T) {
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

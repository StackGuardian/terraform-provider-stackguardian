package workflowgroup_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

const (
	testAccResource = `resource "stackguardian_workflow_group" "%s" {
  resource_name = "%s"
  description   = "Onboarding example  of terraform-provider-stackguardian for WorkflowGroup"
  tags          = ["tf-provider-example", "onboarding"]
}`
	testAccResourceUpdate = `resource "stackguardian_workflow_group" "%s" {
  resource_name = "%s"
  description   = "Onboarding example of terraform-provider-stackguardian for WorkflowGroup"
  tags          = ["tf-provider-example", "onboarding", "update"]
}`
)

func TestAccWorkflowGroup(t *testing.T) {
	workflowGroupResrouceName := "example-workflow-group"
	workflowGroupName := "example-workflow-group"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccResource, workflowGroupResrouceName, workflowGroupName),
			},
			{
				Config: fmt.Sprintf(testAccResourceUpdate, workflowGroupResrouceName, workflowGroupName),
			},
		},
	})
}

func TestAccWorkflowGroupRecreateOnExternalDelete(t *testing.T) {
	workflowGroupResourceName := "example-policy2"
	workflowGroupName := "example-policy2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccResource, workflowGroupResourceName, workflowGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("stackguardian_workflow_group.%s", workflowGroupResourceName), "resource_name", workflowGroupName),
				),
			},
			{
				PreConfig: func() {
					client := acctest.SGClient()
					_, err := client.WorkflowGroups.DeleteWorkflowGroup(context.TODO(), os.Getenv("STACKGUARDIAN_ORG_NAME"), workflowGroupName)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: fmt.Sprintf(testAccResource, workflowGroupResourceName, workflowGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("stackguardian_workflow_group.%s", workflowGroupResourceName), "resource_name", workflowGroupName),
				),
			},
		},
	})
}

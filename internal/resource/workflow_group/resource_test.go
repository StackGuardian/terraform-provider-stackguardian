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
	workflowGroupResrouceName := "wfgrp-example-workflow-group"
	workflowGroupName := "wfgrp-example-workflow-group"

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
	workflowGroupResourceName := "wfgrp-example-workflow-group2"
	workflowGroupName := "wfgrp-example-workflow-group2"

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

func TestAccWorkflowGroupIncompatibleResourceName(t *testing.T) {
	// Test if the resource has name that is not compatible with the
	workflowGroupName := "wfgrp-example-workflow-group3"
	workflowGroupResourceName := "wfgrp example workflow group3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccResource, workflowGroupName, workflowGroupResourceName),
			},
			{
				Config: fmt.Sprintf(testAccResourceUpdate, workflowGroupName, workflowGroupResourceName),
			},
		},
	})
}

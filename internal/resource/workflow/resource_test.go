package workflow_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
	sgoption "github.com/StackGuardian/sg-sdk-go/option"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

var org = os.Getenv("STACKGUARDIAN_ORG_NAME")

func GetClient() *sgclient.Client {
	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	client := sgclient.NewClient(
		sgoption.WithApiKey(fmt.Sprintf("apikey %s", os.Getenv("STACKGUARDIAN_API_KEY"))),
		sgoption.WithBaseURL(os.Getenv("STACKGUARDIAN_API_URI")),
		sgoption.WithHTTPHeader(customHeader),
	)
	return client
}

func createWorkflowGroupFixture(wfGrpName string) error {
	client := GetClient()
	resourceName := wfGrpName
	payload := &sgsdkgo.WorkflowGroup{
		ResourceName: &resourceName,
	}
	_, err := client.WorkflowGroups.CreateWorkflowGroup(context.TODO(), org, payload)
	return err
}

func deleteWorkflowGroupFixture(wfGrpName string) {
	client := GetClient()
	client.WorkflowGroups.DeleteWorkflowGroup(context.TODO(), org, wfGrpName)
}

func deleteWorkflowFixture(wfGrpName, workflowName string) {
	client := GetClient()
	client.Workflows.DeleteWorkflow(context.TODO(), org, workflowName, wfGrpName)
}

func TestAccWorkflow_Basic(t *testing.T) {
	wfGrpName := "tf-provider-workflow-test-wfgrp"
	resourceName := "tf-provider-workflow-basic"
	updatedDescription := "Updated workflow description"

	err := createWorkflowGroupFixture(wfGrpName)
	if err != nil {
		t.Errorf("failed to create workflow group fixture: %s", err.Error())
	}
	defer deleteWorkflowGroupFixture(wfGrpName)     // runs last (registered first)
	defer deleteWorkflowFixture(wfGrpName, resourceName) // runs first (registered second)

	customHeader := http.Header{}
	customHeader.Set("x-sg-internal-auth-orgid", "sg-provider-test")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(customHeader),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccWorkflowConfig(wfGrpName, resourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "resource_name", resourceName),
					resource.TestCheckResourceAttrSet("stackguardian_workflow.test", "id"),
				),
			},
			// Update and Read
			{
				Config: testAccWorkflowConfigUpdated(wfGrpName, resourceName, updatedDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "workflow_group_id", wfGrpName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "resource_name", resourceName),
					resource.TestCheckResourceAttr("stackguardian_workflow.test", "description", updatedDescription),
				),
			},
			// ImportState
			{
				ResourceName:      "stackguardian_workflow.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["stackguardian_workflow.test"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					return fmt.Sprintf("%s/%s", rs.Primary.Attributes["workflow_group_id"], rs.Primary.ID), nil
				},
			},
			// Delete testing automatically occurs
		},
	})
}

func testAccWorkflowConfig(wfGrpName, resourceName string) string {
	return fmt.Sprintf(`
resource "stackguardian_workflow" "test" {
  workflow_group_id = "%s"
  resource_name     = "%s"
  tags              = ["test", "terraform"]
}
`, wfGrpName, resourceName)
}

func testAccWorkflowConfigUpdated(wfGrpName, resourceName, description string) string {
	return fmt.Sprintf(`
resource "stackguardian_workflow" "test" {
  workflow_group_id = "%s"
  resource_name     = "%s"
  description       = "%s"
  tags              = ["test", "terraform", "updated"]
}
`, wfGrpName, resourceName, description)
}

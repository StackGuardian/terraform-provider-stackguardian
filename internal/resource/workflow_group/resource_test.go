package workflowgroup_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
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
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
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
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
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
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
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

func TestAccWorkflowGroupOptionalId(t *testing.T) {
	testResource := `resource "stackguardian_workflow_group" "wfgrp-example-wfgrp4" {
  id = "wfgrp_example_wfgrp4"
  resource_name = "wfgrp example wfgrp4"
  description   = "Onboarding example  of terraform-provider-stackguardian for WorkflowGroup"
  tags          = ["tf-provider-example", "onboarding"]
}`
	testUpdateResource := `resource "stackguardian_workflow_group" "wfgrp-example-wfgrp4" {
  id = "wfgrp_example_wfgrp4"
  resource_name = "wfgrp example wfgrp4 updated"
  description   = "Onboarding example of terraform-provider-stackguardian for WorkflowGroup updated"
  tags          = ["tf-provider-example", "onboarding", "update"]
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{
			{
				Config: testResource,
				//Check:  resource.TestCheckResourceAttr("aws-cloud-connector-example2"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"stackguardian_workflow_group.wfgrp-example-wfgrp4",
						tfjsonpath.New("id"),
						knownvalue.StringExact("wfgrp_example_wfgrp4"),
					),
				},
			},
			{
				Config: testUpdateResource,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"stackguardian_workflow_group.wfgrp-example-wfgrp4",
						tfjsonpath.New("id"),
						knownvalue.StringExact("wfgrp_example_wfgrp4"),
					),
				},
			},
		},
	})
}

func TestAccWorkflowGroupNested_WithIdAndResourceName(t *testing.T) {
	parentName := "wfgrp-example-workflow-group4"
	nestedName := "nested-workflow-group"

	testConfig := fmt.Sprintf(`
resource "stackguardian_workflow_group" "parent" {
  resource_name = "%s"
  description   = "Parent workflow group for nested testing"
  tags          = ["tf-provider-example", "parent"]
}

resource "stackguardian_workflow_group" "nested" {
  id = "${stackguardian_workflow_group.parent.id}/%s"
  resource_name = "${stackguardian_workflow_group.parent.id}/%s"
  description   = "Nested workflow group with id and resource_name"
  tags          = ["tf-provider-example", "nested"]
}
`, parentName, nestedName, nestedName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_group.parent", "resource_name", parentName),
					resource.TestCheckResourceAttrSet("stackguardian_workflow_group.nested", "id"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"stackguardian_workflow_group.nested",
						tfjsonpath.New("id"),
						knownvalue.StringExact(parentName+"/"+nestedName),
					),
				},
			},
		},
	})
}

func TestAccWorkflowGroupNested_WithResourceNameOnly(t *testing.T) {
	parentName := "wfgrp-example-workflow-group5"
	nestedName := "nested-workflow-group-2"

	testConfig := fmt.Sprintf(`
resource "stackguardian_workflow_group" "parent" {
  resource_name = "%s"
  description   = "Parent workflow group for nested testing"
  tags          = ["tf-provider-example", "parent"]
}

resource "stackguardian_workflow_group" "nested" {
  resource_name = "${stackguardian_workflow_group.parent.id}/%s"
  description   = "Nested workflow group with resource_name only"
  tags          = ["tf-provider-example", "nested"]
}
`, parentName, nestedName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("stackguardian_workflow_group.parent", "resource_name", parentName),
					resource.TestCheckResourceAttrSet("stackguardian_workflow_group.nested", "id"),
				),
			},
		},
	})
}

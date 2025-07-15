package role_test

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
	testAccResource = `
resource "stackguardian_workflow_group" "%s" {
  resource_name = "%s"
  description   = "Example of terraform-provider-stackguardian for Workflow Group"
  tags          = ["example-tag"]
}

resource "stackguardian_role" "%s" {
  resource_name = "%s"
  description   = "Example of terraform-provider-stackguardian for a Role"
  tags = [
    "example-org",
  ]

  # Defining allowed permissions for the role
  allowed_permissions = {
    # Permission for accessing a Workflow Group
    "GET/api/v1/orgs/<org>/wfgrps/<wfGrp>/" = { # Replace with your organization name
      name = "GetWorkflowGroup",
      paths = {
        "<wfGrp>" = [
          # Referencing the workflow group resource
          stackguardian_workflow_group.%s.resource_name,
        ]
      }
    }
  }
}`

	testAccResourceUpdate = `
resource "stackguardian_workflow_group" "%s" {
  resource_name = "%s"
  description   = "Example of terraform-provider-stackguardian for Workflow Group"
  tags          = ["example-tag"]
}

resource "stackguardian_role" "%s" {
  resource_name = "%s"
  description   = "Update in Example of terraform-provider-stackguardian for a Role"
  tags = [
    "example-org",
		"update",
  ]

  # Defining allowed permissions for the role
  allowed_permissions = {
    # Permission for accessing a Workflow Group
    "GET/api/v1/orgs/<org>/wfgrps/<wfGrp>/" = { # Replace with your organization name
      name = "GetWorkflowGroup",
      paths = {
        "<wfGrp>" = [
          # Referencing the workflow group resource
          stackguardian_workflow_group.%s.resource_name,
        ]
      }
    }
  }
}`
)

func TestAccRole(t *testing.T) {
	workflowGroupResourceName := "role-example-workflow-group"
	workflowGroupName := "role-example-workflow-group"
	roleResourceName := "role-example-role"
	roleName := "role-example-role"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccResource, workflowGroupResourceName, workflowGroupName, roleResourceName, roleName, workflowGroupName),
			},
			{
				Config: fmt.Sprintf(testAccResourceUpdate, workflowGroupResourceName, workflowGroupName, roleResourceName, roleName, workflowGroupName),
			},
		},
	})
}

func TestAccRoleRecreateOnExternalDelete(t *testing.T) {
	workflowGroupResourceName := "role-example-workflow-group2"
	workflowGroupName := "role-example-workflow-group2"
	roleResourceName := "role-example-role2"
	roleName := "role-example-role2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccResource, workflowGroupResourceName, workflowGroupName, roleResourceName, roleName, workflowGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("stackguardian_role.%s", roleResourceName), "resource_name", roleName),
				),
			},
			{
				PreConfig: func() {
					client := acctest.SGClient()
					err := client.UsersRoles.DeleteRole(context.TODO(), os.Getenv("STACKGUARDIAN_ORG_NAME"), roleName)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: fmt.Sprintf(testAccResource, workflowGroupResourceName, workflowGroupName, roleResourceName, roleName, workflowGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("stackguardian_role.%s", roleResourceName), "resource_name", roleName),
				),
			},
		},
	})
}

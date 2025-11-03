package roleassignment_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	sgsdkgo "github.com/StackGuardian/sg-sdk-go"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
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
}

resource "stackguardian_role_assignment" "%s" {
  user_id     = "%s"
  entity_type = "EMAIL"
  role        = stackguardian_role.%s.resource_name
  send_email  = false
}
`

	testAccResourceUpdate = `
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
}

resource "stackguardian_role_assignment" "%s" {
  user_id     = "%s"
  entity_type = "EMAIL"
  role        = stackguardian_role.%s.resource_name
}
`
)

func TestAccRoleAssignment(t *testing.T) {
	userId := "example.user@domain.com"
	workflowGroupResourceName := "role-assign-example-workflow-group"
	workflowGroupName := "role-assign-example-workflow-group"
	roleResourceName := "role-assign-example-role"
	roleName := "role-assign-example-role"
	roleAssignmentName := "example-role-assignment"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccResource, workflowGroupResourceName, workflowGroupName, roleResourceName, roleName, workflowGroupName, roleAssignmentName, userId, roleName),
			},
			{
				Config: fmt.Sprintf(testAccResourceUpdate, workflowGroupResourceName, workflowGroupName, roleResourceName, roleName, workflowGroupName, roleAssignmentName, userId, roleName),
			},
		},
	})
}

func TestAccRoleAssignmentRecreateOnExternalDelete(t *testing.T) {
	userId := "example.user2@domain.com"
	workflowGroupResourceName := "role-assign-example-workflow-group2"
	workflowGroupName := "role-assign-example-workflow-group2"
	roleResourceName := "role-assign-example-role2"
	roleName := "role-assign-example-role2"
	roleAssignmentName := "example-role-assignment2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccResource, workflowGroupResourceName, workflowGroupName, roleResourceName, roleName, workflowGroupName, roleAssignmentName, userId, roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("stackguardian_role_assignment.%s", roleAssignmentName), "user_id", userId),
				),
			},
			{
				PreConfig: func() {
					client := acctest.SGClient()
					_, err := client.AccessManagement.DeleteUser(context.TODO(), os.Getenv("STACKGUARDIAN_ORG_NAME"), &sgsdkgo.GetorRemoveUserFromOrganization{UserId: &userId})
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: fmt.Sprintf(testAccResource, workflowGroupResourceName, workflowGroupName, roleResourceName, roleName, workflowGroupName, roleAssignmentName, userId, roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("stackguardian_role_assignment.%s", roleAssignmentName), "user_id", userId),
				),
			},
		},
	})
}

func TestAccRoleAssignmentRecreateOnChangeInUserId(t *testing.T) {
	userId := "example.user3@domain.com"
	workflowGroupResourceName := "role-assign-example-workflow-group3"
	workflowGroupName := "role-assign-example-workflow-group3"
	roleResourceName := "role-assign-example-role3"
	roleName := "role-assign-example-role3"
	roleAssignmentName := "example-role-assignment3"
	newUserId := "example.user30@domain.com"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccResource, workflowGroupResourceName, workflowGroupName, roleResourceName, roleName, workflowGroupName, roleAssignmentName, userId, roleName),
			},
			{
				Config: fmt.Sprintf(testAccResource, workflowGroupResourceName, workflowGroupName, roleResourceName, roleName, workflowGroupName, roleAssignmentName, newUserId, roleName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(fmt.Sprintf("stackguardian_role_assignment.%s", roleAssignmentName), plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func TestSendEmail(t *testing.T) {
	userId := "example.user3@domain.com"
	workflowGroupResourceName := "role-assign-example-workflow-group4"
	workflowGroupName := "role-assign-example-workflow-group4"
	roleResourceName := "role-assign-example-role4"
	roleName := "role-assign-example-role4"
	roleAssignmentName := "example-role-assignment4"

	testResource := `resource "stackguardian_workflow_group" "%s" {
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
}

resource "stackguardian_role_assignment" "%s" {
  user_id     = "%s"
  entity_type = "EMAIL"
  role        = stackguardian_role.%s.resource_name
  send_email  = %s
}
	`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testResource, workflowGroupResourceName, workflowGroupName, roleResourceName, roleName, workflowGroupName, roleAssignmentName, userId, roleName, "true"),
			},
		},
	})
}

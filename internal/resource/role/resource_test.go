package role_test

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
	workflowGroupResourceName := acctest.ResourceName("role-example-workflow-group")
	workflowGroupName := workflowGroupResourceName
	roleResourceName := acctest.ResourceName("role-example-role")
	roleName := roleResourceName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
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
	workflowGroupResourceName := acctest.ResourceName("role-example-workflow-group2")
	workflowGroupName := workflowGroupResourceName
	roleResourceName := acctest.ResourceName("role-example-role2")
	roleName := roleResourceName

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
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
					err := client.AccessManagement.DeleteRole(context.TODO(), os.Getenv("STACKGUARDIAN_ORG_NAME"), roleName)
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

func TestRoleEmptyPath(t *testing.T) {
	testResource := `
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
      paths = {}
    }
  }
}`

	roleResourceName := acctest.ResourceName("role-example-role3")
	roleName := roleResourceName
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testResource, roleResourceName, roleName),
			},
		},
	})
}

func TestAccRoleOptionalId(t *testing.T) {
	// Test if the resource has name that is not compatible with the
	roleID := acctest.ResourceName("role-example-role4")
	roleIDResourceName := acctest.ResourceName("role-example-role4-resource-name")

	const roleConfig = `
resource "stackguardian_role" "role-example-role4" {
  id = %q
  resource_name = %q
  description   = %q
  tags = [
    "example-org",
  ]

  # Defining allowed permissions for the role
  allowed_permissions = {
    # Permission for accessing a Workflow Group
    "GET/api/v1/orgs/<org>/wfgrps/<wfGrp>/" = { # Replace with your organization name
      name = "GetWorkflowGroup",
      paths = {}
    }
  }
}`
	testResource := fmt.Sprintf(roleConfig, roleID, roleIDResourceName,
		"Example of terraform-provider-stackguardian for a Role")
	testUpdateResource := fmt.Sprintf(roleConfig, roleID, roleIDResourceName,
		"Example of terraform-provider-stackguardian for a Role updated")

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
						"stackguardian_role.role-example-role4",
						tfjsonpath.New("id"),
						knownvalue.StringExact(roleID),
					),
				},
			},
			{
				Config: testUpdateResource,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"stackguardian_role.role-example-role4",
						tfjsonpath.New("id"),
						knownvalue.StringExact(roleID),
					),
				},
			},
		},
	})
}

func TestAccRoleWithoutAllowedPermissions(t *testing.T) {
	roleResourceName := acctest.ResourceName("role-example-role5")
	roleName := roleResourceName

	testResource := `
resource "stackguardian_role" "%s" {
  resource_name = "%s"
  description   = "Example of terraform-provider-stackguardian for a Role without allowed permissions"
  tags = [
	"example-org",
  ]
}`
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testResource, roleResourceName, roleName),
			},
		},
	})
}

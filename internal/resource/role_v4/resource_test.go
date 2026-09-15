package rolev4_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/StackGuardian/terraform-provider-stackguardian/internal/acctest"
	"github.com/StackGuardian/terraform-provider-stackguardian/internal/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

var org = os.Getenv("STACKGUARDIAN_ORG_NAME")

// Safety-net cleanup: Terraform's own destroy step tears these down in the normal case.
// These exist so a test that fails before reaching destroy (e.g. a failed assertion) doesn't
// leave the workflow group or role behind. Errors are ignored — the resource may already be
// gone, deleted by Terraform's own destroy step. roleID is the role's id, which defaults to
// its resource_name when the config doesn't set an explicit id.
func deleteWorkflowGroupFixture(resourceName string) {
	client := acctest.SGClient()
	client.WorkflowGroups.DeleteWorkflowGroup(context.TODO(), org, resourceName)
}

func deleteRoleFixture(roleID string) {
	client := acctest.SGClient()
	client.AccessManagement.DeleteRole(context.TODO(), org, roleID)
}

const (
	testAccResource = `
resource "stackguardian_workflow_group" "%s" {
  resource_name = "%s"
  description   = "Example of terraform-provider-stackguardian for Workflow Group"
  tags          = ["example-tag"]
}

resource "stackguardian_rolev4" "%s" {
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

resource "stackguardian_rolev4" "%s" {
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
	workflowGroupResourceName := acctest.ResourceName("rolev4-example-workflow-group")
	workflowGroupName := workflowGroupResourceName
	roleResourceName := acctest.ResourceName("rolev4-example-role")
	roleName := roleResourceName

	t.Cleanup(func() {
		deleteRoleFixture(roleName)
		deleteWorkflowGroupFixture(workflowGroupName)
	})

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
	workflowGroupResourceName := acctest.ResourceName("rolev4-example-workflow-group2")
	workflowGroupName := workflowGroupResourceName
	roleResourceName := acctest.ResourceName("rolev4-example-role2")
	roleName := roleResourceName

	t.Cleanup(func() {
		deleteRoleFixture(roleName)
		deleteWorkflowGroupFixture(workflowGroupName)
	})

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
					resource.TestCheckResourceAttr(fmt.Sprintf("stackguardian_rolev4.%s", roleResourceName), "resource_name", roleName),
				),
			},
			{
				PreConfig: func() {
					client := acctest.SGClient()
					err := client.AccessManagement.DeleteRole(context.TODO(), config.Get().OrgName, roleName)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: fmt.Sprintf(testAccResource, workflowGroupResourceName, workflowGroupName, roleResourceName, roleName, workflowGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("stackguardian_rolev4.%s", roleResourceName), "resource_name", roleName),
				),
			},
		},
	})
}

func TestRoleEmptyPath(t *testing.T) {
	testResource := `
resource "stackguardian_rolev4" "%s" {
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

	roleResourceName := acctest.ResourceName("rolev4-example-role3")
	roleName := roleResourceName

	t.Cleanup(func() { deleteRoleFixture(roleName) })

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

func TestAccRoleV4IncompatibleResourceName(t *testing.T) {
	// Test if the resource has name that is not compatible with the
	workflowGroupResourceName := acctest.ResourceName("rolev4-example-workflow-group4")
	workflowGroupName := workflowGroupResourceName
	roleName := acctest.ResourceName("rolev4-assign-example-role4")
	roleResourceName := roleName

	t.Cleanup(func() {
		deleteRoleFixture(roleName)
		deleteWorkflowGroupFixture(workflowGroupName)
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() { acctest.TestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_1_0),
		},
		ProtoV6ProviderFactories: acctest.ProviderFactories(http.Header{}),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccResource, workflowGroupName, workflowGroupResourceName, roleName, roleResourceName, workflowGroupName),
			},
			{
				Config: fmt.Sprintf(testAccResourceUpdate, workflowGroupName, workflowGroupResourceName, roleName, roleResourceName, workflowGroupName),
			},
		},
	})
}

func TestAccRoleOptionalId(t *testing.T) {
	// Test if the resource has name that is not compatible with the
	roleID := acctest.ResourceName("rolev4-example-role5")
	roleIDResourceName := acctest.ResourceName("rolev4-example-role5-resource-name")

	t.Cleanup(func() { deleteRoleFixture(roleID) })

	const roleConfig = `
resource "stackguardian_rolev4" "rolev4-example-role5" {
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
						"stackguardian_rolev4.rolev4-example-role5",
						tfjsonpath.New("id"),
						knownvalue.StringExact(roleID),
					),
				},
			},
			{
				Config: testUpdateResource,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"stackguardian_rolev4.rolev4-example-role5",
						tfjsonpath.New("id"),
						knownvalue.StringExact(roleID),
					),
				},
			},
		},
	})
}

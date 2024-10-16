<<<<<<< HEAD
resource "stackguardian_role" "ONBOARDING-Project01-Developer" {
  resource_name = "ONBOARDING-Project01-Developer"
  description   = "Onboarding example of terraform-provider-stackguardian for Role Developer"
  tags = [
    "demo-org",
  ]
  allowed_permissions = {
    // WF-GROUP
    "GET/api/v1/orgs/demo-org/wfgrps/<wfGrp>/" : {
      "name" : "GetWorkflowGroup",
      "paths" : {
        "<wfGrp>" : [
          resource.stackguardian_workflow_group.ONBOARDING-Project01-Frontend.resource_name,
        ]
      }
    }
  }
}

resource "stackguardian_workflow_group" "ONBOARDING-Project01-Frontend" {
  resource_name = "ONBOARDING-Project01-Frontend"
  description   = "Onboarding example  of terraform-provider-stackguardian for WorkflowGroup"
  tags          = ["tf-provider-example", "onboarding"]
=======
# Example for defining a Role and associating it with a Workflow Group in StackGuardian

resource "stackguardian_workflow_group" "example_workflow_group" {
  resource_name = "example-workflow-group"
  description   = "Example of terraform-provider-stackguardian for Workflow Group"
  tags          = ["example-tag"]
>>>>>>> main
}

resource "stackguardian_role" "example_role" {
  resource_name = "example-role"
  description   = "Example of terraform-provider-stackguardian for Role"

  tags = [
    "developer", "example-role",
  ]

  # Defining permissions for the role
  allowed_permissions = {
    # Permission for accessing a Workflow Group
    "GET/api/v1/orgs/<org>/wfgrps/<wfGrp>/" = { # Replace with your organization name
      name = "GetWorkflowGroup",
      paths = {
        "<wfGrp>" = [
          # Referencing the workflow group resource
          stackguardian_workflow_group.example_workflow_group.resource_name
        ]
      }
    }
  }
}


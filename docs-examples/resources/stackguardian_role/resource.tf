# Example for defining a Role and associating it with a Workflow Group in StackGuardian

resource "stackguardian_workflow_group" "example_workflow_group" {
  resource_name = "example-workflow-group"
  description   = "Example of terraform-provider-stackguardian for Workflow Group"
  tags          = ["example-tag"]
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
    "GET/api/v1/orgs/<organization>/wfgrps/<wfGrp>/": { # Replace with your organization name
      "name" : "GetWorkflowGroup",
      "paths": {
        "<wfGrp>": [
          # Referencing the workflow group resource
          stackguardian_workflow_group.example_workflow_group.resource_name
        ]
      }
    }
  }
}


resource "stackguardian_role_assignment" "ONBOARDING-Project01-Frontend-Developer" {
  user_id     = "frontend.developer.p01@dummy.com"
  entity_type = "EMAIL"
  role        = resource.stackguardian_role.ONBOARDING-Project01-Developer.resource_name
}

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
}

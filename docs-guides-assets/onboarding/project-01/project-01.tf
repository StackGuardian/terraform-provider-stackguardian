// Project-01

terraform {
  required_providers {
    stackguardian = {
      source  = "terraform/provider/stackguardian"
      version = "0.0.1"
    }
  }
}

provider "stackguardian" {
  api_key  = "<YOUR-API-KEY>"                            # Replace this with your API key
  org_name = "<YOUR-ORG-NAME>"                           # Replace this with your organization name
  api_uri  = "https://testapi.qa.stackguardian.io"      # Use testapi instead of production for testing
}


resource "stackguardian_role" "ONBOARDING-Project01-Developer" {
    resource_name = "ONBOARDING-Project01-Developer"
    description = "Onboarding example of terraform-provider-stackguardian for Role Developer"
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
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Backend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-DevOps.resource_name,
          ]
        }
      },

      // WF
      "GET/api/v1/orgs/demo-org/wfgrps/<wfGrp>/wfs/<wf>/" : {
        "name" : "GetWorkflow",
        "paths" : {
          "<wfGrp>" : [
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Frontend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Backend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-DevOps.resource_name,
          ],
          "<wf>" : [
            ".*"
          ]
        }
      },
      "POST/api/v1/orgs/demo-org/wfgrps/<wfGrp>/wfs/" : {
        "name" : "CreateWorkflow",
        "paths" : {
          "<wfGrp>" : [
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Frontend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Backend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-DevOps.resource_name,
          ]
        }
      },
      "PATCH/api/v1/orgs/demo-org/wfgrps/<wfGrp>/wfs/<wf>/" : {
        "name" : "UpdateWorkflow",
        "paths" : {
          "<wfGrp>" : [
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Frontend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Backend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-DevOps.resource_name,
          ],
          "<wf>" : [
            ".*"
          ]
        }
      },
      "DELETE/api/v1/orgs/demo-org/wfgrps/<wfGrp>/wfs/<wf>/" : {
        "name" : "DeleteWorkflow",
        "paths" : {
          "<wfGrp>" : [
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Frontend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Backend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-DevOps.resource_name,
          ],
          "<wf>" : [
            ".*"
          ]
        }
      },

      // WF-RUN
      "GET/api/v1/orgs/demo-org/wfgrps/<wfGrp>/wfs/<wf>/wfruns/<wfRun>/" : {
        "name" : "GetWorkflowRun",
        "paths" : {
          "<wfRun>" : [
            ".*"
          ],
          "<wfGrp>" : [
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Frontend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Backend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-DevOps.resource_name,
          ],
          "<wf>" : [
            ".*"
          ]
        }
      },
      "POST/api/v1/orgs/demo-org/wfgrps/<wfGrp>/wfs/<wf>/wfruns/" : {
        "name" : "CreateWorkflowRun",
        "paths" : {
          "<wfGrp>" : [
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Frontend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Backend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-DevOps.resource_name,
          ],
          "<wf>" : [
            ".*"
          ]
        }
      },
      "DELETE/api/v1/orgs/demo-org/wfgrps/<wfGrp>/wfs/<wf>/wfruns/<wfRun>/" : {
        "name" : "UpdateWorkflowRun",
        "paths" : {
          "<wfRun>" : [
            ".*"
          ],
          "<wfGrp>" : [
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Frontend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Backend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-DevOps.resource_name,
          ],
          "<wf>" : [
            ".*"
          ]
        }
      },
      "POST/api/v1/orgs/demo-org/wfgrps/<wfGrp>/wfs/<wf>/wfruns/<wfRun>/resume/" : {
        "name" : "ResumeWorkflowRun",
        "paths" : {
          "<wfRun>" : [
            ".*"
          ],
          "<wfGrp>" : [
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Frontend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Backend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-DevOps.resource_name,
          ],
          "<wf>" : [
            ".*"
          ]
        }
      },
      "GET/api/v1/orgs/demo-org/wfgrps/<wfGrp>/wfs/<wf>/wfruns/<wfRun>/logs/" : {
        "name" : "GetWorkflowRunLogs",
        "paths" : {
          "<wfRun>" : [
            ".*"
          ],
          "<wfGrp>" : [
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Frontend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Backend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-DevOps.resource_name,
          ],
          "<wf>" : [
            ".*"
          ]
        }
      },

      // WF-RUN-FACTS
      "GET/api/v1/orgs/demo-org/wfgrps/<wfGrp>/wfs/<wf>/wfruns/<wfRun>/wfrunfacts/<wfRunFacts>/" : {
        "name" : "GetWorkflowRunFact",
        "paths" : {
          "<wfGrp>" : [
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Frontend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Backend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-DevOps.resource_name,
          ],
          "<wfRun>" : [
            ".*"
          ],
          "<wf>" : [
            ".*"
          ],
          "<wfRunFacts>" : [
            ".*"
          ]
        }
      },

      // WF-ARTIFACTS
      "GET/api/v1/orgs/demo-org/wfgrps/<wfGrp>/wfs/<wf>/listall_artifacts/" : {
        "name" : "ListWorkflowArtifacts",
        "paths" : {
          "<wfGrp>" : [
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Frontend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Backend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-DevOps.resource_name,
          ],
          "<wf>" : [
            ".*"
          ]
        }
      },

      // WF-OUTPUTS
      "GET/api/v1/orgs/demo-org/wfgrps/<wfGrp>/wfs/<wf>/outputs/" : {
        "name" : "GetWorkflowOutputs",
        "paths" : {
          "<wfGrp>" : [
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Frontend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Backend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-DevOps.resource_name,
          ],
          "<wf>" : [
            ".*"
          ]
        }
      },

      // AUDIT
      "GET/api/v1/orgs/demo-org/audit_logs/" : {
        "name" : "GetAuditLogs",
        "paths" : {}
      },

      // SECRET
      "GET/api/v1/orgs/demo-org/secrets/listall/" : {
        "name" : "ListSecrets",
        "paths" : {}
      },
      "POST/api/v1/orgs/demo-org/secrets/" : {
        "name" : "CreateSecret",
        "paths" : {}
      },
      "PATCH/api/v1/orgs/demo-org/secrets/<secret>/" : {
        "name" : "UpdateSecret",
        "paths" : {}
      },
      "DELETE/api/v1/orgs/demo-org/secrets/<secret>/" : {
        "name" : "DeleteSecret",
        "paths" : {}
      },

      // INTEGRATION
      "GET/api/v1/orgs/demo-org/integrationgroups/<integrationgroup>/" : {
        "name" : "GetIntegrationGroup",
        "paths" : {
          "<integrationgroup>" : [
            ".*"
          ]
        }
      },
      "GET/api/v1/orgs/demo-org/integrationgroups/<integrationgroup>/integrations/<integration>/" : {
        "name" : "GetIntegrationGroupChild",
        "paths" : {
          "<integration>" : [
            "ONBOARDING-Project01"
          ],
          "<integrationgroup>" : [
            ".*"
          ]
        }
      },

    }

  depends_on = [
    stackguardian_workflow_group.ONBOARDING-Project01-Frontend,
    stackguardian_workflow_group.ONBOARDING-Project01-Backend,
    stackguardian_workflow_group.ONBOARDING-Project01-DevOps,
  ]
}

resource "stackguardian_workflow_group" "ONBOARDING-Project01-Backend" {
    resource_name = "ONBOARDING-Project01-Backend"
    description = "Onboarding example  of terraform-provider-stackguardian for WorkflowGroup"
    tags = ["tf-provider-example", "onboarding"]
}

resource "stackguardian_workflow_group" "ONBOARDING-Project01-Frontend" {
    resource_name = "ONBOARDING-Project01-Frontend"
    description = "Onboarding example  of terraform-provider-stackguardian for WorkflowGroup"
    tags = ["tf-provider-example", "onboarding"]
}

resource "stackguardian_workflow_group" "ONBOARDING-Project01-DevOps" {
    resource_name = "ONBOARDING-Project01-DevOps"
    description = "Onboarding example  of terraform-provider-stackguardian for WorkflowGroup"
    tags = ["tf-provider-example", "onboarding"]
}

resource "stackguardian_role_assignment" "ONBOARDING-Project01-Frontend-Developer" {
  user_id = "frontend.developer.p01@dummy.com"
  entity_type = "EMAIL"
  role = resource.stackguardian_role.ONBOARDING-Project01-Developer.resource_name
}



resource "stackguardian_connector" "ONBOARDING-Project01-Cloud-Connector" {
  resource_name = "ONBOARDING-Project01-Cloud-Connector"
  description = "Onboarding example  of terraform-provider-stackguardian for ConnectorCloud"
  settings = {
    kind = "AWS_STATIC",
    config = [{
        aws_access_key_id = "REPLACEME-aws-key",
        aws_secret_access_key = "REPLACEME-aws-key",
        aws_default_region = "us-west-2"
      }]
  }
}


resource "stackguardian_connector" "ONBOARDING-Project01-VCS-Connector" {
  resource_name = "ONBOARDING-Project01-VCS-Connector"
  description = "Onboarding example of terraform-provider-stackguardian for ConnectorVcs"
  settings = {
    kind = "GITLAB_COM",
    config = [{
        gitlab_creds = "REPLACEME-example-user:REPLACEME-example-token"
      }]
  }
}

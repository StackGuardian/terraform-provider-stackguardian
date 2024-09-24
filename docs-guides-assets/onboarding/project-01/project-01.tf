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
  api_key  = ""
  org_name = ""
  api_uri  = ""
}


resource "stackguardian_role" "ONBOARDING-Project01-Developer" {
    resource_name = "ONBOARDING-Project01-Developer"
    description = "Onboarding example of terraform-provider-stackguardian for Role Developer"
    tags = [
      "wicked-hop",
    ]
    allowed_permissions = {
      // WF-GROUP
      "GET/api/v1/orgs/wicked-hop/wfgrps/<wfGrp>/" : {
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
      "GET/api/v1/orgs/wicked-hop/wfgrps/<wfGrp>/wfs/<wf>/" : {
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
      "POST/api/v1/orgs/wicked-hop/wfgrps/<wfGrp>/wfs/" : {
        "name" : "CreateWorkflow",
        "paths" : {
          "<wfGrp>" : [
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Frontend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-Backend.resource_name,
            resource.stackguardian_workflow_group.ONBOARDING-Project01-DevOps.resource_name,
          ]
        }
      },
      "PATCH/api/v1/orgs/wicked-hop/wfgrps/<wfGrp>/wfs/<wf>/" : {
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
      "DELETE/api/v1/orgs/wicked-hop/wfgrps/<wfGrp>/wfs/<wf>/" : {
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
      "GET/api/v1/orgs/wicked-hop/wfgrps/<wfGrp>/wfs/<wf>/wfruns/<wfRun>/" : {
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
      "POST/api/v1/orgs/wicked-hop/wfgrps/<wfGrp>/wfs/<wf>/wfruns/" : {
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
      "DELETE/api/v1/orgs/wicked-hop/wfgrps/<wfGrp>/wfs/<wf>/wfruns/<wfRun>/" : {
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
      "POST/api/v1/orgs/wicked-hop/wfgrps/<wfGrp>/wfs/<wf>/wfruns/<wfRun>/resume/" : {
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
      "GET/api/v1/orgs/wicked-hop/wfgrps/<wfGrp>/wfs/<wf>/wfruns/<wfRun>/logs/" : {
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
      "GET/api/v1/orgs/wicked-hop/wfgrps/<wfGrp>/wfs/<wf>/wfruns/<wfRun>/wfrunfacts/<wfRunFacts>/" : {
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
      "GET/api/v1/orgs/wicked-hop/wfgrps/<wfGrp>/wfs/<wf>/listall_artifacts/" : {
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
      "GET/api/v1/orgs/wicked-hop/wfgrps/<wfGrp>/wfs/<wf>/outputs/" : {
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
      "GET/api/v1/orgs/wicked-hop/audit_logs/" : {
        "name" : "GetAuditLogs",
        "paths" : {}
      },

      // SECRET
      "GET/api/v1/orgs/wicked-hop/secrets/listall/" : {
        "name" : "ListSecrets",
        "paths" : {}
      },
      "POST/api/v1/orgs/wicked-hop/secrets/" : {
        "name" : "CreateSecret",
        "paths" : {}
      },
      "PATCH/api/v1/orgs/wicked-hop/secrets/<secret>/" : {
        "name" : "UpdateSecret",
        "paths" : {}
      },
      "DELETE/api/v1/orgs/wicked-hop/secrets/<secret>/" : {
        "name" : "DeleteSecret",
        "paths" : {}
      },

      // INTEGRATION
      "GET/api/v1/orgs/wicked-hop/integrationgroups/<integrationgroup>/" : {
        "name" : "GetIntegrationGroup",
        "paths" : {
          "<integrationgroup>" : [
            ".*"
          ]
        }
      },
      "GET/api/v1/orgs/wicked-hop/integrationgroups/<integrationgroup>/integrations/<integration>/" : {
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


#Commented until connectors is ready for testing

# resource "stackguardian_connector" "ONBOARDING-Project01-Cloud-Connector" {
#   organization = "demo-org"
#   resource_name = "ONBOARDING-Project01"
#   description = "Onboarding example  of terraform-provider-stackguardian for ConnectorCloud"
#   settings = {
#     kind = "AWS_STATIC",
#     config = jsonencode([
#       {
#         "awsAccessKeyId" : "REPLACEME-aws-key",
#         "awsSecretAccessKey" : "REPLACEME-aws-key",
#         "awsDefaultRegion" : "REPLACEME-us-west-2"
#       }
#     ])
#   }
# }


# resource "stackguardian_connector" "ONBOARDING-Project01-VCS-Connector" {
#   organization = "demo-org"
#   resource_name = "ONBOARDING-Project01"
#   description = "Onboarding example of terraform-provider-stackguardian for ConnectorVcs"
#   settings = {
#     kind = "GITHUB_COM",
#     config = jsonencode([
#       {
#         "gitlabCreds" : "REPLACEME-example-user:REPLACEME-example-token"
#       }
#     ])
#   }
# }

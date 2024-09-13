// Project-01

terraform {
  required_providers {
    stackguardian = {
      source = "terraform/provider/stackguardian"

      # https://developer.hashicorp.com/terraform/language/expressions/version-constraints#version-constraint-behavior
      # NOTE: A prerelease version can be selected only by an exact version constraint.
      version = "0.0.0-dev"
    }
  }
}

provider "stackguardian" {}


resource "stackguardian_role" "ONBOARDING-Project01-Developer" {
  data = jsonencode({
    "ResourceName" : "ONBOARDING-Project01-Developer",
    //"Description" : "Onboarding example of terraform-provider-stackguardian for Role Developer",
    "Actions" : [
      "wicked-hop",
    ],
    "AllowedPermissions" : {

      // WF-GROUP
      "GET/api/v1/orgs/wicked-hop/wfgrps/<wfGrp>/" : {
        "name" : "GetWorkflowGroup",
        "paths" : {
          "<wfGrp>" : [
            "ONBOARDING-Project01-Frontend",
            "ONBOARDING-Project01-Backend",
            "ONBOARDING-Project01-DevOps"
          ]
        }
      },

      // WF
      "GET/api/v1/orgs/wicked-hop/wfgrps/<wfGrp>/wfs/<wf>/" : {
        "name" : "GetWorkflow",
        "paths" : {
          "<wfGrp>" : [
            "ONBOARDING-Project01-Frontend",
            "ONBOARDING-Project01-Backend",
            "ONBOARDING-Project01-DevOps"
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
            "ONBOARDING-Project01-Frontend",
            "ONBOARDING-Project01-Backend",
            "ONBOARDING-Project01-DevOps"
          ]
        }
      },
      "PATCH/api/v1/orgs/wicked-hop/wfgrps/<wfGrp>/wfs/<wf>/" : {
        "name" : "UpdateWorkflow",
        "paths" : {
          "<wfGrp>" : [
            "ONBOARDING-Project01-Frontend",
            "ONBOARDING-Project01-Backend",
            "ONBOARDING-Project01-DevOps"
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
            "ONBOARDING-Project01-Frontend",
            "ONBOARDING-Project01-Backend",
            "ONBOARDING-Project01-DevOps"
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
            "ONBOARDING-Project01-Frontend",
            "ONBOARDING-Project01-Backend",
            "ONBOARDING-Project01-DevOps"
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
            "ONBOARDING-Project01-Frontend",
            "ONBOARDING-Project01-Backend",
            "ONBOARDING-Project01-DevOps"
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
            "ONBOARDING-Project01-Frontend",
            "ONBOARDING-Project01-Backend",
            "ONBOARDING-Project01-DevOps"
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
            "ONBOARDING-Project01-Frontend",
            "ONBOARDING-Project01-Backend",
            "ONBOARDING-Project01-DevOps"
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
            "ONBOARDING-Project01-Frontend",
            "ONBOARDING-Project01-Backend",
            "ONBOARDING-Project01-DevOps"
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
            "ONBOARDING-Project01-Frontend",
            "ONBOARDING-Project01-Backend",
            "ONBOARDING-Project01-DevOps"
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
            "ONBOARDING-Project01-Frontend",
            "ONBOARDING-Project01-Backend",
            "ONBOARDING-Project01-DevOps"
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
            "ONBOARDING-Project01-Frontend",
            "ONBOARDING-Project01-Backend",
            "ONBOARDING-Project01-DevOps"
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

      // POLICY
      "GET/api/v1/orgs/wicked-hop/policies/<policy>/" : {
        "name" : "GetPolicy",
        "paths" : {
          "<policy>" : [
            "ONBOARDING-Project01"
          ]
        }
      }

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

    },
  })

  depends_on = [
    stackguardian_workflow_group.ONBOARDING-Project01-Frontend,
    stackguardian_workflow_group.ONBOARDING-Project01-Backend,
    stackguardian_workflow_group.ONBOARDING-Project01-DevOps,
    stackguardian_policy.ONBOARDING-Project01,
  ]
}

resource "stackguardian_workflow_group" "ONBOARDING-Project01-Backend" {
  data = jsonencode({
    "ResourceName" : "ONBOARDING-Project01-Backend",
    "Description" : "Onboarding example  of terraform-provider-stackguardian for WorkflowGroup",
    "Tags" : ["tf-provider-example", "onboarding"],
    "IsActive" : 1,
  })
}

resource "stackguardian_workflow_group" "ONBOARDING-Project01-Frontend" {
  data = jsonencode({
    "ResourceName" : "ONBOARDING-Project01-Frontend",
    "Description" : "Onboarding example  of terraform-provider-stackguardian for WorkflowGroup",
    "Tags" : ["tf-provider-example", "onboarding"],
    "IsActive" : 1,
  })
}

resource "stackguardian_workflow_group" "ONBOARDING-Project01-DevOps" {
  data = jsonencode({
    "ResourceName" : "ONBOARDING-Project01-DevOps",
    "Description" : "Onboarding example  of terraform-provider-stackguardian for WorkflowGroup",
    "Tags" : ["tf-provider-example", "onboarding"],
    "IsActive" : 1,
  })
}

resource "stackguardian_policy" "ONBOARDING-Project01" {
  data = jsonencode({
    "ResourceName" : "ONBOARDING-Project01",
    "Description" : "Onboarding example  of terraform-provider-stackguardian for Policy",
    "Tags" : ["tf-provider-example", "onboarding"]
  })
}


resource "stackguardian_connector_cloud" "ONBOARDING-Project01" {
  data = jsonencode({
    "ResourceName" : "ONBOARDING-Project01",
    "Tags" : ["tf-provider-example", "onboarding"]
    "Description" : "Onboarding example  of terraform-provider-stackguardian for ConnectorCloud",
    "Settings" : {
      "kind" : "AWS_STATIC",
      "config" : [
        {
          "awsAccessKeyId" : "REPLACEME-aws-key",
          "awsSecretAccessKey" : "REPLACEME-aws-key",
          "awsDefaultRegion" : "REPLACEME-us-west-2"
        }
      ]
    }
  })
}


resource "stackguardian_connector_vcs" "ONBOARDING-Project01" {
  data = jsonencode({
    "ResourceName" : "ONBOARDING-Project01",
    "ResourceType" : "INTEGRATION.GITLAB_COM",
    "Tags" : ["tf-provider-example", "onboarding"]
    "Description" : "Onboarding example of terraform-provider-stackguardian for ConnectorVcs",
    "Settings" : {
      "kind" : "GITLAB_COM",
      "config" : [
        {
          "gitlabCreds" : "REPLACEME-example-user:REPLACEME-example-token"
        }
      ]
    },
  })
}


// --- Non-onboarding resources:

resource "stackguardian_workflow" "ONBOARDING-Project01-DevOps-Wf01" {
  wfgrp = stackguardian_workflow_group.ONBOARDING-Project01-DevOps.id

  data = jsonencode({
    "ResourceName" : "ONBOARDING-Project01-DevOps-Wf01",
    "Description" : "Example of StackGuardian Workflow: Deploy a website from AWS S3",
    "Tags" : ["tf-provider-test", "onboarding"],
    "EnvironmentVariables" : [],
    "DeploymentPlatformConfig" : [{
      "kind" : "AWS_RBAC",
      "config" : {
        "integrationId" : "/integrations/aws"
      }
    }],
    "VCSConfig" : {
      "iacVCSConfig" : {
        "useMarketplaceTemplate" : true,
        "iacTemplate" : "/stackguardian/aws-s3-demo-website",
        "iacTemplateId" : "/stackguardian/aws-s3-demo-website:4"
      },
      "iacInputData" : {
        "schemaType" : "FORM_JSONSCHEMA",
        "data" : {
          "shop_name" : "StackGuardian",
          "bucket_region" : "eu-central-1"
        }
      }
    },
    "Approvers" : [],
    "TerraformConfig" : {
      "managedTerraformState" : true,
      "terraformVersion" : "1.4.6"
    },
    "WfType" : "TERRAFORM",
    "UserSchedules" : []
  })
}

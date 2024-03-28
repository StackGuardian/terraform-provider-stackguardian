// project-test

terraform {
  required_providers {
    stackguardian = {
      source = "terraform/provider/stackguardian"

      # https://developer.hashicorp.com/terraform/language/expressions/version-constraints#version-constraint-behavior
      # NOTE: A prerelease version can be selected only by an exact version constraint.
      version = "0.0.0-dev" #provider-version
    }
  }
}

provider "stackguardian" {}

# Dv: Developer
resource "stackguardian_role" "TPS-OBT-Dv-T000000" {
  data = jsonencode({
    "ResourceName" : "TPS-OBT-Dv-T000000",
    //"Description" : "Onboarding test of terraform-provider-stackguardian for Role Developer",
    "Tags" : ["tf-provider-test", "onboarding"],
    "Actions" : [
      "wicked-hop",
    ],
    "AllowedPermissions" : {

      // WF-GROUP
      "GET/api/v1/orgs/wicked-hop/wfgrps/<wfGrp>/" : {
        "name" : "GetWorkflowGroup",
        "paths" : {
          "<wfGrp>" : [
            "TPS-OBT-Frontend-T000000",
            "TPS-OBT-Backend-T000000",
            "TPS-OBT-DevOps-T000000"
          ]
        }
      },

      // WF
      "GET/api/v1/orgs/wicked-hop/wfgrps/<wfGrp>/wfs/<wf>/" : {
        "name" : "GetWorkflow",
        "paths" : {
          "<wfGrp>" : [
            "TPS-OBT-Frontend-T000000",
            "TPS-OBT-Backend-T000000",
            "TPS-OBT-DevOps-T000000"
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
            "TPS-OBT-Frontend-T000000",
            "TPS-OBT-Backend-T000000",
            "TPS-OBT-DevOps-T000000"
          ]
        }
      },
      "PATCH/api/v1/orgs/wicked-hop/wfgrps/<wfGrp>/wfs/<wf>/" : {
        "name" : "UpdateWorkflow",
        "paths" : {
          "<wfGrp>" : [
            "TPS-OBT-Frontend-T000000",
            "TPS-OBT-Backend-T000000",
            "TPS-OBT-DevOps-T000000"
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
            "TPS-OBT-Frontend-T000000",
            "TPS-OBT-Backend-T000000",
            "TPS-OBT-DevOps-T000000"
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
            "TPS-OBT-Frontend-T000000",
            "TPS-OBT-Backend-T000000",
            "TPS-OBT-DevOps-T000000"
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
            "TPS-OBT-Frontend-T000000",
            "TPS-OBT-Backend-T000000",
            "TPS-OBT-DevOps-T000000"
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
            "TPS-OBT-Frontend-T000000",
            "TPS-OBT-Backend-T000000",
            "TPS-OBT-DevOps-T000000"
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
            "TPS-OBT-Frontend-T000000",
            "TPS-OBT-Backend-T000000",
            "TPS-OBT-DevOps-T000000"
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
            "TPS-OBT-Frontend-T000000",
            "TPS-OBT-Backend-T000000",
            "TPS-OBT-DevOps-T000000"
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
            "TPS-OBT-Frontend-T000000",
            "TPS-OBT-Backend-T000000",
            "TPS-OBT-DevOps-T000000"
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
            "TPS-OBT-Frontend-T000000",
            "TPS-OBT-Backend-T000000",
            "TPS-OBT-DevOps-T000000"
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
            "TPS-OBT-Frontend-T000000",
            "TPS-OBT-Backend-T000000",
            "TPS-OBT-DevOps-T000000"
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
            "TPS-OBT-T000000"
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
            "TPS-OBT-T000000"
          ],
          "<integrationgroup>" : [
            ".*"
          ]
        }
      },

    },
  })

  depends_on = [
    stackguardian_workflow_group.TPS-OBT-Frontend-T000000,
    stackguardian_workflow_group.TPS-OBT-Backend-T000000,
    stackguardian_workflow_group.TPS-OBT-DevOps-T000000,
    stackguardian_policy.TPS-OBT-T000000,
  ]
}

resource "stackguardian_workflow_group" "TPS-OBT-Backend-T000000" {
  data = jsonencode({
    "ResourceName" : "TPS-OBT-Backend-T000000",
    "Description" : "Onboarding test  of terraform-provider-stackguardian for WorkflowGroup",
    "Tags" : ["tf-provider-test", "onboarding"],
    "IsActive" : 1,
  })
}

resource "stackguardian_workflow_group" "TPS-OBT-Frontend-T000000" {
  data = jsonencode({
    "ResourceName" : "TPS-OBT-Frontend-T000000",
    "Description" : "Onboarding test  of terraform-provider-stackguardian for WorkflowGroup",
    "Tags" : ["tf-provider-test", "onboarding"],
    "IsActive" : 1,
  })
}

resource "stackguardian_workflow_group" "TPS-OBT-DevOps-T000000" {
  data = jsonencode({
    "ResourceName" : "TPS-OBT-DevOps-T000000",
    "Description" : "Onboarding test  of terraform-provider-stackguardian for WorkflowGroup",
    "Tags" : ["tf-provider-test", "onboarding"],
    "IsActive" : 1,
  })
}

resource "stackguardian_policy" "TPS-OBT-T000000" {
  data = jsonencode({
    "ResourceName" : "TPS-OBT-T000000",
    "Description" : "Onboarding test  of terraform-provider-stackguardian for Policy",
    "Tags" : ["tf-provider-test", "onboarding"]
  })
}

//
resource "stackguardian_connector_cloud" "TPS-OBT-T000000" {
  data = jsonencode({
    "ResourceName" : "TPS-OBT-T000000",
    "Tags" : ["tf-provider-test", "onboarding"]
    "Description" : "Onboarding test  of terraform-provider-stackguardian for ConnectorCloud",
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
//

resource "stackguardian_connector_vcs" "TPS-OBT-T000000" {
  data = jsonencode({
    "ResourceName" : "TPS-OBT-T000000",
    "ResourceType" : "INTEGRATION.GITLAB_COM",
    "Tags" : ["tf-provider-test", "onboarding"]
    "Description" : "Onboarding test of terraform-provider-stackguardian for ConnectorVcs",
    "Settings" : {
      "kind" : "GITLAB_COM",
      "config" : [
        {
          "gitlabCreds" : "REPLACEME-test-user:REPLACEME-test-token"
        }
      ]
    },
  })
}


// --- Non-onboarding resources:

resource "stackguardian_workflow" "TPS-OBT-DevOps-T000000" {
  wfgrp = stackguardian_workflow_group.TPS-OBT-DevOps-T000000.id

  data = jsonencode({
    "ResourceName": "TPS-OBT-DevOps-T000000",
    "Description": "Example of StackGuardian Workflow: Deploy a website from AWS S3",
    "Tags": ["tf-provider-test", "onboarding"],
    "EnvironmentVariables": [],
    "DeploymentPlatformConfig": [{
      "kind": "AWS_RBAC",
      "config": {
        "integrationId": "/integrations/aws"
      }
    }],
    "VCSConfig": {
      "iacVCSConfig": {
        "useMarketplaceTemplate": true,
        "iacTemplate": "/stackguardian/aws-s3-demo-website",
        "iacTemplateId": "/stackguardian/aws-s3-demo-website:4"
      },
      "iacInputData": {
        "schemaType": "FORM_JSONSCHEMA",
        "data": {
          "shop_name": "StackGuardian",
          "bucket_region": "eu-central-1"
        }
      }
    },
    "Approvers": [],
    "TerraformConfig": {
      "managedTerraformState": true,
      "terraformVersion": "1.4.6"
    },
    "WfType": "TERRAFORM",
    "UserSchedules": []
  })
}

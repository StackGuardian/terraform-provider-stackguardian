resource "stackguardian_role" "test" {
  resource_name = "testing"
  allowed_permissions = {
    "GET/api/v1/orgs/demo-org/wfgrps/<wfGrp>/" : {
      "name" : "GetWorkflowGroup",
      "paths" : {
        "<wfGrp>" : [".*"]
      }
    },
  }
}

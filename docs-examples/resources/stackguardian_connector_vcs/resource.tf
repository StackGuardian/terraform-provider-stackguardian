resource "stackguardian_connector_vcs" "TPS-Example-ConnectorVcs" {
  // integrationgroup = "TPS-Example"
  data = jsonencode({
    "ResourceName" : "TPS-Example-ConnectorVcs",
    "ResourceType" : "INTEGRATION.GITLAB_COM",
    "Tags" : ["tf-provider-example"]
    "Description" : "Example of terraform-provider-stackguardian for ConnectorVcs",
    "Settings" : {
      "kind" : "GITLAB_COM",
      "config" : [
        {
          "gitlabCreds" : "example-user:example-token"
        }
      ]
    },
  })
}

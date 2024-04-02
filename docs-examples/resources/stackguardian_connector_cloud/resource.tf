resource "stackguardian_connector_cloud" "TPS-Example-ConnectorCloud" {
  // integrationgroup = "TPS-Example"
  data = jsonencode({
    "ResourceName" : "TPS-Example-ConnectorCloud",
    "Tags" : ["tf-provider-example"]
    "Description" : "Example of terraform-provider-stackguardian for ConnectorCloud",
    "Settings" : {
      "kind" : "AWS_STATIC",
      "config" : [
        {
          "awsAccessKeyId" : "example-aws-key",
          "awsSecretAccessKey" : "example-aws-key",
          "awsDefaultRegion" : "us-west-2"
        }
      ]
    }
  })
}

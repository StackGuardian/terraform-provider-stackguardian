resource "stackguardian_connector" "aws-cloud-connector-example" {
  resource_name = "aws-rbac-connector"
  description   = "AWS Cloud Connector"

  settings = {
    kind = "AWS_RBAC"

    config = [{
      roleArn         = "arn:aws:iam::209502960327:role/StackGuardian"
      externalId      = "wicked-hop:ElfygiFglfldTwnDFpAScQkvgvHTGV "
      durationSeconds = "3600"
    }]
  }
}

resource "stackguardian_connector" "ONBOARDING-Project01-Cloud-Connector" {
  resource_name = "ONBOARDING-Project01-Cloud-Connector"
  description   = "Onboarding example  of terraform-provider-stackguardian for ConnectorCloud"
  settings = {
    kind = "AWS_STATIC",
    config = [{
      aws_access_key_id     = "REPLACEME-aws-key",
      aws_secret_access_key = "REPLACEME-aws-key",
      aws_default_region    = "us-west-2"
    }]
  }
}

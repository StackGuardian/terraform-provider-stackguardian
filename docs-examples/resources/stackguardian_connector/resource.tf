resource "stackguardian_connector" "example-connector" {
  resource_name = "example-connector"
  description   = "Onboarding example  of terraform-provider-stackguardian for ConnectorCloud"
  settings = {
    kind = "AWS_STATIC",
    config = [{
      aws_access_key_id     = "REPLACEME-aws-key",
      aws_secret_access_key = "REPLACEME-aws-key",
      aws_default_region    = "us-west-2"
    }]
  }
  scope = ["*"]
}

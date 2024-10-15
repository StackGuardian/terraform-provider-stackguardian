resource "stackguardian_connector" "example-connector" {
  resource_name = "example-connector"
  description   = "Onboarding example  of terraform-provider-stackguardian for ConnectorCloud"
  settings = {
    kind = "AWS_STATIC",
    config = [{
      roleArn     = "arn:aws:iam::209502960327:role/StackGuardian"
      externalId = "demo-org:ElfygiFglfldTwnDFpAScQkvgvHTGV "
      durationSeconds    = "3600"
    }]
  }
}

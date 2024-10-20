resource "stackguardian_connector" "aws-cloud-connector-example" {
  resource_name = "aws-rbac-connector"
  description   = "AWS Cloud Connector"

  settings = {
    kind = "AWS_RBAC"

    config = [{
      role_arn         = "arn:aws:iam::209502960327:role/StackGuardian"
      external_id      = "wicked-hop:ElfygiFglfldTwnDFpAScQkvgvHTGV "
      duration_seconds = "3600"
    }]
  }
}

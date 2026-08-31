# `settings.kind` selects the connector type, and each kind takes a different set of
# `settings.config` fields. The kind documentation below lists them all; three common
# ones are shown here.

# Cloud connector using an AWS IAM role. Preferred over static keys.
resource "stackguardian_connector" "aws_rbac" {
  resource_name = "aws-rbac-connector"
  description   = "AWS credentials assumed via IAM role"

  settings = {
    kind = "AWS_RBAC"
    config = [{
      role_arn         = "arn:aws:iam::000000000000:role/StackGuardian"
      external_id      = "my-org:ElfygiFglfldTwnDFpAScQkvgvHTGV"
      duration_seconds = "3600"
    }]
  }
}

# Cloud connector using an Azure service principal.
resource "stackguardian_connector" "azure" {
  resource_name = "azure-connector"
  description   = "Azure service principal credentials"

  settings = {
    kind = "AZURE_STATIC"
    config = [{
      arm_client_id       = "00000000-0000-0000-0000-000000000000"
      arm_client_secret   = "REPLACE-ME"
      arm_subscription_id = "00000000-0000-0000-0000-000000000000"
      arm_tenant_id       = "00000000-0000-0000-0000-000000000000"
    }]
  }
}

# VCS connector, used as `auth` when a workflow clones a private repository.
resource "stackguardian_connector" "github" {
  resource_name = "github-connector"
  description   = "Access to private GitHub repositories"

  settings = {
    kind = "GITHUB_COM"
    config = [{
      github_com_url  = "https://github.com"
      github_http_url = "https://github.com"
    }]
  }
}

# Reference a connector from a workflow as `/integrations/<resource_name>`:
#
#   deployment_platform_config = [{
#     kind   = "AWS_RBAC"
#     config = { integration_id = stackguardian_connector.aws_rbac.id }
#   }]

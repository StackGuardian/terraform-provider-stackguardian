# The connector the runner authenticates to the log bucket with. Referencing it here means
# Terraform creates the connector first and keeps the two in step.
resource "stackguardian_connector" "runner_logs" {
  resource_name = "runner-log-storage"
  description   = "Grants the private runners write access to the log bucket"

  settings = {
    kind = "AWS_RBAC"
    config = [{
      role_arn         = "arn:aws:iam::000000000000:role/StackGuardianRunnerLogs"
      external_id      = "REPLACE-ME"
      duration_seconds = "3600"
    }]
  }
}

resource "stackguardian_runner_group" "example" {
  resource_name = "private-runners"
  description   = "Self-hosted runners for production workloads"

  tags = ["provider", "runnergroup"]

  # Where run logs are written. This is separate from the credentials a workflow
  # deploys with -- those come from the workflow's deployment_platform_config.
  storage_backend_config = {
    type           = "aws_s3"
    aws_region     = "eu-central-1"
    s3_bucket_name = "my-org-runner-logs"
    auth = {
      integration_id = stackguardian_connector.runner_logs.id
    }
  }
}

# Pin a workflow to this group with runner_constraints:
#
#   runner_constraints = {
#     type  = "private"
#     names = [stackguardian_runner_group.example.resource_name]
#   }

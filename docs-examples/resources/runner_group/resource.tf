resource "stackguardian_runner_group" "example-runner-group" {
  resource_name = "example-runner-group"

  description = "RunnerGroup created using provider"

  tags = [
    "provider",
    "runnergroup"
  ]

  storage_backend_config = {
    type           = "aws_s3"
    aws_region     = "eu-central-1"
    s3_bucket_name = "http-proxy-private-runner"
    auth = {
      integration_id = "/integrations/test-connector"
    }
  }
}

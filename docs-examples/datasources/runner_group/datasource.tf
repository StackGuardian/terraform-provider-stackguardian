data "stackguardian_runner_group" "example" {
  resource_name = "test-datasource-runner-group"
}

output "demo-runner-group-output" {
  value = data.stackguardian_runner_group.example.description
}

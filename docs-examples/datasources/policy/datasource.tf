data "stackguardian_policy" "example-policy" {
  resource_name = "example-policy"
}

output "policy-output" {
  value = data.stackguardian_policy.example-policy.description
}

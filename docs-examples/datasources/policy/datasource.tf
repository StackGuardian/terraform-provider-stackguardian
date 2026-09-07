# Read an existing policy to check how it is configured and where it is enforced.
data "stackguardian_policy" "example" {
  resource_name = "require-tags"
}

output "policy_enforced_on" {
  description = "Resource paths this policy applies to."
  value       = data.stackguardian_policy.example.enforced_on
}

output "policy_type" {
  value = data.stackguardian_policy.example.policy_type
}

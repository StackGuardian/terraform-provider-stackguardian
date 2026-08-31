# Inspect an existing role assignment -- useful for auditing which roles a user holds
# before changing them.
data "stackguardian_role_assignment" "example" {
  user_id = "user@example.com"
}

output "assigned_roles" {
  description = "Roles currently granted to this user."
  value       = data.stackguardian_role_assignment.example.roles
}

output "entity_type" {
  description = "Whether the assignment targets an email, an SSO user, or an SSO group."
  value       = data.stackguardian_role_assignment.example.entity_type
}

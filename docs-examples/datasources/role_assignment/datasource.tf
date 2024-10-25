data "stackguardian_role_assignment" "example" {
  user_id = "user-id"
}

output "role-assignment-output" {
  value = data.stackguardian_role_assignment.example.entity_type
}

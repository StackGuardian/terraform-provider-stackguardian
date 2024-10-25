data "stackguardian_role" "example" {
  resource_name = "role-name"
}

output "demo-role-output" {
  value = data.stackguardian_role.example.description
}

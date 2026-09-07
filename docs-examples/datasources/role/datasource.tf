# Look up a role defined elsewhere in the organization, then grant it to a user.
data "stackguardian_role" "developer" {
  resource_name = "developer"
}

resource "stackguardian_role_assignment" "example" {
  user_id    = "user@example.com"
  entity_type = "EMAIL"
  role       = data.stackguardian_role.developer.resource_name
}

# Look up a role defined elsewhere in the organization, then grant it to a user.
data "stackguardian_role" "developer" {
  id = "developer"
}

resource "stackguardian_role_assignment" "example" {
  user_id    = "user@example.com"
  entity_type = "EMAIL"
  role       = data.stackguardian_role.developer.id
}

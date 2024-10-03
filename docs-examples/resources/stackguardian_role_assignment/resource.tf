resource "stackguardian_role_assignment" "testing" {
  user_id     = "frontend.developer.p01@dummy.com"
  entity_type = "EMAIL"
  role        = "testing"
}

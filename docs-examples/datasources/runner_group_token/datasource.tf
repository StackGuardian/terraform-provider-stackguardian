# Fetch a registration token so a self-hosted runner can join the group.
#
# The token is a credential. Mark any output carrying it as sensitive, and prefer passing it
# straight to whatever provisions the runner rather than printing it.
data "stackguardian_runner_group_token" "example" {
  runner_group_id = "private-runners"
}

output "runner_registration_token" {
  value     = data.stackguardian_runner_group_token.example.runner_group_token
  sensitive = true
}

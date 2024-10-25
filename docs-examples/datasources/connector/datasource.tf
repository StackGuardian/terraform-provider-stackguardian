data "stackguardian_connector" "example" {
  resource_name = "testing"
}

output "connector-output" {
  value = data.stackguardian_connector.example.description
}


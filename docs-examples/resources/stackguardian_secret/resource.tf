resource "stackguardian_secret" "TPS-Example-Secret-Name" {
  data = jsonencode({
    "ResourceName" : "TPS-Example-Secret-Name",
    "ResourceValue" : "TPS-Example-Secret-Value"
  })
}

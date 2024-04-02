resource "stackguardian_policy" "TPS-Example-Policy" {
  data = jsonencode({
    "ResourceName" : "TPS-Example-Policy",
    "Description" : "Example of terraform-provider-stackguardian for Policy",
    "Tags" : ["tf-provider-example"]
  })
}

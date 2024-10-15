resource "stackguardian_connector" "aws-cloud-connector-example" {
  resource_name = "aws-cloud-connector-example"
  description   = "Example of terraform-provider-stackguardian for AWS Cloud Connector"

  settings = {
    kind   = "AWS_STATIC" # Type of connector. In this case, AWS_STATIC for static credentials.
    
    config = [{
      aws_access_key_id     = "YOUR_AWS_ACCESS_KEY_ID"       # Replace with your actual AWS Access Key ID
      aws_secret_access_key = "YOUR_AWS_SECRET_ACCESS_KEY"   # Replace with your actual AWS Secret Access Key
      aws_default_region    = "eu-central-1"                 # Specify the AWS region (e.g., eu-central-1)
    }]
  }
}

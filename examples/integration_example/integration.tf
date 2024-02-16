terraform {
  required_providers {
    stackguardian = {
      source = "terraform/provider/stackguardian"
      version = "0.0.0-dev"
    }
  }
}

provider "stackguardian" {
  org_name = "---" // TBD
  api_key  = "---" // TBD
}

resource "stackguardian_integration" "aws-static-integ" {
  data = jsonencode({
  "ResourceName": "aws-static-integ",
  "Description": "",
  "Settings": {
    "kind": "AWS_STATIC",
    "config": [
      {
        "awsAccessKeyId": "vdfvdfvdfvdfvdfvdfv",
        "awsSecretAccessKey": "vdvdfvdfvdvdfvdfvdfv",
        "awsDefaultRegion": "us-west-2"
      }
    ]
  }
})
}

resource "stackguardian_integration" "devops" {
  data = jsonencode({
  "ResourceName": "devops",
  "Settings": {
    "kind": "AZURE_DEVOPS",
    "config": [
      {
        "azureCreds": "dcdscdscdssdcsdc"
      }
    ]
  }
})
}


resource "stackguardian_integration" "gc-integaxcsdcs" {
  data = jsonencode({
  "ResourceName": "gc-integaxcsdcs",
  "Description": "csdcsdcsdc",
  "Settings": {
    "kind": "GCP_STATIC",
    "config": [
      {
        "gcpConfigFileContent": "{\"apple\":true}"
      }
    ]
  }
})
}

resource "stackguardian_integration" "cdcdcdc" {
  data = jsonencode({
  "ResourceName": "cdcdcdc",
  "Description": "",
  "Settings": {
    "kind": "AZURE_STATIC",
    "config": [
      {
        "armTenantId": "dcdcdcdcs",
        "armSubscriptionId": "vsvdfvdfvdfv",
        "armClientId": "vdvdfvdfvdfvfdv",
        "armClientSecret": "vdvfvdfvdfvdfvdfvdfv"
      }
    ]
  }
})
}

resource "stackguardian_integration" "gitlab-integxcsdc" {
  data = jsonencode({
  "ResourceName": "gitlab-integxcsdc",
  "Settings": {
    "kind": "GITLAB_COM",
    "config": [
      {
        "gitlabCreds": "csdcsdcd:csdcsdcsdcsdcd"
      }
    ]
  }
})
}

resource "stackguardian_integration" "rbac-integ" {
  data = jsonencode({
  "ResourceName": "rbac-integ",
  "Description": "",
  "Settings": {
    "kind": "AWS_RBAC",
    "config": [
      {
        "roleArn": "wsdcdscsdcsdcsdcsdcsd",
        "externalId": "demo-org:rEDcTFKAzEFqzpuImnzjqKtOEnILJZ",
        "durationSeconds": "3600"
      }
    ]
  }
})
}

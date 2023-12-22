# Quickstart Instructions for the StackGuardian Provider

Those quickstart instructions lets you setup a new IaC project with the Terraform Provider for StackGuardian.


## Provider Installation

_For now, the StackGuardian Provider is not available on the Terraform Registry,
so it is necessary to add it manually on your system to be able to use it in your IaC Terraform project._

A plarform label, with an OS name and an architecture name, matching the system platform where you will run the terraform provider on, must be selected from the start.
Please select one among the following options:
- `darwin_amd64`
- `darwin_arm64`
- `linux_amd64`
- `linux_arm64`
- `windows_amd64`
- `windows_arm64`

- After selecting one of the available options, set it in the shell. For instance:
```shell
$ export TFSG_OSARCH="linux_amd64"
```

- Go to the [latest release page](https://github.com/StackGuardian/terraform-provider-stackguardian/releases) from the Github repository.
Select a release, pickup its bare version tag without the `v` prefix, and set it in the shell. For instance:
```shell
$ export TFSG_VERSION="0.1.0-aplha1"
```

- Execute the following shell commands to install the provider:
```shell
# Prepare the plugin directory
$ rm -rfv $HOME/.terraform.d/plugins/terraform/provider/stackguardian/${TFSG_VERSION}/${TFSG_OSARCH}
$ mkdir -p $HOME/.terraform.d/plugins/terraform/provider/stackguardian/${TFSG_VERSION}/${TFSG_OSARCH}
$ cd $HOME/.terraform.d/plugins/terraform/provider/stackguardian/${TFSG_VERSION}/${TFSG_OSARCH}

# Fetch the plugin binary from Github
$ wget -q https://github.com/StackGuardian/terraform-provider-stackguardian/releases/download/v${TFSG_VERSION}/terraform-provider-stackguardian_${TFSG_VERSION}_${TFSG_OSARCH}.zip

# Install the plugin binary inside the plugin directory
$ unzip terraform-provider-stackguardian_${TFSG_VERSION}_${TFSG_OSARCH}.zip
$ mv terraform-provider-stackguardian_${TFSG_VERSION}_${TFSG_OSARCH} terraform-provider-stackguardian
```


## Provider Configuration inside project

- Create a new IaC project to setup before being able to define StackGuardian objects.
```shell
$ mkdir -p ~/devel/terraform-stackguardian-quickstart
$ cd ~/devel/terraform-stackguardian-quickstart
```

- Create a new file to declare the provider:
```json
// stackguardian.hcl

terraform {
  required_providers {
    stackguardian = {
      source = "terraform-provider-stackguardian"
    }
  }
}

provider "stackguardian" {}
```
The provider configuration will be passed from environment variables later.

- Check whether the provider was correctly installed with the following commands:
If it is correctly installed, the output will look similar, otherwise it will show an error.
```shell
$ terraform init
[...]
$ terraform providers

Providers required by configuration:
.
└── provider[terraform/provider/stackguardian]
```

Finally, the provider can be configurated from environment variables:
```
$ export STACKGUARDIAN_ORG_NAME="YOUR_SG_ORG"
$ export STACKGUARDIAN_API_KEY="YOUR_SG_KEY"
```

If you do not have any API key for your organization yet, you can generate one on the
StackGuardian App by going to "Organization settings > API Keys".


## Example: Worflow

Finally, you can take inspiration from the [provider examples](./../examples) to create new StackGuardian objects in your organization.

For instance you can create a new workflow

```json
// stackguardian.hcl

resource "stackguardian_tf_provider_workflow" "Workflow_DeployWebsiteS3" {
  wfgrp = "WorkflowGroup_DeployWebsiteS3"

  data = jsonencode({
    "ResourceName": "Workflow_DeployWebsiteS3",
    "Description": "Example of StackGuardian Workflow: Deploy a website from AWS S3",
    "Tags": ["tf-provider-example"],
    "EnvironmentVariables": [],
    "DeploymentPlatformConfig": [{
      "kind": "AWS_RBAC",
      "config": {
        "integrationId": "/integrations/aws"
      }
    }],
    "VCSConfig": {
      "iacVCSConfig": {
        "useMarketplaceTemplate": true,
        "iacTemplate": "/stackguardian/aws-s3-demo-website",
        "iacTemplateId": "/stackguardian/aws-s3-demo-website:11"
      },
      "iacInputData": {
        "schemaType": "FORM_JSONSCHEMA",
        "data": {
          "shop_name": "StackGuardian",
          "bucket_region": "eu-central-1"
        }
      }
    },
    "Approvers": [],
    "TerraformConfig": {
      "managedTerraformState": true,
      "terraformVersion": "1.4.6"
    },
    "WfType": "TERRAFORM",
    "UserSchedules": []
  })
}
```

For a complete example, please refer to the file [docs/stackguardian_workflow.tf](./stackguardian_workflow.tf)

Finally, inspect the plan offered by Terraform, and execute it to create the desired object on StackGuardian:
```shell
$ terraform plan
[...]
$ terraform apply
[...]
```


---

References:
- https://docs.stackguardian.io/docs/getting-started/setup
- https://developer.hashicorp.com/terraform/cli/config/config-file#provider_installation

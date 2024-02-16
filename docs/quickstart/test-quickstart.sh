#!/bin/bash

export TFSG_PROVIDER="terraform/provider/stackguardian"
export TFSG_OSARCH="linux_amd64"
export TFSG_VERSION="1.0.0"

SCRIPT_DIRPATH=$(cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd)

set -eu -o pipefail
set -x

## Provider Installation

# --- Prepare the plugin directory
rm -rfv $HOME/.terraform.d/plugins/${TFSG_PROVIDER}/${TFSG_VERSION}/${TFSG_OSARCH}
mkdir -p $HOME/.terraform.d/plugins/${TFSG_PROVIDER}/${TFSG_VERSION}/${TFSG_OSARCH}
cd $HOME/.terraform.d/plugins/${TFSG_PROVIDER}/${TFSG_VERSION}/${TFSG_OSARCH}

# --- Fetch the plugin binary from Github
wget https://github.com/StackGuardian/terraform-provider-stackguardian/releases/download/v${TFSG_VERSION}/terraform-provider-stackguardian_${TFSG_VERSION}_${TFSG_OSARCH}.zip
### For local testing ### cp ~/Downloads/terraform-provider-stackguardian_${TFSG_VERSION}_${TFSG_OSARCH}.zip .

# --- Install the plugin binary inside the plugin directory
unzip terraform-provider-stackguardian_${TFSG_VERSION}_${TFSG_OSARCH}.zip
rm -v terraform-provider-stackguardian_${TFSG_VERSION}_${TFSG_OSARCH}.zip


## Provider Configuration inside project
rm -rfv ~/tmp/terraform-stackguardian-quickstart
mkdir -p ~/tmp/terraform-stackguardian-quickstart
cp -v ${SCRIPT_DIRPATH}/stackguardian_workflow.tf -t ~/tmp/terraform-stackguardian-quickstart/
cd ~/tmp/terraform-stackguardian-quickstart

# --- The provider configuration should be passed from external environment variables:
# $ export STACKGUARDIAN_ORG_NAME="YOUR_SG_ORG"
# $ export STACKGUARDIAN_API_KEY="YOUR_SG_KEY"

terraform providers
terraform init
terraform version

terraform plan
terraform apply

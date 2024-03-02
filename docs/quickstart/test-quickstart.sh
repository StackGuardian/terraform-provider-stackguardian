#!/bin/bash

usage() {
    echo "Usage: $0 [-p <provider>] [-a <arch>] [-v <version>]" 1>&2;
    exit 1;
}

# TODO: handle provider built locally

while getopts ":p:a:v:" o
do
    case "${o}" in
        p)
            p=${OPTARG}
            ;;
        a)
            a=${OPTARG}
            ;;
        v)
            v=${OPTARG}
            ;;
        *)
            usage
            ;;
    esac
done
shift $((OPTIND-1))

# echo "Args: provider=$p arch=$a version=$v"

export TFSG_PROVIDER="${p:-terraform/provider/stackguardian}"
export TFSG_OSARCH="${a:-linux_amd64}"
export TFSG_VERSION="${v:-0.1.0-rc1}"

echo "Running Example with arguments:"
printenv | grep -E '^TFSG_.*'

SCRIPT_DIRPATH=$(cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd)

set -eu -o pipefail
set -x

## Provider Installation

# --- Prepare the plugin directory
rm -rfv $HOME/.terraform.d/plugins/${TFSG_PROVIDER}/${TFSG_VERSION}/${TFSG_OSARCH}
mkdir -p $HOME/.terraform.d/plugins/${TFSG_PROVIDER}/${TFSG_VERSION}/${TFSG_OSARCH}
cd $HOME/.terraform.d/plugins/${TFSG_PROVIDER}/${TFSG_VERSION}/${TFSG_OSARCH}

# --- Fetch the plugin binary from Github
wget -q https://github.com/StackGuardian/terraform-provider-stackguardian/releases/download/v${TFSG_VERSION}/terraform-provider-stackguardian_${TFSG_VERSION}_${TFSG_OSARCH}.zip
### For local testing ### cp ~/Downloads/terraform-provider-stackguardian_${TFSG_VERSION}_${TFSG_OSARCH}.zip .

# --- Install the plugin binary inside the plugin directory
unzip terraform-provider-stackguardian_${TFSG_VERSION}_${TFSG_OSARCH}.zip
rm -v terraform-provider-stackguardian_${TFSG_VERSION}_${TFSG_OSARCH}.zip


## Provider Configuration inside project
rm -rfv ~/tmp/terraform-stackguardian-quickstart
mkdir -p ~/tmp/terraform-stackguardian-quickstart
cp -v ${SCRIPT_DIRPATH}/stackguardian_workflow.tf -t ~/tmp/terraform-stackguardian-quickstart/
cd ~/tmp/terraform-stackguardian-quickstart

sed -E -i "s/version = \"[[:alnum:]\.\+\_\-]+\" #provider-version/version = \"${TFSG_VERSION}\" #provider-version/" stackguardian_workflow.tf

# --- The provider configuration should be passed from external environment variables:
# $ export STACKGUARDIAN_ORG_NAME="YOUR_SG_ORG"
# $ export STACKGUARDIAN_API_KEY="YOUR_SG_KEY"

terraform providers
terraform init
terraform version

terraform plan
terraform apply -auto-approve
sleep 10
terraform destroy -auto-approve

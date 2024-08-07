#!/bin/bash

# TODO: refactor with test-quickstart to have a single test entrypoint not dependent from
# the location as the directory of the script.

usage() {
    echo "Usage: $0 [-p <provider>] [-a <arch>] [-v <version>] [-f github-release-public|github-release-draft|local-build]" 1>&2;
    exit 1;
}

while getopts ":p:a:v:f:" o
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
        f)
            f=${OPTARG}
            ;;
        *)
            usage
            ;;
    esac
done
shift $((OPTIND-1))

if [[ ! -z "$f" ]]
then
    if [[ "$f" != "github-release-public" && "$f" != "github-release-draft" && "$f" != "local-build" ]]
    then
        usage
    fi
fi

# echo "Args: provider=$p arch=$a version=$v from-origin=$f"

TFSG_VERSION_DEFAULT_GITHUB_RELEASE="0.1.0-rc1"
TFSG_VERSION_DEFAULT_LOCAL_BUILD="0.0.0-dev"

export TFSG_PROVIDER="${p:-terraform/provider/stackguardian}"
export TFSG_OSARCH="${a:-linux_amd64}"
export TFSG_ORIGIN="${f:github-release-draft}"
if [[ "${TFSG_ORIGIN}" == "local-build" ]]
then
    export TFSG_VERSION="${v:-${TFSG_VERSION_DEFAULT_LOCAL_BUILD}}"
else
    export TFSG_VERSION="${v:-${TFSG_VERSION_DEFAULT_GITHUB_RELEASE}}"
fi

echo "Running Example with arguments:"
printenv | grep -E '^TFSG_.*'

SCRIPT_DIRPATH=$(cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd)

set -eu -o pipefail
set -x

## Provider Installation

# --- Installation inside project depending on origin
case "${TFSG_ORIGIN}" in

    github-release-*)
        rm -rfv $HOME/.terraform.d/plugins/${TFSG_PROVIDER}/${TFSG_VERSION}/${TFSG_OSARCH}
        mkdir -p $HOME/.terraform.d/plugins/${TFSG_PROVIDER}/${TFSG_VERSION}/${TFSG_OSARCH}
        cd $HOME/.terraform.d/plugins/${TFSG_PROVIDER}/${TFSG_VERSION}/${TFSG_OSARCH}

        # Fetch the plugin binary from Github depending on release status
        case "${TFSG_ORIGIN}" in
            github-release-public)
                wget https://github.com/StackGuardian/terraform-provider-stackguardian/releases/download/v${TFSG_VERSION}/terraform-provider-stackguardian_${TFSG_VERSION}_${TFSG_OSARCH}.zip
                ### If downloading releases manually: ### cp ~/Downloads/terraform-provider-stackguardian_${TFSG_VERSION}_${TFSG_OSARCH}.zip .
                ;;
            github-release-draft)
                pushd ${SCRIPT_DIRPATH}/../../
                    gh release download --pattern "terraform-provider-stackguardian_${TFSG_VERSION}_${TFSG_OSARCH}.zip" --dir $HOME/.terraform.d/plugins/${TFSG_PROVIDER}/${TFSG_VERSION}/${TFSG_OSARCH} v${TFSG_VERSION}
                popd
                ;;
        esac

        unzip terraform-provider-stackguardian_${TFSG_VERSION}_${TFSG_OSARCH}.zip
        rm -v terraform-provider-stackguardian_${TFSG_VERSION}_${TFSG_OSARCH}.zip
        ls -l terraform-provider-stackguardian_v${TFSG_VERSION}
        ;;

    local-build)
        cd $HOME/.terraform.d/plugins/${TFSG_PROVIDER}/${TFSG_VERSION}/${TFSG_OSARCH}
        ls -l terraform-provider-stackguardian
        ;;

    *)
        usage
        ;;

esac

# --- Bootstrap & Configuration
rm -rfv ~/tmp/terraform-stackguardian-iac-project
mkdir -p ~/tmp/terraform-stackguardian-iac-project
cp -v ${SCRIPT_DIRPATH}/project.tf -t ~/tmp/terraform-stackguardian-iac-project/
cd ~/tmp/terraform-stackguardian-iac-project

# Set the version of the provider inside the terraform config exactly to the version of the downloaded provider.
sed -E -i "s/version = \"[[:alnum:]\.\+\_\-]+\" #provider-version/version = \"${TFSG_VERSION}\" #provider-version/" project.tf

# Randomize the workflow resource id in the config file in order to isolate the test
tf_test_id="T$(date +%s)-R$(printf '%05d' $RANDOM)"
sed -E -i "s/T000000/${tf_test_id}/" project.tf


## Provider Execution Test

# --- The provider configuration should be passed from external environment variables:
# $ export STACKGUARDIAN_ORG_NAME="YOUR_SG_ORG"
# $ export STACKGUARDIAN_API_KEY="YOUR_SG_KEY"

terraform providers
terraform init
terraform version

terraform validate

terraform plan

terraform state list || true

terraform apply -auto-approve
terraform state list

sleep 10

terraform destroy -auto-approve
terraform state list

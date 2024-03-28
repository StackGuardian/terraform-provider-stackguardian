#!/bin/bash

set -e -u -o pipefail -x


# --- Clean all Terraform state files
find . -type f -regextype posix-extended -regex '.+.tfstate(.[[:digit:]]+)?(.backup)?' -exec rm -v {} \+


# --- Import resources
set +e

terraform import stackguardian_workflow_group.ONBOARDING-Project01-Frontend /api/v1/orgs/wicked-hop/wfgrps/ONBOARDING-Project01-Frontend/
terraform import stackguardian_workflow_group.ONBOARDING-Project01-Backend /api/v1/orgs/wicked-hop/wfgrps/ONBOARDING-Project01-Backend/
terraform import stackguardian_workflow_group.ONBOARDING-Project01-DevOps /api/v1/orgs/wicked-hop/wfgrps/ONBOARDING-Project01-DevOps/

terraform import stackguardian_policy.ONBOARDING-Project01 /api/v1/orgs/wicked-hop/policies/ONBOARDING-Project01/
terraform import stackguardian_connector_vcs.ONBOARDING-Project01 /api/v1/orgs/wicked-hop/integrations/ONBOARDING-Project01/

terraform import stackguardian_role.ONBOARDING-Project01-Developer /api/v1/orgs/wicked-hop/roles/ONBOARDING-Project01-Develop/

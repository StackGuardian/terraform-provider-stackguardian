#!/bin/bash

set -e -u -o pipefail -x

res_suffix='' # '-T000000'


# --- Clean all Terraform state files
find . -type f -regextype posix-extended -regex '.+.tfstate(.[[:digit:]]+)?(.backup)?' -exec rm -v {} \+


# --- Import resources
set +e

terraform import stackguardian_workflow_group.TPS-OBT-Frontend${res_suffix} /api/v1/orgs/tps-test-01/wfgrps/TPS-OBT-Frontend${res_suffix}/
terraform import stackguardian_workflow_group.TPS-OBT-Backend${res_suffix} /api/v1/orgs/tps-test-01/wfgrps/TPS-OBT-Backend${res_suffix}/
terraform import stackguardian_workflow_group.TPS-OBT-DevOps${res_suffix} /api/v1/orgs/tps-test-01/wfgrps/TPS-OBT-DevOps${res_suffix}/

terraform import stackguardian_policy.TPS-OBT${res_suffix} /api/v1/orgs/tps-test-01/policies/TPS-OBT${res_suffix}/
terraform import stackguardian_connector_vcs.TPS-OBT${res_suffix} /api/v1/orgs/tps-test-01/integrations/TPS-OBT${res_suffix}/

terraform import stackguardian_role.TPS-OBT-Dv${res_suffix} /api/v1/orgs/tps-test-01/roles/TPS-OBT-Dv${res_suffix}/

terraform import stackguardian_workflow.TPS-OBT-DevOps${res_suffix} /api/v1/orgs/tps-test-01/wfgrps/TPS-OBT-DevOps${res_suffix}/

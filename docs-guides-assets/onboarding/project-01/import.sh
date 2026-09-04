#!/bin/bash
#
# Bring the resources in project-01.tf under Terraform management.
#
# Every StackGuardian resource in this project imports by its bare name -- the
# value of `resource_name` -- not by an API path. The one exception is
# stackguardian_role_assignment, which imports by the user's ID.
#
# Run from this directory, after setting STACKGUARDIAN_API_KEY and
# STACKGUARDIAN_ORG_NAME.

set -e -u -o pipefail -x

# The user whose role assignment is imported below.
: "${SG_USER_ID:?set SG_USER_ID to the user id used in project-01.tf}"

# --- Clean all Terraform state files
find . -type f -regextype posix-extended -regex '.+.tfstate(.[[:digit:]]+)?(.backup)?' -exec rm -v {} \+

# --- Import resources
set +e

terraform import stackguardian_workflow_group.frontend ONBOARDING-Project01-Frontend
terraform import stackguardian_workflow_group.backend  ONBOARDING-Project01-Backend
terraform import stackguardian_workflow_group.devops   ONBOARDING-Project01-DevOps

terraform import stackguardian_connector.cloud ONBOARDING-Project01-Cloud-Connector
terraform import stackguardian_connector.vcs   ONBOARDING-Project01-VCS-Connector

terraform import stackguardian_rolev4.developer ONBOARDING-Project01-Developer

terraform import stackguardian_role_assignment.developer "${SG_USER_ID}"

# A workflow imports as "<workflow group>/<workflow>".
terraform import stackguardian_workflow_git.frontend_deploy \
  ONBOARDING-Project01-Frontend/ONBOARDING-Project01-Frontend-Deploy

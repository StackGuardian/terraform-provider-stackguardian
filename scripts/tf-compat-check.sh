#!/usr/bin/env bash
#
# Terraform CLI compatibility gate.
#
# Builds the provider and makes a Terraform CLI load it and decode its full
# schema, then validates every example configuration under docs-examples/.
#
# Neither `terraform providers schema` nor `terraform validate` calls the
# provider's Configure method, so this needs no API credentials and is safe to
# run on pull requests from forks. What it catches is the realistic way an older
# CLI breaks a Plugin Framework provider: a schema construct the old CLI cannot
# decode over the plugin protocol, or example HCL that needs newer language
# features.
#
# Usage:
#   scripts/tf-compat-check.sh                # uses `terraform` from PATH
#   TERRAFORM_BIN=/path/to/terraform scripts/tf-compat-check.sh
#
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TERRAFORM_BIN="${TERRAFORM_BIN:-terraform}"
PROVIDER_SOURCE="StackGuardian/stackguardian"

if ! command -v "${TERRAFORM_BIN}" >/dev/null 2>&1; then
	echo "error: terraform binary not found (TERRAFORM_BIN=${TERRAFORM_BIN})" >&2
	exit 1
fi

TF_VERSION="$("${TERRAFORM_BIN}" version -json | sed -n 's/.*"terraform_version": *"\([^"]*\)".*/\1/p' | head -1)"
echo "==> Terraform ${TF_VERSION} (${TERRAFORM_BIN})"

WORK_DIR="$(mktemp -d)"
trap 'rm -rf "${WORK_DIR}"' EXIT

echo "==> Building provider"
go build -o "${WORK_DIR}/terraform-provider-stackguardian" "${REPO_ROOT}"

# dev_overrides makes the CLI load the local binary directly, with no registry
# round trip and no `terraform init` — which is also why the generated configs
# below pin no provider version.
cat >"${WORK_DIR}/dev.tfrc" <<EOF
provider_installation {
  dev_overrides {
    "${PROVIDER_SOURCE}" = "${WORK_DIR}"
  }
  direct {}
}
EOF
export TF_CLI_CONFIG_FILE="${WORK_DIR}/dev.tfrc"
export TF_IN_AUTOMATION=1

# The examples under docs-examples/ are documentation snippets: they carry no
# terraform{} block of their own, so each one is copied next to a generated
# header that points at the local build.
write_header() {
	cat >"$1/zz_generated_provider.tf" <<EOF
terraform {
  required_providers {
    stackguardian = {
      source = "${PROVIDER_SOURCE}"
    }
  }
}
EOF
}

echo "==> Decoding provider schema"
SCHEMA_DIR="${WORK_DIR}/schema"
mkdir -p "${SCHEMA_DIR}"
write_header "${SCHEMA_DIR}"
if ! (cd "${SCHEMA_DIR}" && "${TERRAFORM_BIN}" providers schema -json >"${WORK_DIR}/schema.json" 2>"${WORK_DIR}/schema.err"); then
	echo "FAIL: provider schema does not decode under Terraform ${TF_VERSION}" >&2
	cat "${WORK_DIR}/schema.err" >&2
	exit 1
fi
echo "    ok ($(wc -c <"${WORK_DIR}/schema.json" | tr -d ' ') bytes)"

echo "==> Validating examples"
passed=0
failed=0
failures=()

while IFS= read -r example_dir; do
	name="${example_dir#"${REPO_ROOT}"/}"
	case_dir="${WORK_DIR}/cases/$(echo "${name}" | tr '/' '_')"
	mkdir -p "${case_dir}"
	cp "${example_dir}"/*.tf "${case_dir}/"
	write_header "${case_dir}"

	if (cd "${case_dir}" && "${TERRAFORM_BIN}" validate -no-color >"${case_dir}/out.txt" 2>&1); then
		passed=$((passed + 1))
	else
		failed=$((failed + 1))
		failures+=("${name}")
		echo "    FAIL ${name}"
		sed 's/^/        /' "${case_dir}/out.txt"
	fi
done < <(find "${REPO_ROOT}/docs-examples" -name '*.tf' -exec dirname {} \; | sort -u)

echo "==> ${passed} passed, ${failed} failed on Terraform ${TF_VERSION}"
if ((failed > 0)); then
	printf 'failed: %s\n' "${failures[@]}" >&2
	exit 1
fi

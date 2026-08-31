#!/usr/bin/env bash
#
# Validate every documentation example against the provider schema.
#
# Each file under docs-examples/ is a *fragment*: it declares only the resource or
# data source, with no terraform{} or provider{} block, because that is what
# tfplugindocs embeds into a docs page. To validate one we copy it into a scratch
# directory and inject the missing blocks. Files under docs-guides-assets/ are
# standalone projects that carry their own blocks, so they are validated as-is.
#
# This runs `validate` only: no API calls, no credentials, no state. It catches
# unknown resource types, unknown attributes, bad references and type errors --
# every example defect found in the August 2026 documentation sweep.
#
# The provider is served from a local filesystem mirror, so the examples can
# declare the real registry source (StackGuardian/stackguardian) exactly as a
# customer would copy them.
#
# TF_CLI may be set to `tofu`; the toolchain otherwise defaults to `terraform`
# because tfplugindocs itself shells out to the Terraform CLI.

set -euo pipefail

REPO_ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
TF_CLI="${TF_CLI:-terraform}"
VERSION="0.0.0-dev"
NAMESPACE="StackGuardian"
TYPE="stackguardian"

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

OS_ARCH="$(go env GOOS)_$(go env GOARCH)"
MIRROR="$WORK/mirror"
PLUGIN_DIR="$MIRROR/registry.terraform.io/$(echo "$NAMESPACE" | tr '[:upper:]' '[:lower:]')/$TYPE/$VERSION/$OS_ARCH"

echo "==> building provider"
mkdir -p "$PLUGIN_DIR"
( cd "$REPO_ROOT" && go build -o "$PLUGIN_DIR/terraform-provider-${TYPE}_v${VERSION}" )

# A filesystem mirror lets `source = "StackGuardian/stackguardian"` resolve locally.
cat > "$WORK/cli.tfrc" <<EOF
provider_installation {
  filesystem_mirror {
    path    = "$MIRROR"
    include = ["registry.terraform.io/*/*"]
  }
  direct {
    exclude = ["registry.terraform.io/*/*"]
  }
}
EOF
export TF_CLI_CONFIG_FILE="$WORK/cli.tfrc"
export TF_IN_AUTOMATION=1

pass=0
fail=0
failed_files=()

validate_dir() {
  local dir="$1" label="$2"
  if ! out="$("$TF_CLI" -chdir="$dir" init -backend=false -input=false -no-color 2>&1)"; then
    echo "FAIL (init) $label"
    echo "$out" | sed 's/^/      /' | tail -20
    fail=$((fail + 1)); failed_files+=("$label"); return
  fi
  if ! out="$("$TF_CLI" -chdir="$dir" validate -no-color 2>&1)"; then
    echo "FAIL $label"
    echo "$out" | sed 's/^/      /' | tail -30
    fail=$((fail + 1)); failed_files+=("$label"); return
  fi
  echo "  ok  $label"
  pass=$((pass + 1))
}

echo "==> validating docs-examples fragments"
n=0
while IFS= read -r f; do
  n=$((n + 1))
  rel="${f#"$REPO_ROOT"/}"
  dir="$WORK/frag-$n"
  mkdir -p "$dir"
  cp "$f" "$dir/main.tf"
  # Fragments never declare these; inject them so the file can be type-checked.
  cat > "$dir/provider.tf" <<EOF
terraform {
  required_providers {
    stackguardian = {
      source  = "$NAMESPACE/$TYPE"
      version = "$VERSION"
    }
  }
}

provider "stackguardian" {
  api_key  = "validate-only"
  org_name = "validate-only"
  api_uri  = "https://api.app.stackguardian.io"
}
EOF
  validate_dir "$dir" "$rel"
done < <(find "$REPO_ROOT/docs-examples" -name '*.tf' | sort)

echo "==> validating standalone guide projects"
while IFS= read -r d; do
  rel="${d#"$REPO_ROOT"/}"
  dir="$WORK/proj-$(echo "$rel" | tr '/' '-')"
  mkdir -p "$dir"
  cp "$d"/*.tf "$dir/"

  # A customer copies these verbatim, so the source address must be the real
  # registry one. Check it before validating -- otherwise the version rewrite
  # below would hide a wrong source.
  # -i because wrong sources have shown up with every capitalisation
  # (stackguardian/StackGuardian, terraform.local/local/StackGuardian, ...).
  bad_source="$(grep -hoEi 'source[[:space:]]*=[[:space:]]*"[^"]*stackguardian[^"]*"' "$dir"/*.tf \
                | grep -vF "\"$NAMESPACE/$TYPE\"" || true)"
  if [ -n "$bad_source" ]; then
    echo "FAIL $rel"
    echo "      provider source must be \"$NAMESPACE/$TYPE\", found:"
    echo "$bad_source" | sed 's/^/        /'
    fail=$((fail + 1)); failed_files+=("$rel"); continue
  fi

  # Pin to the locally built provider so validation needs no network. The real
  # constraint the file ships with is checked by review, not by this script.
  perl -0pi -e "s/(source\s*=\s*\"$NAMESPACE\/$TYPE\"\s*\n\s*version\s*=\s*)\"[^\"]*\"/\${1}\"$VERSION\"/g" "$dir"/*.tf

  validate_dir "$dir" "$rel"
done < <(find "$REPO_ROOT/docs-guides-assets" -name '*.tf' -exec dirname {} \; \
         | grep -v '/project-test$' | sort -u)

echo
echo "==> $pass passed, $fail failed"
if [ "$fail" -gt 0 ]; then
  printf '    %s\n' "${failed_files[@]}"
  exit 1
fi

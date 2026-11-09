TEST?=$$(go list ./... | grep -v 'vendor')
TEST_ACC=$$(grep -rl 'func TestAcc' --include='*_test.go' . | xargs dirname | sort -u)
HOSTNAME=terraform
NAMESPACE=provider
NAME=stackguardian
BINARY=terraform-provider-${NAME}
VERSION=0.0.0-dev
OS_ARCH=linux_amd64

# Optional local credentials (gitignored). Only the API variables are exported:
# TF_ACC is left out so `make test` stays a unit-test run; test-acc and
# test-crud set it themselves.
-include .env
export STACKGUARDIAN_API_KEY STACKGUARDIAN_API_URI STACKGUARDIAN_ORG_NAME TEST_AZURE_STORAGE_BACKEND_ACCESS_KEY

default: install

clean: clean-examples

clean-examples:
	find examples/ -type d -name '.terraform' -exec rm -rv {} \+
	find examples/ -type f -name '.terraform.lock.hcl' -exec rm -v {} \+
	find examples/ -type f -regextype posix-extended -regex '.+.tfstate(.[[:digit:]]+)?(.backup)?' -exec rm -v {} \+

build:
	go build -o ${BINARY}

release:
	goreleaser release --rm-dist --snapshot --skip-publish  --skip-sign

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

clean-test-cache:
	go clean -testcache

test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

# -p=8 runs up to 8 packages concurrently so the two heaviest packages
# (workflow_template_revision ~327s and workflow_from_template ~247s)
# overlap instead of landing in separate batches. Safe now that every test
# uses acctest.ResourceName for unique identifiers.
# -parallel=1 keeps t.Parallel() tests sequential within each package (none
# of the TestAcc functions use t.Parallel() today, but the flag is a guard).
test-acc:
	TF_ACC=1 go test -p=8 -parallel=1 $(TEST_ACC) -v $(TESTARGS) -timeout=15m

# Create -> update -> import -> (delete outside Terraform) -> destroy for every
# resource, reported as a resource x operation matrix. Same credentials as
# test-acc. Results land in $(CRUD_OUT)/ (matrix.md, matrix.json, raw go-test.json).
# Narrow with TESTARGS, e.g. TESTARGS='-run TestAccPolicy_CRUD'.
#
# The gate fails if either go test or the matrix tool fails. The matrix tool
# runs even after go test fails, so a failed run still gets its artifacts --
# but its exit status alone must not decide the gate: it only sees tests that
# reported a result, so a build failure or a timed-out run would otherwise
# look green.
CRUD_OUT ?= .crud-results
test-crud:
	@mkdir -p $(CRUD_OUT)
	status=0; \
	TF_ACC=1 go test -p=8 -parallel=1 ./internal/resource/... -run '_CRUD$$' -json -count=1 -timeout=20m $(TESTARGS) > $(CRUD_OUT)/go-test.json || status=$$?; \
	go run ./scripts/crud-matrix $(CRUD_OUT)/go-test.json --markdown $(CRUD_OUT)/matrix.md --json $(CRUD_OUT)/matrix.json || status=$$?; \
	exit $$status

# Terraform CLI compatibility gate. Needs no API credentials: it only makes the
# CLI load the provider and decode its schema, then validates docs-examples/.
# Override the CLI under test with TERRAFORM_BIN=/path/to/terraform.
tf-compat-check:
	bash scripts/tf-compat-check.sh

# Scans the provider's dependencies for known vulnerabilities reachable from
# its code. The version pin keeps local and CI runs identical.
govulncheck:
	go install golang.org/x/vuln/cmd/govulncheck@v1.9.0
	govulncheck ./...

test-acc-local:
	set -a; . ./.env; set +a; TF_ACC=1 go test -parallel=1 $(TEST) -v $(TESTARGS) -timeout=15m

# Reports resources an earlier acceptance run left behind. Read-only: it lists
# what carries the test prefix and deletes nothing.
test-acc-sweep:
	SG_ACC_SWEEP=1 go test ./internal/acctest -run TestSweepOrphans -v -count=1

# Deletes them. Only touches names carrying the acctest prefix.
test-acc-sweep-apply:
	SG_ACC_SWEEP=1 SG_ACC_SWEEP_APPLY=1 go test ./internal/acctest -run TestSweepOrphans -v -count=1

test-examples-quickstart:
	bash docs-guides-assets/quickstart/test-quickstart.sh $(ARGS)

test-examples-onboarding:
	bash docs-guides-assets/onboarding/project-test/test-onboarding.sh $(ARGS)

# Writes the "Building this with AI" section into every page template: which
# skill to load for that page, and the mistakes its resource invites. The
# wording lives in scripts/sync-ai-skills-sections rather than in 47 pages, so
# a correction lands everywhere at once; the script replaces whatever sits
# between the AI-SKILLS markers, so this is idempotent. Never edit the
# generated section in a template by hand.
docs-sync-ai-skills:
	go run ./scripts/sync-ai-skills-sections

# The banner step is temporary, for as long as the docs revamp is in progress.
# Delete the line below and re-run to remove the banner from every page.
docs-generate: docs-sync-ai-skills
	tfplugindocs generate \
		--provider-name stackguardian --website-source-dir docs-templates
	bash scripts/inject-docs-banner.sh

docs-validate:
	tfplugindocs validate --provider-name stackguardian

# Fails if docs/ is not what docs-templates/ + the current schemas generate.
# Without this, a schema change silently leaves the published docs stale.
docs-check: docs-generate
	@git diff --exit-code -- docs/ \
		|| { echo ""; echo "ERROR: docs/ is out of date. Run 'make docs-generate' and commit the result."; exit 1; }

# Type-checks every documentation example against the provider schema: the files
# under docs-examples/ and docs-guides-assets/, plus every ```terraform block
# embedded in page and guide prose (see scripts/extract-doc-blocks).
# No credentials and no API calls -- `terraform validate` only.
docs-validate-examples:
	bash scripts/validate-examples.sh

tools-install:
	cd tools; go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

gh-workflow-test-provider:
	act \
		--workflows ${PWD}/.github/workflows/test.yaml \
		--job provider-project_test \
		--secret SG_PRD_API_URI=${SG_PRD_API_URI} \
		--secret SG_PRD_API_KEY=${SG_PRD_API_KEY} \
		--secret SG_PRD_ORG_NAME=${SG_PRD_ORG_NAME} \
		push \
		;

gh-workflow-test-provider-mock-stg-as-prd:
	act \
		--workflows ${PWD}/.github/workflows/test.yaml \
		--job provider-project_test \
		--secret SG_PRD_API_URI=${SG_STG_API_URI} \
		--secret SG_PRD_API_KEY=${SG_STG_API_KEY} \
		--secret SG_PRD_ORG_NAME=${SG_STG_ORG_NAME} \
		push \
		;

#		--local-repository StackGuardian/terraform-provider-stackguardian@devel=${PWD} \#
gh-workflow-test-api-stg:
	act \
		--workflows ${PWD}/.github/workflows/test-api-stg.yaml \
		--secret SG_STG_API_URI=${SG_STG_API_URI} \
		--secret SG_STG_API_KEY=${SG_STG_API_KEY} \
		--secret SG_STG_ORG_NAME=${SG_STG_ORG_NAME} \
		workflow_dispatch \
		;

#		--local-repository StackGuardian/terraform-provider-stackguardian@devel=${PWD} \#
gh-workflow-test-api-prd:
	act \
		--workflows ${PWD}/.github/workflows/test-api-prd.yaml \
		--secret SG_PRD_API_URI=${SG_PRD_API_URI} \
		--secret SG_PRD_API_KEY=${SG_PRD_API_KEY} \
		--secret SG_PRD_ORG_NAME=${SG_PRD_ORG_NAME} \
		workflow_dispatch \
		;

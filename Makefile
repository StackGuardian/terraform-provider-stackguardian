TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=terraform
NAMESPACE=provider
NAME=stackguardian
BINARY=terraform-provider-${NAME}
VERSION=0.0.0-dev
OS_ARCH=linux_amd64

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

test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

test-acc:
	TF_ACC=1 go test -parallel=1 $(TEST) -v $(TESTARGS) -timeout=15m

test-examples-quickstart:
	bash docs-guides-assets/quickstart/test-quickstart.sh $(ARGS)

test-examples-onboarding:
	bash docs-guides-assets/onboarding/project-test/test-onboarding.sh $(ARGS)

docs-generate:
	tfplugindocs generate \
		--website-source-dir docs-templates

docs-validate:
	tfplugindocs validate

tools-install:
	cd tools; go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

gh-workflow-test-provider:
	act \
		--workflows ${PWD}/.github/workflows/test.yaml \
		--job provider-project_test \
		--secret STACKGUARDIAN_API_KEY=${SG_PRD_API_KEY} \
		--secret STACKGUARDIAN_ORG_NAME=${SG_PRD_ORG_NAME} \
		--secret SG_PRD_API_KEY=${SG_PRD_API_KEY} \
		--secret SG_PRD_ORG_NAME=${SG_PRD_ORG_NAME} \
		--secret SG_STG_API_URI=${SG_STG_API_URI} \
		--secret SG_STG_API_KEY=${SG_STG_API_KEY} \
		--secret SG_STG_ORG_NAME=${SG_STG_ORG_NAME} \
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

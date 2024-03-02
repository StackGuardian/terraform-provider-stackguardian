TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=terraform
NAMESPACE=provider
NAME=stackguardian
BINARY=terraform-provider-${NAME}
VERSION=0.0.0-dev
OS_ARCH=linux_amd64

default: install

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
	TF_ACC=1 STACKGUARDIAN_ORG_NAME=wicked-hop go test -parallel=1 $(TEST) -v $(TESTARGS) -timeout=15m

test-example:
	bash docs/quickstart/test-quickstart.sh $(ARGS)


docstf:
	tfplugindocs generate --rendered-website-dir docs/reference

gh-workflow:
	act --job run-tests \
		--secret STACKGUARDIAN_ORG_NAME=${STACKGUARDIAN_ORG_NAME} \
		--secret STACKGUARDIAN_API_KEY=${STACKGUARDIAN_API_KEY} \
		;

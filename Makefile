NAME=terraform-provider-pbs
VERSION=0.1.0

default: build

.PHONY: build
build:
	go build -o ${NAME} .

.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/micah/pbs/${VERSION}/darwin_amd64
	cp ${NAME} ~/.terraform.d/plugins/registry.terraform.io/micah/pbs/${VERSION}/darwin_amd64/

.PHONY: test
test:
	go test ./...

.PHONY: test-unit
test-unit:
	go test ./...

.PHONY: testacc
testacc:
	TF_ACC=1 go test ./fwprovider/... -v

.PHONY: lint
lint:
	golangci-lint run

.PHONY: fmt
fmt:
	gofmt -s -w .
	go mod tidy

.PHONY: clean
clean:
	rm -f ${NAME}

.PHONY: docs
docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build    - Build the provider binary"
	@echo "  install  - Build and install the provider locally"
	@echo "  test     - Run unit tests"
	@echo "  testacc  - Run acceptance tests"
	@echo "  lint     - Run linter"
	@echo "  fmt      - Format code"
	@echo "  clean    - Remove built binaries"
	@echo "  docs     - Generate documentation"
	@echo "  help     - Show this help message"
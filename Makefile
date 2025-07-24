export GO111MODULE=on
export CGO_ENABLED=1
export GOPRIVATE=github.com/eroshiva

POC_VERSION := $(shell git rev-parse --abbrev-ref HEAD)
DOCKER_REPOSITORY := eroshiva
GOLANGCI_LINTERS_VERSION := v2.2.2
BUF_VERSION := v1.55.1
GRPC_GATEWAY_VERSION := v2.27.1

.PHONY: help build
help: # Credits to https://gist.github.com/prwhite/8168133 for this handy oneliner
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

buf-install: ## Installs buf to convert protobuf into Golang code
	go install github.com/bufbuild/buf/cmd/buf@${BUF_VERSION}

buf-generate: buf-install buf-update ## Generates Golang-driven bindings out of Protobuf
	buf generate --path api/v1/monitoring.proto

buf-update: ## Generates Golang-driven bindings out of Protobuf
	buf dep update

buf-lint: ## Runs linters against Protobuf
	buf lint

buf-breaking: ## Checks Protobuf schema on breaking changes
	buf breaking --against '.git#branch=main'

build: go-tidy build-api ## Builds all code

build-api: ## Build the Go binary for gRPC API Gateway
	go build -mod=vendor -o build/_output/api-gateway ./cmd/api-gateway.go

deps: buf-install go-linters-install ## Installs developer prerequisites for this project
	go get github.com/grpc-ecosystem/grpc-gateway/v2@${GRPC_GATEWAY_VERSION}

go-linters-install: ## Install linters locally for verification
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin ${GOLANGCI_LINTERS_VERSION}

go-linters: go-linters-install ## Perform linting to verify codebase
	golangci-lint run --timeout 5m

govulncheck-install: ## Installs latest govulncheck tool
	go install golang.org/x/vuln/cmd/govulncheck@latest

govulncheck: govulncheck-install ## Runs govulncheck on the current codebase
	govulncheck ./...

go-vet: ## Searching for suspicious constructs in Go code
	go vet ./...

go-test: ## Run unit tests present in the codebase
	mkdir -p tmp
	go test -coverprofile=./tmp/test-cover.out -race ./...

test-ci: buf-generate buf-lint buf-breaking build go-vet govulncheck go-linters go-test ## Test the whole codebase (mimics CI/CD)

run: build-api ## Runs compiled API Gateway instance
	./build/_output/api-gateway

image: ## Builds a Docker image for API Gateway
	docker build . -f build/Dockerfile \
		-t ${DOCKER_REPOSITORY}/api-gateway:${POC_VERSION}

docker-run: image ## Runs compiled binary in a Docker container
	docker run --net=host --rm ${DOCKER_REPOSITORY}/api-gateway:${POC_VERSION}

kind: image ## Builds Docker image for API Gateway and loads it to the currently configured kind cluster
	@if [ "`kind get clusters`" = '' ]; then echo "no kind cluster found" && exit 1; fi
	kind load docker-image ${DOCKER_REPOSITORY}/api-gateway:${POC_VERSION}

go-tidy: ## Runs go mod related commands
	go mod tidy
	go mod vendor

clean: ## Remove all the build artifacts
	rm -rf ./build/_output ./vendor ./tmp
	go clean -testcache

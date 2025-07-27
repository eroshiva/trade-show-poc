export GO111MODULE=on
export CGO_ENABLED=1
export GOPRIVATE=github.com/eroshiva

POC_NAME := monitoring
POC_VERSION := $(shell git rev-parse --abbrev-ref HEAD)
DOCKER_REPOSITORY := eroshiva
GOLANGCI_LINTERS_VERSION := v2.3.0
GOFUMPT_VERSION := v0.8.0
BUF_VERSION := v1.55.1
GRPC_GATEWAY_VERSION := v2.27.1
PROTOC_GEN_ENT_VERSION := v0.6.0
DOCKER_POSTGRESQL_NAME := monitoring-postgresql
DOCKER_POSTGRESQL_VERSION := 15

# Postgres DB configuration and credentials for testing. This mimics the Aurora
# production environment.
export PGUSER=admin
export PGHOST=localhost
export PGDATABASE=postgres
export PGPORT=5432
export PGPASSWORD=pass
export PGSSLMODE=disable

.PHONY: help
help: # Credits to https://gist.github.com/prwhite/8168133 for this handy oneliner
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

atlas-install: ## Installs Atlas tool for generating migrations
	curl -sSf https://atlasgo.sh | sh

buf-install: ## Installs buf to convert protobuf into Golang code
	go install github.com/bufbuild/buf/cmd/buf@${BUF_VERSION}

buf-generate: clean-vendor buf-install buf-update ## Generates Golang-driven bindings out of Protobuf
	mkdir -p internal/ent/schema
	buf generate --exclude-path api/v1/ent --path api/v1/monitoring.proto --path pkg/mocks/simulator.proto

buf-update: ## Updates the buf dependencies
	buf dep update

buf-lint: ## Runs linters against Protobuf
	buf lint --path api/v1/monitoring.proto

buf-breaking: ## Checks Protobuf schema on breaking changes
	buf breaking --against '.git#branch=main'

generate: buf-generate ## Generates all necessary code bindings
	go generate ./internal/ent

build: go-tidy build-monitoring ## Builds all code

build-monitoring: ## Build the Go binary for network device monitoring service
	go build -mod=vendor -o build/_output/${POC_NAME} ./cmd/monitoring.go

deps: buf-install go-linters-install atlas-install ## Installs developer prerequisites for this project
	go get github.com/grpc-ecosystem/grpc-gateway/v2@${GRPC_GATEWAY_VERSION}
	go install entgo.io/contrib/entproto/cmd/protoc-gen-ent@${PROTOC_GEN_ENT_VERSION}
	go install mvdan.cc/gofumpt@${GOFUMPT_VERSION}

atlas-inspect: ## Inspect connection with DB with atlas
	atlas schema inspect --url "postgresql://${PGUSER}:${PGPASSWORD}@localhost:${PGPORT}/${PGDATABASE}?search_path=public" --format "OK"

migration-apply: ## Uploads migration to the running DB instance
	$(MAKE) db-start
	sleep 5;
	atlas migrate apply --dir file://internal/ent/migrate/migrations \
      --url postgresql://${PGUSER}:${PGPASSWORD}@localhost:${PGPORT}/${PGDATABASE}?search_path=public

migration-hash: ## Hashes the atlas checksum to correspond to the migration
	atlas migrate hash --dir file://internal/ent/migrate/migrations

migration-generate: ## Generate DB migration "make migration-generate MIGRATION=<migration-name>"
	@if test -z $(MIGRATION); then echo "Please specify migration name" && exit 1; fi
	$(MAKE) db-start
	sleep 5;
	atlas migrate diff $(MIGRATION) \
  		--dir "file://internal/ent/migrate/migrations" \
  		--to "ent://internal/ent/schema" \
  		--dev-url "docker://postgres/15/${PGDATABASE}?search_path=public"
	$(MAKE) db-stop

db-start: ## Starts PostgreSQL Docker instance with uploaded migration
	- $(MAKE) db-stop
	docker run --name ${DOCKER_POSTGRESQL_NAME} --rm -p ${PGPORT}:${PGPORT} -e POSTGRES_PASSWORD=${PGPASSWORD} -e POSTGRES_DB=${PGDATABASE} -e POSTGRES_USER=${PGUSER} -d postgres:$(DOCKER_POSTGRESQL_VERSION)

db-stop: ## Stops PostgreSQL Docker instance
	docker stop ${DOCKER_POSTGRESQL_NAME}

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

go-test: db-start ## Run unit tests present in the codebase
	mkdir -p tmp
	sleep 5;
	go test -coverprofile=./tmp/test-cover.out -race ./...
	$(MAKE) db-stop

test-ci: generate buf-lint buf-breaking build go-vet govulncheck go-linters go-test ## Test the whole codebase (mimics CI/CD)

run: go-tidy build-monitoring db-start ## Runs compiled network device monitoring service
	sleep 5;
	./build/_output/${POC_NAME}

bring-up-db: migration-apply ## Start DB and upload migrations to it

image: ## Builds a Docker image for API Gateway
	docker build . -f build/Dockerfile \
		-t ${DOCKER_REPOSITORY}/${POC_NAME}:${POC_VERSION}

docker-run: image bring-up-db ## Runs compiled binary in a Docker container
	docker run --net=host --rm ${DOCKER_REPOSITORY}/${POC_NAME}:${POC_VERSION}

kind: image ## Builds Docker image for API Gateway and loads it to the currently configured kind cluster
	@if [ "`kind get clusters`" = '' ]; then echo "no kind cluster found" && exit 1; fi
	kind load docker-image ${DOCKER_REPOSITORY}/${POC_NAME}:${POC_VERSION}

go-tidy: ## Runs go mod related commands
	go mod tidy
	go mod vendor

clean-vendor: ## Cleans only vendor folder
	rm -rf ./vendor

clean: ## Remove all the build artifacts
	rm -rf ./build/_output ./vendor ./tmp
	go clean -testcache

export GO111MODULE=on
export CGO_ENABLED=1
export GOPRIVATE=github.com/eroshiva

POC_NAME := monitoring
POC_SIMULATOR_NAME := nd-simulator
POC_VERSION := 0.1.0 # $(shell git rev-parse --abbrev-ref HEAD)
DOCKER_REPOSITORY := eroshiva
GOLANGCI_LINTERS_VERSION := v2.3.0
GOFUMPT_VERSION := v0.8.0
BUF_VERSION := v1.55.1
GRPC_GATEWAY_VERSION := v2.27.1
PROTOC_GEN_ENT_VERSION := v0.6.0
KIND_VERSION := v0.29.0
DOCKER_POSTGRESQL_NAME := monitoring-postgresql
DOCKER_POSTGRESQL_VERSION := 15

KUBE_NAMESPACE := monitoring-system

# Postgres DB configuration and credentials for testing. This mimics the Aurora
# production environment.
export PGHOST=localhost
export PGPORT=5432
export PGSSLMODE=disable
export PGDATABASE=postgres
export PGUSER=admin
export PGPASSWORD=pass

.PHONY: help
help: # Credits to https://gist.github.com/prwhite/8168133 for this handy oneliner
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

atlas-install: ## Installs Atlas tool for generating migrations
	curl -sSf https://atlasgo.sh | sh

buf-install: ## Installs buf to convert protobuf into Golang code
	go install github.com/bufbuild/buf/cmd/buf@${BUF_VERSION}

buf-generate: clean-vendor buf-install buf-update ## Generates Golang-driven bindings out of Protobuf
	mkdir -p internal/ent/schema
	buf generate --exclude-path api/v1/ent --path api/v1/monitoring.proto

buf-generate-simulator-api: clean-vendor buf-install buf-update ## Generates Golang-driven bindings out of Protobuf for Network Device Simulator
	mkdir -p internal/ent/schema
	buf generate --path pkg/mocks/simulator.proto

buf-update: ## Updates the buf dependencies
	buf dep update

buf-lint: ## Runs linters against Protobuf
	buf lint --path api/v1/monitoring.proto

buf-breaking: ## Checks Protobuf schema on breaking changes
	buf breaking --against '.git#branch=main'

generate: buf-generate ## Generates all necessary code bindings
	go generate ./internal/ent

build: go-tidy build-monitoring build-simulator ## Builds all code

build-monitoring: ## Build the Go binary for network device monitoring service
	go build -mod=vendor -o build/_output/${POC_NAME} ./cmd/monitoring.go

build-simulator: ## Build the Go binary for network device simulator
	go build -mod=vendor -o build/_output/${POC_SIMULATOR_NAME} ./cmd/simulator/simulator.go

deps: buf-install go-linters-install atlas-install kind-install ## Installs developer prerequisites for this project
	go get github.com/grpc-ecosystem/grpc-gateway/v2@${GRPC_GATEWAY_VERSION}
	go install entgo.io/contrib/entproto/cmd/protoc-gen-ent@${PROTOC_GEN_ENT_VERSION}
	go install mvdan.cc/gofumpt@${GOFUMPT_VERSION}

kind-install: ## Installs kind to the system
	go install sigs.k8s.io/kind@${KIND_VERSION}

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

run-simulator: go-tidy ## Runs Network Device Simulator with default values
	go run cmd/simulator/simulator.go

run-cli-add-device: go-tidy ## Runs helper CLI tool and adds one device to the monitoring service
	go run cmd/helper-cli/helper-cli.go --addDevice

run-cli-add-devices: go-tidy ## Runs helper CLI tool and adds bunch of the devices specified in the config.json
	go run cmd/helper-cli/helper-cli.go --addDevices

run-cli-delete-device: go-tidy ## Runs helper CLI tool and deletes device by specified resource ID
	go run cmd/helper-cli/helper-cli.go --deleteDevice --deleteDeviceID=${DEVICE_ID}

run-cli-delete-all-devices: go-tidy ## Runs helper CLI tool and deletes all of the devices present in the monitoring service
	go run cmd/helper-cli/helper-cli.go --deleteAllDevices

run-cli-get-status: go-tidy ## Runs helper CLI tool and retrieves device status by specified resource ID
	go run cmd/helper-cli/helper-cli.go --getStatus --deviceID=${DEVICE_ID}

run-cli-get-all-statuses: go-tidy ## Runs helper CLI tool and retrieves all statuses for all network devices
	go run cmd/helper-cli/helper-cli.go --getAllStatuses

run-cli-update-devices: go-tidy ## Runs helper CLI tool and updates all network devices specified in the config.json
	go run cmd/helper-cli/helper-cli.go --updateDevices

run-cli-swap-devices: go-tidy ## Runs helper CLI tool and swaps all network devices in the controller with specified in the config.json
	go run cmd/helper-cli/helper-cli.go --swapDevices

run-cli-get-summary: go-tidy ## Runs helper CLI tool and retrieves a brief summary of all network devices present in the system
	go run cmd/helper-cli/helper-cli.go --getSummary

run-rest-get-summary: ## Runs CURL command and returns summary of network devices
	curl -v http://localhost:50052/v1/monitoring/summary

run-rest-get-devices: ## Runs CURL command and returns a list of network devices
	curl -v http://localhost:50052/v1/monitoring/devices

bring-up-db: migration-apply ## Start DB and upload migrations to it

image: ## Builds a Docker image for Network Device monitoring service
	docker build . -f build/Dockerfile \
		-t ${DOCKER_REPOSITORY}/${POC_NAME}:${POC_VERSION}

image-simulator: ## Builds a Docker image for Network Device simulator
	docker build . -f build/simulator/Dockerfile \
		-t ${DOCKER_REPOSITORY}/${POC_SIMULATOR_NAME}:${POC_VERSION}

images: image image-simulator ## Builds Docker images for monitoring service and for device simulator

docker-run: image bring-up-db ## Runs compiled binary in a Docker container
	docker run --net=host --rm ${DOCKER_REPOSITORY}/${POC_NAME}:${POC_VERSION}

kind: images ## Builds Docker image for API Gateway and loads it to the currently configured kind cluster
	@if [ "`kind get clusters`" = '' ]; then echo "no kind cluster found" && exit 1; fi
	kind load docker-image ${DOCKER_REPOSITORY}/${POC_NAME}:${POC_VERSION}
	kind load docker-image ${DOCKER_REPOSITORY}/${POC_SIMULATOR_NAME}:${POC_VERSION}

create-cluster: delete-cluster ## Creates cluster with KinD
	kind create cluster

delete-cluster: ## Removes KinD cluster
	kind delete cluster

kubectl-delete-namespace: ## Deletes namespace with kubectl
	kubectl delete namespace ${KUBE_NAMESPACE}

deploy-device-simulator: ## Deploys Network Device simulator Helm charts
	helm upgrade --install device-simulator ./helm-charts/network-device-simulator --namespace ${KUBE_NAMESPACE} --create-namespace --wait

deploy-device-monitoring: ## Deploys Network Device Monitoring service Helm charts
	helm upgrade --install device-monitoring ./helm-charts/network-device-monitoring --namespace ${KUBE_NAMESPACE} --create-namespace --wait

update-device-monitoring-charts: ## Updates dependencies for a Network Device Monitoring service charts (i.e., pull PostgreSQL dependency)
	helm dependency update ./helm-charts/network-device-monitoring

helm-test-simulator: ## Runs helm testo for a Network Device simulator
	helm test device-simulator --namespace ${KUBE_NAMESPACE}

helm-test-monitoring: ## Runs helm testo for a Network Device monitoring
	helm test device-monitoring --namespace ${KUBE_NAMESPACE}

poc: kind-install create-cluster go-tidy kind update-device-monitoring-charts deploy-device-monitoring deploy-device-simulator ## Runs PoC in Kubernetes cluster

poc-test: ## Runs PoC in Kubernetes cluster
poc-test: kind-install create-cluster go-tidy kind update-device-monitoring-charts deploy-device-monitoring deploy-device-simulator helm-test-monitoring helm-test-simulator delete-cluster

go-tidy: ## Runs go mod related commands
	go mod tidy
	go mod vendor

clean-vendor: ## Cleans only vendor folder
	rm -rf ./vendor

clean: ## Remove all the build artifacts
	rm -rf ./build/_output ./vendor ./tmp
	go clean -testcache

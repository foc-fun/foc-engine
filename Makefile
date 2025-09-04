build: build-engine build-contracts

build-engine:
	@echo "Building engine..."
	go build -o bin ./...

build-contracts:
	@echo "Building contracts..."
	cd onchain && scarb build

# Docker Local Development Commands

## Devnet Environment
docker-devnet-build:
	@echo "Building devnet Docker images..."
	docker compose -f docker-compose-devnet.yml build

docker-devnet-up:
	@echo "Starting devnet environment..."
	docker compose -f docker-compose-devnet.yml up

docker-devnet-down:
	@echo "Stopping devnet environment..."
	docker compose -f docker-compose-devnet.yml down

docker-devnet-clean:
	@echo "Cleaning devnet environment (removing volumes)..."
	docker compose -f docker-compose-devnet.yml down --volumes

## Sepolia Environment
docker-sepolia-build:
	@echo "Building Sepolia Docker images..."
	docker compose -f docker-compose-sepolia.yml build

docker-sepolia-up:
	@echo "Starting Sepolia environment..."
	docker compose -f docker-compose-sepolia.yml up

docker-sepolia-down:
	@echo "Stopping Sepolia environment..."
	docker compose -f docker-compose-sepolia.yml down

docker-sepolia-clean:
	@echo "Cleaning Sepolia environment (removing volumes)..."
	docker compose -f docker-compose-sepolia.yml down --volumes

## Mainnet Environment
docker-mainnet-build:
	@echo "Building Mainnet Docker images..."
	docker compose -f docker-compose-mainnet.yml build

docker-mainnet-up:
	@echo "Starting Mainnet environment..."
	docker compose -f docker-compose-mainnet.yml up

docker-mainnet-down:
	@echo "Stopping Mainnet environment..."
	docker compose -f docker-compose-mainnet.yml down

docker-mainnet-clean:
	@echo "Cleaning Mainnet environment (removing volumes)..."
	docker compose -f docker-compose-mainnet.yml down --volumes

# Production Docker Commands

## Build all production images for both networks
docker-build: docker-build-sepolia docker-build-mainnet docker-build-standalone

## Build Sepolia production images
docker-build-sepolia:
	$(eval APP_VERSION := $(shell cat infra/Chart.yaml | yq eval '.appVersion' -))
	$(eval COMMIT_SHA := $(shell git rev-parse --short HEAD))
	@echo "Building Sepolia docker images with version $(APP_VERSION)-$(COMMIT_SHA)"
	@echo "Building Sepolia API..."
	docker build . -f dockerfiles/Dockerfile.sepolia.api -t "brandonjroberts/foc-engine-api:sepolia-$(APP_VERSION)-$(COMMIT_SHA)" -t "brandonjroberts/foc-engine-api:sepolia-latest"
	@echo "Building Sepolia Indexer..."
	docker build . -f dockerfiles/Dockerfile.sepolia.indexer -t "brandonjroberts/foc-engine-indexer:sepolia-$(APP_VERSION)-$(COMMIT_SHA)" -t "brandonjroberts/foc-engine-indexer:sepolia-latest"

## Build Mainnet production images
docker-build-mainnet:
	$(eval APP_VERSION := $(shell cat infra/Chart.yaml | yq eval '.appVersion' -))
	$(eval COMMIT_SHA := $(shell git rev-parse --short HEAD))
	@echo "Building Mainnet docker images with version $(APP_VERSION)-$(COMMIT_SHA)"
	@echo "Building Mainnet API..."
	docker build . -f dockerfiles/Dockerfile.mainnet.api -t "brandonjroberts/foc-engine-api:mainnet-$(APP_VERSION)-$(COMMIT_SHA)" -t "brandonjroberts/foc-engine-api:mainnet-latest"
	@echo "Building Mainnet Indexer..."
	docker build . -f dockerfiles/Dockerfile.mainnet.indexer -t "brandonjroberts/foc-engine-indexer:mainnet-$(APP_VERSION)-$(COMMIT_SHA)" -t "brandonjroberts/foc-engine-indexer:mainnet-latest"

## Build standalone indexer (network-agnostic)
docker-build-standalone:
	$(eval APP_VERSION := $(shell cat infra/Chart.yaml | yq eval '.appVersion' -))
	$(eval COMMIT_SHA := $(shell git rev-parse --short HEAD))
	@echo "Building standalone indexer with version $(APP_VERSION)-$(COMMIT_SHA)"
	docker build . -f dockerfiles/Dockerfile.standalone-indexer -t "brandonjroberts/foc-engine-standalone-indexer:$(APP_VERSION)-$(COMMIT_SHA)" -t "brandonjroberts/foc-engine-standalone-indexer:latest"

## Push all production images
docker-push: docker-push-sepolia docker-push-mainnet docker-push-standalone

## Push Sepolia images
docker-push-sepolia:
	$(eval APP_VERSION := $(shell cat infra/Chart.yaml | yq eval '.appVersion' -))
	$(eval COMMIT_SHA := $(shell git rev-parse --short HEAD))
	@echo "Pushing Sepolia docker images with version $(APP_VERSION)-$(COMMIT_SHA)"
	docker push "brandonjroberts/foc-engine-api:sepolia-$(APP_VERSION)-$(COMMIT_SHA)"
	docker push "brandonjroberts/foc-engine-api:sepolia-latest"
	docker push "brandonjroberts/foc-engine-indexer:sepolia-$(APP_VERSION)-$(COMMIT_SHA)"
	docker push "brandonjroberts/foc-engine-indexer:sepolia-latest"

## Push Mainnet images
docker-push-mainnet:
	$(eval APP_VERSION := $(shell cat infra/Chart.yaml | yq eval '.appVersion' -))
	$(eval COMMIT_SHA := $(shell git rev-parse --short HEAD))
	@echo "Pushing Mainnet docker images with version $(APP_VERSION)-$(COMMIT_SHA)"
	docker push "brandonjroberts/foc-engine-api:mainnet-$(APP_VERSION)-$(COMMIT_SHA)"
	docker push "brandonjroberts/foc-engine-api:mainnet-latest"
	docker push "brandonjroberts/foc-engine-indexer:mainnet-$(APP_VERSION)-$(COMMIT_SHA)"
	docker push "brandonjroberts/foc-engine-indexer:mainnet-latest"

## Push standalone indexer
docker-push-standalone:
	$(eval APP_VERSION := $(shell cat infra/Chart.yaml | yq eval '.appVersion' -))
	$(eval COMMIT_SHA := $(shell git rev-parse --short HEAD))
	@echo "Pushing standalone indexer with version $(APP_VERSION)-$(COMMIT_SHA)"
	docker push "brandonjroberts/foc-engine-standalone-indexer:$(APP_VERSION)-$(COMMIT_SHA)"
	docker push "brandonjroberts/foc-engine-standalone-indexer:latest"

# Sepolia Environment Commands
helm-install-sepolia:
	$(eval COMMIT_SHA := $(shell git rev-parse --short HEAD))
	@echo "Installing Sepolia deployment..."
	helm install -f infra/values-sepolia.yaml --set paymaster.apiKey=$(AVNU_API_KEY) --set standaloneIndexer.rpcUrl=$(STANDALONE_INDEXER_RPC_URL) --set deployments.sha=$(COMMIT_SHA) foc-engine-sepolia infra --namespace foc-engine-sepolia --create-namespace

helm-template-sepolia:
	$(eval COMMIT_SHA := $(shell git rev-parse --short HEAD))
	@echo "Templating Sepolia helm chart..."
	helm template -f infra/values-sepolia.yaml --set paymaster.apiKey=$(AVNU_API_KEY) --set standaloneIndexer.rpcUrl=$(STANDALONE_INDEXER_RPC_URL) --set deployments.sha=$(COMMIT_SHA) foc-engine-sepolia infra --namespace foc-engine-sepolia

helm-upgrade-sepolia:
	$(eval COMMIT_SHA := $(shell git rev-parse --short HEAD))
	@echo "Upgrading Sepolia deployment..."
	helm upgrade -f infra/values-sepolia.yaml --set paymaster.apiKey=$(AVNU_API_KEY) --set standaloneIndexer.rpcUrl=$(STANDALONE_INDEXER_RPC_URL) --set deployments.sha=$(COMMIT_SHA) foc-engine-sepolia infra --namespace foc-engine-sepolia

helm-uninstall-sepolia:
	@echo "Uninstalling Sepolia helm release..."
	helm uninstall foc-engine-sepolia --namespace foc-engine-sepolia

# Mainnet Environment Commands
helm-install-mainnet:
	$(eval COMMIT_SHA := $(shell git rev-parse --short HEAD))
	@echo "Installing Mainnet deployment..."
	helm install -f infra/values-mainnet.yaml --set paymaster.apiKey=$(AVNU_API_KEY) --set standaloneIndexer.rpcUrl=$(STANDALONE_INDEXER_RPC_URL) --set deployments.sha=$(COMMIT_SHA) foc-engine-mainnet infra --namespace foc-engine-mainnet --create-namespace

helm-template-mainnet:
	$(eval COMMIT_SHA := $(shell git rev-parse --short HEAD))
	@echo "Templating Mainnet helm chart..."
	helm template -f infra/values-mainnet.yaml --set paymaster.apiKey=$(AVNU_API_KEY) --set standaloneIndexer.rpcUrl=$(STANDALONE_INDEXER_RPC_URL) --set deployments.sha=$(COMMIT_SHA) foc-engine-mainnet infra --namespace foc-engine-mainnet

helm-upgrade-mainnet:
	$(eval COMMIT_SHA := $(shell git rev-parse --short HEAD))
	@echo "Upgrading Mainnet deployment..."
	helm upgrade -f infra/values-mainnet.yaml --set paymaster.apiKey=$(AVNU_API_KEY) --set standaloneIndexer.rpcUrl=$(STANDALONE_INDEXER_RPC_URL) --set deployments.sha=$(COMMIT_SHA) foc-engine-mainnet infra --namespace foc-engine-mainnet

helm-uninstall-mainnet:
	@echo "Uninstalling Mainnet helm release..."
	helm uninstall foc-engine-mainnet --namespace foc-engine-mainnet

# Legacy commands (for backwards compatibility - defaults to sepolia)
helm-uninstall:
	@echo "Uninstalling legacy helm release..."
	helm uninstall foc-engine-infra

helm-install:
	@echo "DEPRECATED: Use 'make helm-install-sepolia' or 'make helm-install-mainnet' instead"
	$(eval COMMIT_SHA := $(shell git rev-parse --short HEAD))
	@echo "Installing legacy deployment (sepolia config)..."
	helm install -f infra/values-sepolia.yaml --set paymaster.apiKey=$(AVNU_API_KEY) --set standaloneIndexer.rpcUrl=$(STANDALONE_INDEXER_RPC_URL) --set deployments.sha=$(COMMIT_SHA) foc-engine-infra infra

helm-template:
	@echo "DEPRECATED: Use 'make helm-template-sepolia' or 'make helm-template-mainnet' instead"
	$(eval COMMIT_SHA := $(shell git rev-parse --short HEAD))
	@echo "Templating legacy helm chart (sepolia config)..."
	helm template -f infra/values-sepolia.yaml --set paymaster.apiKey=$(AVNU_API_KEY) --set standaloneIndexer.rpcUrl=$(STANDALONE_INDEXER_RPC_URL) --set deployments.sha=$(COMMIT_SHA) foc-engine-infra infra

helm-upgrade:
	@echo "DEPRECATED: Use 'make helm-upgrade-sepolia' or 'make helm-upgrade-mainnet' instead"
	$(eval COMMIT_SHA := $(shell git rev-parse --short HEAD))
	@echo "Upgrading legacy helm release (sepolia config)..."
	helm upgrade -f infra/values-sepolia.yaml --set paymaster.apiKey=$(AVNU_API_KEY) --set standaloneIndexer.rpcUrl=$(STANDALONE_INDEXER_RPC_URL) --set deployments.sha=$(COMMIT_SHA) foc-engine-infra infra

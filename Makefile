build: build-engine build-contracts

build-engine:
	@echo "Building engine..."
	go build -o bin ./...

build-contracts:
	@echo "Building contracts..."
	cd onchain && scarb build

docker-build:
	$(eval APP_VERSION := $(shell cat infra/Chart.yaml | yq eval '.appVersion' -))
	$(eval COMMIT_SHA := $(shell git rev-parse --short HEAD))
	@echo "Building docker images with version $(APP_VERSION)-$(COMMIT_SHA)"
	@echo "Building api..."
	docker build . -f dockerfiles/Dockerfile.api.prod -t "brandonjroberts/foc-engine-api:$(APP_VERSION)-$(COMMIT_SHA)"
	@echo "Building indexer..."
	docker build . -f dockerfiles/Dockerfile.indexer.prod -t "brandonjroberts/foc-engine-indexer:$(APP_VERSION)-$(COMMIT_SHA)"

docker-push:
	$(eval APP_VERSION := $(shell cat infra/Chart.yaml | yq eval '.appVersion' -))
	$(eval COMMIT_SHA := $(shell git rev-parse --short HEAD))
	@echo "Pushing docker images with version $(APP_VERSION)-$(COMMIT_SHA)"
	docker push "brandonjroberts/foc-engine-api:$(APP_VERSION)-$(COMMIT_SHA)"
	docker push "brandonjroberts/foc-engine-indexer:$(APP_VERSION)-$(COMMIT_SHA)"

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

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

helm-uninstall:
	@echo "Uninstalling helm release..."
	helm uninstall foc-engine-infra

helm-install:
	$(eval COMMIT_SHA := $(shell git rev-parse --short HEAD))
	@echo "Installing helm chart..."
	helm install --set paymaster.apiKey=$(AVNU_API_KEY) --set deployments.sha=$(COMMIT_SHA) foc-engine-infra infra

helm-template:
	$(eval COMMIT_SHA := $(shell git rev-parse --short HEAD))
	@echo "Templating helm chart..."
	helm template --set paymaster.apiKey=$(AVNU_API_KEY) --set deployments.sha=$(COMMIT_SHA) foc-engine-infra infra

helm-upgrade:
	$(eval COMMIT_SHA := $(shell git rev-parse --short HEAD))
	@echo "Upgrading helm release..."
	helm upgrade --set paymaster.apiKey=$(AVNU_API_KEY) --set deployments.sha=$(COMMIT_SHA) foc-engine-infra infra

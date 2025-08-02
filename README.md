<div align="center">
  <img src="resources/logo.png" alt="foc_engine_logo" height="300"/>

  ***Let's make Starknet Magic***
</div>

## Overview

FOC Engine is a Starknet-based application that provides API and indexer services for interacting with the Starknet blockchain. The project includes:

- **API Service**: REST API server running on port 8080
- **Indexer Service**: Starknet event indexer running on port 8085
- **Smart Contracts**: Cairo contracts for on-chain functionality
- **Infrastructure**: Docker Compose for local development, Helm charts for Kubernetes deployment

## Dependencies

The following dependencies must be installed to run the foc-engine:
- docker
- docker compose
- cmdline tools: `jq`, `yq`

### Development Dependencies
- Go 1.23.4+
- [Scarb](https://docs.swmansion.com/scarb/) (Cairo package manager)
- [snforge](https://foundry-rs.github.io/starknet-foundry/) (for contract testing)

## Installation

### Option 1: Install using [asdf](https://asdf-vm.com/)
```bash
asdf plugin add foc-engine https://github.com/b-j-roberts/asdf-foc-engine.git
asdf install foc-engine latest
asdf global foc-engine latest
```

### Option 2: Clone and build from source
```bash
git clone git@github.com:b-j-roberts/foc-engine.git
cd foc-engine
docker compose -f docker-compose-devnet.yml build
```

## Configuration

Create a `.env` file from the example:
```bash
cp .env.example .env
```

Required environment variables:
- `STARKNET_KEYSTORE`: Path to Starknet keystore
- `STARKNET_ACCOUNT`: Starknet account address
- `AVNU_PAYMASTER_API_KEY`: API key for AVNU paymaster service

## Running

### Using the CLI tool
```bash
foc-engine run      # Run the engine
foc-engine clean    # Clean up resources
foc-engine version  # Show version
foc-engine help     # Display help
```

### Docker Compose (Local Development)

Start all services:
```bash
docker compose -f docker-compose-devnet.yml up
```

Fresh restart with clean state:
```bash
docker compose -f docker-compose-devnet.yml down --volumes
docker compose -f docker-compose-devnet.yml build
docker compose -f docker-compose-devnet.yml up
```

Run on Sepolia testnet:
```bash
docker compose -f docker-compose-sepolia.yml up
```

## Development

### Building

Build everything:
```bash
make build
```

Build specific components:
```bash
make build-engine     # Build Go binaries
make build-contracts  # Build Cairo contracts
```

### Docker Images

Build and push Docker images:
```bash
make docker-build
make docker-push
```

### Testing

Run contract tests:
```bash
cd onchain
scarb test
```

## Deployment

### Kubernetes with Helm

Install:
```bash
make helm-install
```

Upgrade existing deployment:
```bash
make helm-upgrade
```

Uninstall:
```bash
make helm-uninstall
```

Preview generated manifests:
```bash
make helm-template
```

## Architecture

The project consists of:

- **Backend Services** (Go):
  - API service for external interactions
  - Indexer service for blockchain event processing
  - MongoDB for data persistence
  - Redis for caching

- **Smart Contracts** (Cairo):
  - Located in `onchain/` directory
  - Built with Scarb
  - Tested with snforge

- **Infrastructure**:
  - Docker Compose configurations for local and testnet environments
  - Helm charts for Kubernetes deployments
  - Production-ready Dockerfiles

## Services

When running locally, the following services are available:

- **API**: http://localhost:8080
- **Indexer**: http://localhost:8085
- **MongoDB**: localhost:27017
- **Redis**: localhost:6379
- **Starknet Devnet**: http://localhost:5050

## License

[Add license information here]
